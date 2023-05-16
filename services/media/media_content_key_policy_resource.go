package media

import (
	b64 "encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/mediaservices/mgmt/2021-05-01/media"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/media/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/media/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceMediaContentKeyPolicy() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMediaContentKeyPolicyCreateUpdate,
		Read:   resourceMediaContentKeyPolicyRead,
		Update: resourceMediaContentKeyPolicyCreateUpdate,
		Delete: resourceMediaContentKeyPolicyDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ContentKeyPolicyID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[-a-zA-Z0-9(_)]{1,128}$"),
					"Content Key Policy name must be 1 - 128 characters long, can contain letters, numbers, underscores, and hyphens (but the first and last character must be a letter or number).",
				),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"media_services_account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.AccountName,
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"policy_option": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"clear_key_configuration_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},

						"widevine_configuration_template": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						//lintignore:XS003
						"playready_configuration_license": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"allow_test_devices": {
										Type:     pluginsdk.TypeBool,
										Optional: true,
									},

									"begin_date": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},

									"content_key_location_from_header_enabled": {
										Type:     pluginsdk.TypeBool,
										Optional: true,
									},

									"content_key_location_from_key_id": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsUUID,
									},

									"content_type": {
										Type:     pluginsdk.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(media.ContentKeyPolicyPlayReadyContentTypeUltraVioletDownload),
											string(media.ContentKeyPolicyPlayReadyContentTypeUltraVioletStreaming),
											string(media.ContentKeyPolicyPlayReadyContentTypeUnspecified),
										}, false),
									},

									"expiration_date": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},

									"grace_period": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validation.StringIsNotEmpty,
									},

									"license_type": {
										Type:     pluginsdk.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(media.ContentKeyPolicyPlayReadyLicenseTypeNonPersistent),
											string(media.ContentKeyPolicyPlayReadyLicenseTypePersistent),
										}, false),
									},

									//lintignore:XS003
									"play_right": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &pluginsdk.Resource{
											Schema: map[string]*pluginsdk.Schema{
												"agc_and_color_stripe_restriction": {
													Type:         pluginsdk.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 3),
												},

												"allow_passing_video_content_to_unknown_output": {
													Type:     pluginsdk.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														string(media.ContentKeyPolicyPlayReadyUnknownOutputPassingOptionAllowed),
														string(media.ContentKeyPolicyPlayReadyUnknownOutputPassingOptionAllowedWithVideoConstriction),
														string(media.ContentKeyPolicyPlayReadyUnknownOutputPassingOptionNotAllowed),
													}, false),
												},

												"analog_video_opl": {
													Type:         pluginsdk.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntInSlice([]int{100, 150, 200}),
												},

												"compressed_digital_audio_opl": {
													Type:         pluginsdk.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntInSlice([]int{100, 150, 200}),
												},

												"digital_video_only_content_restriction": {
													Type:     pluginsdk.TypeBool,
													Optional: true,
												},

												"first_play_expiration": {
													Type:         pluginsdk.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringIsNotEmpty,
												},

												"image_constraint_for_analog_component_video_restriction": {
													Type:     pluginsdk.TypeBool,
													Optional: true,
												},

												"image_constraint_for_analog_computer_monitor_restriction": {
													Type:     pluginsdk.TypeBool,
													Optional: true,
												},

												"scms_restriction": {
													Type:         pluginsdk.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 3),
												},

												"uncompressed_digital_audio_opl": {
													Type:         pluginsdk.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntInSlice([]int{100, 150, 250, 300}),
												},

												"uncompressed_digital_video_opl": {
													Type:         pluginsdk.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntInSlice([]int{100, 250, 270, 300}),
												},
											},
										},
									},
									"relative_begin_date": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},

									"relative_expiration_date": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},
								},
							},
						},
						//lintignore:XS003
						"fairplay_configuration": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"ask": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"pfx": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"pfx_password": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									//lintignore:XS003
									"offline_rental_configuration": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &pluginsdk.Resource{
											Schema: map[string]*pluginsdk.Schema{
												"playback_duration_seconds": {
													Type:         pluginsdk.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"storage_duration_seconds": {
													Type:         pluginsdk.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
											},
										},
									},
									"rental_and_lease_key_type": {
										Type:     pluginsdk.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(media.ContentKeyPolicyFairPlayRentalAndLeaseKeyTypeDualExpiry),
											string(media.ContentKeyPolicyFairPlayRentalAndLeaseKeyTypePersistentLimited),
											string(media.ContentKeyPolicyFairPlayRentalAndLeaseKeyTypePersistentUnlimited),
											string(media.ContentKeyPolicyFairPlayRentalAndLeaseKeyTypeUndefined),
										}, false),
									},
									"rental_duration_seconds": {
										Type:         pluginsdk.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						//lintignore:XS003
						"token_restriction": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"audience": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"issuer": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"token_type": {
										Type:     pluginsdk.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(media.ContentKeyPolicyRestrictionTokenTypeJwt),
											string(media.ContentKeyPolicyRestrictionTokenTypeSwt),
										}, false),
									},
									"primary_symmetric_token_key": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsBase64,
										Sensitive:    true,
									},
									"primary_rsa_token_key_exponent": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
										Sensitive:    true,
									},
									"primary_rsa_token_key_modulus": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
										Sensitive:    true,
									},
									"primary_x509_token_key_raw": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
										Sensitive:    true,
									},
									"open_id_connect_discovery_document": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									//lintignore:XS003
									"required_claim": {
										Type:     pluginsdk.TypeList,
										Optional: true,
										Elem: &pluginsdk.Resource{
											Schema: map[string]*pluginsdk.Schema{
												"type": {
													Type:         pluginsdk.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringIsNotEmpty,
												},
												"value": {
													Type:         pluginsdk.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringIsNotEmpty,
												},
											},
										},
									},
								},
							},
						},
						"open_restriction_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceMediaContentKeyPolicyCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.ContentKeyPoliciesClient
	subscriptionID := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceID := parse.NewContentKeyPolicyID(subscriptionID, d.Get("resource_group_name").(string), d.Get("media_services_account_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceID.ResourceGroup, resourceID.MediaserviceName, resourceID.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of %s: %+v", resourceID, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_media_content_key_policy", resourceID.ID())
		}
	}

	parameters := media.ContentKeyPolicy{
		ContentKeyPolicyProperties: &media.ContentKeyPolicyProperties{},
	}

	if description, ok := d.GetOk("description"); ok {
		parameters.ContentKeyPolicyProperties.Description = utils.String(description.(string))
	}

	if v, ok := d.GetOk("policy_option"); ok {
		options, err := expandPolicyOptions(v.(*pluginsdk.Set).List())
		if err != nil {
			return err
		}
		parameters.ContentKeyPolicyProperties.Options = options
	}

	_, err := client.CreateOrUpdate(ctx, resourceID.ResourceGroup, resourceID.MediaserviceName, resourceID.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", resourceID, err)
	}

	d.SetId(resourceID.ID())

	return resourceMediaContentKeyPolicyRead(d, meta)
}

func resourceMediaContentKeyPolicyRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.ContentKeyPoliciesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ContentKeyPolicyID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.GetPolicyPropertiesWithSecrets(ctx, id.ResourceGroup, id.MediaserviceName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s was not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("media_services_account_name", id.MediaserviceName)
	d.Set("description", resp.Description)

	if resp.Options != nil {
		options, err := flattenPolicyOptions(resp.Options)
		if err != nil {
			return err
		}

		d.Set("policy_option", options)
	}

	return nil
}

func resourceMediaContentKeyPolicyDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.ContentKeyPoliciesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ContentKeyPolicyID(d.Id())
	if err != nil {
		return err
	}

	if _, err = client.Delete(ctx, id.ResourceGroup, id.MediaserviceName, id.Name); err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	return nil
}

func expandPolicyOptions(input []interface{}) (*[]media.ContentKeyPolicyOption, error) {
	results := make([]media.ContentKeyPolicyOption, 0)

	for _, policyOptionRaw := range input {
		policyOption := policyOptionRaw.(map[string]interface{})

		restriction, err := expandRestriction(policyOption)
		if err != nil {
			return nil, err
		}

		configuration, err := expandConfiguration(policyOption)
		if err != nil {
			return nil, err
		}

		contentKeyPolicyOption := media.ContentKeyPolicyOption{
			Restriction:   restriction,
			Configuration: configuration,
		}

		if v := policyOption["name"]; v != nil {
			contentKeyPolicyOption.Name = utils.String(v.(string))
		}

		results = append(results, contentKeyPolicyOption)
	}

	return &results, nil
}

func flattenPolicyOptions(input *[]media.ContentKeyPolicyOption) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}

	results := make([]interface{}, 0)
	for _, option := range *input {
		name := ""
		if option.Name != nil {
			name = *option.Name
		}

		clearKeyConfigurationEnabled := false
		playReadyLicense := make([]interface{}, 0)
		widevineTemplate := ""
		fairplayConfiguration := make([]interface{}, 0)
		if v := option.Configuration; v != nil {
			switch v.(type) {
			case media.ContentKeyPolicyClearKeyConfiguration:
				clearKeyConfigurationEnabled = true
			case media.ContentKeyPolicyWidevineConfiguration:
				wideVineConfiguration, ok := v.AsContentKeyPolicyWidevineConfiguration()
				if !ok {
					return nil, fmt.Errorf("content key configuration was not a Widevine Configuration")
				}

				if wideVineConfiguration.WidevineTemplate != nil {
					widevineTemplate = *wideVineConfiguration.WidevineTemplate
				}
			case media.ContentKeyPolicyFairPlayConfiguration:
				fairPlayConfiguration, ok := v.AsContentKeyPolicyFairPlayConfiguration()
				if !ok {
					return nil, fmt.Errorf("content key configuration was not a Fairplay Configuration")
				}
				fairplayConfiguration = flattenFairplayConfiguration(fairPlayConfiguration)
			case media.ContentKeyPolicyPlayReadyConfiguration:
				playReadyConfiguration, ok := v.AsContentKeyPolicyPlayReadyConfiguration()
				if !ok {
					return nil, fmt.Errorf("content key configuration was not a Playready Configuration")
				}
				if playReadyConfiguration.Licenses != nil {
					license, err := flattenPlayReadyLicenses(playReadyConfiguration.Licenses)
					if err != nil {
						return nil, err
					}
					playReadyLicense = license
				}
			}
		}

		openRestrictionEnabled := false
		tokenRestriction := make([]interface{}, 0)
		if v := option.Restriction; v != nil {
			switch v.(type) {
			case media.ContentKeyPolicyOpenRestriction:
				openRestrictionEnabled = true
			case media.ContentKeyPolicyTokenRestriction:
				token, ok := v.AsContentKeyPolicyTokenRestriction()
				if !ok {
					return nil, fmt.Errorf("content key restriction was not a Token Restriction")
				}
				restriction, err := flattenTokenRestriction(token)
				if err != nil {
					return nil, err
				}
				tokenRestriction = restriction
			}
		}

		results = append(results, map[string]interface{}{
			"name":                            name,
			"clear_key_configuration_enabled": clearKeyConfigurationEnabled,
			"playready_configuration_license": playReadyLicense,
			"widevine_configuration_template": widevineTemplate,
			"fairplay_configuration":          fairplayConfiguration,
			"open_restriction_enabled":        openRestrictionEnabled,
			"token_restriction":               tokenRestriction,
		})
	}

	return results, nil
}

func expandRestriction(option map[string]interface{}) (media.BasicContentKeyPolicyRestriction, error) {
	restrictionCount := 0
	restrictionType := ""
	if option["open_restriction_enabled"] != nil && option["open_restriction_enabled"].(bool) {
		restrictionCount++
		restrictionType = string(media.OdataTypeBasicContentKeyPolicyRestrictionOdataTypeMicrosoftMediaContentKeyPolicyOpenRestriction)
	}
	if option["token_restriction"] != nil && len(option["token_restriction"].([]interface{})) > 0 && option["token_restriction"].([]interface{})[0] != nil {
		restrictionCount++
		restrictionType = string(media.OdataTypeBasicContentKeyPolicyRestrictionOdataTypeMicrosoftMediaContentKeyPolicyTokenRestriction)
	}

	if restrictionCount == 0 {
		return nil, fmt.Errorf("policy_option must contain at least one type of restriction: open_restriction_enabled or token_restriction.")
	}

	if restrictionCount > 1 {
		return nil, fmt.Errorf("more than one type of restriction in the same policy_option is not allowed.")
	}

	switch restrictionType {
	case string(media.OdataTypeBasicContentKeyPolicyRestrictionOdataTypeMicrosoftMediaContentKeyPolicyOpenRestriction):
		openRestriction := &media.ContentKeyPolicyOpenRestriction{
			OdataType: media.OdataTypeBasicContentKeyPolicyRestrictionOdataTypeMicrosoftMediaContentKeyPolicyOpenRestriction,
		}
		return openRestriction, nil
	case string(media.OdataTypeBasicContentKeyPolicyRestrictionOdataTypeMicrosoftMediaContentKeyPolicyTokenRestriction):
		tokenRestrictions := option["token_restriction"].([]interface{})
		tokenRestriction := tokenRestrictions[0].(map[string]interface{})
		contentKeyPolicyTokenRestriction := &media.ContentKeyPolicyTokenRestriction{
			OdataType: media.OdataTypeBasicContentKeyPolicyRestrictionOdataTypeMicrosoftMediaContentKeyPolicyTokenRestriction,
		}
		if tokenRestriction["audience"] != nil && tokenRestriction["audience"].(string) != "" {
			contentKeyPolicyTokenRestriction.Audience = utils.String(tokenRestriction["audience"].(string))
		}
		if tokenRestriction["issuer"] != nil && tokenRestriction["issuer"].(string) != "" {
			contentKeyPolicyTokenRestriction.Issuer = utils.String(tokenRestriction["issuer"].(string))
		}
		if tokenRestriction["token_type"] != nil && tokenRestriction["token_type"].(string) != "" {
			contentKeyPolicyTokenRestriction.RestrictionTokenType = media.ContentKeyPolicyRestrictionTokenType(tokenRestriction["token_type"].(string))
		}
		if tokenRestriction["open_id_connect_discovery_document"] != nil && tokenRestriction["open_id_connect_discovery_document"].(string) != "" {
			contentKeyPolicyTokenRestriction.OpenIDConnectDiscoveryDocument = utils.String(tokenRestriction["open_id_connect_discovery_document"].(string))
		}
		if v := tokenRestriction["required_claim"]; v != nil {
			contentKeyPolicyTokenRestriction.RequiredClaims = expandRequiredClaims(v.([]interface{}))
		}
		primaryVerificationKey, err := expandVerificationKey(tokenRestriction)
		if err != nil {
			return nil, err
		}
		contentKeyPolicyTokenRestriction.PrimaryVerificationKey = primaryVerificationKey

		return contentKeyPolicyTokenRestriction, nil
	default:
		return nil, fmt.Errorf("policy_option must contain at least one type of restriction: open_restriction_enabled or token_restriction.")
	}
}

func flattenTokenRestriction(input *media.ContentKeyPolicyTokenRestriction) ([]interface{}, error) {
	if input == nil {
		return make([]interface{}, 0), nil
	}

	audience := ""
	if input.Audience != nil {
		audience = *input.Audience
	}

	issuer := ""
	if input.Issuer != nil {
		issuer = *input.Issuer
	}

	openIDConnectDiscoveryDocument := ""
	if input.OpenIDConnectDiscoveryDocument != nil {
		openIDConnectDiscoveryDocument = *input.OpenIDConnectDiscoveryDocument
	}

	requiredClaims := make([]interface{}, 0)
	if input.RequiredClaims != nil {
		requiredClaims = flattenRequiredClaims(input.RequiredClaims)
	}

	symmetricToken := ""
	rsaTokenKeyExponent := ""
	rsaTokenKeyModulus := ""
	x509TokenBodyRaw := ""
	if v := input.PrimaryVerificationKey; v != nil {
		switch v.(type) {
		case media.ContentKeyPolicySymmetricTokenKey:
			symmetricTokenKey, ok := v.AsContentKeyPolicySymmetricTokenKey()
			if !ok {
				return nil, fmt.Errorf("token key was not Symmetric Token Key")
			}

			if symmetricTokenKey.KeyValue != nil {
				symmetricToken = b64.StdEncoding.EncodeToString(*symmetricTokenKey.KeyValue)
			}
		case media.ContentKeyPolicyRsaTokenKey:
			rsaTokenKey, ok := v.AsContentKeyPolicyRsaTokenKey()
			if !ok {
				return nil, fmt.Errorf("token key was not Rsa Token Key")
			}

			if rsaTokenKey.Exponent != nil {
				rsaTokenKeyExponent = string(*rsaTokenKey.Exponent)
			}

			if rsaTokenKey.Modulus != nil {
				rsaTokenKeyModulus = string(*rsaTokenKey.Modulus)
			}
		case media.ContentKeyPolicyX509CertificateTokenKey:
			x509CertificateTokenKey, ok := v.AsContentKeyPolicyX509CertificateTokenKey()
			if !ok {
				return nil, fmt.Errorf("token key was not x509Certificate Token Key")
			}

			if x509CertificateTokenKey.RawBody != nil {
				x509TokenBodyRaw = string(*x509CertificateTokenKey.RawBody)
			}
		}
	}

	return []interface{}{
		map[string]interface{}{
			"audience":                           audience,
			"issuer":                             issuer,
			"token_type":                         string(input.RestrictionTokenType),
			"open_id_connect_discovery_document": openIDConnectDiscoveryDocument,
			"required_claim":                     requiredClaims,
			"primary_symmetric_token_key":        symmetricToken,
			"primary_x509_token_key_raw":         x509TokenBodyRaw,
			"primary_rsa_token_key_exponent":     rsaTokenKeyExponent,
			"primary_rsa_token_key_modulus":      rsaTokenKeyModulus,
		},
	}, nil
}

func expandConfiguration(input map[string]interface{}) (media.BasicContentKeyPolicyConfiguration, error) {
	configurationCount := 0
	configurationType := ""
	if input["clear_key_configuration_enabled"] != nil && input["clear_key_configuration_enabled"].(bool) {
		configurationCount++
		configurationType = string(media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyClearKeyConfiguration)
	}
	if input["widevine_configuration_template"] != nil && input["widevine_configuration_template"].(string) != "" {
		configurationCount++
		configurationType = string(media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyWidevineConfiguration)
	}
	if input["fairplay_configuration"] != nil && len(input["fairplay_configuration"].([]interface{})) > 0 && input["fairplay_configuration"].([]interface{})[0] != nil {
		configurationCount++
		configurationType = string(media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyFairPlayConfiguration)
	}

	if input["playready_configuration_license"] != nil && len(input["playready_configuration_license"].([]interface{})) > 0 {
		configurationCount++
		configurationType = string(media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyPlayReadyConfiguration)
	}

	if configurationCount == 0 {
		return nil, fmt.Errorf("policy_option must contain at least one type of configuration: clear_key_configuration_enabled , widevine_configuration_template, playready_configuration_license or fairplay_configuration.")
	}

	if configurationCount > 1 {
		return nil, fmt.Errorf("more than one type of configuration in the same policy_option is not allowed.")
	}

	switch configurationType {
	case string(media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyClearKeyConfiguration):
		clearKeyConfiguration := &media.ContentKeyPolicyClearKeyConfiguration{
			OdataType: media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyClearKeyConfiguration,
		}
		return clearKeyConfiguration, nil
	case string(media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyWidevineConfiguration):
		wideVineConfiguration := &media.ContentKeyPolicyWidevineConfiguration{
			OdataType:        media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyWidevineConfiguration,
			WidevineTemplate: utils.String(input["widevine_configuration_template"].(string)),
		}
		return wideVineConfiguration, nil
	case string(media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyFairPlayConfiguration):
		fairplayConfiguration, err := expandFairplayConfiguration(input["fairplay_configuration"].([]interface{}))
		if err != nil {
			return nil, err
		}
		return fairplayConfiguration, nil
	case string(media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyPlayReadyConfiguration):
		playReadyConfiguration := &media.ContentKeyPolicyPlayReadyConfiguration{
			OdataType: media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyPlayReadyConfiguration,
		}

		if input["playready_configuration_license"] != nil {
			licenses, err := expandPlayReadyLicenses(input["playready_configuration_license"].([]interface{}))
			if err != nil {
				return nil, err
			}
			playReadyConfiguration.Licenses = licenses
		}
		return playReadyConfiguration, nil

	default:
		return nil, fmt.Errorf("policy_option must contain at least one type of configuration: clear_key_configuration_enabled , widevine_configuration_template, playready_configuration_license or fairplay_configuration.")
	}
}

func expandVerificationKey(input map[string]interface{}) (media.BasicContentKeyPolicyRestrictionTokenKey, error) {
	verificationKeyCount := 0
	verificationKeyType := ""
	if input["primary_symmetric_token_key"] != nil && input["primary_symmetric_token_key"].(string) != "" {
		verificationKeyCount++
		verificationKeyType = string(media.OdataTypeBasicContentKeyPolicyRestrictionTokenKeyOdataTypeMicrosoftMediaContentKeyPolicySymmetricTokenKey)
	}
	if (input["primary_rsa_token_key_exponent"] != nil && input["primary_rsa_token_key_exponent"].(string) != "") || (input["primary_rsa_token_key_modulus"] != nil && input["primary_rsa_token_key_modulus"].(string) != "") {
		verificationKeyCount++
		verificationKeyType = string(media.OdataTypeBasicContentKeyPolicyRestrictionTokenKeyOdataTypeMicrosoftMediaContentKeyPolicyRsaTokenKey)
	}

	if input["primary_x509_token_key_raw"] != nil && input["primary_x509_token_key_raw"].(string) != "" {
		verificationKeyCount++
		verificationKeyType = string(media.OdataTypeBasicContentKeyPolicyRestrictionTokenKeyOdataTypeMicrosoftMediaContentKeyPolicyX509CertificateTokenKey)
	}

	if verificationKeyCount > 1 {
		return nil, fmt.Errorf("more than one type of token key in the same token_restriction is not allowed.")
	}

	switch verificationKeyType {
	case string(media.OdataTypeBasicContentKeyPolicyRestrictionTokenKeyOdataTypeMicrosoftMediaContentKeyPolicySymmetricTokenKey):
		symmetricTokenKey := &media.ContentKeyPolicySymmetricTokenKey{
			OdataType: media.OdataTypeBasicContentKeyPolicyRestrictionTokenKeyOdataTypeMicrosoftMediaContentKeyPolicySymmetricTokenKey,
		}

		if input["primary_symmetric_token_key"] != nil && input["primary_symmetric_token_key"].(string) != "" {
			keyValue, _ := b64.StdEncoding.DecodeString(input["primary_symmetric_token_key"].(string))
			symmetricTokenKey.KeyValue = &keyValue
		}
		return symmetricTokenKey, nil
	case string(media.OdataTypeBasicContentKeyPolicyRestrictionTokenKeyOdataTypeMicrosoftMediaContentKeyPolicyRsaTokenKey):
		rsaTokenKey := &media.ContentKeyPolicyRsaTokenKey{
			OdataType: media.OdataTypeBasicContentKeyPolicyRestrictionTokenKeyOdataTypeMicrosoftMediaContentKeyPolicyRsaTokenKey,
		}
		if input["primary_rsa_token_key_exponent"] != nil && input["primary_rsa_token_key_exponent"].(string) != "" {
			exponent := []byte(input["primary_rsa_token_key_exponent"].(string))
			rsaTokenKey.Exponent = &exponent
		}
		if input["primary_rsa_token_key_modulus"] != nil && input["primary_rsa_token_key_modulus"].(string) != "" {
			modulus := []byte(input["primary_rsa_token_key_modulus"].(string))
			rsaTokenKey.Modulus = &modulus
		}
		return rsaTokenKey, nil
	case string(media.OdataTypeBasicContentKeyPolicyRestrictionTokenKeyOdataTypeMicrosoftMediaContentKeyPolicyX509CertificateTokenKey):
		x509CertificateTokenKey := &media.ContentKeyPolicyX509CertificateTokenKey{
			OdataType: media.OdataTypeBasicContentKeyPolicyRestrictionTokenKeyOdataTypeMicrosoftMediaContentKeyPolicyX509CertificateTokenKey,
		}

		if input["primary_x509_token_key_raw"] != nil && input["primary_x509_token_key_raw"].(string) != "" {
			rawBody := []byte(input["primary_x509_token_key_raw"].(string))
			x509CertificateTokenKey.RawBody = &rawBody
		}
		return x509CertificateTokenKey, nil
	default:
		return nil, nil
	}
}

func expandRequiredClaims(input []interface{}) *[]media.ContentKeyPolicyTokenClaim {
	results := make([]media.ContentKeyPolicyTokenClaim, 0)

	for _, tokenClaimRaw := range input {
		if tokenClaimRaw == nil {
			continue
		}
		tokenClaim := tokenClaimRaw.(map[string]interface{})

		claimType := ""
		if v := tokenClaim["type"]; v != nil {
			claimType = v.(string)
		}

		claimValue := ""
		if v := tokenClaim["value"]; v != nil {
			claimValue = v.(string)
		}

		contentPolicyTokenClaim := media.ContentKeyPolicyTokenClaim{
			ClaimType:  &claimType,
			ClaimValue: &claimValue,
		}

		results = append(results, contentPolicyTokenClaim)
	}

	return &results
}

func flattenRequiredClaims(input *[]media.ContentKeyPolicyTokenClaim) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)
	for _, tokenClaim := range *input {
		claimValue := ""
		if tokenClaim.ClaimValue != nil {
			claimValue = *tokenClaim.ClaimValue
		}

		claimType := ""
		if tokenClaim.ClaimType != nil {
			claimType = *tokenClaim.ClaimType
		}

		results = append(results, map[string]interface{}{
			"value": claimValue,
			"type":  claimType,
		})
	}

	return results
}

func expandRentalConfiguration(input []interface{}) *media.ContentKeyPolicyFairPlayOfflineRentalConfiguration {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	rentalConfiguration := input[0].(map[string]interface{})
	playbackDuration := utils.Int64(int64(rentalConfiguration["playback_duration_seconds"].(int)))
	storageDuration := utils.Int64(int64(rentalConfiguration["storage_duration_seconds"].(int)))
	return &media.ContentKeyPolicyFairPlayOfflineRentalConfiguration{
		PlaybackDurationSeconds: playbackDuration,
		StorageDurationSeconds:  storageDuration,
	}
}

func flattenRentalConfiguration(input *media.ContentKeyPolicyFairPlayOfflineRentalConfiguration) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	playbackDurationSeconds := 0
	if input.PlaybackDurationSeconds != nil {
		playbackDurationSeconds = int(*input.PlaybackDurationSeconds)
	}

	storageDurationSeconds := 0
	if input.StorageDurationSeconds != nil {
		storageDurationSeconds = int(*input.StorageDurationSeconds)
	}

	return []interface{}{map[string]interface{}{
		"playback_duration_seconds": playbackDurationSeconds,
		"storage_duration_seconds":  storageDurationSeconds,
	}}
}

func expandFairplayConfiguration(input []interface{}) (*media.ContentKeyPolicyFairPlayConfiguration, error) {
	fairplayConfiguration := &media.ContentKeyPolicyFairPlayConfiguration{
		OdataType: media.OdataTypeBasicContentKeyPolicyConfigurationOdataTypeMicrosoftMediaContentKeyPolicyFairPlayConfiguration,
	}

	fairplay := input[0].(map[string]interface{})
	if fairplay["rental_duration_seconds"] != nil {
		fairplayConfiguration.RentalDuration = utils.Int64(int64(fairplay["rental_duration_seconds"].(int)))
	}

	if fairplay["offline_rental_configuration"] != nil {
		fairplayConfiguration.OfflineRentalConfiguration = expandRentalConfiguration(fairplay["offline_rental_configuration"].([]interface{}))
	}

	if fairplay["rental_and_lease_key_type"] != nil {
		fairplayConfiguration.RentalAndLeaseKeyType = media.ContentKeyPolicyFairPlayRentalAndLeaseKeyType(fairplay["rental_and_lease_key_type"].(string))
	}

	if fairplay["ask"] != nil && fairplay["ask"].(string) != "" {
		askBytes, err := hex.DecodeString(fairplay["ask"].(string))
		if err != nil {
			return nil, err
		}
		fairplayConfiguration.Ask = &askBytes
	}

	if fairplay["pfx"] != nil && fairplay["pfx"].(string) != "" {
		fairplayConfiguration.FairPlayPfx = utils.String(fairplay["pfx"].(string))
	}

	if fairplay["pfx_password"] != nil && fairplay["pfx_password"].(string) != "" {
		fairplayConfiguration.FairPlayPfxPassword = utils.String(fairplay["pfx_password"].(string))
	}

	return fairplayConfiguration, nil
}

func flattenFairplayConfiguration(input *media.ContentKeyPolicyFairPlayConfiguration) []interface{} {
	rentalDuration := 0
	if input.RentalDuration != nil {
		rentalDuration = int(*input.RentalDuration)
	}

	offlineRentalConfiguration := make([]interface{}, 0)
	if input.OfflineRentalConfiguration != nil {
		offlineRentalConfiguration = flattenRentalConfiguration(input.OfflineRentalConfiguration)
	}

	pfx := ""
	if input.FairPlayPfx != nil {
		pfx = *input.FairPlayPfx
	}

	pfxPassword := ""
	if input.FairPlayPfxPassword != nil {
		pfxPassword = *input.FairPlayPfxPassword
	}

	ask := ""
	if input.Ask != nil {
		ask = hex.EncodeToString(*input.Ask)
	}

	return []interface{}{
		map[string]interface{}{
			"rental_duration_seconds":      rentalDuration,
			"offline_rental_configuration": offlineRentalConfiguration,
			"rental_and_lease_key_type":    string(input.RentalAndLeaseKeyType),
			"pfx":                          pfx,
			"pfx_password":                 pfxPassword,
			"ask":                          ask,
		},
	}
}

func expandPlayReadyLicenses(input []interface{}) (*[]media.ContentKeyPolicyPlayReadyLicense, error) {
	results := make([]media.ContentKeyPolicyPlayReadyLicense, 0)

	for _, licenseRaw := range input {
		if licenseRaw == nil {
			continue
		}
		license := licenseRaw.(map[string]interface{})
		playReadyLicense := media.ContentKeyPolicyPlayReadyLicense{}

		if v := license["allow_test_devices"]; v != nil {
			playReadyLicense.AllowTestDevices = utils.Bool(v.(bool))
		}

		if v := license["begin_date"]; v != nil && v != "" {
			beginDate, err := date.ParseTime(time.RFC3339, v.(string))
			if err != nil {
				return nil, err
			}
			playReadyLicense.BeginDate = &date.Time{
				Time: beginDate,
			}
		}

		locationFromHeader := false
		if v := license["content_key_location_from_header_enabled"]; v != nil && v != "" {
			playReadyLicense.ContentKeyLocation = media.ContentKeyPolicyPlayReadyContentEncryptionKeyFromHeader{
				OdataType: media.OdataTypeMicrosoftMediaContentKeyPolicyPlayReadyContentEncryptionKeyFromHeader,
			}
			locationFromHeader = true
		}

		if v := license["content_key_location_from_key_id"]; v != nil && v != "" {
			if locationFromHeader {
				return nil, fmt.Errorf("playready_configuration_license only support one key location at time, you must to specify content_key_location_from_header_enabled or content_key_location_from_key_id but not both at the same time")
			}

			keyID := uuid.FromStringOrNil(v.(string))
			playReadyLicense.ContentKeyLocation = media.ContentKeyPolicyPlayReadyContentEncryptionKeyFromKeyIdentifier{
				OdataType: media.OdataTypeMicrosoftMediaContentKeyPolicyPlayReadyContentEncryptionKeyFromHeader,
				KeyID:     &keyID,
			}
		}

		if v := license["content_type"]; v != nil && v != "" {
			playReadyLicense.ContentType = media.ContentKeyPolicyPlayReadyContentType(v.(string))
		}

		if v := license["expiration_date"]; v != nil && v != "" {
			expirationDate, err := date.ParseTime(time.RFC3339, v.(string))
			if err != nil {
				return nil, err
			}
			playReadyLicense.ExpirationDate = &date.Time{
				Time: expirationDate,
			}
		}

		if v := license["grace_period"]; v != nil && v != "" {
			playReadyLicense.GracePeriod = utils.String(v.(string))
		}

		if v := license["license_type"]; v != nil && v != "" {
			playReadyLicense.LicenseType = media.ContentKeyPolicyPlayReadyLicenseType(v.(string))
		}

		if v := license["play_right"]; v != nil {
			playReadyLicense.PlayRight = expandPlayRight(v.([]interface{}))
		}

		if v := license["relative_begin_date"]; v != nil && v != "" {
			playReadyLicense.RelativeBeginDate = utils.String(v.(string))
		}

		if v := license["relative_expiration_date"]; v != nil && v != "" {
			playReadyLicense.RelativeExpirationDate = utils.String(v.(string))
		}

		results = append(results, playReadyLicense)
	}

	return &results, nil
}

func flattenPlayReadyLicenses(input *[]media.ContentKeyPolicyPlayReadyLicense) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}

	results := make([]interface{}, 0)
	for _, v := range *input {
		allowTestDevices := false
		if v.AllowTestDevices != nil {
			allowTestDevices = *v.AllowTestDevices
		}

		beginDate := ""
		if v.BeginDate != nil {
			beginDate = v.BeginDate.Format(time.RFC3339)
		}

		locationFromHeaderEnabled := false
		locationFromKeyID := ""
		if v.ContentKeyLocation != nil {
			switch v.ContentKeyLocation.(type) {
			case media.ContentKeyPolicyPlayReadyContentEncryptionKeyFromHeader:
				locationFromHeaderEnabled = true
			case media.ContentKeyPolicyPlayReadyContentEncryptionKeyFromKeyIdentifier:
				keyLocation, ok := v.ContentKeyLocation.AsContentKeyPolicyPlayReadyContentEncryptionKeyFromKeyIdentifier()
				if !ok {
					return nil, fmt.Errorf("Content key Play ready location was not a Content Encryption Key from Key Identifier")
				}
				locationFromKeyID = keyLocation.KeyID.String()
			}
		}

		expirationDate := ""
		if v.ExpirationDate != nil {
			expirationDate = v.ExpirationDate.Format(time.RFC3339)
		}

		gracePeriod := ""
		if v.GracePeriod != nil {
			gracePeriod = *v.GracePeriod
		}

		playRight := make([]interface{}, 0)
		if v.PlayRight != nil {
			playRight = flattenPlayRight(v.PlayRight)
		}

		relativeBeginDate := ""
		if v.RelativeBeginDate != nil {
			relativeBeginDate = *v.RelativeBeginDate
		}

		relativeExpirationDate := ""
		if v.RelativeExpirationDate != nil {
			relativeExpirationDate = *v.RelativeExpirationDate
		}

		results = append(results, map[string]interface{}{
			"allow_test_devices": allowTestDevices,
			"begin_date":         beginDate,
			"content_key_location_from_header_enabled": locationFromHeaderEnabled,
			"content_key_location_from_key_id":         locationFromKeyID,
			"content_type":                             string(v.ContentType),
			"expiration_date":                          expirationDate,
			"grace_period":                             gracePeriod,
			"license_type":                             string(v.LicenseType),
			"relative_begin_date":                      relativeBeginDate,
			"relative_expiration_date":                 relativeExpirationDate,
			"play_right":                               playRight,
		})
	}

	return results, nil
}

func expandPlayRight(input []interface{}) *media.ContentKeyPolicyPlayReadyPlayRight {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	playRight := &media.ContentKeyPolicyPlayReadyPlayRight{}
	playRightConfiguration := input[0].(map[string]interface{})

	if v := playRightConfiguration["agc_and_color_stripe_restriction"]; v != nil {
		playRight.AgcAndColorStripeRestriction = utils.Int32(int32(v.(int)))
	}

	if v := playRightConfiguration["allow_passing_video_content_to_unknown_output"]; v != nil {
		playRight.AllowPassingVideoContentToUnknownOutput = media.ContentKeyPolicyPlayReadyUnknownOutputPassingOption(v.(string))
	}

	if v := playRightConfiguration["analog_video_opl"]; v != nil && v != 0 {
		playRight.AnalogVideoOpl = utils.Int32(int32(v.(int)))
	}

	if v := playRightConfiguration["compressed_digital_audio_opl"]; v != nil && v != 0 {
		playRight.CompressedDigitalAudioOpl = utils.Int32(int32(v.(int)))
	}

	if v := playRightConfiguration["digital_video_only_content_restriction"]; v != nil {
		playRight.DigitalVideoOnlyContentRestriction = utils.Bool(v.(bool))
	}

	if v := playRightConfiguration["first_play_expiration"]; v != nil && v != "" {
		playRight.FirstPlayExpiration = utils.String(v.(string))
	}

	if v := playRightConfiguration["image_constraint_for_analog_component_video_restriction"]; v != nil {
		playRight.ImageConstraintForAnalogComponentVideoRestriction = utils.Bool(v.(bool))
	}

	if v := playRightConfiguration["image_constraint_for_analog_computer_monitor_restriction"]; v != nil {
		playRight.ImageConstraintForAnalogComputerMonitorRestriction = utils.Bool(v.(bool))
	}

	if v := playRightConfiguration["scms_restriction"]; v != nil {
		playRight.ScmsRestriction = utils.Int32(int32(v.(int)))
	}
	if v := playRightConfiguration["uncompressed_digital_audio_opl"]; v != nil && v != 0 {
		playRight.UncompressedDigitalAudioOpl = utils.Int32(int32(v.(int)))
	}

	if v := playRightConfiguration["uncompressed_digital_video_opl"]; v != nil && v != 0 {
		playRight.UncompressedDigitalVideoOpl = utils.Int32(int32(v.(int)))
	}

	return playRight
}

func flattenPlayRight(input *media.ContentKeyPolicyPlayReadyPlayRight) []interface{} {
	agcStripeRestriction := 0
	if input.AgcAndColorStripeRestriction != nil {
		agcStripeRestriction = int(*input.AgcAndColorStripeRestriction)
	}

	analogVideoOpl := 0
	if input.AnalogVideoOpl != nil {
		analogVideoOpl = int(*input.AnalogVideoOpl)
	}

	compressedDigitalAudioOpl := 0
	if input.AnalogVideoOpl != nil {
		compressedDigitalAudioOpl = int(*input.CompressedDigitalAudioOpl)
	}

	digitalVideoOnlyContentRestriction := false
	if input.DigitalVideoOnlyContentRestriction != nil {
		digitalVideoOnlyContentRestriction = *input.DigitalVideoOnlyContentRestriction
	}

	firstPlayExpiration := ""
	if input.FirstPlayExpiration != nil {
		firstPlayExpiration = *input.FirstPlayExpiration
	}

	imageConstraintForAnalogComponentVideoRestriction := false
	if input.ImageConstraintForAnalogComponentVideoRestriction != nil {
		imageConstraintForAnalogComponentVideoRestriction = *input.ImageConstraintForAnalogComponentVideoRestriction
	}

	imageConstraintForAnalogComputerMonitorRestriction := false
	if input.ImageConstraintForAnalogComputerMonitorRestriction != nil {
		imageConstraintForAnalogComputerMonitorRestriction = *input.ImageConstraintForAnalogComputerMonitorRestriction
	}

	scmsRestriction := 0
	if input.ScmsRestriction != nil {
		scmsRestriction = int(*input.ScmsRestriction)
	}

	uncompressedDigitalAudioOpl := 0
	if input.UncompressedDigitalAudioOpl != nil {
		uncompressedDigitalAudioOpl = int(*input.UncompressedDigitalAudioOpl)
	}

	uncompressedDigitalVideoOpl := 0
	if input.UncompressedDigitalVideoOpl != nil {
		uncompressedDigitalVideoOpl = int(*input.UncompressedDigitalVideoOpl)
	}

	return []interface{}{
		map[string]interface{}{
			"agc_and_color_stripe_restriction":                         agcStripeRestriction,
			"allow_passing_video_content_to_unknown_output":            string(input.AllowPassingVideoContentToUnknownOutput),
			"analog_video_opl":                                         analogVideoOpl,
			"compressed_digital_audio_opl":                             compressedDigitalAudioOpl,
			"digital_video_only_content_restriction":                   digitalVideoOnlyContentRestriction,
			"first_play_expiration":                                    firstPlayExpiration,
			"image_constraint_for_analog_component_video_restriction":  imageConstraintForAnalogComponentVideoRestriction,
			"image_constraint_for_analog_computer_monitor_restriction": imageConstraintForAnalogComputerMonitorRestriction,
			"scms_restriction":                                         scmsRestriction,
			"uncompressed_digital_audio_opl":                           uncompressedDigitalAudioOpl,
			"uncompressed_digital_video_opl":                           uncompressedDigitalVideoOpl,
		},
	}
}
