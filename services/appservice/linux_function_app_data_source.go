package appservice

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2021-02-01/web"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/services/appservice/helpers"
	"github.com/hashicorp/terraform-provider-azurerm/services/appservice/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/appservice/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type LinuxFunctionAppDataSource struct{}

type LinuxFunctionAppDataSourceModel struct {
	Name               string `tfschema:"name"`
	ResourceGroup      string `tfschema:"resource_group_name"`
	Location           string `tfschema:"location"`
	ServicePlanId      string `tfschema:"service_plan_id"`
	StorageAccountName string `tfschema:"storage_account_name"`

	StorageAccountKey       string `tfschema:"storage_account_access_key"`
	StorageUsesMSI          bool   `tfschema:"storage_uses_managed_identity"` // Storage uses MSI not account key
	StorageKeyVaultSecretID string `tfschema:"storage_key_vault_secret_id"`

	AppSettings               map[string]string                    `tfschema:"app_settings"`
	AuthSettings              []helpers.AuthSettings               `tfschema:"auth_settings"`
	Backup                    []helpers.Backup                     `tfschema:"backup"` // Not supported on Dynamic or Basic plans
	BuiltinLogging            bool                                 `tfschema:"builtin_logging_enabled"`
	ClientCertEnabled         bool                                 `tfschema:"client_certificate_enabled"`
	ClientCertMode            string                               `tfschema:"client_certificate_mode"`
	ConnectionStrings         []helpers.ConnectionString           `tfschema:"connection_string"`
	DailyMemoryTimeQuota      int                                  `tfschema:"daily_memory_time_quota"`
	Enabled                   bool                                 `tfschema:"enabled"`
	FunctionExtensionsVersion string                               `tfschema:"functions_extension_version"`
	ForceDisableContentShare  bool                                 `tfschema:"content_share_force_disabled"`
	HttpsOnly                 bool                                 `tfschema:"https_only"`
	SiteConfig                []helpers.SiteConfigLinuxFunctionApp `tfschema:"site_config"`
	StickySettings            []helpers.StickySettings             `tfschema:"sticky_settings"`
	Tags                      map[string]string                    `tfschema:"tags"`

	CustomDomainVerificationId    string   `tfschema:"custom_domain_verification_id"`
	DefaultHostname               string   `tfschema:"default_hostname"`
	Kind                          string   `tfschema:"kind"`
	OutboundIPAddresses           string   `tfschema:"outbound_ip_addresses"`
	OutboundIPAddressList         []string `tfschema:"outbound_ip_address_list"`
	PossibleOutboundIPAddresses   string   `tfschema:"possible_outbound_ip_addresses"`
	PossibleOutboundIPAddressList []string `tfschema:"possible_outbound_ip_address_list"`

	SiteCredentials []helpers.SiteCredential `tfschema:"site_credential"`
}

func (d LinuxFunctionAppDataSource) ModelObject() interface{} {
	return &LinuxFunctionAppDataSourceModel{}
}

func (d LinuxFunctionAppDataSource) ResourceType() string {
	return "azurerm_linux_function_app"
}

func (d LinuxFunctionAppDataSource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return validate.FunctionAppID
}

func (d LinuxFunctionAppDataSource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validate.WebAppName,
		},

		"resource_group_name": commonschema.ResourceGroupNameForDataSource(),
	}
}

func (d LinuxFunctionAppDataSource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"location": commonschema.LocationComputed(),

		"service_plan_id": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"storage_account_name": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"storage_account_access_key": {
			Type:      pluginsdk.TypeString,
			Sensitive: true,
			Computed:  true,
		},

		"storage_uses_managed_identity": {
			Type:     pluginsdk.TypeBool,
			Computed: true,
		},

		"storage_key_vault_secret_id": {
			Type:        pluginsdk.TypeString,
			Computed:    true,
			Description: "The Key Vault Secret ID, including version, that contains the Connection String used to connect to the storage account for this Function App.",
		},

		"app_settings": {
			Type:     pluginsdk.TypeMap,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"auth_settings": helpers.AuthSettingsSchemaComputed(),

		"backup": helpers.BackupSchemaComputed(),

		"builtin_logging_enabled": {
			Type:     pluginsdk.TypeBool,
			Computed: true,
		},

		"client_certificate_enabled": {
			Type:     pluginsdk.TypeBool,
			Computed: true,
		},

		"client_certificate_mode": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"connection_string": helpers.ConnectionStringSchemaComputed(),

		"daily_memory_time_quota": {
			Type:     pluginsdk.TypeInt,
			Computed: true,
		},

		"enabled": {
			Type:     pluginsdk.TypeBool,
			Computed: true,
		},

		"content_share_force_disabled": {
			Type:     pluginsdk.TypeBool,
			Computed: true,
		},

		"functions_extension_version": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"https_only": {
			Type:     pluginsdk.TypeBool,
			Computed: true,
		},

		"identity": commonschema.SystemAssignedUserAssignedIdentityComputed(),

		"site_config": helpers.SiteConfigSchemaLinuxFunctionAppComputed(),

		"tags": tags.SchemaDataSource(),

		"custom_domain_verification_id": {
			Type:      pluginsdk.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		"default_hostname": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"kind": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"outbound_ip_addresses": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"outbound_ip_address_list": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"possible_outbound_ip_addresses": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"possible_outbound_ip_address_list": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"site_credential": helpers.SiteCredentialSchema(),

		"sticky_settings": helpers.StickySettingsComputedSchema(),
	}
}

func (d LinuxFunctionAppDataSource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 25 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.AppService.WebAppsClient
			subscriptionId := metadata.Client.Account.SubscriptionId

			var linuxFunctionApp LinuxFunctionAppDataSourceModel
			if err := metadata.Decode(&linuxFunctionApp); err != nil {
				return err
			}

			id := parse.NewFunctionAppID(subscriptionId, linuxFunctionApp.ResourceGroup, linuxFunctionApp.Name)

			functionApp, err := client.Get(ctx, id.ResourceGroup, id.SiteName)
			if err != nil {
				if utils.ResponseWasNotFound(functionApp.Response) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("reading Linux %s: %+v", id, err)
			}

			if functionApp.SiteProperties == nil {
				return fmt.Errorf("reading properties of Linux %s", id)
			}
			props := *functionApp.SiteProperties

			appSettingsResp, err := client.ListApplicationSettings(ctx, id.ResourceGroup, id.SiteName)
			if err != nil {
				return fmt.Errorf("reading App Settings for Linux %s: %+v", id, err)
			}

			connectionStrings, err := client.ListConnectionStrings(ctx, id.ResourceGroup, id.SiteName)
			if err != nil {
				return fmt.Errorf("reading Connection String information for Linux %s: %+v", id, err)
			}

			stickySettings, err := client.ListSlotConfigurationNames(ctx, id.ResourceGroup, id.SiteName)
			if err != nil {
				return fmt.Errorf("reading Sticky Settings for Linux %s: %+v", id, err)
			}

			siteCredentialsFuture, err := client.ListPublishingCredentials(ctx, id.ResourceGroup, id.SiteName)
			if err != nil {
				return fmt.Errorf("listing Site Publishing Credential information for Linux %s: %+v", id, err)
			}

			if err := siteCredentialsFuture.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for Site Publishing Credential information for Linux %s: %+v", id, err)
			}
			siteCredentials, err := siteCredentialsFuture.Result(*client)
			if err != nil {
				return fmt.Errorf("reading Site Publishing Credential information for Linux %s: %+v", id, err)
			}

			auth, err := client.GetAuthSettings(ctx, id.ResourceGroup, id.SiteName)
			if err != nil {
				return fmt.Errorf("reading Auth Settings for Linux %s: %+v", id, err)
			}

			backup, err := client.GetBackupConfiguration(ctx, id.ResourceGroup, id.SiteName)
			if err != nil {
				if !utils.ResponseWasNotFound(backup.Response) {
					return fmt.Errorf("reading Backup Settings for Linux %s: %+v", id, err)
				}
			}

			logs, err := client.GetDiagnosticLogsConfiguration(ctx, id.ResourceGroup, id.SiteName)
			if err != nil {
				return fmt.Errorf("reading logs configuration for Linux %s: %+v", id, err)
			}

			state := LinuxFunctionAppDataSourceModel{
				Name:                 id.SiteName,
				ResourceGroup:        id.ResourceGroup,
				ServicePlanId:        utils.NormalizeNilableString(props.ServerFarmID),
				Location:             location.NormalizeNilable(functionApp.Location),
				Enabled:              utils.NormaliseNilableBool(functionApp.Enabled),
				ClientCertMode:       string(functionApp.ClientCertMode),
				DailyMemoryTimeQuota: int(utils.NormaliseNilableInt32(props.DailyMemoryTimeQuota)),
				StickySettings:       helpers.FlattenStickySettings(stickySettings.SlotConfigNames),
				Tags:                 tags.ToTypedObject(functionApp.Tags),
				Kind:                 utils.NormalizeNilableString(functionApp.Kind),
			}

			configResp, err := client.GetConfiguration(ctx, id.ResourceGroup, id.SiteName)
			if err != nil {
				return fmt.Errorf("making Read request on AzureRM Function App Configuration %q: %+v", id.SiteName, err)
			}

			siteConfig, err := helpers.FlattenSiteConfigLinuxFunctionApp(configResp.SiteConfig)
			if err != nil {
				return fmt.Errorf("reading Site Config for Linux %s: %+v", id, err)
			}
			state.SiteConfig = []helpers.SiteConfigLinuxFunctionApp{*siteConfig}

			state.unpackLinuxFunctionAppSettings(appSettingsResp, metadata)

			state.ConnectionStrings = helpers.FlattenConnectionStrings(connectionStrings)

			state.SiteCredentials = helpers.FlattenSiteCredentials(siteCredentials)

			state.AuthSettings = helpers.FlattenAuthSettings(auth)

			state.Backup = helpers.FlattenBackupConfig(backup)

			state.SiteConfig[0].AppServiceLogs = helpers.FlattenFunctionAppAppServiceLogs(logs)

			state.HttpsOnly = utils.NormaliseNilableBool(functionApp.HTTPSOnly)
			state.ClientCertEnabled = utils.NormaliseNilableBool(functionApp.ClientCertEnabled)

			metadata.SetID(id)

			if err := metadata.Encode(&state); err != nil {
				return fmt.Errorf("encoding: %+v", err)
			}

			flattenedIdentity, err := flattenIdentity(functionApp.Identity)
			if err != nil {
				return fmt.Errorf("flattening `identity`: %+v", err)
			}
			if err := metadata.ResourceData.Set("identity", flattenedIdentity); err != nil {
				return fmt.Errorf("setting `identity`: %+v", err)
			}

			return nil
		},
	}
}

func (m *LinuxFunctionAppDataSourceModel) unpackLinuxFunctionAppSettings(input web.StringDictionary, metadata sdk.ResourceMetaData) {
	if input.Properties == nil {
		return
	}

	appSettings := make(map[string]string)
	var dockerSettings helpers.ApplicationStackDocker
	m.BuiltinLogging = false

	for k, v := range input.Properties {
		switch k {
		case "FUNCTIONS_EXTENSION_VERSION":
			m.FunctionExtensionsVersion = utils.NormalizeNilableString(v)

		case "WEBSITE_NODE_DEFAULT_VERSION": // Note - This is only set if it's not the default of 12, but we collect it from LinuxFxVersion so can discard it here
		case "WEBSITE_CONTENTAZUREFILECONNECTIONSTRING":
			if _, ok := metadata.ResourceData.GetOk("app_settings.WEBSITE_CONTENTAZUREFILECONNECTIONSTRING"); ok {
				appSettings[k] = utils.NormalizeNilableString(v)
			}
		case "WEBSITE_CONTENTSHARE":
			if _, ok := metadata.ResourceData.GetOk("app_settings.WEBSITE_CONTENTSHARE"); ok {
				appSettings[k] = utils.NormalizeNilableString(v)
			}
		case "WEBSITE_HTTPLOGGING_RETENTION_DAYS":
		case "FUNCTIONS_WORKER_RUNTIME":
			if m.SiteConfig[0].ApplicationStack != nil {
				m.SiteConfig[0].ApplicationStack[0].CustomHandler = strings.EqualFold(*v, "custom")
			}

		case "DOCKER_REGISTRY_SERVER_URL":
			dockerSettings.RegistryURL = utils.NormalizeNilableString(v)

		case "DOCKER_REGISTRY_SERVER_USERNAME":
			dockerSettings.RegistryUsername = utils.NormalizeNilableString(v)

		case "DOCKER_REGISTRY_SERVER_PASSWORD":
			dockerSettings.RegistryPassword = utils.NormalizeNilableString(v)

		// case "WEBSITES_ENABLE_APP_SERVICE_STORAGE": // TODO - Support this as a configurable bool, default `false` - Ref: https://docs.microsoft.com/en-us/azure/app-service/faq-app-service-linux#i-m-using-my-own-custom-container--i-want-the-platform-to-mount-an-smb-share-to-the---home---directory-

		case "APPINSIGHTS_INSTRUMENTATIONKEY":
			m.SiteConfig[0].AppInsightsInstrumentationKey = utils.NormalizeNilableString(v)

		case "APPLICATIONINSIGHTS_CONNECTION_STRING":
			m.SiteConfig[0].AppInsightsConnectionString = utils.NormalizeNilableString(v)

		case "AzureWebJobsStorage":
			if v != nil && strings.HasPrefix(*v, "@Microsoft.KeyVault") {
				trimmed := strings.TrimPrefix(strings.TrimSuffix(*v, ")"), "@Microsoft.KeyVault(")
				m.StorageKeyVaultSecretID = trimmed
			} else {
				m.StorageAccountName, m.StorageAccountKey = helpers.ParseWebJobsStorageString(v)
			}

		case "AzureWebJobsDashboard":
			m.BuiltinLogging = true

		case "WEBSITE_HEALTHCHECK_MAXPINGFAILURES":
			i, _ := strconv.Atoi(utils.NormalizeNilableString(v))
			m.SiteConfig[0].HealthCheckEvictionTime = utils.NormaliseNilableInt(&i)

		default:
			appSettings[k] = utils.NormalizeNilableString(v)
		}
	}

	if dockerSettings.RegistryURL != "" {
		appStack := make([]helpers.ApplicationStackLinuxFunctionApp, 0)
		docker, _ := helpers.DecodeFunctionAppDockerFxString(m.SiteConfig[0].LinuxFxVersion, dockerSettings)
		appStack = append(appStack, helpers.ApplicationStackLinuxFunctionApp{Docker: docker})
		m.SiteConfig[0].ApplicationStack = appStack
	}

	m.AppSettings = appSettings
}
