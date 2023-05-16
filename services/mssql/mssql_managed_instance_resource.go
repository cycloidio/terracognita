package mssql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v5.0/sql"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/identity"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	maintenanceParse "github.com/hashicorp/terraform-provider-azurerm/services/maintenance/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/mssql/validate"
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/sql/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type MsSqlManagedInstanceModel struct {
	AdministratorLogin           string                    `tfschema:"administrator_login"`
	AdministratorLoginPassword   string                    `tfschema:"administrator_login_password"`
	Collation                    string                    `tfschema:"collation"`
	DnsZonePartnerId             string                    `tfschema:"dns_zone_partner_id"`
	Fqdn                         string                    `tfschema:"fqdn"`
	Identity                     []identity.SystemAssigned `tfschema:"identity"`
	LicenseType                  string                    `tfschema:"license_type"`
	Location                     string                    `tfschema:"location"`
	MaintenanceConfigurationName string                    `tfschema:"maintenance_configuration_name"`
	MinimumTlsVersion            string                    `tfschema:"minimum_tls_version"`
	Name                         string                    `tfschema:"name"`
	ProxyOverride                string                    `tfschema:"proxy_override"`
	PublicDataEndpointEnabled    bool                      `tfschema:"public_data_endpoint_enabled"`
	ResourceGroupName            string                    `tfschema:"resource_group_name"`
	SkuName                      string                    `tfschema:"sku_name"`
	StorageAccountType           string                    `tfschema:"storage_account_type"`
	StorageSizeInGb              int                       `tfschema:"storage_size_in_gb"`
	SubnetId                     string                    `tfschema:"subnet_id"`
	Tags                         map[string]string         `tfschema:"tags"`
	TimezoneId                   string                    `tfschema:"timezone_id"`
	VCores                       int                       `tfschema:"vcores"`
}

var _ sdk.Resource = MsSqlManagedInstanceResource{}
var _ sdk.ResourceWithUpdate = MsSqlManagedInstanceResource{}
var _ sdk.ResourceWithCustomizeDiff = MsSqlManagedInstanceResource{}

type MsSqlManagedInstanceResource struct{}

func (r MsSqlManagedInstanceResource) ResourceType() string {
	return "azurerm_mssql_managed_instance"
}

func (r MsSqlManagedInstanceResource) ModelObject() interface{} {
	return &MsSqlManagedInstanceModel{}
}

func (r MsSqlManagedInstanceResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return validate.ManagedInstanceID
}

func (r MsSqlManagedInstanceResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.ValidateMsSqlServerName,
		},

		"location": azure.SchemaLocation(),

		"resource_group_name": azure.SchemaResourceGroupName(),

		"sku_name": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				"BC_Gen4",
				"BC_Gen5",
				"BC_Gen8IH",
				"BC_Gen8IM",
				"GP_Gen4",
				"GP_Gen5",
				"GP_Gen8IH",
				"GP_Gen8IM",
			}, false),
		},

		"administrator_login": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"administrator_login_password": {
			Type:         schema.TypeString,
			Required:     true,
			Sensitive:    true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"license_type": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				"LicenseIncluded",
				"BasePrice",
			}, true),
		},

		"storage_size_in_gb": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(32, 8192),
		},

		"subnet_id": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: networkValidate.SubnetID,
		},

		"vcores": {
			Type:     schema.TypeInt,
			Required: true,
			ValidateFunc: validation.IntInSlice([]int{
				4,
				8,
				16,
				24,
				32,
				40,
				64,
				80,
			}),
		},

		"collation": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "SQL_Latin1_General_CP1_CI_AS",
			ValidateFunc: validation.StringIsNotEmpty,
			ForceNew:     true,
		},

		"dns_zone_partner_id": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validate.ManagedInstanceID,
		},

		"identity": commonschema.SystemAssignedIdentityOptional(),

		"maintenance_configuration_name": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "SQL_Default",
			ValidateFunc: validation.StringInSlice([]string{"SQL_Default", "SQL_EastUS_MI_1", "SQL_EastUS2_MI_1", "SQL_WestUS2_MI_1", "SQL_SoutheastAsia_MI_1", "SQL_AustraliaEast_MI_1", "SQL_NorthEurope_MI_1", "SQL_SouthCentralUS_MI_1",
				"SQL_UKSouth_MI_1", "SQL_WestEurope_MI_1", "SQL_EastUS_MI_2", "SQL_EastUS2_MI_2", "SQL_WestUS2_MI_2", "SQL_SoutheastAsia_MI_2", "SQL_NorthEurope_MI_2", "SQL_SouthCentralUS_MI_2",
				"SQL_UKSouth_MI_2", "SQL_WestEurope_MI_2", "SQL_AustraliaSoutheast_MI_1", "SQL_BrazilSouth_MI_1", "SQL_CanadaCentral_MI_1", "SQL_CanadaEast_MI_1", "SQL_CentralUS_MI_1", "SQL_EastAsia_MI_1",
				"SQL_FranceCentral_MI_1", "SQL_GermanyWestCentral_MI_1", "SQL_CentralIndia_MI_1", "SQL_JapanEast_MI_1", "SQL_JapanWest_MI_1", "SQL_NorthCentralUS_MI_1", "SQL_UKWest_MI_1", "SQL_WestUS_MI_1",
				"SQL_AustraliaSoutheast_MI_2", "SQL_BrazilSouth_MI_2", "SQL_CanadaCentral_MI_2", "SQL_CanadaEast_MI_2", "SQL_CentralUS_MI_2", "SQL_EastAsia_MI_2", "SQL_FranceCentral_MI_2", "SQL_GermanyWestCentral_MI_2",
				"SQL_CentralIndia_MI_2", "SQL_JapanEast_MI_2", "SQL_JapanWest_MI_2", "SQL_NorthCentralUS_MI_2", "SQL_UKWest_MI_2", "SQL_WestUS_MI_2", "SQL_KoreaCentral_MI_1", "SQL_KoreaCentral_MI_2",
				"SQL_WestCentralUS_MI_1", "SQL_WestCentralUS_MI_2", "SQL_UAENorth_MI_1", "SQL_SwitzerlandWest_MI_1", "SQL_SwitzerlandNorth_MI_1", "SQL_UAENorth_MI_2", "SQL_SwitzerlandWest_MI_2",
				"SQL_SwitzerlandNorth_MI_2", "SQL_FranceSouth_MI_1", "SQL_FranceSouth_MI_2", "SQL_SouthAfricaNorth_MI_1", "SQL_KoreaSouth_MI_1", "SQL_UAECentral_MI_1", "SQL_SouthAfricaNorth_MI_2",
				"SQL_KoreaSouth_MI_2", "SQL_UAECentral_MI_2", "SQL_SouthIndia_MI_1", "SQL_SouthIndia_MI_2", "SQL_AustraliaCentral_MI_1", "SQL_AustraliaCentral2_MI_1", "SQL_AustraliaCentral_MI_2",
				"SQL_AustraliaCentral2_MI_2", "SQL_WestIndia_MI_1", "SQL_WestIndia_MI_2", "SQL_SouthAfricaWest_MI_1", "SQL_SouthAfricaWest_MI_2", "SQL_GermanyNorth_MI_1", "SQL_GermanyNorth_MI_2", "SQL_NorwayEast_MI_1",
				"SQL_BrazilSoutheast_MI_1", "SQL_NorwayWest_MI_1", "SQL_WestUS3_MI_1", "SQL_NorwayEast_MI_2", "SQL_BrazilSoutheast_MI_2", "SQL_NorwayWest_MI_2", "SQL_WestUS3_MI_2"}, false),
		},

		"minimum_tls_version": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "1.2",
			ValidateFunc: validation.StringInSlice([]string{
				"1.0",
				"1.1",
				"1.2",
			}, false),
		},

		"proxy_override": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  string(sql.ManagedInstanceProxyOverrideDefault),
			ValidateFunc: validation.StringInSlice([]string{
				string(sql.ManagedInstanceProxyOverrideDefault),
				string(sql.ManagedInstanceProxyOverrideRedirect),
				string(sql.ManagedInstanceProxyOverrideProxy),
			}, false),
		},

		"public_data_endpoint_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},

		"storage_account_type": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			ForceNew: true,
			Default:  string(sql.StorageAccountTypeGRS),
			ValidateFunc: validation.StringInSlice([]string{
				string(sql.StorageAccountTypeGRS),
				string(sql.StorageAccountTypeLRS),
				string(sql.StorageAccountTypeZRS),
			}, false),
		},

		"tags": tags.Schema(),

		"timezone_id": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "UTC",
			ValidateFunc: validation.StringIsNotEmpty,
			ForceNew:     true,
		},
	}
}

func (r MsSqlManagedInstanceResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"fqdn": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func (r MsSqlManagedInstanceResource) CustomizeDiff() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			rd := metadata.ResourceDiff

			// dns_zone_partner_id can only be set on init
			if oldVal, newVal := rd.GetChange("dns_zone_partner_id"); oldVal.(string) == "" && newVal.(string) != "" {
				if err := rd.ForceNew("dns_zone_partner_id"); err != nil {
					return err
				}
			}

			// system-assigned identity can't be removed due to https://github.com/Azure/azure-rest-api-specs/issues/16838
			if oldVal, newVal := rd.GetChange("identity.#"); oldVal.(int) == 1 && newVal.(int) == 0 {
				if err := rd.ForceNew("identity"); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func (r MsSqlManagedInstanceResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 24 * time.Hour,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.MSSQL.ManagedInstancesClient
			subscriptionId := metadata.Client.Account.SubscriptionId

			var model MsSqlManagedInstanceModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			id := parse.NewManagedInstanceID(subscriptionId, model.ResourceGroupName, model.Name)

			metadata.Logger.Infof("Import check for %s", id)
			existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
			if err != nil && !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}

			if !utils.ResponseWasNotFound(existing.Response) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			sku, err := r.expandSkuName(model.SkuName)
			if err != nil {
				return fmt.Errorf("expanding `sku_name` for SQL Managed Instance Server %q: %v", id.ID(), err)
			}

			maintenanceConfigId := maintenanceParse.NewPublicMaintenanceConfigurationID(subscriptionId, model.MaintenanceConfigurationName)

			parameters := sql.ManagedInstance{
				Sku:      sku,
				Identity: r.expandIdentity(model.Identity),
				Location: utils.String(location.Normalize(model.Location)),
				ManagedInstanceProperties: &sql.ManagedInstanceProperties{
					AdministratorLogin:         utils.String(model.AdministratorLogin),
					AdministratorLoginPassword: utils.String(model.AdministratorLoginPassword),
					Collation:                  utils.String(model.Collation),
					DNSZonePartner:             utils.String(model.DnsZonePartnerId),
					LicenseType:                sql.ManagedInstanceLicenseType(model.LicenseType),
					MaintenanceConfigurationID: utils.String(maintenanceConfigId.ID()),
					MinimalTLSVersion:          utils.String(model.MinimumTlsVersion),
					ProxyOverride:              sql.ManagedInstanceProxyOverride(model.ProxyOverride),
					PublicDataEndpointEnabled:  utils.Bool(model.PublicDataEndpointEnabled),
					StorageAccountType:         sql.StorageAccountType(model.StorageAccountType),
					StorageSizeInGB:            utils.Int32(int32(model.StorageSizeInGb)),
					SubnetID:                   utils.String(model.SubnetId),
					TimezoneID:                 utils.String(model.TimezoneId),
					VCores:                     utils.Int32(int32(model.VCores)),
				},
				Tags: tags.FromTypedObject(model.Tags),
			}

			metadata.Logger.Infof("Creating %s", id)

			future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters)
			if err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
				if response.WasConflict(future.Response()) {
					return fmt.Errorf("sql managed instance names need to be globally unique and %q is already in use", id.Name)
				}

				return fmt.Errorf("waiting for creation of %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (r MsSqlManagedInstanceResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 24 * time.Hour,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.MSSQL.ManagedInstancesClient

			id, err := parse.ManagedInstanceID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			metadata.Logger.Infof("Decoding state for %s", id)
			var state MsSqlManagedInstanceModel
			if err := metadata.Decode(&state); err != nil {
				return err
			}

			metadata.Logger.Infof("Updating %s", id)

			sku, err := r.expandSkuName(state.SkuName)
			if err != nil {
				return fmt.Errorf("expanding `sku_name` for SQL Managed Instance Server %q: %v", id.ID(), err)
			}

			properties := sql.ManagedInstance{
				Sku:      sku,
				Identity: r.expandIdentity(state.Identity),
				Location: utils.String(location.Normalize(state.Location)),
				ManagedInstanceProperties: &sql.ManagedInstanceProperties{
					DNSZonePartner:            utils.String(state.DnsZonePartnerId),
					LicenseType:               sql.ManagedInstanceLicenseType(state.LicenseType),
					MinimalTLSVersion:         utils.String(state.MinimumTlsVersion),
					ProxyOverride:             sql.ManagedInstanceProxyOverride(state.ProxyOverride),
					PublicDataEndpointEnabled: utils.Bool(state.PublicDataEndpointEnabled),
					StorageSizeInGB:           utils.Int32(int32(state.StorageSizeInGb)),
					VCores:                    utils.Int32(int32(state.VCores)),
				},
				Tags: tags.FromTypedObject(state.Tags),
			}

			if metadata.ResourceData.HasChange("maintenance_configuration_name") {
				maintenanceConfigId := maintenanceParse.NewPublicMaintenanceConfigurationID(id.SubscriptionId, state.MaintenanceConfigurationName)
				properties.MaintenanceConfigurationID = utils.String(maintenanceConfigId.ID())
			}

			if metadata.ResourceData.HasChange("administrator_password") {
				properties.AdministratorLoginPassword = utils.String(state.AdministratorLoginPassword)
			}

			metadata.Logger.Infof("Updating %s", id)

			future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, properties)
			if err != nil {
				return fmt.Errorf("updating %s: %+v", id, err)
			}

			if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for update of %s: %+v", id, err)
			}

			return nil
		},
	}
}

func (r MsSqlManagedInstanceResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.MSSQL.ManagedInstancesClient

			id, err := parse.ManagedInstanceID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			metadata.Logger.Infof("Decoding state for %s", id)
			var state MsSqlManagedInstanceModel
			if err := metadata.Decode(&state); err != nil {
				return err
			}

			existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
			if err != nil {
				if utils.ResponseWasNotFound(existing.Response) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %v", id, err)
			}

			model := MsSqlManagedInstanceModel{
				Name:              id.Name,
				Location:          location.NormalizeNilable(existing.Location),
				ResourceGroupName: id.ResourceGroup,
				Identity:          r.flattenIdentity(existing.Identity),
				Tags:              tags.ToTypedObject(existing.Tags),

				// This value is not returned, so we'll just set whatever is in the state/config
				AdministratorLoginPassword: state.AdministratorLoginPassword,
			}

			if sku := existing.Sku; sku != nil && sku.Name != nil {
				model.SkuName = r.normalizeSku(*sku.Name)
			}

			if props := existing.ManagedInstanceProperties; props != nil {
				model.LicenseType = string(props.LicenseType)
				model.ProxyOverride = string(props.ProxyOverride)
				model.StorageAccountType = string(props.StorageAccountType)

				if props.AdministratorLogin != nil {
					model.AdministratorLogin = *props.AdministratorLogin
				}
				if props.Collation != nil {
					model.Collation = *props.Collation
				}
				if props.FullyQualifiedDomainName != nil {
					model.Fqdn = *props.FullyQualifiedDomainName
				}
				if props.MaintenanceConfigurationID != nil {
					maintenanceConfigId, err := maintenanceParse.PublicMaintenanceConfigurationID(*props.MaintenanceConfigurationID)
					if err != nil {
						return err
					}
					model.MaintenanceConfigurationName = maintenanceConfigId.Name
				}
				if props.MinimalTLSVersion != nil {
					model.MinimumTlsVersion = *props.MinimalTLSVersion
				}
				if props.PublicDataEndpointEnabled != nil {
					model.PublicDataEndpointEnabled = *props.PublicDataEndpointEnabled
				}
				if props.StorageSizeInGB != nil {
					model.StorageSizeInGb = int(*props.StorageSizeInGB)
				}
				if props.SubnetID != nil {
					model.SubnetId = *props.SubnetID
				}
				if props.TimezoneID != nil {
					model.TimezoneId = *props.TimezoneID
				}
				if props.VCores != nil {
					model.VCores = int(*props.VCores)
				}
			}

			return metadata.Encode(&model)
		},
	}
}

func (r MsSqlManagedInstanceResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 24 * time.Hour,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.MSSQL.ManagedInstancesClient

			id, err := parse.ManagedInstanceID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
			if err != nil {
				return fmt.Errorf("deleting %s: %+v", id, err)
			}

			if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for deletion of %s: %+v", id, err)
			}

			return nil
		},
	}
}

func (r MsSqlManagedInstanceResource) expandIdentity(input []identity.SystemAssigned) *sql.ResourceIdentity {
	if len(input) == 0 {
		return nil
	}

	// Workaround for issue https://github.com/Azure/azure-rest-api-specs/issues/16838
	if input[0].Type == identity.TypeNone {
		return nil
	}

	return &sql.ResourceIdentity{
		Type: sql.IdentityType(input[0].Type),
	}
}

func (r MsSqlManagedInstanceResource) flattenIdentity(input *sql.ResourceIdentity) []identity.SystemAssigned {
	if input == nil {
		return nil
	}

	principalId := ""
	if input.PrincipalID != nil {
		principalId = input.PrincipalID.String()
	}

	tenantId := ""
	if input.TenantID != nil {
		tenantId = input.TenantID.String()
	}

	return []identity.SystemAssigned{{
		Type:        identity.Type(input.Type),
		PrincipalId: principalId,
		TenantId:    tenantId,
	}}
}

func (r MsSqlManagedInstanceResource) expandSkuName(skuName string) (*sql.Sku, error) {
	parts := strings.Split(skuName, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("sku_name (%s) has the wrong number of parts (%d) after splitting on _", skuName, len(parts))
	}

	var tier string
	switch parts[0] {
	case "GP":
		tier = "GeneralPurpose"
	case "BC":
		tier = "BusinessCritical"
	default:
		return nil, fmt.Errorf("sku_name %s has unknown sku tier %s", skuName, parts[0])
	}

	return &sql.Sku{
		Name:   utils.String(skuName),
		Tier:   utils.String(tier),
		Family: utils.String(parts[1]),
	}, nil
}

func (r MsSqlManagedInstanceResource) normalizeSku(sku string) string {
	switch sku {
	case "MIBC64G8IH":
		return "BC_Gen8IH"
	case "MIBC64G8IM":
		return "BC_Gen8IM"
	case "MIGP4G8IH":
		return "GP_Gen8IH"
	case "MIGP4G8IM":
		return "GP_Gen8IM"
	}

	return sku
}
