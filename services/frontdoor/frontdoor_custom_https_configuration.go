package frontdoor

import (
	"github.com/hashicorp/terraform-provider-azurerm/services/frontdoor/sdk/2020-05-01/frontdoors"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

func schemaCustomHttpsConfiguration() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"certificate_source": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Default:  string(frontdoors.FrontDoorCertificateSourceFrontDoor),
			ValidateFunc: validation.StringInSlice([]string{
				string(frontdoors.FrontDoorCertificateSourceAzureKeyVault),
				string(frontdoors.FrontDoorCertificateSourceFrontDoor),
			}, false),
		},
		"minimum_tls_version": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
		"provisioning_state": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
		"provisioning_substate": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
		// NOTE: None of these attributes are valid if
		//       certificate_source is set to FrontDoor
		"azure_key_vault_certificate_secret_name": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},
		"azure_key_vault_certificate_secret_version": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},
		"azure_key_vault_certificate_vault_id": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},
	}
}

type flattenedCustomHttpsConfiguration struct {
	CustomHTTPSConfiguration       []interface{}
	CustomHTTPSProvisioningEnabled bool
}

func flattenCustomHttpsConfiguration(properties *frontdoors.FrontendEndpointProperties) flattenedCustomHttpsConfiguration {
	result := flattenedCustomHttpsConfiguration{
		CustomHTTPSConfiguration:       make([]interface{}, 0),
		CustomHTTPSProvisioningEnabled: false,
	}

	if properties == nil {
		return result
	}

	if config := properties.CustomHttpsConfiguration; config != nil {
		certificateSource := string(frontdoors.FrontDoorCertificateSourceFrontDoor)
		keyVaultCertificateVaultId := ""
		keyVaultCertificateSecretName := ""
		keyVaultCertificateSecretVersion := ""
		provisioningState := ""
		provisioningSubstate := ""

		if config.CertificateSource == frontdoors.FrontDoorCertificateSourceAzureKeyVault {
			if vault := config.KeyVaultCertificateSourceParameters; vault != nil {
				certificateSource = string(frontdoors.FrontDoorCertificateSourceAzureKeyVault)

				if vault.Vault != nil && vault.Vault.Id != nil {
					keyVaultCertificateVaultId = *vault.Vault.Id
				}

				if vault.SecretName != nil {
					keyVaultCertificateSecretName = *vault.SecretName
				}

				if vault.SecretVersion != nil {
					keyVaultCertificateSecretVersion = *vault.SecretVersion
				}
			}
		}

		if properties.CustomHttpsProvisioningState != nil && *properties.CustomHttpsProvisioningState != "" {
			provisioningState = string(*properties.CustomHttpsProvisioningState)
			if properties.CustomHttpsProvisioningState != nil && *properties.CustomHttpsProvisioningState == frontdoors.CustomHttpsProvisioningStateEnabled || *properties.CustomHttpsProvisioningState == frontdoors.CustomHttpsProvisioningStateEnabling {
				result.CustomHTTPSProvisioningEnabled = true

				if properties.CustomHttpsProvisioningSubstate != nil && *properties.CustomHttpsProvisioningSubstate != "" {
					provisioningSubstate = string(*properties.CustomHttpsProvisioningSubstate)
				}
			}

			// Only return a CustomHTTPSConfiguration if CustomHTTPSConfiguration
			// is enabled
			if result.CustomHTTPSProvisioningEnabled {
				if certificateSource == string(frontdoors.FrontDoorCertificateSourceFrontDoor) {
					result.CustomHTTPSConfiguration = append(result.CustomHTTPSConfiguration, map[string]interface{}{
						"certificate_source":    certificateSource,
						"minimum_tls_version":   string(config.MinimumTlsVersion),
						"provisioning_state":    provisioningState,
						"provisioning_substate": provisioningSubstate,
					})
				} else {
					result.CustomHTTPSConfiguration = append(result.CustomHTTPSConfiguration, map[string]interface{}{
						"azure_key_vault_certificate_vault_id":       keyVaultCertificateVaultId,
						"azure_key_vault_certificate_secret_name":    keyVaultCertificateSecretName,
						"azure_key_vault_certificate_secret_version": keyVaultCertificateSecretVersion,
						"certificate_source":                         certificateSource,
						"minimum_tls_version":                        string(config.MinimumTlsVersion),
						"provisioning_state":                         provisioningState,
						"provisioning_substate":                      provisioningSubstate,
					})
				}
			}
		}
	}

	return result
}
