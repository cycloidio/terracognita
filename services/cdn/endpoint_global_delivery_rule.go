package cdn

import (
	"github.com/Azure/azure-sdk-for-go/services/cdn/mgmt/2020-09-01/cdn"
	"github.com/hashicorp/terraform-provider-azurerm/services/cdn/deliveryruleactions"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func endpointGlobalDeliveryRule() *pluginsdk.Schema {
	//lintignore:XS003
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"cache_expiration_action": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem:     deliveryruleactions.CacheExpiration(),
				},

				"cache_key_query_string_action": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem:     deliveryruleactions.CacheKeyQueryString(),
				},

				"modify_request_header_action": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					Elem:     deliveryruleactions.ModifyRequestHeader(),
				},

				"modify_response_header_action": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					Elem:     deliveryruleactions.ModifyResponseHeader(),
				},

				"url_redirect_action": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem:     deliveryruleactions.URLRedirect(),
				},

				"url_rewrite_action": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem:     deliveryruleactions.URLRewrite(),
				},
			},
		},
	}
}

func expandArmCdnEndpointGlobalDeliveryRule(rule map[string]interface{}) (*cdn.DeliveryRule, error) {
	deliveryRule := cdn.DeliveryRule{
		Name:  utils.String("Global"),
		Order: utils.Int32(0),
	}

	actions, err := expandDeliveryRuleActions(rule)
	if err != nil {
		return nil, err
	}
	deliveryRule.Actions = &actions

	return &deliveryRule, nil
}

func flattenArmCdnEndpointGlobalDeliveryRule(deliveryRule cdn.DeliveryRule) (*map[string]interface{}, error) {
	actions, err := flattenDeliveryRuleActions(deliveryRule.Actions)
	if err != nil {
		return nil, err
	}

	output := make(map[string]interface{})
	for key, value := range *actions {
		output[key] = value
	}
	return &output, nil
}
