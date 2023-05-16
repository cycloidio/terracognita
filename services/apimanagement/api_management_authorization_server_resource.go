package apimanagement

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2021-08-01/apimanagement"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/schemaz"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceApiManagementAuthorizationServer() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceApiManagementAuthorizationServerCreateUpdate,
		Read:   resourceApiManagementAuthorizationServerRead,
		Update: resourceApiManagementAuthorizationServerCreateUpdate,
		Delete: resourceApiManagementAuthorizationServerDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.AuthorizationServerID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": schemaz.SchemaApiManagementChildName(),

			"api_management_name": schemaz.SchemaApiManagementName(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"authorization_endpoint": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"authorization_methods": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						string(apimanagement.AuthorizationMethodDELETE),
						string(apimanagement.AuthorizationMethodGET),
						string(apimanagement.AuthorizationMethodHEAD),
						string(apimanagement.AuthorizationMethodOPTIONS),
						string(apimanagement.AuthorizationMethodPATCH),
						string(apimanagement.AuthorizationMethodPOST),
						string(apimanagement.AuthorizationMethodPUT),
						string(apimanagement.AuthorizationMethodTRACE),
					}, false),
				},
				Set: pluginsdk.HashString,
			},

			"client_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"client_registration_endpoint": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"display_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"grant_types": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						string(apimanagement.GrantTypeAuthorizationCode),
						string(apimanagement.GrantTypeClientCredentials),
						string(apimanagement.GrantTypeImplicit),
						string(apimanagement.GrantTypeResourceOwnerPassword),
					}, false),
				},
				Set: pluginsdk.HashString,
			},

			// Optional
			"bearer_token_sending_methods": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						string(apimanagement.BearerTokenSendingMethodAuthorizationHeader),
						string(apimanagement.BearerTokenSendingMethodQuery),
					}, false),
				},
				Set: pluginsdk.HashString,
			},

			"client_authentication_method": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						string(apimanagement.ClientAuthenticationMethodBasic),
						string(apimanagement.ClientAuthenticationMethodBody),
					}, false),
				},
				Set: pluginsdk.HashString,
			},

			"client_secret": {
				Type:      pluginsdk.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"default_scope": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"resource_owner_username": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"resource_owner_password": {
				Type:      pluginsdk.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"support_state": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"token_body_parameter": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"value": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"token_endpoint": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceApiManagementAuthorizationServerCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.AuthorizationServersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewAuthorizationServerID(subscriptionId, d.Get("resource_group_name").(string), d.Get("api_management_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.ServiceName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_api_management_authorization_server", id.ID())
		}
	}

	authorizationEndpoint := d.Get("authorization_endpoint").(string)
	clientId := d.Get("client_id").(string)
	clientRegistrationEndpoint := d.Get("client_registration_endpoint").(string)
	displayName := d.Get("display_name").(string)
	grantTypesRaw := d.Get("grant_types").(*pluginsdk.Set).List()
	grantTypes := expandApiManagementAuthorizationServerGrantTypes(grantTypesRaw)

	clientAuthenticationMethodsRaw := d.Get("client_authentication_method").(*pluginsdk.Set).List()
	clientAuthenticationMethods := expandApiManagementAuthorizationServerClientAuthenticationMethods(clientAuthenticationMethodsRaw)
	clientSecret := d.Get("client_secret").(string)
	defaultScope := d.Get("default_scope").(string)
	description := d.Get("description").(string)
	resourceOwnerPassword := d.Get("resource_owner_password").(string)
	resourceOwnerUsername := d.Get("resource_owner_username").(string)
	supportState := d.Get("support_state").(bool)
	tokenBodyParametersRaw := d.Get("token_body_parameter").([]interface{})
	tokenBodyParameters := expandApiManagementAuthorizationServerTokenBodyParameters(tokenBodyParametersRaw)

	params := apimanagement.AuthorizationServerContract{
		AuthorizationServerContractProperties: &apimanagement.AuthorizationServerContractProperties{
			// Required
			AuthorizationEndpoint:      utils.String(authorizationEndpoint),
			ClientID:                   utils.String(clientId),
			ClientRegistrationEndpoint: utils.String(clientRegistrationEndpoint),
			DisplayName:                utils.String(displayName),
			GrantTypes:                 grantTypes,

			// Optional
			ClientAuthenticationMethod: clientAuthenticationMethods,
			ClientSecret:               utils.String(clientSecret),
			DefaultScope:               utils.String(defaultScope),
			Description:                utils.String(description),
			ResourceOwnerPassword:      utils.String(resourceOwnerPassword),
			ResourceOwnerUsername:      utils.String(resourceOwnerUsername),
			SupportState:               utils.Bool(supportState),
			TokenBodyParameters:        tokenBodyParameters,
		},
	}

	authorizationMethodsRaw := d.Get("authorization_methods").(*pluginsdk.Set).List()
	if len(authorizationMethodsRaw) > 0 {
		authorizationMethods := expandApiManagementAuthorizationServerAuthorizationMethods(authorizationMethodsRaw)
		params.AuthorizationServerContractProperties.AuthorizationMethods = authorizationMethods
	}

	bearerTokenSendingMethodsRaw := d.Get("bearer_token_sending_methods").(*pluginsdk.Set).List()
	if len(bearerTokenSendingMethodsRaw) > 0 {
		bearerTokenSendingMethods := expandApiManagementAuthorizationServerBearerTokenSendingMethods(bearerTokenSendingMethodsRaw)
		params.AuthorizationServerContractProperties.BearerTokenSendingMethods = bearerTokenSendingMethods
	}

	if tokenEndpoint := d.Get("token_endpoint").(string); tokenEndpoint != "" {
		params.AuthorizationServerContractProperties.TokenEndpoint = utils.String(tokenEndpoint)
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.ServiceName, id.Name, params, ""); err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceApiManagementAuthorizationServerRead(d, meta)
}

func resourceApiManagementAuthorizationServerRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.AuthorizationServersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AuthorizationServerID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ServiceName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %s does not exist - removing from state!", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("api_management_name", id.ServiceName)
	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if props := resp.AuthorizationServerContractProperties; props != nil {
		d.Set("authorization_endpoint", props.AuthorizationEndpoint)
		d.Set("client_id", props.ClientID)
		d.Set("client_registration_endpoint", props.ClientRegistrationEndpoint)
		d.Set("default_scope", props.DefaultScope)
		d.Set("description", props.Description)
		d.Set("display_name", props.DisplayName)
		d.Set("support_state", props.SupportState)
		d.Set("token_endpoint", props.TokenEndpoint)

		// TODO: Read properties from api, https://github.com/Azure/azure-rest-api-specs/issues/14128
		d.Set("resource_owner_password", d.Get("resource_owner_password").(string))
		d.Set("resource_owner_username", d.Get("resource_owner_username").(string))

		if err := d.Set("authorization_methods", flattenApiManagementAuthorizationServerAuthorizationMethods(props.AuthorizationMethods)); err != nil {
			return fmt.Errorf("flattening `authorization_methods`: %+v", err)
		}

		if err := d.Set("bearer_token_sending_methods", flattenApiManagementAuthorizationServerBearerTokenSendingMethods(props.BearerTokenSendingMethods)); err != nil {
			return fmt.Errorf("flattening `bearer_token_sending_methods`: %+v", err)
		}

		if err := d.Set("client_authentication_method", flattenApiManagementAuthorizationServerClientAuthenticationMethods(props.ClientAuthenticationMethod)); err != nil {
			return fmt.Errorf("flattening `client_authentication_method`: %+v", err)
		}

		if err := d.Set("grant_types", flattenApiManagementAuthorizationServerGrantTypes(props.GrantTypes)); err != nil {
			return fmt.Errorf("flattening `grant_types`: %+v", err)
		}

		if err := d.Set("token_body_parameter", flattenApiManagementAuthorizationServerTokenBodyParameters(props.TokenBodyParameters)); err != nil {
			return fmt.Errorf("flattening `token_body_parameter`: %+v", err)
		}
	}

	return nil
}

func resourceApiManagementAuthorizationServerDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.AuthorizationServersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AuthorizationServerID(d.Id())
	if err != nil {
		return err
	}

	if resp, err := client.Delete(ctx, id.ResourceGroup, id.ServiceName, id.Name, ""); err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("deleting %s: %s", *id, err)
		}
	}

	return nil
}

func expandApiManagementAuthorizationServerGrantTypes(input []interface{}) *[]apimanagement.GrantType {
	outputs := make([]apimanagement.GrantType, 0)

	for _, v := range input {
		grantType := apimanagement.GrantType(v.(string))
		outputs = append(outputs, grantType)
	}

	return &outputs
}

func flattenApiManagementAuthorizationServerGrantTypes(input *[]apimanagement.GrantType) []interface{} {
	outputs := make([]interface{}, 0)
	if input == nil {
		return outputs
	}

	for _, v := range *input {
		outputs = append(outputs, string(v))
	}

	return outputs
}

func expandApiManagementAuthorizationServerAuthorizationMethods(input []interface{}) *[]apimanagement.AuthorizationMethod {
	outputs := make([]apimanagement.AuthorizationMethod, 0)

	for _, v := range input {
		grantType := apimanagement.AuthorizationMethod(v.(string))
		outputs = append(outputs, grantType)
	}

	return &outputs
}

func flattenApiManagementAuthorizationServerAuthorizationMethods(input *[]apimanagement.AuthorizationMethod) []interface{} {
	outputs := make([]interface{}, 0)
	if input == nil {
		return outputs
	}

	for _, v := range *input {
		outputs = append(outputs, string(v))
	}

	return outputs
}

func expandApiManagementAuthorizationServerBearerTokenSendingMethods(input []interface{}) *[]apimanagement.BearerTokenSendingMethod {
	outputs := make([]apimanagement.BearerTokenSendingMethod, 0)

	for _, v := range input {
		grantType := apimanagement.BearerTokenSendingMethod(v.(string))
		outputs = append(outputs, grantType)
	}

	return &outputs
}

func flattenApiManagementAuthorizationServerBearerTokenSendingMethods(input *[]apimanagement.BearerTokenSendingMethod) []interface{} {
	outputs := make([]interface{}, 0)
	if input == nil {
		return outputs
	}

	for _, v := range *input {
		outputs = append(outputs, string(v))
	}

	return outputs
}

func expandApiManagementAuthorizationServerClientAuthenticationMethods(input []interface{}) *[]apimanagement.ClientAuthenticationMethod {
	outputs := make([]apimanagement.ClientAuthenticationMethod, 0)

	for _, v := range input {
		grantType := apimanagement.ClientAuthenticationMethod(v.(string))
		outputs = append(outputs, grantType)
	}

	return &outputs
}

func flattenApiManagementAuthorizationServerClientAuthenticationMethods(input *[]apimanagement.ClientAuthenticationMethod) []interface{} {
	outputs := make([]interface{}, 0)
	if input == nil {
		return outputs
	}

	for _, v := range *input {
		outputs = append(outputs, string(v))
	}

	return outputs
}

func expandApiManagementAuthorizationServerTokenBodyParameters(input []interface{}) *[]apimanagement.TokenBodyParameterContract {
	outputs := make([]apimanagement.TokenBodyParameterContract, 0)

	for _, v := range input {
		vs := v.(map[string]interface{})
		name := vs["name"].(string)
		value := vs["value"].(string)

		output := apimanagement.TokenBodyParameterContract{
			Name:  utils.String(name),
			Value: utils.String(value),
		}
		outputs = append(outputs, output)
	}

	return &outputs
}

func flattenApiManagementAuthorizationServerTokenBodyParameters(input *[]apimanagement.TokenBodyParameterContract) []interface{} {
	outputs := make([]interface{}, 0)
	if input == nil {
		return outputs
	}

	for _, v := range *input {
		output := make(map[string]interface{})

		if v.Name != nil {
			output["name"] = *v.Name
		}

		if v.Value != nil {
			output["value"] = *v.Value
		}

		outputs = append(outputs, output)
	}

	return outputs
}
