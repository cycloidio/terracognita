package netapp

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"

	"github.com/Azure/azure-sdk-for-go/services/netapp/mgmt/2021-10-01/netapp"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/netapp/parse"
	netAppValidate "github.com/hashicorp/terraform-provider-azurerm/services/netapp/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceNetAppVolume() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceNetAppVolumeCreate,
		Read:   resourceNetAppVolumeRead,
		Update: resourceNetAppVolumeUpdate,
		Delete: resourceNetAppVolumeDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(60 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(60 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
		},
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VolumeID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: netAppValidate.VolumeName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: netAppValidate.AccountName,
			},

			"pool_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: netAppValidate.PoolName,
			},

			"volume_path": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: netAppValidate.VolumePath,
			},

			"service_level": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(netapp.ServiceLevelPremium),
					string(netapp.ServiceLevelStandard),
					string(netapp.ServiceLevelUltra),
				}, false),
			},

			"subnet_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"create_from_snapshot_resource_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: netAppValidate.SnapshotID,
			},

			"network_features": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(netapp.NetworkFeaturesBasic),
					string(netapp.NetworkFeaturesStandard),
				}, false),
			},

			"protocols": {
				Type:     pluginsdk.TypeSet,
				ForceNew: true,
				Optional: true,
				Computed: true,
				MaxItems: 2,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"NFSv3",
						"NFSv4.1",
						"CIFS",
					}, false),
				},
			},

			"security_style": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Unix", // Using hardcoded values instead of SDK enum since no matter what case is passed,
					"Ntfs", // ANF changes casing to Pascal case in the backend. Please refer to https://github.com/Azure/azure-sdk-for-go/issues/14684
				}, false),
			},

			"storage_quota_in_gb": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(100, 102400),
			},

			"throughput_in_mibps": {
				Type:     pluginsdk.TypeFloat,
				Optional: true,
				Computed: true,
			},

			"export_policy_rule": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"rule_index": {
							Type:         pluginsdk.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 5),
						},

						"allowed_clients": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validate.CIDR,
							},
						},

						"protocols_enabled": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							MinItems: 1,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"NFSv3",
									"NFSv4.1",
									"CIFS",
								}, false),
							},
						},

						"unix_read_only": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},

						"unix_read_write": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},

						"root_access_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"tags": tags.Schema(),

			"mount_ip_addresses": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"snapshot_directory_visible": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"data_protection_replication": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"endpoint_type": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Default:  "dst",
							ValidateFunc: validation.StringInSlice([]string{
								"dst",
							}, false),
						},

						"remote_volume_location": azure.SchemaLocation(),

						"remote_volume_resource_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},

						"replication_frequency": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"10minutes",
								"daily",
								"hourly",
							}, false),
						},
					},
				},
			},

			"data_protection_snapshot_policy": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"snapshot_policy_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
					},
				},
			},
		},
	}
}

func resourceNetAppVolumeCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NetApp.VolumeClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewVolumeID(subscriptionId, d.Get("resource_group_name").(string), d.Get("account_name").(string), d.Get("pool_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.NetAppAccountName, id.CapacityPoolName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_netapp_volume", id.ID())
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	volumePath := d.Get("volume_path").(string)
	serviceLevel := d.Get("service_level").(string)
	subnetID := d.Get("subnet_id").(string)

	networkFeatures := d.Get("network_features").(string)
	if networkFeatures == "" {
		networkFeatures = string(netapp.NetworkFeaturesBasic)
	}

	protocols := d.Get("protocols").(*pluginsdk.Set).List()
	if len(protocols) == 0 {
		protocols = append(protocols, "NFSv3")
	}

	// Handling security style property
	securityStyle := d.Get("security_style").(string)
	if strings.EqualFold(securityStyle, "unix") && len(protocols) == 1 && strings.EqualFold(protocols[0].(string), "cifs") {
		return fmt.Errorf("unix security style cannot be used in a CIFS enabled volume for %s", id)
	}
	if strings.EqualFold(securityStyle, "ntfs") && len(protocols) == 1 && (strings.EqualFold(protocols[0].(string), "nfsv3") || strings.EqualFold(protocols[0].(string), "nfsv4.1")) {
		return fmt.Errorf("ntfs security style cannot be used in a NFSv3/NFSv4.1 enabled volume for %s", id)
	}

	storageQuotaInGB := int64(d.Get("storage_quota_in_gb").(int) * 1073741824)

	exportPolicyRuleRaw := d.Get("export_policy_rule").([]interface{})
	exportPolicyRule := expandNetAppVolumeExportPolicyRule(exportPolicyRuleRaw)

	dataProtectionReplicationRaw := d.Get("data_protection_replication").([]interface{})
	dataProtectionSnapshotPolicyRaw := d.Get("data_protection_snapshot_policy").([]interface{})

	dataProtectionReplication := expandNetAppVolumeDataProtectionReplication(dataProtectionReplicationRaw)
	dataProtectionSnapshotPolicy := expandNetAppVolumeDataProtectionSnapshotPolicy(dataProtectionSnapshotPolicyRaw)

	authorizeReplication := false
	volumeType := ""
	if dataProtectionReplication != nil && dataProtectionReplication.Replication != nil && strings.ToLower(string(dataProtectionReplication.Replication.EndpointType)) == "dst" {
		authorizeReplication = true
		volumeType = "DataProtection"
	}

	// Validating that snapshot policies are not being created in a data protection volume
	if dataProtectionSnapshotPolicy.Snapshot != nil && volumeType != "" {
		return fmt.Errorf("snapshot policy cannot be enabled on a data protection volume, NetApp Volume %q (Resource Group %q)", id.Name, id.ResourceGroup)
	}

	snapshotDirectoryVisible := d.Get("snapshot_directory_visible").(bool)

	// Handling volume creation from snapshot case
	snapshotResourceID := d.Get("create_from_snapshot_resource_id").(string)
	snapshotID := ""
	if snapshotResourceID != "" {
		// Get snapshot ID GUID value
		parsedSnapshotResourceID, err := parse.SnapshotID(snapshotResourceID)
		if err != nil {
			return fmt.Errorf("parsing snapshotResourceID %q: %+v", snapshotResourceID, err)
		}

		snapshotClient := meta.(*clients.Client).NetApp.SnapshotClient
		snapshotResponse, err := snapshotClient.Get(
			ctx,
			parsedSnapshotResourceID.ResourceGroup,
			parsedSnapshotResourceID.NetAppAccountName,
			parsedSnapshotResourceID.CapacityPoolName,
			parsedSnapshotResourceID.VolumeName,
			parsedSnapshotResourceID.Name,
		)
		if err != nil {
			return fmt.Errorf("getting snapshot from NetApp Volume %q (Resource Group %q): %+v", parsedSnapshotResourceID.VolumeName, parsedSnapshotResourceID.ResourceGroup, err)
		}
		snapshotID = *snapshotResponse.SnapshotID

		// Validate if properties that cannot be changed matches (protocols, subnet_id, location, resource group, account_name, pool_name, service_level)
		sourceVolume, err := client.Get(
			ctx,
			parsedSnapshotResourceID.ResourceGroup,
			parsedSnapshotResourceID.NetAppAccountName,
			parsedSnapshotResourceID.CapacityPoolName,
			parsedSnapshotResourceID.VolumeName,
		)
		if err != nil {
			return fmt.Errorf("getting source NetApp Volume (snapshot's parent resource) %q (Resource Group %q): %+v", parsedSnapshotResourceID.VolumeName, parsedSnapshotResourceID.ResourceGroup, err)
		}

		parsedVolumeID, err := parse.VolumeID(*sourceVolume.ID)
		if err != nil {
			return fmt.Errorf("parsing Source Volume ID: %s", err)
		}
		propertyMismatch := []string{}
		if !ValidateSlicesEquality(*sourceVolume.ProtocolTypes, *utils.ExpandStringSlice(protocols), false) {
			propertyMismatch = append(propertyMismatch, "protocols")
		}
		if !strings.EqualFold(*sourceVolume.SubnetID, subnetID) {
			propertyMismatch = append(propertyMismatch, "subnet_id")
		}
		if !strings.EqualFold(*sourceVolume.Location, location) {
			propertyMismatch = append(propertyMismatch, "location")
		}
		if !strings.EqualFold(string(sourceVolume.ServiceLevel), serviceLevel) {
			propertyMismatch = append(propertyMismatch, "service_level")
		}
		if !strings.EqualFold(parsedVolumeID.ResourceGroup, id.ResourceGroup) {
			propertyMismatch = append(propertyMismatch, "resource_group_name")
		}
		if !strings.EqualFold(parsedVolumeID.NetAppAccountName, id.NetAppAccountName) {
			propertyMismatch = append(propertyMismatch, "account_name")
		}
		if !strings.EqualFold(parsedVolumeID.CapacityPoolName, id.CapacityPoolName) {
			propertyMismatch = append(propertyMismatch, "pool_name")
		}
		if len(propertyMismatch) > 0 {
			return fmt.Errorf("following NetApp Volume properties on new Volume from Snapshot does not match Snapshot's source %s: %s", id, strings.Join(propertyMismatch, ", "))
		}
	}

	parameters := netapp.Volume{
		Location: utils.String(location),
		VolumeProperties: &netapp.VolumeProperties{
			CreationToken:   utils.String(volumePath),
			ServiceLevel:    netapp.ServiceLevel(serviceLevel),
			SubnetID:        utils.String(subnetID),
			NetworkFeatures: netapp.NetworkFeatures(networkFeatures),
			ProtocolTypes:   utils.ExpandStringSlice(protocols),
			SecurityStyle:   netapp.SecurityStyle(securityStyle),
			UsageThreshold:  utils.Int64(storageQuotaInGB),
			ExportPolicy:    exportPolicyRule,
			VolumeType:      utils.String(volumeType),
			SnapshotID:      utils.String(snapshotID),
			DataProtection: &netapp.VolumePropertiesDataProtection{
				Replication: dataProtectionReplication.Replication,
				Snapshot:    dataProtectionSnapshotPolicy.Snapshot,
			},
			SnapshotDirectoryVisible: utils.Bool(snapshotDirectoryVisible),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if throughputMibps, ok := d.GetOk("throughput_in_mibps"); ok {
		parameters.VolumeProperties.ThroughputMibps = utils.Float(throughputMibps.(float64))
	}

	future, err := client.CreateOrUpdate(ctx, parameters, id.ResourceGroup, id.NetAppAccountName, id.CapacityPoolName, id.Name)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the creation of %s: %+v", id, err)
	}

	// Waiting for volume be completely provisioned
	if err := waitForVolumeCreateOrUpdate(ctx, client, id); err != nil {
		return err
	}

	// If this is a data replication secondary volume, authorize replication on primary volume
	if authorizeReplication {
		replVolID, err := parse.VolumeID(*dataProtectionReplication.Replication.RemoteVolumeResourceID)
		if err != nil {
			return err
		}

		future, err := client.AuthorizeReplication(
			ctx,
			replVolID.ResourceGroup,
			replVolID.NetAppAccountName,
			replVolID.CapacityPoolName,
			replVolID.Name,
			netapp.AuthorizeRequest{
				RemoteVolumeResourceID: utils.String(id.ID()),
			},
		)
		if err != nil {
			return fmt.Errorf("cannot authorize volume replication: %v", err)
		}

		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("cannot get authorize volume replication future response: %v", err)
		}

		// Wait for volume replication authorization to complete
		log.Printf("[DEBUG] Waiting for replication authorization on NetApp Volume Provisioning Service %q (Resource Group %q) to complete", id.Name, id.ResourceGroup)
		if err := waitForReplAuthorization(ctx, client, id); err != nil {
			return err
		}
	}

	d.SetId(id.ID())

	return resourceNetAppVolumeRead(d, meta)
}

func resourceNetAppVolumeUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NetApp.VolumeClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VolumeID(d.Id())
	if err != nil {
		return err
	}

	shouldUpdate := false
	update := netapp.VolumePatch{
		VolumePatchProperties: &netapp.VolumePatchProperties{},
	}

	if d.HasChange("storage_quota_in_gb") {
		shouldUpdate = true
		storageQuotaInBytes := int64(d.Get("storage_quota_in_gb").(int) * 1073741824)
		update.VolumePatchProperties.UsageThreshold = utils.Int64(storageQuotaInBytes)
	}

	if d.HasChange("export_policy_rule") {
		shouldUpdate = true
		exportPolicyRuleRaw := d.Get("export_policy_rule").([]interface{})
		exportPolicyRule := expandNetAppVolumeExportPolicyRulePatch(exportPolicyRuleRaw)
		update.VolumePatchProperties.ExportPolicy = exportPolicyRule
	}

	if d.HasChange("data_protection_snapshot_policy") {
		// Validating that snapshot policies are not being created in a data protection volume
		dataProtectionReplicationRaw := d.Get("data_protection_replication").([]interface{})
		dataProtectionReplication := expandNetAppVolumeDataProtectionReplication(dataProtectionReplicationRaw)

		if dataProtectionReplication != nil && dataProtectionReplication.Replication != nil && strings.ToLower(string(dataProtectionReplication.Replication.EndpointType)) == "dst" {
			return fmt.Errorf("snapshot policy cannot be enabled on a data protection volume, NetApp Volume %q (Resource Group %q)", id.Name, id.ResourceGroup)
		}

		shouldUpdate = true
		dataProtectionSnapshotPolicyRaw := d.Get("data_protection_snapshot_policy").([]interface{})
		dataProtectionSnapshotPolicy := expandNetAppVolumeDataProtectionSnapshotPolicyPatch(dataProtectionSnapshotPolicyRaw)
		update.VolumePatchProperties.DataProtection = dataProtectionSnapshotPolicy
	}

	if d.HasChange("throughput_in_mibps") {
		shouldUpdate = true
		throughputMibps := d.Get("throughput_in_mibps")
		update.VolumePatchProperties.ThroughputMibps = utils.Float(throughputMibps.(float64))
	}

	if d.HasChange("tags") {
		shouldUpdate = true
		tagsRaw := d.Get("tags").(map[string]interface{})
		update.Tags = tags.Expand(tagsRaw)
	}

	if shouldUpdate {
		future, err := client.Update(ctx, update, id.ResourceGroup, id.NetAppAccountName, id.CapacityPoolName, id.Name)
		if err != nil {
			return fmt.Errorf("updating Volume %q: %+v", id.Name, err)
		}
		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("waiting for the update of %s: %+v", id, err)
		}

		// Wait for volume to complete update
		if err := waitForVolumeCreateOrUpdate(ctx, client, *id); err != nil {
			return err
		}
	}

	return resourceNetAppVolumeRead(d, meta)
}

func resourceNetAppVolumeRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NetApp.VolumeClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VolumeID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.NetAppAccountName, id.CapacityPoolName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s was not found - removing from state", *id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("account_name", id.NetAppAccountName)
	d.Set("pool_name", id.CapacityPoolName)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}
	if props := resp.VolumeProperties; props != nil {
		d.Set("volume_path", props.CreationToken)
		d.Set("service_level", props.ServiceLevel)
		d.Set("subnet_id", props.SubnetID)
		d.Set("network_features", props.NetworkFeatures)
		d.Set("protocols", props.ProtocolTypes)
		d.Set("security_style", props.SecurityStyle)
		d.Set("snapshot_directory_visible", props.SnapshotDirectoryVisible)
		d.Set("throughput_in_mibps", props.ThroughputMibps)
		if props.UsageThreshold != nil {
			d.Set("storage_quota_in_gb", *props.UsageThreshold/1073741824)
		}
		if err := d.Set("export_policy_rule", flattenNetAppVolumeExportPolicyRule(props.ExportPolicy)); err != nil {
			return fmt.Errorf("setting `export_policy_rule`: %+v", err)
		}
		if err := d.Set("mount_ip_addresses", flattenNetAppVolumeMountIPAddresses(props.MountTargets)); err != nil {
			return fmt.Errorf("setting `mount_ip_addresses`: %+v", err)
		}
		if err := d.Set("data_protection_replication", flattenNetAppVolumeDataProtectionReplication(props.DataProtection)); err != nil {
			return fmt.Errorf("setting `data_protection_replication`: %+v", err)
		}
		if err := d.Set("data_protection_snapshot_policy", flattenNetAppVolumeDataProtectionSnapshotPolicy(props.DataProtection)); err != nil {
			return fmt.Errorf("setting `data_protection_snapshot_policy`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceNetAppVolumeDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NetApp.VolumeClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VolumeID(d.Id())
	if err != nil {
		return err
	}

	// Removing replication if present
	dataProtectionReplicationRaw := d.Get("data_protection_replication").([]interface{})
	dataProtectionReplication := expandNetAppVolumeDataProtectionReplication(dataProtectionReplicationRaw)

	if replicaVolumeId := id; dataProtectionReplication != nil && dataProtectionReplication.Replication != nil {
		if dataProtectionReplication.Replication.RemoteVolumeResourceID == nil {
			return fmt.Errorf("remote volume id was nil")
		}

		if strings.ToLower(string(dataProtectionReplication.Replication.EndpointType)) != "dst" {
			// This is the case where primary volume started the deletion, in this case, to be consistent we will remove replication from secondary
			replicaVolumeId, err = parse.VolumeID(*dataProtectionReplication.Replication.RemoteVolumeResourceID)
			if err != nil {
				return err
			}
		}

		// Checking replication status before deletion, it need to be broken before proceeding with deletion
		if res, err := client.ReplicationStatusMethod(ctx, replicaVolumeId.ResourceGroup, replicaVolumeId.NetAppAccountName, replicaVolumeId.CapacityPoolName, replicaVolumeId.Name); err == nil {
			// Wait for replication state = "mirrored"
			if strings.ToLower(string(res.MirrorState)) == "uninitialized" {
				if err := waitForReplMirrorState(ctx, client, *replicaVolumeId, "mirrored"); err != nil {
					return fmt.Errorf("waiting for replica %s to become 'mirrored': %+v", *replicaVolumeId, err)
				}
			}

			// Breaking replication
			_, err = client.BreakReplication(ctx,
				replicaVolumeId.ResourceGroup,
				replicaVolumeId.NetAppAccountName,
				replicaVolumeId.CapacityPoolName,
				replicaVolumeId.Name,
				&netapp.BreakReplicationRequest{
					ForceBreakReplication: utils.Bool(true),
				})

			if err != nil {
				return fmt.Errorf("breaking replication for %s: %+v", *replicaVolumeId, err)
			}

			// Waiting for replication be in broken state
			log.Printf("[DEBUG] Waiting for the replication of %s to be in broken state", *replicaVolumeId)
			if err := waitForReplMirrorState(ctx, client, *replicaVolumeId, "broken"); err != nil {
				return fmt.Errorf("waiting for the breaking of replication for %s: %+v", *replicaVolumeId, err)
			}
		}

		// Deleting replication and waiting for it to fully complete the operation
		future, err := client.DeleteReplication(ctx, replicaVolumeId.ResourceGroup, replicaVolumeId.NetAppAccountName, replicaVolumeId.CapacityPoolName, replicaVolumeId.Name)
		if err != nil {
			return fmt.Errorf("deleting replicate %s: %+v", *replicaVolumeId, err)
		}

		log.Printf("[DEBUG] Waiting for the replica of %s to be deleted", replicaVolumeId)
		if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("waiting for the replica %s to be deleted: %+v", *replicaVolumeId, err)
		}
		if err := waitForReplicationDeletion(ctx, client, *replicaVolumeId); err != nil {
			return fmt.Errorf("waiting for the replica %s to be deleted: %+v", *replicaVolumeId, err)
		}
	}

	// Deleting volume and waiting for it fo fully complete the operation
	future, err := client.Delete(ctx, id.ResourceGroup, id.NetAppAccountName, id.CapacityPoolName, id.Name, utils.Bool(true))
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	log.Printf("[DEBUG] Waiting for %s to be deleted", *id)
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %q: %+v", id, err)
	}
	if err := waitForVolumeDeletion(ctx, client, *id); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}

func waitForVolumeCreateOrUpdate(ctx context.Context, client *netapp.VolumesClient, id parse.VolumeId) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("context had no deadline")
	}
	stateConf := &pluginsdk.StateChangeConf{
		ContinuousTargetOccurence: 5,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending:                   []string{"204", "404"},
		Target:                    []string{"200", "202"},
		Refresh:                   netappVolumeStateRefreshFunc(ctx, client, id),
		Timeout:                   time.Until(deadline),
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for %s to finish creating: %+v", id, err)
	}

	return nil
}

func waitForReplAuthorization(ctx context.Context, client *netapp.VolumesClient, id parse.VolumeId) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("context had no deadline")
	}
	stateConf := &pluginsdk.StateChangeConf{
		ContinuousTargetOccurence: 5,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending:                   []string{"204", "404", "400"}, // TODO: Remove 400 when bug is fixed on RP side, where replicationStatus returns 400 at some point during authorization process
		Target:                    []string{"200", "202"},
		Refresh:                   netappVolumeReplicationStateRefreshFunc(ctx, client, id),
		Timeout:                   time.Until(deadline),
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for replication authorization NetApp Volume Provisioning Service %q (Resource Group %q) to complete: %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}

func waitForReplMirrorState(ctx context.Context, client *netapp.VolumesClient, id parse.VolumeId, desiredState string) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("context had no deadline")
	}
	stateConf := &pluginsdk.StateChangeConf{
		ContinuousTargetOccurence: 5,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending:                   []string{"200"}, // 200 means mirror state is still Mirrored
		Target:                    []string{"204"}, // 204 means mirror state is <> than Mirrored
		Refresh:                   netappVolumeReplicationMirrorStateRefreshFunc(ctx, client, id, desiredState),
		Timeout:                   time.Until(deadline),
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for %s to be in the state %q: %+v", id, desiredState, err)
	}

	return nil
}

func waitForReplicationDeletion(ctx context.Context, client *netapp.VolumesClient, id parse.VolumeId) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("context had no deadline")
	}

	stateConf := &pluginsdk.StateChangeConf{
		ContinuousTargetOccurence: 5,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending:                   []string{"200", "202", "400"}, // TODO: Remove 400 when bug is fixed on RP side, where replicationStatus returns 400 while it is in "Deleting" state
		Target:                    []string{"404"},
		Refresh:                   netappVolumeReplicationStateRefreshFunc(ctx, client, id),
		Timeout:                   time.Until(deadline),
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for Replication of %s to be deleted: %+v", id, err)
	}

	return nil
}

func waitForVolumeDeletion(ctx context.Context, client *netapp.VolumesClient, id parse.VolumeId) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("context had no deadline")
	}
	stateConf := &pluginsdk.StateChangeConf{
		ContinuousTargetOccurence: 5,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending:                   []string{"200", "202"},
		Target:                    []string{"204", "404"},
		Refresh:                   netappVolumeStateRefreshFunc(ctx, client, id),
		Timeout:                   time.Until(deadline),
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for %s to be deleted: %+v", id, err)
	}

	return nil
}

func netappVolumeStateRefreshFunc(ctx context.Context, client *netapp.VolumesClient, id parse.VolumeId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, id.ResourceGroup, id.NetAppAccountName, id.CapacityPoolName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(res.Response) {
				return nil, "", fmt.Errorf("retrieving NetApp Volume %q (Resource Group %q): %s", id.Name, id.ResourceGroup, err)
			}
		}

		return res, strconv.Itoa(res.StatusCode), nil
	}
}

func netappVolumeReplicationMirrorStateRefreshFunc(ctx context.Context, client *netapp.VolumesClient, id parse.VolumeId, desiredState string) pluginsdk.StateRefreshFunc {
	validStates := []string{"mirrored", "broken", "uninitialized"}

	return func() (interface{}, string, error) {
		// Possible Mirror States to be used as desiredStates:
		// mirrored, broken or uninitialized
		if !utils.SliceContainsValue(validStates, strings.ToLower(desiredState)) {
			return nil, "", fmt.Errorf("Invalid desired mirror state was passed to check mirror replication state (%s), possible values: (%+v)", desiredState, netapp.PossibleMirrorStateValues())
		}

		res, err := client.ReplicationStatusMethod(ctx, id.ResourceGroup, id.NetAppAccountName, id.CapacityPoolName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(res.Response) {
				return nil, "", fmt.Errorf("retrieving replication status information from NetApp Volume %q (Resource Group %q): %s", id.Name, id.ResourceGroup, err)
			}
		}

		// TODO: fix this refresh function to use strings instead of fake status codes
		// Setting 200 as default response
		response := 200
		if strings.EqualFold(string(res.MirrorState), desiredState) {
			// return 204 if state matches desired state
			response = 204
		}

		return res, strconv.Itoa(response), nil
	}
}

func netappVolumeReplicationStateRefreshFunc(ctx context.Context, client *netapp.VolumesClient, id parse.VolumeId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.ReplicationStatusMethod(ctx, id.ResourceGroup, id.NetAppAccountName, id.CapacityPoolName, id.Name)
		if err != nil {
			if res.StatusCode == 400 && (strings.Contains(strings.ToLower(err.Error()), "deleting") || strings.Contains(strings.ToLower(err.Error()), "volume replication missing or deleted")) {
				// This error can be ignored until a bug is fixed on RP side that it is returning 400 while the replication is in "Deleting" process
				// TODO: remove this workaround when above bug is fixed
			} else if !utils.ResponseWasNotFound(res.Response) {
				return nil, "", fmt.Errorf("retrieving replication status from NetApp Volume %q (Resource Group %q): %s", id.Name, id.ResourceGroup, err)
			}
		}

		return res, strconv.Itoa(res.StatusCode), nil
	}
}

func expandNetAppVolumeExportPolicyRule(input []interface{}) *netapp.VolumePropertiesExportPolicy {
	results := make([]netapp.ExportPolicyRule, 0)
	for _, item := range input {
		if item != nil {
			v := item.(map[string]interface{})
			ruleIndex := int32(v["rule_index"].(int))
			allowedClients := strings.Join(*utils.ExpandStringSlice(v["allowed_clients"].(*pluginsdk.Set).List()), ",")

			cifsEnabled := false
			nfsv3Enabled := false
			nfsv41Enabled := false

			if vpe := v["protocols_enabled"]; vpe != nil {
				protocolsEnabled := vpe.([]interface{})
				if len(protocolsEnabled) != 0 {
					for _, protocol := range protocolsEnabled {
						if protocol != nil {
							switch strings.ToLower(protocol.(string)) {
							case "cifs":
								cifsEnabled = true
							case "nfsv3":
								nfsv3Enabled = true
							case "nfsv4.1":
								nfsv41Enabled = true
							}
						}
					}
				}
			}

			unixReadOnly := v["unix_read_only"].(bool)
			unixReadWrite := v["unix_read_write"].(bool)
			rootAccessEnabled := v["root_access_enabled"].(bool)

			result := netapp.ExportPolicyRule{
				AllowedClients: utils.String(allowedClients),
				Cifs:           utils.Bool(cifsEnabled),
				Nfsv3:          utils.Bool(nfsv3Enabled),
				Nfsv41:         utils.Bool(nfsv41Enabled),
				RuleIndex:      utils.Int32(ruleIndex),
				UnixReadOnly:   utils.Bool(unixReadOnly),
				UnixReadWrite:  utils.Bool(unixReadWrite),
				HasRootAccess:  utils.Bool(rootAccessEnabled),
			}

			results = append(results, result)
		}
	}

	return &netapp.VolumePropertiesExportPolicy{
		Rules: &results,
	}
}

func expandNetAppVolumeExportPolicyRulePatch(input []interface{}) *netapp.VolumePatchPropertiesExportPolicy {
	results := make([]netapp.ExportPolicyRule, 0)
	for _, item := range input {
		if item != nil {
			v := item.(map[string]interface{})
			ruleIndex := int32(v["rule_index"].(int))
			allowedClients := strings.Join(*utils.ExpandStringSlice(v["allowed_clients"].(*pluginsdk.Set).List()), ",")

			cifsEnabled := false
			nfsv3Enabled := false
			nfsv41Enabled := false

			if vpe := v["protocols_enabled"]; vpe != nil {
				protocolsEnabled := vpe.([]interface{})
				if len(protocolsEnabled) != 0 {
					for _, protocol := range protocolsEnabled {
						if protocol != nil {
							switch strings.ToLower(protocol.(string)) {
							case "cifs":
								cifsEnabled = true
							case "nfsv3":
								nfsv3Enabled = true
							case "nfsv4.1":
								nfsv41Enabled = true
							}
						}
					}
				}
			}

			unixReadOnly := v["unix_read_only"].(bool)
			unixReadWrite := v["unix_read_write"].(bool)
			rootAccessEnabled := v["root_access_enabled"].(bool)

			result := netapp.ExportPolicyRule{
				AllowedClients: utils.String(allowedClients),
				Cifs:           utils.Bool(cifsEnabled),
				Nfsv3:          utils.Bool(nfsv3Enabled),
				Nfsv41:         utils.Bool(nfsv41Enabled),
				RuleIndex:      utils.Int32(ruleIndex),
				UnixReadOnly:   utils.Bool(unixReadOnly),
				UnixReadWrite:  utils.Bool(unixReadWrite),
				HasRootAccess:  utils.Bool(rootAccessEnabled),
			}

			results = append(results, result)
		}
	}

	return &netapp.VolumePatchPropertiesExportPolicy{
		Rules: &results,
	}
}

func expandNetAppVolumeDataProtectionReplication(input []interface{}) *netapp.VolumePropertiesDataProtection {
	if len(input) == 0 || input[0] == nil {
		return &netapp.VolumePropertiesDataProtection{}
	}

	replicationObject := netapp.ReplicationObject{}

	replicationRaw := input[0].(map[string]interface{})

	if v, ok := replicationRaw["endpoint_type"]; ok {
		replicationObject.EndpointType = netapp.EndpointType(v.(string))
	}
	if v, ok := replicationRaw["remote_volume_location"]; ok {
		replicationObject.RemoteVolumeRegion = utils.String(v.(string))
	}
	if v, ok := replicationRaw["remote_volume_resource_id"]; ok {
		replicationObject.RemoteVolumeResourceID = utils.String(v.(string))
	}
	if v, ok := replicationRaw["replication_frequency"]; ok {
		replicationObject.ReplicationSchedule = netapp.ReplicationSchedule(translateTFSchedule(v.(string)))
	}

	return &netapp.VolumePropertiesDataProtection{
		Replication: &replicationObject,
	}
}

func expandNetAppVolumeDataProtectionSnapshotPolicy(input []interface{}) *netapp.VolumePropertiesDataProtection {
	if len(input) == 0 || input[0] == nil {
		return &netapp.VolumePropertiesDataProtection{}
	}

	snapshotObject := netapp.VolumeSnapshotProperties{}

	snapshotRaw := input[0].(map[string]interface{})

	if v, ok := snapshotRaw["snapshot_policy_id"]; ok {
		snapshotObject.SnapshotPolicyID = utils.String(v.(string))
	}

	return &netapp.VolumePropertiesDataProtection{
		Snapshot: &snapshotObject,
	}
}

func expandNetAppVolumeDataProtectionSnapshotPolicyPatch(input []interface{}) *netapp.VolumePatchPropertiesDataProtection {
	if len(input) == 0 || input[0] == nil {
		return &netapp.VolumePatchPropertiesDataProtection{}
	}

	snapshotObject := netapp.VolumeSnapshotProperties{}

	snapshotRaw := input[0].(map[string]interface{})

	if v, ok := snapshotRaw["snapshot_policy_id"]; ok {
		snapshotObject.SnapshotPolicyID = utils.String(v.(string))
	}

	return &netapp.VolumePatchPropertiesDataProtection{
		Snapshot: &snapshotObject,
	}
}

func flattenNetAppVolumeExportPolicyRule(input *netapp.VolumePropertiesExportPolicy) []interface{} {
	results := make([]interface{}, 0)
	if input == nil || input.Rules == nil {
		return results
	}

	for _, item := range *input.Rules {
		ruleIndex := int32(0)
		if v := item.RuleIndex; v != nil {
			ruleIndex = *v
		}
		allowedClients := []string{}
		if v := item.AllowedClients; v != nil {
			allowedClients = strings.Split(*v, ",")
		}

		protocolsEnabled := []string{}
		if v := item.Cifs; v != nil {
			if *v {
				protocolsEnabled = append(protocolsEnabled, "CIFS")
			}
		}
		if v := item.Nfsv3; v != nil {
			if *v {
				protocolsEnabled = append(protocolsEnabled, "NFSv3")
			}
		}
		if v := item.Nfsv41; v != nil {
			if *v {
				protocolsEnabled = append(protocolsEnabled, "NFSv4.1")
			}
		}
		unixReadOnly := false
		if v := item.UnixReadOnly; v != nil {
			unixReadOnly = *v
		}
		unixReadWrite := false
		if v := item.UnixReadWrite; v != nil {
			unixReadWrite = *v
		}
		rootAccessEnabled := false
		if v := item.HasRootAccess; v != nil {
			rootAccessEnabled = *v
		}

		result := map[string]interface{}{
			"rule_index":          ruleIndex,
			"allowed_clients":     utils.FlattenStringSlice(&allowedClients),
			"unix_read_only":      unixReadOnly,
			"unix_read_write":     unixReadWrite,
			"root_access_enabled": rootAccessEnabled,
			"protocols_enabled":   utils.FlattenStringSlice(&protocolsEnabled),
		}
		results = append(results, result)
	}

	return results
}

func flattenNetAppVolumeMountIPAddresses(input *[]netapp.MountTargetProperties) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		if item.IPAddress != nil {
			results = append(results, item.IPAddress)
		}
	}

	return results
}

func flattenNetAppVolumeDataProtectionReplication(input *netapp.VolumePropertiesDataProtection) []interface{} {
	if input == nil || input.Replication == nil {
		return []interface{}{}
	}

	if strings.ToLower(string(input.Replication.EndpointType)) == "" || strings.ToLower(string(input.Replication.EndpointType)) != "dst" {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"endpoint_type":             strings.ToLower(string(input.Replication.EndpointType)),
			"remote_volume_location":    location.NormalizeNilable(input.Replication.RemoteVolumeRegion),
			"remote_volume_resource_id": input.Replication.RemoteVolumeResourceID,
			"replication_frequency":     translateSDKSchedule(strings.ToLower(string(input.Replication.ReplicationSchedule))),
		},
	}
}

func flattenNetAppVolumeDataProtectionSnapshotPolicy(input *netapp.VolumePropertiesDataProtection) []interface{} {
	if input == nil || input.Snapshot == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"snapshot_policy_id": input.Snapshot.SnapshotPolicyID,
		},
	}
}

func translateTFSchedule(scheduleName string) string {
	if strings.EqualFold(scheduleName, "10minutes") {
		return "_10minutely"
	}

	return scheduleName
}

func translateSDKSchedule(scheduleName string) string {
	if strings.EqualFold(scheduleName, "_10minutely") {
		return "10minutes"
	}

	return scheduleName
}
