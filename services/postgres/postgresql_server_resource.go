package postgres

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2020-01-01/postgresql"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/identity"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/postgres/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/postgres/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

const (
	postgreSQLServerResourceName = "azurerm_postgresql_server"
)

var skuList = []string{
	"B_Gen4_1",
	"B_Gen4_2",
	"B_Gen5_1",
	"B_Gen5_2",
	"GP_Gen4_2",
	"GP_Gen4_4",
	"GP_Gen4_8",
	"GP_Gen4_16",
	"GP_Gen4_32",
	"GP_Gen5_2",
	"GP_Gen5_4",
	"GP_Gen5_8",
	"GP_Gen5_16",
	"GP_Gen5_32",
	"GP_Gen5_64",
	"MO_Gen5_2",
	"MO_Gen5_4",
	"MO_Gen5_8",
	"MO_Gen5_16",
	"MO_Gen5_32",
}

func resourcePostgreSQLServer() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourcePostgreSQLServerCreate,
		Read:   resourcePostgreSQLServerRead,
		Update: resourcePostgreSQLServerUpdate,
		Delete: resourcePostgreSQLServerDelete,

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := parse.ServerID(id)
			return err
		}, func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) ([]*pluginsdk.ResourceData, error) {
			client := meta.(*clients.Client).Postgres.ServersClient

			id, err := parse.ServerID(d.Id())
			if err != nil {
				return []*pluginsdk.ResourceData{d}, err
			}

			resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
			if err != nil {
				return []*pluginsdk.ResourceData{d}, fmt.Errorf("reading PostgreSQL Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
			}

			d.Set("create_mode", "Default")
			if resp.ReplicationRole != nil && *resp.ReplicationRole != "Master" && *resp.ReplicationRole != "None" {
				d.Set("create_mode", resp.ReplicationRole)

				sourceServerId, err := parse.ServerID(*resp.MasterServerID)
				if err != nil {
					return []*pluginsdk.ResourceData{d}, fmt.Errorf("parsing Postgres Main Server ID : %v", err)
				}
				d.Set("creation_source_server_id", sourceServerId.ID())
			}

			return []*pluginsdk.ResourceData{d}, nil
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(60 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(60 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ServerName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"sku_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(skuList, false),
			},

			"version": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(postgresql.NineFullStopFive),
					string(postgresql.NineFullStopSix),
					string(postgresql.OneOne),
					string(postgresql.OneZero),
					string(postgresql.OneZeroFullStopZero),
				}, false),
			},

			"administrator_login": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.All(validation.StringIsNotWhiteSpace, validate.AdminUsernames),
			},

			"administrator_login_password": {
				Type:      pluginsdk.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"auto_grow_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"backup_retention_days": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(7, 35),
			},

			"geo_redundant_backup_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},

			"create_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(postgresql.CreateModeDefault),
				ValidateFunc: validation.StringInSlice([]string{
					string(postgresql.CreateModeDefault),
					string(postgresql.CreateModeGeoRestore),
					string(postgresql.CreateModePointInTimeRestore),
					string(postgresql.CreateModeReplica),
				}, false),
			},

			"creation_source_server_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.ServerID,
			},

			"identity": commonschema.SystemAssignedIdentityOptional(),

			"infrastructure_encryption_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"public_network_access_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"restore_point_in_time": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},

			"storage_mb": {
				Type:     pluginsdk.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.IntBetween(5120, 16777216),
					validation.IntDivisibleBy(1024),
				),
			},

			"ssl_minimal_tls_version_enforced": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(postgresql.TLS12),
				ValidateFunc: validation.StringInSlice([]string{
					string(postgresql.TLSEnforcementDisabled),
					string(postgresql.TLS10),
					string(postgresql.TLS11),
					string(postgresql.TLS12),
				}, false),
			},

			"ssl_enforcement_enabled": {
				Type:     pluginsdk.TypeBool,
				Required: true,
			},

			"threat_detection_policy": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							AtLeastOneOf: []string{
								"threat_detection_policy.0.enabled", "threat_detection_policy.0.disabled_alerts", "threat_detection_policy.0.email_account_admins",
								"threat_detection_policy.0.email_addresses", "threat_detection_policy.0.retention_days", "threat_detection_policy.0.storage_account_access_key",
								"threat_detection_policy.0.storage_endpoint",
							},
						},

						"disabled_alerts": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Set:      pluginsdk.HashString,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"Sql_Injection",
									"Sql_Injection_Vulnerability",
									"Access_Anomaly",
									"Data_Exfiltration",
									"Unsafe_Action",
								}, false),
							},
							AtLeastOneOf: []string{
								"threat_detection_policy.0.enabled", "threat_detection_policy.0.disabled_alerts", "threat_detection_policy.0.email_account_admins",
								"threat_detection_policy.0.email_addresses", "threat_detection_policy.0.retention_days", "threat_detection_policy.0.storage_account_access_key",
								"threat_detection_policy.0.storage_endpoint",
							},
						},

						"email_account_admins": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							AtLeastOneOf: []string{
								"threat_detection_policy.0.enabled", "threat_detection_policy.0.disabled_alerts", "threat_detection_policy.0.email_account_admins",
								"threat_detection_policy.0.email_addresses", "threat_detection_policy.0.retention_days", "threat_detection_policy.0.storage_account_access_key",
								"threat_detection_policy.0.storage_endpoint",
							},
						},

						"email_addresses": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
								// todo email validation in code
							},
							Set: pluginsdk.HashString,
							AtLeastOneOf: []string{
								"threat_detection_policy.0.enabled", "threat_detection_policy.0.disabled_alerts", "threat_detection_policy.0.email_account_admins",
								"threat_detection_policy.0.email_addresses", "threat_detection_policy.0.retention_days", "threat_detection_policy.0.storage_account_access_key",
								"threat_detection_policy.0.storage_endpoint",
							},
						},

						"retention_days": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							AtLeastOneOf: []string{
								"threat_detection_policy.0.enabled", "threat_detection_policy.0.disabled_alerts", "threat_detection_policy.0.email_account_admins",
								"threat_detection_policy.0.email_addresses", "threat_detection_policy.0.retention_days", "threat_detection_policy.0.storage_account_access_key",
								"threat_detection_policy.0.storage_endpoint",
							},
						},

						"storage_account_access_key": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
							AtLeastOneOf: []string{
								"threat_detection_policy.0.enabled", "threat_detection_policy.0.disabled_alerts", "threat_detection_policy.0.email_account_admins",
								"threat_detection_policy.0.email_addresses", "threat_detection_policy.0.retention_days", "threat_detection_policy.0.storage_account_access_key",
								"threat_detection_policy.0.storage_endpoint",
							},
						},

						"storage_endpoint": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							AtLeastOneOf: []string{
								"threat_detection_policy.0.enabled", "threat_detection_policy.0.disabled_alerts", "threat_detection_policy.0.email_account_admins",
								"threat_detection_policy.0.email_addresses", "threat_detection_policy.0.retention_days", "threat_detection_policy.0.storage_account_access_key",
								"threat_detection_policy.0.storage_endpoint",
							},
						},
					},
				},
			},

			"fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},

		CustomizeDiff: pluginsdk.CustomDiffWithAll(
			pluginsdk.ForceNewIfChange("sku_name", func(ctx context.Context, old, new, meta interface{}) bool {
				oldTier := strings.Split(old.(string), "_")
				newTier := strings.Split(new.(string), "_")
				// If the sku tier was not changed, we don't need ForceNew
				if oldTier[0] == newTier[0] {
					return false
				}
				// Basic tier could not be changed to other tiers
				if oldTier[0] == "B" || newTier[0] == "B" {
					return true
				}
				return false
			}),
			pluginsdk.ForceNewIfChange("create_mode", func(ctx context.Context, old, new, meta interface{}) bool {
				oldMode := postgresql.CreateMode(old.(string))
				newMode := postgresql.CreateMode(new.(string))
				// Instance could not be changed from Default to Replica
				if oldMode == postgresql.CreateModeDefault && newMode == postgresql.CreateModeReplica {
					return true
				}
				return false
			}),
		),
	}
}

func resourcePostgreSQLServerCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Postgres.ServersClient
	securityClient := meta.(*clients.Client).Postgres.ServerSecurityAlertPoliciesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureRM PostgreSQL Server creation.")

	id := parse.NewServerID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
		}
	}

	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_postgresql_server", id.ID())
	}

	mode := postgresql.CreateMode(d.Get("create_mode").(string))
	source := d.Get("creation_source_server_id").(string)
	version := postgresql.ServerVersion(d.Get("version").(string))

	sku, err := expandServerSkuName(d.Get("sku_name").(string))
	if err != nil {
		return fmt.Errorf("expanding `sku_name`: %+v", err)
	}

	infraEncrypt := postgresql.InfrastructureEncryptionEnabled
	if v := d.Get("infrastructure_encryption_enabled"); !v.(bool) {
		infraEncrypt = postgresql.InfrastructureEncryptionDisabled
	}

	publicAccess := postgresql.PublicNetworkAccessEnumEnabled
	if v := d.Get("public_network_access_enabled"); !v.(bool) {
		publicAccess = postgresql.PublicNetworkAccessEnumDisabled
	}

	ssl := postgresql.SslEnforcementEnumEnabled
	if v := d.Get("ssl_enforcement_enabled"); !v.(bool) {
		ssl = postgresql.SslEnforcementEnumDisabled
	}

	tlsMin := postgresql.MinimalTLSVersionEnum(d.Get("ssl_minimal_tls_version_enforced").(string))
	if ssl == postgresql.SslEnforcementEnumDisabled && tlsMin != postgresql.TLSEnforcementDisabled {
		return fmt.Errorf("`ssl_minimal_tls_version_enforced` must be set to `TLSEnforcementDisabled` if `ssl_enforcement_enabled` is set to `false`")
	}

	storage := expandPostgreSQLStorageProfile(d)

	var props postgresql.BasicServerPropertiesForCreate
	switch mode {
	case postgresql.CreateModeDefault:
		admin := d.Get("administrator_login").(string)
		pass := d.Get("administrator_login_password").(string)
		if admin == "" {
			return fmt.Errorf("`administrator_login` must not be empty when `create_mode` is `default`")
		}
		if pass == "" {
			return fmt.Errorf("`administrator_login_password` must not be empty when `create_mode` is `default`")
		}

		if _, ok := d.GetOk("restore_point_in_time"); ok {
			return fmt.Errorf("`restore_point_in_time` cannot be set when `create_mode` is `default`")
		}

		// check admin
		props = &postgresql.ServerPropertiesForDefaultCreate{
			AdministratorLogin:         &admin,
			AdministratorLoginPassword: &pass,
			CreateMode:                 mode,
			InfrastructureEncryption:   infraEncrypt,
			PublicNetworkAccess:        publicAccess,
			MinimalTLSVersion:          tlsMin,
			SslEnforcement:             ssl,
			StorageProfile:             storage,
			Version:                    version,
		}
	case postgresql.CreateModePointInTimeRestore:
		v, ok := d.GetOk("restore_point_in_time")
		if !ok || v.(string) == "" {
			return fmt.Errorf("restore_point_in_time must be set when create_mode is PointInTimeRestore")
		}
		time, _ := time.Parse(time.RFC3339, v.(string)) // should be validated by the schema

		// d.GetOk cannot identify whether user sets the property that is bool type and has default value. So it has to identify it using `d.GetRawConfig()`
		if v := d.GetRawConfig().AsValueMap()["public_network_access_enabled"]; !v.IsNull() {
			return fmt.Errorf("`public_network_access_enabled` doesn't support PointInTimeRestore mode")
		}

		props = &postgresql.ServerPropertiesForRestore{
			CreateMode:     mode,
			SourceServerID: &source,
			RestorePointInTime: &date.Time{
				Time: time,
			},
			InfrastructureEncryption: infraEncrypt,
			MinimalTLSVersion:        tlsMin,
			SslEnforcement:           ssl,
			StorageProfile:           storage,
			Version:                  version,
		}
	case postgresql.CreateModeGeoRestore:
		props = &postgresql.ServerPropertiesForGeoRestore{
			CreateMode:               mode,
			SourceServerID:           &source,
			InfrastructureEncryption: infraEncrypt,
			PublicNetworkAccess:      publicAccess,
			MinimalTLSVersion:        tlsMin,
			SslEnforcement:           ssl,
			StorageProfile:           storage,
			Version:                  version,
		}
	case postgresql.CreateModeReplica:
		props = &postgresql.ServerPropertiesForReplica{
			CreateMode:               mode,
			SourceServerID:           &source,
			InfrastructureEncryption: infraEncrypt,
			PublicNetworkAccess:      publicAccess,
			MinimalTLSVersion:        tlsMin,
			SslEnforcement:           ssl,
			Version:                  version,
		}
	}

	expandedIdentity, err := expandServerIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return fmt.Errorf("expanding `identity`: %+v", err)
	}
	server := postgresql.ServerForCreate{
		Identity:   expandedIdentity,
		Location:   utils.String(location.Normalize(d.Get("location").(string))),
		Properties: props,
		Sku:        sku,
		Tags:       tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	future, err := client.Create(ctx, id.ResourceGroup, id.Name, server)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	log.Printf("[DEBUG] Waiting for %s to become available", id)
	stateConf := &pluginsdk.StateChangeConf{
		Pending:    []string{string(postgresql.ServerStateInaccessible)},
		Target:     []string{string(postgresql.ServerStateReady)},
		Refresh:    postgreSqlStateRefreshFunc(ctx, client, id),
		MinTimeout: 15 * time.Second,
		Timeout:    d.Timeout(pluginsdk.TimeoutCreate),
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for %s to become available: %+v", id, err)
	}

	d.SetId(id.ID())

	if v, ok := d.GetOk("threat_detection_policy"); ok {
		alert := expandSecurityAlertPolicy(v)
		if alert != nil {
			future, err := securityClient.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, *alert)
			if err != nil {
				return fmt.Errorf("updataing security alert policy for %s: %v", id, err)
			}

			if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for update of security alert policy for %s: %+v", id, err)
			}
		}
	}

	// Issue tracking the REST API update failure: https://github.com/Azure/azure-rest-api-specs/issues/14117
	if mode == postgresql.CreateModeReplica {
		log.Printf("[INFO] updating `public_network_access_enabled` for %s", id)
		properties := postgresql.ServerUpdateParameters{
			ServerUpdateParametersProperties: &postgresql.ServerUpdateParametersProperties{
				PublicNetworkAccess: publicAccess,
			},
		}

		future, err := client.Update(ctx, id.ResourceGroup, id.Name, properties)
		if err != nil {
			return fmt.Errorf("updating Public Network Access for Replica %q: %+v", id, err)
		}

		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("waiting for update of Public Network Access for Replica %q: %+v", id, err)
		}
	}

	return resourcePostgreSQLServerRead(d, meta)
}

func resourcePostgreSQLServerUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Postgres.ServersClient
	securityClient := meta.(*clients.Client).Postgres.ServerSecurityAlertPoliciesClient
	replicasClient := meta.(*clients.Client).Postgres.ReplicasClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	// TODO: support for Delta updates

	id, err := parse.ServerID(d.Id())
	if err != nil {
		return fmt.Errorf("parsing Postgres Server ID : %v", err)
	}

	// Locks for upscaling of replicas
	mode := postgresql.CreateMode(d.Get("create_mode").(string))
	primaryID := id.String()
	if mode == postgresql.CreateModeReplica {
		primaryID = d.Get("creation_source_server_id").(string)

		// Wait for possible restarts triggered by scaling primary (and its replicas)
		log.Printf("[DEBUG] Waiting for %s to become available", *id)
		stateConf := &pluginsdk.StateChangeConf{
			Pending:    []string{string(postgresql.ServerStateInaccessible), "Restarting"},
			Target:     []string{string(postgresql.ServerStateReady)},
			Refresh:    postgreSqlStateRefreshFunc(ctx, client, *id),
			MinTimeout: 15 * time.Second,
			Timeout:    d.Timeout(pluginsdk.TimeoutCreate),
		}

		if _, err = stateConf.WaitForStateContext(ctx); err != nil {
			return fmt.Errorf("waiting for %s to become available: %+v", *id, err)
		}
	}
	locks.ByID(primaryID)
	defer locks.UnlockByID(primaryID)

	sku, err := expandServerSkuName(d.Get("sku_name").(string))
	if err != nil {
		return fmt.Errorf("expanding `sku_name`: %v", err)
	}

	if d.HasChange("sku_name") && mode != postgresql.CreateModeReplica {
		oldRaw, newRaw := d.GetChange("sku_name")
		old := oldRaw.(string)
		new := newRaw.(string)

		if indexOfSku(old) < indexOfSku(new) {
			listReplicas, err := replicasClient.ListByServer(ctx, id.ResourceGroup, id.Name)
			if err != nil {
				return fmt.Errorf("listing replicas for %s: %+v", *id, err)
			}

			propertiesReplica := postgresql.ServerUpdateParameters{
				Sku: sku,
			}
			for _, replica := range *listReplicas.Value {
				replicaId, err := parse.ServerID(*replica.ID)
				if err != nil {
					return fmt.Errorf("parsing Postgres Server Replica ID : %v", err)
				}
				future, err := client.Update(ctx, replicaId.ResourceGroup, replicaId.Name, propertiesReplica)
				if err != nil {
					return fmt.Errorf("updating SKU for Replica %s: %+v", *replicaId, err)
				}

				if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
					return fmt.Errorf("waiting for SKU update for Replica %s: %+v", *replicaId, err)
				}
			}
		}
	}

	ssl := postgresql.SslEnforcementEnumEnabled
	if v := d.Get("ssl_enforcement_enabled"); !v.(bool) {
		ssl = postgresql.SslEnforcementEnumDisabled
	}

	tlsMin := postgresql.MinimalTLSVersionEnum(d.Get("ssl_minimal_tls_version_enforced").(string))

	if ssl == postgresql.SslEnforcementEnumDisabled && tlsMin != postgresql.TLSEnforcementDisabled {
		return fmt.Errorf("`ssl_minimal_tls_version_enforced` must be set to `TLSEnforcementDisabled` if `ssl_enforcement_enabled` is set to `false`")
	}

	expandedIdentity, err := expandServerIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return fmt.Errorf("expanding `identity`: %+v", err)
	}

	properties := postgresql.ServerUpdateParameters{
		Identity: expandedIdentity,
		ServerUpdateParametersProperties: &postgresql.ServerUpdateParametersProperties{
			SslEnforcement:    ssl,
			MinimalTLSVersion: tlsMin,
			StorageProfile:    expandPostgreSQLStorageProfile(d),
			Version:           postgresql.ServerVersion(d.Get("version").(string)),
		},
		Sku:  sku,
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if mode == postgresql.CreateModePointInTimeRestore {
		// d.GetOk cannot identify whether user sets the property that is bool type and has default value. So it has to identify it using `d.GetRawConfig()`
		if v := d.GetRawConfig().AsValueMap()["public_network_access_enabled"]; !v.IsNull() {
			return fmt.Errorf("`public_network_access_enabled` doesn't support PointInTimeRestore mode")
		}
	} else {
		publicAccess := postgresql.PublicNetworkAccessEnumEnabled
		if v := d.Get("public_network_access_enabled"); !v.(bool) {
			publicAccess = postgresql.PublicNetworkAccessEnumDisabled
		}
		properties.ServerUpdateParametersProperties.PublicNetworkAccess = publicAccess
	}

	oldCreateMode, newCreateMode := d.GetChange("create_mode")
	replicaUpdatedToDefault := postgresql.CreateMode(oldCreateMode.(string)) == postgresql.CreateModeReplica && postgresql.CreateMode(newCreateMode.(string)) == postgresql.CreateModeDefault
	if replicaUpdatedToDefault {
		properties.ServerUpdateParametersProperties.ReplicationRole = utils.String("None")
	}

	// Update Admin Password in the separate call when Replication is stopped: https://github.com/Azure/azure-rest-api-specs/issues/16898
	if d.HasChange("administrator_login_password") && !replicaUpdatedToDefault {
		properties.ServerUpdateParametersProperties.AdministratorLoginPassword = utils.String(d.Get("administrator_login_password").(string))
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.Name, properties)
	if err != nil {
		return fmt.Errorf("updating %s: %+v", *id, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of %s: %+v", *id, err)
	}

	// Update Admin Password in a separate call when Replication is stopped: https://github.com/Azure/azure-rest-api-specs/issues/16898
	if d.HasChange("administrator_login_password") && replicaUpdatedToDefault {
		properties.ServerUpdateParametersProperties.AdministratorLoginPassword = utils.String(d.Get("administrator_login_password").(string))

		future, err := client.Update(ctx, id.ResourceGroup, id.Name, properties)
		if err != nil {
			return fmt.Errorf("updating Admin Password of %q: %+v", id, err)
		}
		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("waiting for Admin Password update of %q: %+v", id, err)
		}
	}

	if v, ok := d.GetOk("threat_detection_policy"); ok {
		alert := expandSecurityAlertPolicy(v)
		if alert != nil {
			future, err := securityClient.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, *alert)
			if err != nil {
				return fmt.Errorf("updating security alert policy for %s: %+v", *id, err)
			}

			if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for update of security alert policy for %s: %+v", *id, err)
			}
		}
	}

	return resourcePostgreSQLServerRead(d, meta)
}

func resourcePostgreSQLServerRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Postgres.ServersClient
	securityClient := meta.(*clients.Client).Postgres.ServerSecurityAlertPoliciesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServerID(d.Id())
	if err != nil {
		return fmt.Errorf("parsing Postgres Server ID : %v", err)
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[WARN] %s was not found - removing from state", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	tier := postgresql.Basic
	if sku := resp.Sku; sku != nil {
		d.Set("sku_name", sku.Name)
		tier = sku.Tier
	}

	if err := d.Set("identity", flattenServerIdentity(resp.Identity)); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}

	if props := resp.ServerProperties; props != nil {
		d.Set("administrator_login", props.AdministratorLogin)
		d.Set("ssl_minimal_tls_version_enforced", props.MinimalTLSVersion)
		d.Set("version", string(props.Version))

		d.Set("infrastructure_encryption_enabled", props.InfrastructureEncryption == postgresql.InfrastructureEncryptionEnabled)
		d.Set("public_network_access_enabled", props.PublicNetworkAccess == postgresql.PublicNetworkAccessEnumEnabled)
		d.Set("ssl_enforcement_enabled", props.SslEnforcement == postgresql.SslEnforcementEnumEnabled)

		if storage := props.StorageProfile; storage != nil {
			d.Set("storage_mb", storage.StorageMB)
			d.Set("backup_retention_days", storage.BackupRetentionDays)
			d.Set("auto_grow_enabled", storage.StorageAutogrow == postgresql.StorageAutogrowEnabled)
			d.Set("geo_redundant_backup_enabled", storage.GeoRedundantBackup == postgresql.Enabled)
		}

		// Computed
		d.Set("fqdn", props.FullyQualifiedDomainName)
	}

	// the basic does not support threat detection policies
	if tier == postgresql.GeneralPurpose || tier == postgresql.MemoryOptimized {
		secResp, err := securityClient.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil && !utils.ResponseWasNotFound(secResp.Response) {
			return fmt.Errorf("making read request to postgres server security alert policy: %+v", err)
		}

		if !utils.ResponseWasNotFound(secResp.Response) {
			block := flattenSecurityAlertPolicy(secResp.SecurityAlertPolicyProperties, d.Get("threat_detection_policy.0.storage_account_access_key").(string))
			if err := d.Set("threat_detection_policy", block); err != nil {
				return fmt.Errorf("setting `threat_detection_policy`: %+v", err)
			}
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourcePostgreSQLServerDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Postgres.ServersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServerID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}

func indexOfSku(skuName string) int {
	for k, v := range skuList {
		if skuName == v {
			return k
		}
	}
	return -1 // not found.
}

func expandServerSkuName(skuName string) (*postgresql.Sku, error) {
	parts := strings.Split(skuName, "_")
	if len(parts) != 3 {
		return nil, fmt.Errorf("sku_name (%s) has the wrong number of parts (%d) after splitting on _", skuName, len(parts))
	}

	var tier postgresql.SkuTier
	switch parts[0] {
	case "B":
		tier = postgresql.Basic
	case "GP":
		tier = postgresql.GeneralPurpose
	case "MO":
		tier = postgresql.MemoryOptimized
	default:
		return nil, fmt.Errorf("sku_name %s has unknown sku tier %s", skuName, parts[0])
	}

	capacity, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("cannot convert skuname %s capcity %s to int", skuName, parts[2])
	}

	return &postgresql.Sku{
		Name:     utils.String(skuName),
		Tier:     tier,
		Capacity: utils.Int32(int32(capacity)),
		Family:   utils.String(parts[1]),
	}, nil
}

func expandPostgreSQLStorageProfile(d *pluginsdk.ResourceData) *postgresql.StorageProfile {
	storage := postgresql.StorageProfile{}

	// now override whatever we may have from the block with the top level properties
	if v, ok := d.GetOk("auto_grow_enabled"); ok {
		storage.StorageAutogrow = postgresql.StorageAutogrowDisabled
		if v.(bool) {
			storage.StorageAutogrow = postgresql.StorageAutogrowEnabled
		}
	}

	if v, ok := d.GetOk("backup_retention_days"); ok {
		storage.BackupRetentionDays = utils.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("geo_redundant_backup_enabled"); ok {
		storage.GeoRedundantBackup = postgresql.Disabled
		if v.(bool) {
			storage.GeoRedundantBackup = postgresql.Enabled
		}
	}

	if v, ok := d.GetOk("storage_mb"); ok {
		storage.StorageMB = utils.Int32(int32(v.(int)))
	}

	return &storage
}

func expandSecurityAlertPolicy(i interface{}) *postgresql.ServerSecurityAlertPolicy {
	slice := i.([]interface{})
	if len(slice) == 0 {
		return nil
	}

	block := slice[0].(map[string]interface{})

	state := postgresql.ServerSecurityAlertPolicyStateEnabled
	if !block["enabled"].(bool) {
		state = postgresql.ServerSecurityAlertPolicyStateDisabled
	}

	props := &postgresql.SecurityAlertPolicyProperties{
		State: state,
	}

	if v, ok := block["disabled_alerts"]; ok {
		props.DisabledAlerts = utils.ExpandStringSlice(v.(*pluginsdk.Set).List())
	}

	if v, ok := block["email_addresses"]; ok {
		props.EmailAddresses = utils.ExpandStringSlice(v.(*pluginsdk.Set).List())
	}

	if v, ok := block["email_account_admins"]; ok {
		props.EmailAccountAdmins = utils.Bool(v.(bool))
	}

	if v, ok := block["retention_days"]; ok {
		props.RetentionDays = utils.Int32(int32(v.(int)))
	}

	if v, ok := block["storage_account_access_key"]; ok && v.(string) != "" {
		props.StorageAccountAccessKey = utils.String(v.(string))
	}

	if v, ok := block["storage_endpoint"]; ok && v.(string) != "" {
		props.StorageEndpoint = utils.String(v.(string))
	}

	return &postgresql.ServerSecurityAlertPolicy{
		SecurityAlertPolicyProperties: props,
	}
}

func flattenSecurityAlertPolicy(props *postgresql.SecurityAlertPolicyProperties, accessKey string) interface{} {
	if props == nil {
		return nil
	}

	// check if its an empty block as in its never been set before
	if props.DisabledAlerts != nil && len(*props.DisabledAlerts) == 1 && (*props.DisabledAlerts)[0] == "" &&
		props.EmailAddresses != nil && len(*props.EmailAddresses) == 1 && (*props.EmailAddresses)[0] == "" &&
		props.StorageAccountAccessKey != nil && *props.StorageAccountAccessKey == "" &&
		props.StorageEndpoint != nil && *props.StorageEndpoint == "" &&
		props.RetentionDays != nil && *props.RetentionDays == 0 &&
		props.EmailAccountAdmins != nil && !*props.EmailAccountAdmins &&
		props.State == postgresql.ServerSecurityAlertPolicyStateDisabled {
		return nil
	}

	block := map[string]interface{}{}

	block["enabled"] = props.State == postgresql.ServerSecurityAlertPolicyStateEnabled

	block["disabled_alerts"] = flattenSecurityAlertPolicySet(props.DisabledAlerts)
	block["email_addresses"] = flattenSecurityAlertPolicySet(props.EmailAddresses)

	if v := props.EmailAccountAdmins; v != nil {
		block["email_account_admins"] = *v
	}
	if v := props.RetentionDays; v != nil {
		block["retention_days"] = *v
	}
	if v := props.StorageEndpoint; v != nil {
		block["storage_endpoint"] = *v
	}

	block["storage_account_access_key"] = accessKey

	return []interface{}{block}
}

func expandServerIdentity(input []interface{}) (*postgresql.ResourceIdentity, error) {
	expanded, err := identity.ExpandSystemAssigned(input)
	if err != nil {
		return nil, err
	}

	if expanded.Type == identity.TypeNone {
		return nil, nil
	}

	return &postgresql.ResourceIdentity{
		Type: postgresql.IdentityType(string(expanded.Type)),
	}, nil
}

func flattenServerIdentity(input *postgresql.ResourceIdentity) []interface{} {
	var transition *identity.SystemAssigned

	if input != nil {
		transition = &identity.SystemAssigned{
			Type: identity.Type(string(input.Type)),
		}
		if input.PrincipalID != nil {
			transition.PrincipalId = input.PrincipalID.String()
		}
		if input.TenantID != nil {
			transition.TenantId = input.TenantID.String()
		}
	}

	return identity.FlattenSystemAssigned(transition)
}

func flattenSecurityAlertPolicySet(input *[]string) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	// When empty, `disabledAlerts` and `emailAddresses` are returned as `[""]` by the api. We'll catch that here and return
	// an empty interface to set.
	attr := *input
	if len(attr) == 1 && attr[0] == "" {
		return make([]interface{}, 0)
	}

	return utils.FlattenStringSlice(input)
}

func postgreSqlStateRefreshFunc(ctx context.Context, client *postgresql.ServersClient, id parse.ServerId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if !utils.ResponseWasNotFound(res.Response) && err != nil {
			return nil, "", fmt.Errorf("retrieving status of %s: %+v", id, err)
		}

		// This is an issue with the RP, there is a 10 to 15 second lag before the
		// service will actually return the server
		if utils.ResponseWasNotFound(res.Response) {
			return res, string(postgresql.ServerStateInaccessible), nil
		}

		if res.ServerProperties != nil && res.ServerProperties.UserVisibleState != "" {
			return res, string(res.ServerProperties.UserVisibleState), nil
		}

		return res, string(postgresql.ServerStateInaccessible), nil
	}
}
