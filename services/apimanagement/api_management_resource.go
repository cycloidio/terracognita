package apimanagement

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2021-08-01/apimanagement"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/identity"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/zones"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/schemaz"
	apimValidate "github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

var (
	apimBackendProtocolSsl3                  = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Backend.Protocols.Ssl30"
	apimBackendProtocolTls10                 = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Backend.Protocols.Tls10"
	apimBackendProtocolTls11                 = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Backend.Protocols.Tls11"
	apimFrontendProtocolSsl3                 = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Protocols.Ssl30"
	apimFrontendProtocolTls10                = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Protocols.Tls10"
	apimFrontendProtocolTls11                = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Protocols.Tls11"
	apimTripleDesCiphers                     = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TripleDes168"
	apimHttp2Protocol                        = "Microsoft.WindowsAzure.ApiManagement.Gateway.Protocols.Server.Http2"
	apimTlsEcdheEcdsaWithAes256CbcShaCiphers = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
	apimTlsEcdheEcdsaWithAes128CbcShaCiphers = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
	apimTlsEcdheRsaWithAes256CbcShaCiphers   = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
	apimTlsEcdheRsaWithAes128CbcShaCiphers   = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
	apimTlsRsaWithAes128GcmSha256Ciphers     = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_128_GCM_SHA256"
	apimTlsRsaWithAes256CbcSha256Ciphers     = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_256_CBC_SHA256"
	apimTlsRsaWithAes128CbcSha256Ciphers     = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_128_CBC_SHA256"
	apimTlsRsaWithAes256CbcShaCiphers        = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_256_CBC_SHA"
	apimTlsRsaWithAes128CbcShaCiphers        = "Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_128_CBC_SHA"
)

func resourceApiManagementService() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceApiManagementServiceCreateUpdate,
		Read:   resourceApiManagementServiceRead,
		Update: resourceApiManagementServiceCreateUpdate,
		Delete: resourceApiManagementServiceDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ApiManagementID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(3 * time.Hour),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(3 * time.Hour),
			Delete: pluginsdk.DefaultTimeout(3 * time.Hour),
		},

		Schema: resourceApiManagementSchema(),

		// we can only change `virtual_network_type` from None to Internal Or External, Else the subnet can not be destroyed cause “InUseSubnetCannotBeDeleted” for 3 hours
		// we can not change the subnet from subnet1 to subnet2 either, Else the subnet1 can not be destroyed cause “InUseSubnetCannotBeDeleted” for 3 hours
		// Issue: https://github.com/Azure/azure-rest-api-specs/issues/10395
		CustomizeDiff: pluginsdk.CustomDiffWithAll(
			pluginsdk.ForceNewIfChange("virtual_network_type", func(ctx context.Context, old, new, meta interface{}) bool {
				return !(old.(string) == string(apimanagement.VirtualNetworkTypeNone) &&
					(new.(string) == string(apimanagement.VirtualNetworkTypeInternal) ||
						new.(string) == string(apimanagement.VirtualNetworkTypeExternal)))
			}),

			pluginsdk.ForceNewIfChange("virtual_network_configuration", func(ctx context.Context, old, new, meta interface{}) bool {
				return !(len(old.([]interface{})) == 0 && len(new.([]interface{})) > 0)
			}),
		),
	}
}

func resourceApiManagementSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": schemaz.SchemaApiManagementName(),

		"resource_group_name": azure.SchemaResourceGroupName(),

		"location": azure.SchemaLocation(),

		"publisher_name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: apimValidate.ApiManagementServicePublisherName,
		},

		"publisher_email": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: apimValidate.ApiManagementServicePublisherEmail,
		},

		"sku_name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: apimValidate.ApimSkuName(),
		},

		"identity": commonschema.SystemAssignedUserAssignedIdentityOptional(),

		"virtual_network_type": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Default:  string(apimanagement.VirtualNetworkTypeNone),
			ValidateFunc: validation.StringInSlice([]string{
				string(apimanagement.VirtualNetworkTypeNone),
				string(apimanagement.VirtualNetworkTypeExternal),
				string(apimanagement.VirtualNetworkTypeInternal),
			}, false),
		},

		"virtual_network_configuration": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"subnet_id": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: azure.ValidateResourceID,
					},
				},
			},
		},

		"client_certificate_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  false,
		},

		"gateway_disabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  false,
		},

		"min_api_version": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"notification_sender_email": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Computed: true,
		},

		"additional_location": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"location": commonschema.LocationWithoutForceNew(),

					"virtual_network_configuration": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"subnet_id": {
									Type:         pluginsdk.TypeString,
									Required:     true,
									ValidateFunc: azure.ValidateResourceID,
								},
							},
						},
					},

					"capacity": {
						Type:         pluginsdk.TypeInt,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.IntBetween(0, 12),
					},

					"zones": commonschema.ZonesMultipleOptionalForceNew(),

					"gateway_regional_url": {
						Type:     pluginsdk.TypeString,
						Computed: true,
					},

					"public_ip_addresses": {
						Type: pluginsdk.TypeList,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
						Computed: true,
					},

					"public_ip_address_id": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: azure.ValidateResourceID,
					},

					"private_ip_addresses": {
						Type: pluginsdk.TypeList,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
						Computed: true,
					},
				},
			},
		},

		"certificate": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 10,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"encoded_certificate": {
						Type:      pluginsdk.TypeString,
						Required:  true,
						Sensitive: true,
					},

					"certificate_password": {
						Type:      pluginsdk.TypeString,
						Optional:  true,
						Sensitive: true,
					},

					"store_name": {
						Type:     pluginsdk.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							string(apimanagement.StoreNameCertificateAuthority),
							string(apimanagement.StoreNameRoot),
						}, false),
					},

					"expiry": {
						Type:     pluginsdk.TypeString,
						Computed: true,
					},

					"subject": {
						Type:     pluginsdk.TypeString,
						Computed: true,
					},

					"thumbprint": {
						Type:     pluginsdk.TypeString,
						Computed: true,
					},
				},
			},
		},

		"protocols": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"enable_http2": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
						// TODO 4.0: change this from enable_* to *_enabled
					},
				},
			},
		},

		"security": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					// TODO 4.0: change this from enable_* to *_enabled
					"enable_backend_ssl30": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					// TODO 4.0: change this from enable_* to *_enabled
					"enable_backend_tls10": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					// TODO 4.0: change this from enable_* to *_enabled
					"enable_backend_tls11": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},

					// TODO 4.0: change this from enable_* to *_enabled
					"enable_frontend_ssl30": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},

					// TODO 4.0: change this from enable_* to *_enabled
					"enable_frontend_tls10": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},

					// TODO 4.0: change this from enable_* to *_enabled
					"enable_frontend_tls11": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},

					"triple_des_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
					},

					"tls_ecdhe_ecdsa_with_aes256_cbc_sha_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					"tls_ecdhe_ecdsa_with_aes128_cbc_sha_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					"tls_ecdhe_rsa_with_aes256_cbc_sha_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					"tls_ecdhe_rsa_with_aes128_cbc_sha_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					"tls_rsa_with_aes128_gcm_sha256_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					"tls_rsa_with_aes256_cbc_sha256_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					"tls_rsa_with_aes128_cbc_sha256_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					"tls_rsa_with_aes256_cbc_sha_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
					"tls_rsa_with_aes128_cbc_sha_ciphers_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  false,
					},
				},
			},
		},

		"hostname_configuration": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"management": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Resource{
							Schema: apiManagementResourceHostnameSchema(),
						},
						AtLeastOneOf: []string{"hostname_configuration.0.management", "hostname_configuration.0.portal", "hostname_configuration.0.developer_portal", "hostname_configuration.0.proxy", "hostname_configuration.0.scm"},
					},
					"portal": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Resource{
							Schema: apiManagementResourceHostnameSchema(),
						},
						AtLeastOneOf: []string{"hostname_configuration.0.management", "hostname_configuration.0.portal", "hostname_configuration.0.developer_portal", "hostname_configuration.0.proxy", "hostname_configuration.0.scm"},
					},
					"developer_portal": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Resource{
							Schema: apiManagementResourceHostnameSchema(),
						},
						AtLeastOneOf: []string{"hostname_configuration.0.management", "hostname_configuration.0.portal", "hostname_configuration.0.developer_portal", "hostname_configuration.0.proxy", "hostname_configuration.0.scm"},
					},
					"proxy": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Resource{
							Schema: apiManagementResourceHostnameProxySchema(),
						},
						AtLeastOneOf: []string{"hostname_configuration.0.management", "hostname_configuration.0.portal", "hostname_configuration.0.developer_portal", "hostname_configuration.0.proxy", "hostname_configuration.0.scm"},
					},
					"scm": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Resource{
							Schema: apiManagementResourceHostnameSchema(),
						},
						AtLeastOneOf: []string{"hostname_configuration.0.management", "hostname_configuration.0.portal", "hostname_configuration.0.developer_portal", "hostname_configuration.0.proxy", "hostname_configuration.0.scm"},
					},
				},
			},
		},

		//lintignore:XS003
		"policy": {
			Type:       pluginsdk.TypeList,
			Optional:   true,
			Computed:   true,
			MaxItems:   1,
			ConfigMode: pluginsdk.SchemaConfigModeAttr,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"xml_content": {
						Type:             pluginsdk.TypeString,
						Optional:         true,
						Computed:         true,
						ConflictsWith:    []string{"policy.0.xml_link"},
						DiffSuppressFunc: XmlWithDotNetInterpolationsDiffSuppress,
					},

					"xml_link": {
						Type:          pluginsdk.TypeString,
						Optional:      true,
						ConflictsWith: []string{"policy.0.xml_content"},
					},
				},
			},
		},

		"sign_in": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"enabled": {
						Type:     pluginsdk.TypeBool,
						Required: true,
					},
				},
			},
		},

		"sign_up": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"enabled": {
						Type:     pluginsdk.TypeBool,
						Required: true,
					},

					"terms_of_service": {
						Type:     pluginsdk.TypeList,
						Required: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"enabled": {
									Type:     pluginsdk.TypeBool,
									Required: true,
								},
								"consent_required": {
									Type:     pluginsdk.TypeBool,
									Required: true,
								},
								"text": {
									Type:     pluginsdk.TypeString,
									Optional: true,
								},
							},
						},
					},
				},
			},
		},

		"zones": commonschema.ZonesMultipleOptionalForceNew(),

		"gateway_url": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"management_api_url": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"gateway_regional_url": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"public_ip_addresses": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"public_ip_address_id": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: azure.ValidateResourceID,
		},

		"public_network_access_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  true,
		},

		"private_ip_addresses": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"portal_url": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"developer_portal_url": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"scm_url": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"tenant_access": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"enabled": {
						Type:     pluginsdk.TypeBool,
						Required: true,
					},
					"tenant_id": {
						Type:     pluginsdk.TypeString,
						Computed: true,
					},
					"primary_key": {
						Type:      pluginsdk.TypeString,
						Computed:  true,
						Sensitive: true,
					},
					"secondary_key": {
						Type:      pluginsdk.TypeString,
						Computed:  true,
						Sensitive: true,
					},
				},
			},
		},

		"tags": tags.Schema(),
	}
}

func resourceApiManagementServiceCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ServiceClient
	apiClient := meta.(*clients.Client).ApiManagement.ApiClient
	deletedServicesClient := meta.(*clients.Client).ApiManagement.DeletedServicesClient
	productsClient := meta.(*clients.Client).ApiManagement.ProductsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	sku := expandAzureRmApiManagementSkuName(d)

	log.Printf("[INFO] preparing arguments for API Management Service creation.")

	id := parse.NewApiManagementID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.ServiceName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_api_management", id.ID())
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	publicIpAddressId := d.Get("public_ip_address_id").(string)
	notificationSenderEmail := d.Get("notification_sender_email").(string)
	virtualNetworkType := d.Get("virtual_network_type").(string)

	customProperties, err := expandApiManagementCustomProperties(d, sku.Name == apimanagement.SkuTypeConsumption)
	if err != nil {
		return err
	}
	certificates := expandAzureRmApiManagementCertificates(d)

	publicNetworkAccess := apimanagement.PublicNetworkAccessEnabled
	if !d.Get("public_network_access_enabled").(bool) {
		publicNetworkAccess = apimanagement.PublicNetworkAccessDisabled
	}

	if d.IsNewResource() {
		// before creating check to see if the resource exists in the soft delete state
		softDeleted, err := deletedServicesClient.GetByName(ctx, id.ServiceName, location)
		if err != nil {
			// If Terraform lacks permission to read at the Subscription we'll get 403, not 404
			if !utils.ResponseWasNotFound(softDeleted.Response) && !utils.ResponseWasForbidden(softDeleted.Response) {
				return fmt.Errorf("checking for the presence of an existing Soft-Deleted API Management %q (Location %q): %+v", id.ServiceName, location, err)
			}
		}

		// if so, does the user want us to recover it?
		if !utils.ResponseWasNotFound(softDeleted.Response) && !utils.ResponseWasForbidden(softDeleted.Response) {
			if !meta.(*clients.Client).Features.ApiManagement.RecoverSoftDeleted {
				// this exists but the users opted out, so they must import this it out-of-band
				return fmt.Errorf(optedOutOfRecoveringSoftDeletedApiManagementErrorFmt(id.ServiceName, location))
			}

			// First recover the deleted API Management, since all other properties are ignored during a restore operation
			// (don't set the ID just yet to avoid tainting on failure)
			params := apimanagement.ServiceResource{
				Location: utils.String(location),
				ServiceProperties: &apimanagement.ServiceProperties{
					Restore: utils.Bool(true),
				},
			}

			if _, err = client.CreateOrUpdate(ctx, id.ResourceGroup, id.ServiceName, params); err != nil {
				return fmt.Errorf("recovering %s: %+v", id, err)
			}

			// Wait for the ProvisioningState to become "Succeeded" before attempting to update
			log.Printf("[DEBUG] Waiting for %s to become ready", id)
			deadline, ok := ctx.Deadline()
			if !ok {
				return fmt.Errorf("context had no deadline")
			}
			stateConf := &pluginsdk.StateChangeConf{
				Pending:                   []string{"Deleted", "Activating", "Updating", "Unknown"},
				Target:                    []string{"Succeeded", "Ready"},
				Refresh:                   apiManagementRefreshFunc(ctx, client, id.ServiceName, id.ResourceGroup),
				MinTimeout:                1 * time.Minute,
				ContinuousTargetOccurence: 2,
				Timeout:                   time.Until(deadline),
			}

			if _, err = stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf("waiting for %s to become ready: %+v", id, err)
			}
		}
	}

	properties := apimanagement.ServiceResource{
		Location: utils.String(location),
		ServiceProperties: &apimanagement.ServiceProperties{
			PublisherName:       pointer.FromString(d.Get("publisher_name").(string)),
			PublisherEmail:      pointer.FromString(d.Get("publisher_email").(string)),
			PublicNetworkAccess: publicNetworkAccess,
			CustomProperties:    customProperties,
			Certificates:        certificates,
		},
		Tags: tags.Expand(t),
		Sku:  sku,
	}

	if _, ok := d.GetOk("hostname_configuration"); ok {
		properties.ServiceProperties.HostnameConfigurations = expandAzureRmApiManagementHostnameConfigurations(d)
	}

	// intentionally not gated since we specify a default value (of None) in the expand, which we need on updates
	identityRaw := d.Get("identity").([]interface{})
	identity, err := expandIdentity(identityRaw)
	if err != nil {
		return fmt.Errorf("expanding `identity`: %+v", err)
	}
	properties.Identity = identity

	if _, ok := d.GetOk("additional_location"); ok {
		var err error
		properties.ServiceProperties.AdditionalLocations, err = expandAzureRmApiManagementAdditionalLocations(d, sku)
		if err != nil {
			return err
		}
	}

	if notificationSenderEmail != "" {
		properties.ServiceProperties.NotificationSenderEmail = &notificationSenderEmail
	}

	if virtualNetworkType != "" {
		properties.ServiceProperties.VirtualNetworkType = apimanagement.VirtualNetworkType(virtualNetworkType)

		if virtualNetworkType != string(apimanagement.VirtualNetworkTypeNone) {
			virtualNetworkConfiguration := expandAzureRmApiManagementVirtualNetworkConfigurations(d)
			if virtualNetworkConfiguration == nil {
				return fmt.Errorf("You must specify 'virtual_network_configuration' when 'virtual_network_type' is %q", virtualNetworkType)
			}
			properties.ServiceProperties.VirtualNetworkConfiguration = virtualNetworkConfiguration
		}
	}

	if publicIpAddressId != "" {
		if sku.Name != apimanagement.SkuTypePremium && sku.Name != apimanagement.SkuTypeDeveloper {
			if virtualNetworkType == string(apimanagement.VirtualNetworkTypeNone) {
				return fmt.Errorf("`public_ip_address_id` is only supported when sku type is `Developer` or `Premium`, and the APIM instance is deployed in a virtual network.")
			}
		}
		properties.ServiceProperties.PublicIPAddressID = utils.String(publicIpAddressId)
	}

	if d.HasChange("client_certificate_enabled") {
		enableClientCertificate := d.Get("client_certificate_enabled").(bool)
		if enableClientCertificate && sku.Name != apimanagement.SkuTypeConsumption {
			return fmt.Errorf("`client_certificate_enabled` is only supported when sku type is `Consumption`")
		}
		properties.ServiceProperties.EnableClientCertificate = utils.Bool(enableClientCertificate)
	}

	gateWayDisabled := d.Get("gateway_disabled").(bool)
	if gateWayDisabled && len(*properties.AdditionalLocations) == 0 {
		return fmt.Errorf("`gateway_disabled` is only supported when `additional_location` is set")
	}
	properties.ServiceProperties.DisableGateway = utils.Bool(gateWayDisabled)

	if v, ok := d.GetOk("min_api_version"); ok {
		properties.ServiceProperties.APIVersionConstraint = &apimanagement.APIVersionConstraint{
			MinAPIVersion: utils.String(v.(string)),
		}
	}

	if v := d.Get("zones").(*schema.Set).List(); len(v) > 0 {
		if sku.Name != apimanagement.SkuTypePremium {
			return fmt.Errorf("`zones` is only supported when sku type is `Premium`")
		}

		if publicIpAddressId == "" {
			return fmt.Errorf("`public_ip_address` must be specified when `zones` are provided")
		}
		zones := zones.Expand(v)
		properties.Zones = &zones
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.ServiceName, properties)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation/update of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	// Remove sample products and APIs after creating (v3.0 behaviour)
	if d.IsNewResource() {
		apis := make([]apimanagement.APIContract, 0)

		for apisIter, err := apiClient.ListByService(ctx, id.ResourceGroup, id.ServiceName, "", nil, nil, "", nil); apisIter.NotDone(); err = apisIter.NextWithContext(ctx) {
			if err != nil {
				return fmt.Errorf("listing APIs after creation of %s: %+v", id, err)
			}
			if apisIter.Response().IsEmpty() {
				break
			}
			if apisList := apisIter.Values(); apisList != nil {
				apis = append(apis, apisList...)
			}
		}

		for _, api := range apis {
			if api.ID == nil {
				continue
			}
			apiId, err := parse.ApiID(*api.ID)
			if err != nil {
				return fmt.Errorf("parsing API ID: %+v", err)
			}
			log.Printf("[DEBUG] Deleting %s", apiId)
			if resp, err := apiClient.Delete(ctx, apiId.ResourceGroup, apiId.ServiceName, apiId.Name, "", utils.Bool(true)); err != nil && !utils.ResponseWasNotFound(resp) {
				return fmt.Errorf("deleting %s: %+v", apiId, err)
			}
		}

		products := make([]apimanagement.ProductContract, 0)

		for productsIter, err := productsClient.ListByService(ctx, id.ResourceGroup, id.ServiceName, "", nil, nil, nil, ""); productsIter.NotDone(); err = productsIter.NextWithContext(ctx) {
			if err != nil {
				return fmt.Errorf("listing products after creation of %s: %+v", id, err)
			}
			if productsIter.Response().IsEmpty() {
				break
			}
			if productList := productsIter.Values(); products != nil {
				products = append(products, productList...)
			}
		}

		for _, product := range products {
			if product.ID == nil {
				continue
			}
			productId, err := parse.ProductID(*product.ID)
			if err != nil {
				return fmt.Errorf("parsing product ID: %+v", err)
			}
			log.Printf("[DEBUG] Deleting %s", productId)
			if resp, err := productsClient.Delete(ctx, productId.ResourceGroup, productId.ServiceName, productId.Name, "", utils.Bool(true)); err != nil && !utils.ResponseWasNotFound(resp) {
				return fmt.Errorf("deleting %s: %+v", productId, err)
			}
		}

	}

	signInSettingsRaw := d.Get("sign_in").([]interface{})
	if sku.Name == apimanagement.SkuTypeConsumption && len(signInSettingsRaw) > 0 {
		return fmt.Errorf("`sign_in` is not support for sku tier `Consumption`")
	}
	if sku.Name != apimanagement.SkuTypeConsumption {
		signInSettings := expandApiManagementSignInSettings(signInSettingsRaw)
		signInClient := meta.(*clients.Client).ApiManagement.SignInClient
		if _, err := signInClient.CreateOrUpdate(ctx, id.ResourceGroup, id.ServiceName, signInSettings, ""); err != nil {
			return fmt.Errorf(" setting Sign In settings for %s: %+v", id, err)
		}
	}

	signUpSettingsRaw := d.Get("sign_up").([]interface{})
	if sku.Name == apimanagement.SkuTypeConsumption && len(signInSettingsRaw) > 0 {
		return fmt.Errorf("`sign_up` is not support for sku tier `Consumption`")
	}
	if sku.Name != apimanagement.SkuTypeConsumption {
		signUpSettings := expandApiManagementSignUpSettings(signUpSettingsRaw)
		signUpClient := meta.(*clients.Client).ApiManagement.SignUpClient
		if _, err := signUpClient.CreateOrUpdate(ctx, id.ResourceGroup, id.ServiceName, signUpSettings, ""); err != nil {
			return fmt.Errorf(" setting Sign Up settings for %s: %+v", id, err)
		}
	}

	policyClient := meta.(*clients.Client).ApiManagement.PolicyClient
	policiesRaw := d.Get("policy").([]interface{})
	policy, err := expandApiManagementPolicies(policiesRaw)
	if err != nil {
		return err
	}

	if d.HasChange("policy") {
		// remove the existing policy
		if resp, err := policyClient.Delete(ctx, id.ResourceGroup, id.ServiceName, ""); err != nil {
			if !utils.ResponseWasNotFound(resp) {
				return fmt.Errorf("removing Policies from %s: %+v", id, err)
			}
		}

		// then add the new one, if it exists
		if policy != nil {
			if _, err := policyClient.CreateOrUpdate(ctx, id.ResourceGroup, id.ServiceName, *policy, ""); err != nil {
				return fmt.Errorf(" setting Policies for %s: %+v", id, err)
			}
		}
	}

	tenantAccessRaw := d.Get("tenant_access").([]interface{})
	if sku.Name == apimanagement.SkuTypeConsumption && len(tenantAccessRaw) > 0 {
		return fmt.Errorf("`tenant_access` is not supported for sku tier `Consumption`")
	}
	if sku.Name != apimanagement.SkuTypeConsumption && d.HasChange("tenant_access") {
		tenantAccessInformationParametersRaw := d.Get("tenant_access").([]interface{})
		tenantAccessInformationParameters := expandApiManagementTenantAccessSettings(tenantAccessInformationParametersRaw)
		tenantAccessClient := meta.(*clients.Client).ApiManagement.TenantAccessClient
		if _, err := tenantAccessClient.Update(ctx, id.ResourceGroup, id.ServiceName, tenantAccessInformationParameters, "access", ""); err != nil {
			return fmt.Errorf(" updating tenant access settings for %s: %+v", id, err)
		}
	}

	return resourceApiManagementServiceRead(d, meta)
}

func resourceApiManagementServiceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ServiceClient
	signInClient := meta.(*clients.Client).ApiManagement.SignInClient
	signUpClient := meta.(*clients.Client).ApiManagement.SignUpClient
	tenantAccessClient := meta.(*clients.Client).ApiManagement.TenantAccessClient
	environment := meta.(*clients.Client).Account.Environment
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ApiManagementID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ServiceName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("%s was not found - removing from state!", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("making Read request on %s: %+v", *id, err)
	}

	policyClient := meta.(*clients.Client).ApiManagement.PolicyClient
	policy, err := policyClient.Get(ctx, id.ResourceGroup, id.ServiceName, apimanagement.PolicyExportFormatXML)
	if err != nil {
		if !utils.ResponseWasNotFound(policy.Response) {
			return fmt.Errorf("retrieving Policy for %s: %+v", *id, err)
		}
	}

	d.Set("name", id.ServiceName)
	d.Set("resource_group_name", id.ResourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	identity, err := flattenIdentity(resp.Identity)
	if err != nil {
		return fmt.Errorf("flattening `identity`: %+v", err)
	}
	if err := d.Set("identity", identity); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}

	if props := resp.ServiceProperties; props != nil {
		d.Set("publisher_email", props.PublisherEmail)
		d.Set("publisher_name", props.PublisherName)
		d.Set("notification_sender_email", props.NotificationSenderEmail)
		d.Set("gateway_url", props.GatewayURL)
		d.Set("gateway_regional_url", props.GatewayRegionalURL)
		d.Set("portal_url", props.PortalURL)
		d.Set("developer_portal_url", props.DeveloperPortalURL)
		d.Set("management_api_url", props.ManagementAPIURL)
		d.Set("scm_url", props.ScmURL)
		d.Set("public_ip_addresses", props.PublicIPAddresses)
		d.Set("public_ip_address_id", props.PublicIPAddressID)
		d.Set("public_network_access_enabled", props.PublicNetworkAccess == apimanagement.PublicNetworkAccessEnabled)
		d.Set("private_ip_addresses", props.PrivateIPAddresses)
		d.Set("virtual_network_type", props.VirtualNetworkType)
		d.Set("client_certificate_enabled", props.EnableClientCertificate)
		d.Set("gateway_disabled", props.DisableGateway)

		d.Set("certificate", flattenAPIManagementCertificates(d, props.Certificates))

		if resp.Sku != nil && resp.Sku.Name != "" {
			if err := d.Set("security", flattenApiManagementSecurityCustomProperties(props.CustomProperties, resp.Sku.Name == apimanagement.SkuTypeConsumption)); err != nil {
				return fmt.Errorf("setting `security`: %+v", err)
			}
		}

		if err := d.Set("protocols", flattenApiManagementProtocolsCustomProperties(props.CustomProperties)); err != nil {
			return fmt.Errorf("setting `protocols`: %+v", err)
		}

		apimHostNameSuffix := environment.APIManagementHostNameSuffix
		hostnameConfigs := flattenApiManagementHostnameConfigurations(props.HostnameConfigurations, d, id.ServiceName, apimHostNameSuffix)
		if err := d.Set("hostname_configuration", hostnameConfigs); err != nil {
			return fmt.Errorf("setting `hostname_configuration`: %+v", err)
		}

		if err := d.Set("additional_location", flattenApiManagementAdditionalLocations(props.AdditionalLocations)); err != nil {
			return fmt.Errorf("setting `additional_location`: %+v", err)
		}

		if err := d.Set("virtual_network_configuration", flattenApiManagementVirtualNetworkConfiguration(props.VirtualNetworkConfiguration)); err != nil {
			return fmt.Errorf("setting `virtual_network_configuration`: %+v", err)
		}

		var minApiVersion string
		if props.APIVersionConstraint != nil && props.APIVersionConstraint.MinAPIVersion != nil {
			minApiVersion = *props.APIVersionConstraint.MinAPIVersion
		}
		d.Set("min_api_version", minApiVersion)

	}

	if err := d.Set("sku_name", flattenApiManagementServiceSkuName(resp.Sku)); err != nil {
		return fmt.Errorf("setting `sku_name`: %+v", err)
	}

	if err := d.Set("policy", flattenApiManagementPolicies(d, policy)); err != nil {
		return fmt.Errorf("setting `policy`: %+v", err)
	}

	d.Set("zones", zones.Flatten(resp.Zones))

	if resp.Sku.Name != apimanagement.SkuTypeConsumption {
		signInSettings, err := signInClient.Get(ctx, id.ResourceGroup, id.ServiceName)
		if err != nil {
			return fmt.Errorf("retrieving Sign In Settings for %s: %+v", *id, err)
		}
		if err := d.Set("sign_in", flattenApiManagementSignInSettings(signInSettings)); err != nil {
			return fmt.Errorf("setting `sign_in`: %+v", err)
		}

		signUpSettings, err := signUpClient.Get(ctx, id.ResourceGroup, id.ServiceName)
		if err != nil {
			return fmt.Errorf("retrieving Sign Up Settings for %s: %+v", *id, err)
		}

		if err := d.Set("sign_up", flattenApiManagementSignUpSettings(signUpSettings)); err != nil {
			return fmt.Errorf("setting `sign_up`: %+v", err)
		}
	} else {
		d.Set("sign_in", []interface{}{})
		d.Set("sign_up", []interface{}{})
	}

	if resp.Sku.Name != apimanagement.SkuTypeConsumption {
		tenantAccessInformationContract, err := tenantAccessClient.ListSecrets(ctx, id.ResourceGroup, id.ServiceName, "access")
		if err != nil {
			return fmt.Errorf("retrieving tenant access properties for %s: %+v", *id, err)
		}
		if err := d.Set("tenant_access", flattenApiManagementTenantAccessSettings(tenantAccessInformationContract)); err != nil {
			return fmt.Errorf("setting `tenant_access`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceApiManagementServiceDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ServiceClient
	deletedServicesClient := meta.(*clients.Client).ApiManagement.DeletedServicesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ApiManagementID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting %s", *id)
	future, err := client.Delete(ctx, id.ResourceGroup, id.ServiceName)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
		}
	}

	// Purge the soft deleted Api Management permanently if the feature flag is enabled
	if meta.(*clients.Client).Features.ApiManagement.PurgeSoftDeleteOnDestroy {
		log.Printf("[DEBUG] %s marked for purge - executing purge", *id)
		_, err := deletedServicesClient.GetByName(ctx, id.ServiceName, azure.NormalizeLocation(d.Get("location").(string)))
		if err != nil {
			return err
		}
		future, err := deletedServicesClient.Purge(ctx, id.ServiceName, azure.NormalizeLocation(d.Get("location").(string)))
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Waiting for purge of %s..", *id)
		err = future.WaitForCompletionRef(ctx, deletedServicesClient.Client)
		if err != nil {
			return fmt.Errorf("purging %s: %+v", *id, err)
		}
		log.Printf("[DEBUG] Purged %s.", *id)
		return nil
	}

	return nil
}

func apiManagementRefreshFunc(ctx context.Context, client *apimanagement.ServiceClient, serviceName, resourceGroup string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Checking to see if API Management Service %q (Resource Group: %q) is available..", serviceName, resourceGroup)

		resp, err := client.Get(ctx, resourceGroup, serviceName)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				log.Printf("[DEBUG] Retrieving API Management %q (Resource Group: %q) returned 404.", serviceName, resourceGroup)
				return nil, "NotFound", nil
			}

			return nil, "", fmt.Errorf("polling for the state of the API Management Service %q (Resource Group: %q): %+v", serviceName, resourceGroup, err)
		}

		state := ""
		if props := resp.ServiceProperties; props != nil {
			if props.ProvisioningState != nil {
				state = *props.ProvisioningState
			}
		}

		return resp, state, nil
	}
}

func expandAzureRmApiManagementHostnameConfigurations(d *pluginsdk.ResourceData) *[]apimanagement.HostnameConfiguration {
	results := make([]apimanagement.HostnameConfiguration, 0)
	vs := d.Get("hostname_configuration")
	if vs == nil {
		return &results
	}
	hostnameVs := vs.([]interface{})

	for _, hostnameRawVal := range hostnameVs {
		// hostnameRawVal is guaranteed to be non-nil as there is AtLeastOneOf constraint on its containing properties.
		hostnameV := hostnameRawVal.(map[string]interface{})

		managementVs := hostnameV["management"].([]interface{})
		for _, managementV := range managementVs {
			v := managementV.(map[string]interface{})
			output := expandApiManagementCommonHostnameConfiguration(v, apimanagement.HostnameTypeManagement)
			results = append(results, output)
		}

		portalVs := hostnameV["portal"].([]interface{})
		for _, portalV := range portalVs {
			v := portalV.(map[string]interface{})
			output := expandApiManagementCommonHostnameConfiguration(v, apimanagement.HostnameTypePortal)
			results = append(results, output)
		}

		developerPortalVs := hostnameV["developer_portal"].([]interface{})
		for _, developerPortalV := range developerPortalVs {
			v := developerPortalV.(map[string]interface{})
			output := expandApiManagementCommonHostnameConfiguration(v, apimanagement.HostnameTypeDeveloperPortal)
			results = append(results, output)
		}

		proxyVs := hostnameV["proxy"].([]interface{})
		for _, proxyV := range proxyVs {
			v := proxyV.(map[string]interface{})
			output := expandApiManagementCommonHostnameConfiguration(v, apimanagement.HostnameTypeProxy)
			if value, ok := v["default_ssl_binding"]; ok {
				output.DefaultSslBinding = utils.Bool(value.(bool))
			}
			results = append(results, output)
		}

		scmVs := hostnameV["scm"].([]interface{})
		for _, scmV := range scmVs {
			v := scmV.(map[string]interface{})
			output := expandApiManagementCommonHostnameConfiguration(v, apimanagement.HostnameTypeScm)
			results = append(results, output)
		}
	}

	return &results
}

func expandApiManagementCommonHostnameConfiguration(input map[string]interface{}, hostnameType apimanagement.HostnameType) apimanagement.HostnameConfiguration {
	output := apimanagement.HostnameConfiguration{
		Type: hostnameType,
	}
	if v, ok := input["certificate"]; ok {
		if v.(string) != "" {
			output.EncodedCertificate = utils.String(v.(string))
		}
	}
	if v, ok := input["certificate_password"]; ok {
		if v.(string) != "" {
			output.CertificatePassword = utils.String(v.(string))
		}
	}
	if v, ok := input["host_name"]; ok {
		if v.(string) != "" {
			output.HostName = utils.String(v.(string))
		}
	}
	if v, ok := input["key_vault_id"]; ok {
		if v.(string) != "" {
			output.KeyVaultID = utils.String(v.(string))
		}
	}

	if v, ok := input["negotiate_client_certificate"]; ok {
		output.NegotiateClientCertificate = utils.Bool(v.(bool))
	}

	if v, ok := input["ssl_keyvault_identity_client_id"].(string); ok && v != "" {
		output.IdentityClientID = utils.String(v)
	}

	return output
}

func flattenApiManagementHostnameConfigurations(input *[]apimanagement.HostnameConfiguration, d *pluginsdk.ResourceData, name, apimHostNameSuffix string) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	managementResults := make([]interface{}, 0)
	portalResults := make([]interface{}, 0)
	developerPortalResults := make([]interface{}, 0)
	proxyResults := make([]interface{}, 0)
	scmResults := make([]interface{}, 0)

	for _, config := range *input {
		output := make(map[string]interface{})

		if config.HostName != nil {
			output["host_name"] = *config.HostName
		}

		// There'll always be a default custom domain with hostName "apim_name.azure-api.net" and Type "Proxy", which should be ignored
		if *config.HostName == strings.ToLower(name)+"."+apimHostNameSuffix && config.Type == apimanagement.HostnameTypeProxy {
			continue
		}

		if config.NegotiateClientCertificate != nil {
			output["negotiate_client_certificate"] = *config.NegotiateClientCertificate
		}

		if config.KeyVaultID != nil {
			output["key_vault_id"] = *config.KeyVaultID
		}

		if config.IdentityClientID != nil {
			output["ssl_keyvault_identity_client_id"] = *config.IdentityClientID
		}

		if config.Certificate != nil {
			if config.Certificate.Expiry != nil && !config.Certificate.Expiry.IsZero() {
				output["expiry"] = config.Certificate.Expiry.Format(time.RFC3339)
			}

			if config.Certificate.Thumbprint != nil {
				output["thumbprint"] = *config.Certificate.Thumbprint
			}

			if config.Certificate.Subject != nil {
				output["subject"] = *config.Certificate.Subject
			}
		}

		var configType string
		switch strings.ToLower(string(config.Type)) {
		case strings.ToLower(string(apimanagement.HostnameTypeProxy)):
			// only set SSL binding for proxy types
			if config.DefaultSslBinding != nil {
				output["default_ssl_binding"] = *config.DefaultSslBinding
			}
			proxyResults = append(proxyResults, output)
			configType = "proxy"

		case strings.ToLower(string(apimanagement.HostnameTypeManagement)):
			managementResults = append(managementResults, output)
			configType = "management"

		case strings.ToLower(string(apimanagement.HostnameTypePortal)):
			portalResults = append(portalResults, output)
			configType = "portal"

		case strings.ToLower(string(apimanagement.HostnameTypeDeveloperPortal)):
			developerPortalResults = append(developerPortalResults, output)
			configType = "developer_portal"

		case strings.ToLower(string(apimanagement.HostnameTypeScm)):
			scmResults = append(scmResults, output)
			configType = "scm"
		}

		existingHostnames := d.Get("hostname_configuration").([]interface{})
		if len(existingHostnames) > 0 && configType != "" {
			v := existingHostnames[0].(map[string]interface{})

			if valsRaw, ok := v[configType]; ok {
				vals := valsRaw.([]interface{})
				schemaz.CopyCertificateAndPassword(vals, *config.HostName, output)
			}
		}
	}

	if len(managementResults) == 0 && len(portalResults) == 0 && len(developerPortalResults) == 0 && len(proxyResults) == 0 && len(scmResults) == 0 {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"management":       managementResults,
			"portal":           portalResults,
			"developer_portal": developerPortalResults,
			"proxy":            proxyResults,
			"scm":              scmResults,
		},
	}
}

func expandAzureRmApiManagementCertificates(d *pluginsdk.ResourceData) *[]apimanagement.CertificateConfiguration {
	vs := d.Get("certificate").([]interface{})

	results := make([]apimanagement.CertificateConfiguration, 0)

	for _, v := range vs {
		config := v.(map[string]interface{})

		certBase64 := config["encoded_certificate"].(string)
		storeName := apimanagement.StoreName(config["store_name"].(string))

		cert := apimanagement.CertificateConfiguration{
			EncodedCertificate: utils.String(certBase64),
			StoreName:          storeName,
		}

		if certPassword := config["certificate_password"]; certPassword != nil {
			cert.CertificatePassword = utils.String(certPassword.(string))
		}

		results = append(results, cert)
	}

	return &results
}

func expandAzureRmApiManagementAdditionalLocations(d *pluginsdk.ResourceData, sku *apimanagement.ServiceSkuProperties) (*[]apimanagement.AdditionalLocation, error) {
	inputLocations := d.Get("additional_location").([]interface{})
	parentVnetConfig := d.Get("virtual_network_configuration").([]interface{})

	additionalLocations := make([]apimanagement.AdditionalLocation, 0)

	for _, v := range inputLocations {
		config := v.(map[string]interface{})
		location := azure.NormalizeLocation(config["location"].(string))

		if config["capacity"].(int) > 0 {
			sku.Capacity = utils.Int32(int32(config["capacity"].(int)))
		}

		additionalLocation := apimanagement.AdditionalLocation{
			Location: utils.String(location),
			Sku:      sku,
		}

		childVnetConfig := config["virtual_network_configuration"].([]interface{})
		switch {
		case len(childVnetConfig) == 0 && len(parentVnetConfig) > 0:
			return nil, fmt.Errorf("`virtual_network_configuration` must be specified in any `additional_location` block when top-level `virtual_network_configuration` is supplied")
		case len(childVnetConfig) > 0 && len(parentVnetConfig) == 0:
			return nil, fmt.Errorf("`virtual_network_configuration` must be empty in all `additional_location` blocks when top-level `virtual_network_configuration` is not supplied")
		case len(childVnetConfig) > 0 && len(parentVnetConfig) > 0:
			v := childVnetConfig[0].(map[string]interface{})
			subnetResourceId := v["subnet_id"].(string)
			additionalLocation.VirtualNetworkConfiguration = &apimanagement.VirtualNetworkConfiguration{
				SubnetResourceID: &subnetResourceId,
			}
		}

		publicIPAddressID := config["public_ip_address_id"].(string)
		if publicIPAddressID != "" {
			if sku.Name != apimanagement.SkuTypePremium {
				if len(childVnetConfig) == 0 {
					return nil, fmt.Errorf("`public_ip_address_id` for an additional location is only supported when sku type is `Premium`, and the APIM instance is deployed in a virtual network.")
				}
			}
			additionalLocation.PublicIPAddressID = &publicIPAddressID
		}

		zones := zones.Expand(d.Get("zones").(*schema.Set).List())
		if len(zones) > 0 {
			additionalLocation.Zones = &zones
		}

		additionalLocations = append(additionalLocations, additionalLocation)
	}

	return &additionalLocations, nil
}

func flattenApiManagementAdditionalLocations(input *[]apimanagement.AdditionalLocation) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, prop := range *input {
		var publicIPAddresses []string
		if prop.PublicIPAddresses != nil {
			publicIPAddresses = *prop.PublicIPAddresses
		}

		publicIpAddressId := ""
		if prop.PublicIPAddressID != nil {
			publicIpAddressId = *prop.PublicIPAddressID
		}

		var privateIPAddresses []string
		if prop.PrivateIPAddresses != nil {
			privateIPAddresses = *prop.PrivateIPAddresses
		}

		var capacity *int32
		if prop.Sku.Capacity != nil {
			capacity = prop.Sku.Capacity
		}

		gatewayRegionalUrl := ""
		if prop.GatewayRegionalURL != nil {
			gatewayRegionalUrl = *prop.GatewayRegionalURL
		}

		results = append(results, map[string]interface{}{
			"capacity":                      capacity,
			"gateway_regional_url":          gatewayRegionalUrl,
			"location":                      location.NormalizeNilable(prop.Location),
			"private_ip_addresses":          privateIPAddresses,
			"public_ip_address_id":          publicIpAddressId,
			"public_ip_addresses":           publicIPAddresses,
			"virtual_network_configuration": flattenApiManagementVirtualNetworkConfiguration(prop.VirtualNetworkConfiguration),
			"zones":                         zones.Flatten(prop.Zones),
		})
	}

	return results
}

func expandIdentity(input []interface{}) (*apimanagement.ServiceIdentity, error) {
	expanded, err := identity.ExpandSystemAndUserAssignedMap(input)
	if err != nil {
		return nil, err
	}

	out := apimanagement.ServiceIdentity{
		Type: apimanagement.ApimIdentityType(string(expanded.Type)),
	}
	if expanded.Type == identity.TypeUserAssigned || expanded.Type == identity.TypeSystemAssignedUserAssigned {
		out.UserAssignedIdentities = make(map[string]*apimanagement.UserIdentityProperties)
		for k := range expanded.IdentityIds {
			out.UserAssignedIdentities[k] = &apimanagement.UserIdentityProperties{
				// intentionally empty
			}
		}
	}
	return &out, nil
}

func flattenIdentity(input *apimanagement.ServiceIdentity) (*[]interface{}, error) {
	var transform *identity.SystemAndUserAssignedMap

	if input != nil {
		transform = &identity.SystemAndUserAssignedMap{
			Type:        identity.Type(string(input.Type)),
			IdentityIds: make(map[string]identity.UserAssignedIdentityDetails),
		}
		if input.PrincipalID != nil {
			transform.PrincipalId = input.PrincipalID.String()
		}
		if input.TenantID != nil {
			transform.TenantId = input.TenantID.String()
		}
		for k, v := range input.UserAssignedIdentities {
			transform.IdentityIds[k] = identity.UserAssignedIdentityDetails{
				ClientId:    v.ClientID,
				PrincipalId: v.PrincipalID,
			}
		}
	}

	return identity.FlattenSystemAndUserAssignedMap(transform)
}

func expandAzureRmApiManagementSkuName(d *pluginsdk.ResourceData) *apimanagement.ServiceSkuProperties {
	vs := d.Get("sku_name").(string)

	if len(vs) == 0 {
		return nil
	}

	name, capacity, err := azure.SplitSku(vs)
	if err != nil {
		return nil
	}

	return &apimanagement.ServiceSkuProperties{
		Name:     apimanagement.SkuType(name),
		Capacity: utils.Int32(capacity),
	}
}

func flattenApiManagementServiceSkuName(input *apimanagement.ServiceSkuProperties) string {
	if input == nil {
		return ""
	}

	return fmt.Sprintf("%s_%d", string(input.Name), *input.Capacity)
}

func expandApiManagementCustomProperties(d *pluginsdk.ResourceData, skuIsConsumption bool) (map[string]*string, error) {
	backendProtocolSsl3 := false
	backendProtocolTls10 := false
	backendProtocolTls11 := false
	frontendProtocolSsl3 := false
	frontendProtocolTls10 := false
	frontendProtocolTls11 := false
	tripleDesCiphers := false
	tlsEcdheEcdsaWithAes256CbcShaCiphers := false
	tlsEcdheEcdsaWithAes128CbcShaCiphers := false
	tlsEcdheRsaWithAes256CbcShaCiphers := false
	tlsEcdheRsaWithAes128CbcShaCiphers := false
	tlsRsaWithAes128GcmSha256Ciphers := false
	tlsRsaWithAes256CbcSha256Ciphers := false
	tlsRsaWithAes128CbcSha256Ciphers := false
	tlsRsaWithAes256CbcShaCiphers := false
	tlsRsaWithAes128CbcShaCiphers := false

	if vs := d.Get("security").([]interface{}); len(vs) > 0 {
		v := vs[0].(map[string]interface{})
		backendProtocolSsl3 = v["enable_backend_ssl30"].(bool)
		backendProtocolTls10 = v["enable_backend_tls10"].(bool)
		backendProtocolTls11 = v["enable_backend_tls11"].(bool)
		frontendProtocolSsl3 = v["enable_frontend_ssl30"].(bool)
		frontendProtocolTls10 = v["enable_frontend_tls10"].(bool)
		frontendProtocolTls11 = v["enable_frontend_tls11"].(bool)

		if v, exists := v["triple_des_ciphers_enabled"]; exists {
			tripleDesCiphers = v.(bool)
		}

		tlsEcdheEcdsaWithAes256CbcShaCiphers = v["tls_ecdhe_ecdsa_with_aes256_cbc_sha_ciphers_enabled"].(bool)
		tlsEcdheEcdsaWithAes128CbcShaCiphers = v["tls_ecdhe_ecdsa_with_aes128_cbc_sha_ciphers_enabled"].(bool)
		tlsEcdheRsaWithAes256CbcShaCiphers = v["tls_ecdhe_rsa_with_aes256_cbc_sha_ciphers_enabled"].(bool)
		tlsEcdheRsaWithAes128CbcShaCiphers = v["tls_ecdhe_rsa_with_aes128_cbc_sha_ciphers_enabled"].(bool)
		tlsRsaWithAes128GcmSha256Ciphers = v["tls_rsa_with_aes128_gcm_sha256_ciphers_enabled"].(bool)
		tlsRsaWithAes256CbcSha256Ciphers = v["tls_rsa_with_aes256_cbc_sha256_ciphers_enabled"].(bool)
		tlsRsaWithAes128CbcSha256Ciphers = v["tls_rsa_with_aes128_cbc_sha256_ciphers_enabled"].(bool)
		tlsRsaWithAes256CbcShaCiphers = v["tls_rsa_with_aes256_cbc_sha_ciphers_enabled"].(bool)
		tlsRsaWithAes128CbcShaCiphers = v["tls_rsa_with_aes128_cbc_sha_ciphers_enabled"].(bool)

		if skuIsConsumption && frontendProtocolSsl3 {
			return nil, fmt.Errorf("`enable_frontend_ssl30` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tripleDesCiphers {
			return nil, fmt.Errorf("`enable_triple_des_ciphers` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tlsEcdheEcdsaWithAes256CbcShaCiphers {
			return nil, fmt.Errorf("`tls_ecdhe_ecdsa_with_aes256_cbc_sha_ciphers_enabled` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tlsEcdheEcdsaWithAes128CbcShaCiphers {
			return nil, fmt.Errorf("`tls_ecdhe_ecdsa_with_aes128_cbc_sha_ciphers_enabled` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tlsEcdheRsaWithAes256CbcShaCiphers {
			return nil, fmt.Errorf("`tls_ecdhe_rsa_with_aes256_cbc_sha_ciphers_enabled` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tlsEcdheRsaWithAes128CbcShaCiphers {
			return nil, fmt.Errorf("`tls_ecdhe_rsa_with_aes128_cbc_sha_ciphers_enabled` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tlsRsaWithAes128GcmSha256Ciphers {
			return nil, fmt.Errorf("`tls_rsa_with_aes128_gcm_sha256_ciphers_enabled` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tlsRsaWithAes256CbcSha256Ciphers {
			return nil, fmt.Errorf("`tls_rsa_with_aes256_cbc_sha256_ciphers_enabled` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tlsRsaWithAes128CbcSha256Ciphers {
			return nil, fmt.Errorf("`tls_rsa_with_aes128_cbc_sha256_ciphers_enabled` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tlsRsaWithAes256CbcShaCiphers {
			return nil, fmt.Errorf("`tls_rsa_with_aes256_cbc_sha_ciphers_enabled` is not support for Sku Tier `Consumption`")
		}

		if skuIsConsumption && tlsRsaWithAes128CbcShaCiphers {
			return nil, fmt.Errorf("`tls_rsa_with_aes128_cbc_sha_ciphers_enabled` is not support for Sku Tier `Consumption`")
		}
	}

	customProperties := map[string]*string{
		apimBackendProtocolSsl3:   utils.String(strconv.FormatBool(backendProtocolSsl3)),
		apimBackendProtocolTls10:  utils.String(strconv.FormatBool(backendProtocolTls10)),
		apimBackendProtocolTls11:  utils.String(strconv.FormatBool(backendProtocolTls11)),
		apimFrontendProtocolTls10: utils.String(strconv.FormatBool(frontendProtocolTls10)),
		apimFrontendProtocolTls11: utils.String(strconv.FormatBool(frontendProtocolTls11)),
	}

	if !skuIsConsumption {
		customProperties[apimFrontendProtocolSsl3] = utils.String(strconv.FormatBool(frontendProtocolSsl3))
		customProperties[apimTripleDesCiphers] = utils.String(strconv.FormatBool(tripleDesCiphers))
		customProperties[apimTlsEcdheEcdsaWithAes256CbcShaCiphers] = utils.String(strconv.FormatBool(tlsEcdheEcdsaWithAes256CbcShaCiphers))
		customProperties[apimTlsEcdheEcdsaWithAes128CbcShaCiphers] = utils.String(strconv.FormatBool(tlsEcdheEcdsaWithAes128CbcShaCiphers))
		customProperties[apimTlsEcdheRsaWithAes256CbcShaCiphers] = utils.String(strconv.FormatBool(tlsEcdheRsaWithAes256CbcShaCiphers))
		customProperties[apimTlsEcdheRsaWithAes128CbcShaCiphers] = utils.String(strconv.FormatBool(tlsEcdheRsaWithAes128CbcShaCiphers))
		customProperties[apimTlsRsaWithAes128GcmSha256Ciphers] = utils.String(strconv.FormatBool(tlsRsaWithAes128GcmSha256Ciphers))
		customProperties[apimTlsRsaWithAes256CbcSha256Ciphers] = utils.String(strconv.FormatBool(tlsRsaWithAes256CbcSha256Ciphers))
		customProperties[apimTlsRsaWithAes128CbcSha256Ciphers] = utils.String(strconv.FormatBool(tlsRsaWithAes128CbcSha256Ciphers))
		customProperties[apimTlsRsaWithAes256CbcShaCiphers] = utils.String(strconv.FormatBool(tlsRsaWithAes256CbcShaCiphers))
		customProperties[apimTlsRsaWithAes128CbcShaCiphers] = utils.String(strconv.FormatBool(tlsRsaWithAes128CbcShaCiphers))
	}

	if vp := d.Get("protocols").([]interface{}); len(vp) > 0 {
		vpr := vp[0].(map[string]interface{})
		enableHttp2 := vpr["enable_http2"].(bool)
		customProperties[apimHttp2Protocol] = utils.String(strconv.FormatBool(enableHttp2))
	}

	return customProperties, nil
}

func expandAzureRmApiManagementVirtualNetworkConfigurations(d *pluginsdk.ResourceData) *apimanagement.VirtualNetworkConfiguration {
	vs := d.Get("virtual_network_configuration").([]interface{})
	if len(vs) == 0 {
		return nil
	}

	v := vs[0].(map[string]interface{})
	subnetResourceId := v["subnet_id"].(string)

	return &apimanagement.VirtualNetworkConfiguration{
		SubnetResourceID: &subnetResourceId,
	}
}

func flattenApiManagementSecurityCustomProperties(input map[string]*string, skuIsConsumption bool) []interface{} {
	output := make(map[string]interface{})

	output["enable_backend_ssl30"] = parseApiManagementNilableDictionary(input, apimBackendProtocolSsl3)
	output["enable_backend_tls10"] = parseApiManagementNilableDictionary(input, apimBackendProtocolTls10)
	output["enable_backend_tls11"] = parseApiManagementNilableDictionary(input, apimBackendProtocolTls11)
	output["enable_frontend_tls10"] = parseApiManagementNilableDictionary(input, apimFrontendProtocolTls10)
	output["enable_frontend_tls11"] = parseApiManagementNilableDictionary(input, apimFrontendProtocolTls11)

	if !skuIsConsumption {
		output["enable_frontend_ssl30"] = parseApiManagementNilableDictionary(input, apimFrontendProtocolSsl3)
		output["triple_des_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTripleDesCiphers)
		output["tls_ecdhe_ecdsa_with_aes256_cbc_sha_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTlsEcdheEcdsaWithAes256CbcShaCiphers)
		output["tls_ecdhe_ecdsa_with_aes128_cbc_sha_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTlsEcdheEcdsaWithAes128CbcShaCiphers)
		output["tls_ecdhe_rsa_with_aes256_cbc_sha_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTlsEcdheRsaWithAes256CbcShaCiphers)
		output["tls_ecdhe_rsa_with_aes128_cbc_sha_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTlsEcdheRsaWithAes128CbcShaCiphers)
		output["tls_rsa_with_aes128_gcm_sha256_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTlsRsaWithAes128GcmSha256Ciphers)
		output["tls_rsa_with_aes256_cbc_sha256_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTlsRsaWithAes256CbcSha256Ciphers)
		output["tls_rsa_with_aes128_cbc_sha256_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTlsRsaWithAes128CbcSha256Ciphers)
		output["tls_rsa_with_aes256_cbc_sha_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTlsRsaWithAes256CbcShaCiphers)
		output["tls_rsa_with_aes128_cbc_sha_ciphers_enabled"] = parseApiManagementNilableDictionary(input, apimTlsRsaWithAes128CbcShaCiphers)
	}

	return []interface{}{output}
}

func flattenApiManagementProtocolsCustomProperties(input map[string]*string) []interface{} {
	output := make(map[string]interface{})

	output["enable_http2"] = parseApiManagementNilableDictionary(input, apimHttp2Protocol)

	return []interface{}{output}
}

func flattenApiManagementVirtualNetworkConfiguration(input *apimanagement.VirtualNetworkConfiguration) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	virtualNetworkConfiguration := make(map[string]interface{})

	if input.SubnetResourceID != nil {
		virtualNetworkConfiguration["subnet_id"] = *input.SubnetResourceID
	}

	return []interface{}{virtualNetworkConfiguration}
}

func parseApiManagementNilableDictionary(input map[string]*string, key string) bool {
	log.Printf("Parsing value for %q", key)

	v, ok := input[key]
	if !ok {
		log.Printf("%q was not found in the input - returning `false` as the default value", key)
		return false
	}

	val, err := strconv.ParseBool(*v)
	if err != nil {
		log.Printf(" parsing %q (key %q) as bool: %+v - assuming false", key, *v, err)
		return false
	}

	return val
}

func expandApiManagementSignInSettings(input []interface{}) apimanagement.PortalSigninSettings {
	enabled := false

	if len(input) > 0 {
		vs := input[0].(map[string]interface{})
		enabled = vs["enabled"].(bool)
	}

	return apimanagement.PortalSigninSettings{
		PortalSigninSettingProperties: &apimanagement.PortalSigninSettingProperties{
			Enabled: utils.Bool(enabled),
		},
	}
}

func flattenApiManagementSignInSettings(input apimanagement.PortalSigninSettings) []interface{} {
	enabled := false

	if props := input.PortalSigninSettingProperties; props != nil {
		if props.Enabled != nil {
			enabled = *props.Enabled
		}
	}

	return []interface{}{
		map[string]interface{}{
			"enabled": enabled,
		},
	}
}

func expandApiManagementSignUpSettings(input []interface{}) apimanagement.PortalSignupSettings {
	if len(input) == 0 {
		return apimanagement.PortalSignupSettings{
			PortalSignupSettingsProperties: &apimanagement.PortalSignupSettingsProperties{
				Enabled: utils.Bool(false),
				TermsOfService: &apimanagement.TermsOfServiceProperties{
					ConsentRequired: utils.Bool(false),
					Enabled:         utils.Bool(false),
					Text:            utils.String(""),
				},
			},
		}
	}

	vs := input[0].(map[string]interface{})

	props := apimanagement.PortalSignupSettingsProperties{
		Enabled: utils.Bool(vs["enabled"].(bool)),
	}

	termsOfServiceRaw := vs["terms_of_service"].([]interface{})
	if len(termsOfServiceRaw) > 0 {
		termsOfServiceVs := termsOfServiceRaw[0].(map[string]interface{})
		props.TermsOfService = &apimanagement.TermsOfServiceProperties{
			Enabled:         utils.Bool(termsOfServiceVs["enabled"].(bool)),
			ConsentRequired: utils.Bool(termsOfServiceVs["consent_required"].(bool)),
			Text:            utils.String(termsOfServiceVs["text"].(string)),
		}
	}

	return apimanagement.PortalSignupSettings{
		PortalSignupSettingsProperties: &props,
	}
}

func flattenApiManagementSignUpSettings(input apimanagement.PortalSignupSettings) []interface{} {
	enabled := false
	termsOfService := make([]interface{}, 0)

	if props := input.PortalSignupSettingsProperties; props != nil {
		if props.Enabled != nil {
			enabled = *props.Enabled
		}

		if tos := props.TermsOfService; tos != nil {
			output := make(map[string]interface{})

			if tos.Enabled != nil {
				output["enabled"] = *tos.Enabled
			}

			if tos.ConsentRequired != nil {
				output["consent_required"] = *tos.ConsentRequired
			}

			if tos.Text != nil {
				output["text"] = *tos.Text
			}

			termsOfService = append(termsOfService, output)
		}
	}

	return []interface{}{
		map[string]interface{}{
			"enabled":          enabled,
			"terms_of_service": termsOfService,
		},
	}
}

func expandApiManagementPolicies(input []interface{}) (*apimanagement.PolicyContract, error) {
	if len(input) == 0 || input[0] == nil {
		return nil, nil
	}

	vs := input[0].(map[string]interface{})
	xmlContent := vs["xml_content"].(string)
	xmlLink := vs["xml_link"].(string)

	if xmlContent != "" {
		return &apimanagement.PolicyContract{
			PolicyContractProperties: &apimanagement.PolicyContractProperties{
				Format: apimanagement.PolicyContentFormatRawxml,
				Value:  utils.String(xmlContent),
			},
		}, nil
	}

	if xmlLink != "" {
		return &apimanagement.PolicyContract{
			PolicyContractProperties: &apimanagement.PolicyContractProperties{
				Format: apimanagement.PolicyContentFormatXMLLink,
				Value:  utils.String(xmlLink),
			},
		}, nil
	}

	return nil, fmt.Errorf("Either `xml_content` or `xml_link` should be set if the `policy` block is defined.")
}

func flattenApiManagementPolicies(d *pluginsdk.ResourceData, input apimanagement.PolicyContract) []interface{} {
	xmlContent := ""
	if props := input.PolicyContractProperties; props != nil {
		if props.Value != nil {
			xmlContent = *props.Value
		}
	}

	// if there's no policy assigned, we set this to an empty list
	if xmlContent == "" {
		return []interface{}{}
	}

	output := map[string]interface{}{
		"xml_content": xmlContent,
		"xml_link":    "",
	}

	// when you submit an `xml_link` to the API, the API downloads this link and stores it as `xml_content`
	// as such we need to retrieve this value from the state if it's present
	if existing, ok := d.GetOk("policy"); ok {
		existingVs := existing.([]interface{})
		if len(existingVs) > 0 {
			existingV := existingVs[0].(map[string]interface{})
			output["xml_link"] = existingV["xml_link"].(string)
		}
	}

	return []interface{}{output}
}

func expandApiManagementTenantAccessSettings(input []interface{}) apimanagement.AccessInformationUpdateParameters {
	enabled := false

	if len(input) > 0 {
		vs := input[0].(map[string]interface{})
		enabled = vs["enabled"].(bool)
	}

	return apimanagement.AccessInformationUpdateParameters{
		AccessInformationUpdateParameterProperties: &apimanagement.AccessInformationUpdateParameterProperties{
			Enabled: utils.Bool(enabled),
		},
	}
}

func flattenApiManagementTenantAccessSettings(input apimanagement.AccessInformationSecretsContract) []interface{} {
	result := make(map[string]interface{})

	result["enabled"] = *input.Enabled

	if input.ID != nil {
		result["tenant_id"] = *input.ID
	}

	if input.PrimaryKey != nil {
		result["primary_key"] = *input.PrimaryKey
	}

	if input.SecondaryKey != nil {
		result["secondary_key"] = *input.SecondaryKey

	}

	return []interface{}{result}
}

func flattenAPIManagementCertificates(d *pluginsdk.ResourceData, inputs *[]apimanagement.CertificateConfiguration) []interface{} {
	if inputs == nil || len(*inputs) == 0 {
		return []interface{}{}
	}

	outputs := []interface{}{}
	for i, input := range *inputs {
		var expiry, subject, thumbprint, pwd, encodedCertificate string
		if v, ok := d.GetOk(fmt.Sprintf("certificate.%d.certificate_password", i)); ok {
			pwd = v.(string)
		}

		if v, ok := d.GetOk(fmt.Sprintf("certificate.%d.encoded_certificate", i)); ok {
			encodedCertificate = v.(string)
		}

		if input.Certificate.Expiry != nil && !input.Certificate.Expiry.IsZero() {
			expiry = input.Certificate.Expiry.Format(time.RFC3339)
		}

		if input.Certificate.Thumbprint != nil {
			thumbprint = *input.Certificate.Thumbprint
		}

		if input.Certificate.Subject != nil {
			subject = *input.Certificate.Subject
		}

		output := map[string]interface{}{
			"certificate_password": pwd,
			"encoded_certificate":  encodedCertificate,
			"store_name":           string(input.StoreName),
			"expiry":               expiry,
			"subject":              subject,
			"thumbprint":           thumbprint,
		}
		outputs = append(outputs, output)
	}
	return outputs
}

func optedOutOfRecoveringSoftDeletedApiManagementErrorFmt(name, location string) string {
	message := `
An existing soft-deleted API Management exists with the Name %q in the location %q, however
automatically recovering this API Management has been disabled via the "features" block.

Terraform can automatically recover the soft-deleted API Management when this behaviour is
enabled within the "features" block (located within the "provider" block) - more
information can be found here:

https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs#features

Alternatively you can manually recover this (e.g. using the Azure CLI) and then import
this into Terraform via "terraform import", or pick a different name/location.
`
	return fmt.Sprintf(message, name, location)
}
