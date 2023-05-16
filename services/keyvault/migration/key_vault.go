package migration

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/set"
)

var _ pluginsdk.StateUpgrade = KeyVaultV0ToV1{}

type KeyVaultV0ToV1 struct{}

func (KeyVaultV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"location": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"resource_group_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"sku": {
			Type:     pluginsdk.TypeList,
			Required: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
				},
			},
		},

		"vault_uri": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"tenant_id": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"access_policy": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MinItems: 1,
			MaxItems: 16,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"tenant_id": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
					"object_id": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
					"application_id": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},
					"certificate_permissions": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
					"key_permissions": {
						Type:     pluginsdk.TypeList,
						Required: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
					"secret_permissions": {
						Type:     pluginsdk.TypeList,
						Required: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
				},
			},
		},

		"enabled_for_deployment": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"enabled_for_disk_encryption": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"enabled_for_template_deployment": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"tags": {
			Type:     pluginsdk.TypeMap,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},
	}
}

func (KeyVaultV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		inputAccessPolicies := rawState["access_policy"].([]interface{})
		if len(inputAccessPolicies) == 0 {
			return rawState, nil
		}

		outputAccessPolicies := make([]interface{}, 0)
		for _, accessPolicy := range inputAccessPolicies {
			policy := accessPolicy.(map[string]interface{})

			if v, ok := policy["certificate_permissions"]; ok {
				inputCertificatePermissions := v.([]interface{})
				outputCertificatePermissions := make([]string, 0)
				for _, p := range inputCertificatePermissions {
					permission := p.(string)
					if strings.ToLower(permission) == "all" {
						outputCertificatePermissions = append(outputCertificatePermissions, "create")
						outputCertificatePermissions = append(outputCertificatePermissions, "delete")
						outputCertificatePermissions = append(outputCertificatePermissions, "deleteissuers")
						outputCertificatePermissions = append(outputCertificatePermissions, "get")
						outputCertificatePermissions = append(outputCertificatePermissions, "getissuers")
						outputCertificatePermissions = append(outputCertificatePermissions, "import")
						outputCertificatePermissions = append(outputCertificatePermissions, "list")
						outputCertificatePermissions = append(outputCertificatePermissions, "listissuers")
						outputCertificatePermissions = append(outputCertificatePermissions, "managecontacts")
						outputCertificatePermissions = append(outputCertificatePermissions, "manageissuers")
						outputCertificatePermissions = append(outputCertificatePermissions, "setissuers")
						outputCertificatePermissions = append(outputCertificatePermissions, "update")
						break
					}
				}

				if len(outputCertificatePermissions) > 0 {
					policy["certificate_permissions"] = outputCertificatePermissions
				}
			}

			if v, ok := policy["key_permissions"]; ok {
				inputKeyPermissions := v.([]interface{})
				outputKeyPermissions := make([]string, 0)
				for _, p := range inputKeyPermissions {
					permission := p.(string)
					if strings.ToLower(permission) == "all" {
						outputKeyPermissions = append(outputKeyPermissions, "backup")
						outputKeyPermissions = append(outputKeyPermissions, "create")
						outputKeyPermissions = append(outputKeyPermissions, "decrypt")
						outputKeyPermissions = append(outputKeyPermissions, "delete")
						outputKeyPermissions = append(outputKeyPermissions, "encrypt")
						outputKeyPermissions = append(outputKeyPermissions, "get")
						outputKeyPermissions = append(outputKeyPermissions, "import")
						outputKeyPermissions = append(outputKeyPermissions, "list")
						outputKeyPermissions = append(outputKeyPermissions, "purge")
						outputKeyPermissions = append(outputKeyPermissions, "recover")
						outputKeyPermissions = append(outputKeyPermissions, "restore")
						outputKeyPermissions = append(outputKeyPermissions, "sign")
						outputKeyPermissions = append(outputKeyPermissions, "unwrapKey")
						outputKeyPermissions = append(outputKeyPermissions, "update")
						outputKeyPermissions = append(outputKeyPermissions, "verify")
						outputKeyPermissions = append(outputKeyPermissions, "wrapKey")
						break
					}
				}

				if len(outputKeyPermissions) > 0 {
					policy["key_permissions"] = outputKeyPermissions
				}
			}

			if v, ok := policy["secret_permissions"]; ok {
				inputSecretPermissions := v.([]interface{})
				outputSecretPermissions := make([]string, 0)
				for _, p := range inputSecretPermissions {
					permission := p.(string)
					if strings.ToLower(permission) == "all" {
						outputSecretPermissions = append(outputSecretPermissions, "backup")
						outputSecretPermissions = append(outputSecretPermissions, "delete")
						outputSecretPermissions = append(outputSecretPermissions, "get")
						outputSecretPermissions = append(outputSecretPermissions, "list")
						outputSecretPermissions = append(outputSecretPermissions, "purge")
						outputSecretPermissions = append(outputSecretPermissions, "recover")
						outputSecretPermissions = append(outputSecretPermissions, "restore")
						outputSecretPermissions = append(outputSecretPermissions, "set")
						break
					}
				}

				if len(outputSecretPermissions) > 0 {
					policy["secret_permissions"] = outputSecretPermissions
				}
			}

			outputAccessPolicies = append(outputAccessPolicies, policy)
		}

		rawState["access_policy"] = outputAccessPolicies
		return rawState, nil
	}
}

var _ pluginsdk.StateUpgrade = KeyVaultV1ToV2{}

type KeyVaultV1ToV2 struct{}

func (KeyVaultV1ToV2) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"location": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"resource_group_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"sku_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"tenant_id": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"access_policy": {
			Type:       pluginsdk.TypeList,
			ConfigMode: pluginsdk.SchemaConfigModeAttr,
			Optional:   true,
			Computed:   true,
			MaxItems:   1024,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"tenant_id": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
					"object_id": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
					"application_id": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},
					"certificate_permissions": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
					"key_permissions": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
					"secret_permissions": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
					"storage_permissions": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
				},
			},
		},

		"enabled_for_deployment": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"enabled_for_disk_encryption": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"enabled_for_template_deployment": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"enable_rbac_authorization": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"network_acls": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"default_action": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
					"bypass": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
					"ip_rules": {
						Type:     pluginsdk.TypeSet,
						Optional: true,
						Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
						Set:      pluginsdk.HashString,
					},
					"virtual_network_subnet_ids": {
						Type:     pluginsdk.TypeSet,
						Optional: true,
						Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
						Set:      set.HashStringIgnoreCase,
					},
				},
			},
		},

		"purge_protection_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"soft_delete_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Computed: true,
		},

		"soft_delete_retention_days": {
			Type:     pluginsdk.TypeInt,
			Optional: true,
		},

		"contact": {
			Type:     pluginsdk.TypeSet,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"email": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
					"name": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},
					"phone": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},
				},
			},
		},

		"tags": {
			Type:     pluginsdk.TypeMap,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		// Computed
		"vault_uri": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}

func (KeyVaultV1ToV2) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		// @tombuildsstuff: this is an int in the schema but was previously set into the
		// state as `*int32` - so using `.(int)` causes:
		// panic: interface conversion: interface {} is float64, not int
		// so I guess we try both?
		oldVal := 0
		if v, ok := rawState["soft_delete_retention_days"]; ok {
			if val, ok := v.(float64); ok {
				oldVal = int(val)
			}
			if val, ok := v.(int); ok {
				oldVal = val
			}
		}

		if oldVal == 0 {
			// The Azure API force-upgraded all Key Vaults so that Soft Delete was enabled on 2020-12-15
			// As a part of this, all Key Vaults got a default "retention days" of 90 days, however
			// for both newly created and upgraded key vaults, this value isn't returned unless it's
			// explicitly set by a user.
			//
			// As such we have to default this to a value of 90 days, which whilst assuming this default is
			// less than ideal, unfortunately there's little choice otherwise as this isn't returned.
			// Whilst the API Documentation doesn't show this default, it's listed here:
			//
			// > Once a secret, key, certificate, or key vault is deleted, it will remain recoverable
			// > for a configurable period of 7 to 90 calendar days. If no configuration is specified
			// > the default recovery period will be set to 90 days
			// https://docs.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview
			//
			// Notably this value cannot be updated once it's initially been configured, meaning that we
			// must not send this during creation if it's the default value, to allow users to change
			// this value later. This also means we can't use Terraform's "ForceNew" here without breaking
			// the one-time change.
			//
			// Hopefully in time this behaviour is fixed, but for now I'm not sure what else we can do.
			//
			// - @tombuildsstuff
			rawState["soft_delete_retention_days"] = 90
		}
		return rawState, nil
	}
}
