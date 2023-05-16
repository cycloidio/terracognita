package servicebus

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicebus/mgmt/2021-06-01-preview/servicebus"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

func expandAuthorizationRuleRights(d *pluginsdk.ResourceData) *[]servicebus.AccessRights {
	rights := make([]servicebus.AccessRights, 0)

	if d.Get("listen").(bool) {
		rights = append(rights, servicebus.AccessRightsListen)
	}

	if d.Get("send").(bool) {
		rights = append(rights, servicebus.AccessRightsSend)
	}

	if d.Get("manage").(bool) {
		rights = append(rights, servicebus.AccessRightsManage)
	}

	return &rights
}

func flattenAuthorizationRuleRights(rights *[]servicebus.AccessRights) (listen, send, manage bool) {
	// zero (initial) value for a bool in go is false

	if rights != nil {
		for _, right := range *rights {
			switch right {
			case servicebus.AccessRightsListen:
				listen = true
			case servicebus.AccessRightsSend:
				send = true
			case servicebus.AccessRightsManage:
				manage = true
			default:
				log.Printf("[DEBUG] Unknown Authorization Rule Right '%s'", right)
			}
		}
	}

	return listen, send, manage
}

func authorizationRuleSchemaFrom(s map[string]*pluginsdk.Schema) map[string]*pluginsdk.Schema {
	s["listen"] = &pluginsdk.Schema{
		Type:     pluginsdk.TypeBool,
		Optional: true,
		Default:  false,
	}
	s["send"] = &pluginsdk.Schema{
		Type:     pluginsdk.TypeBool,
		Optional: true,
		Default:  false,
	}
	s["manage"] = &pluginsdk.Schema{
		Type:     pluginsdk.TypeBool,
		Optional: true,
		Default:  false,
	}
	s["primary_key"] = &pluginsdk.Schema{
		Type:      pluginsdk.TypeString,
		Computed:  true,
		Sensitive: true,
	}
	s["primary_connection_string"] = &pluginsdk.Schema{
		Type:      pluginsdk.TypeString,
		Computed:  true,
		Sensitive: true,
	}
	s["secondary_key"] = &pluginsdk.Schema{
		Type:      pluginsdk.TypeString,
		Computed:  true,
		Sensitive: true,
	}
	s["secondary_connection_string"] = &pluginsdk.Schema{
		Type:      pluginsdk.TypeString,
		Computed:  true,
		Sensitive: true,
	}
	s["primary_connection_string_alias"] = &pluginsdk.Schema{
		Type:      pluginsdk.TypeString,
		Computed:  true,
		Sensitive: true,
	}
	s["secondary_connection_string_alias"] = &pluginsdk.Schema{
		Type:      pluginsdk.TypeString,
		Computed:  true,
		Sensitive: true,
	}
	return s
}

func authorizationRuleCustomizeDiff(ctx context.Context, d *pluginsdk.ResourceDiff, _ interface{}) error {
	listen, hasListen := d.GetOk("listen")
	send, hasSend := d.GetOk("send")
	manage, hasManage := d.GetOk("manage")

	if !hasListen && !hasSend && !hasManage {
		return fmt.Errorf("One of the `listen`, `send` or `manage` properties needs to be set")
	}

	if manage.(bool) && (!listen.(bool) || !send.(bool)) {
		return fmt.Errorf("if `manage` is set both `listen` and `send` must be set to true too")
	}

	return nil
}

func waitForPairedNamespaceReplication(ctx context.Context, meta interface{}, resourceGroup, namespaceName string, timeout time.Duration) error {
	namespaceClient := meta.(*clients.Client).ServiceBus.NamespacesClient
	namespace, err := namespaceClient.Get(ctx, resourceGroup, namespaceName)

	if !strings.EqualFold(string(namespace.Sku.Name), "Premium") {
		return err
	}

	disasterRecoveryClient := meta.(*clients.Client).ServiceBus.DisasterRecoveryConfigsClient
	disasterRecoveryResponse, err := disasterRecoveryClient.List(ctx, resourceGroup, namespaceName)
	if disasterRecoveryResponse.Values() == nil {
		return err
	}

	if len(disasterRecoveryResponse.Values()) != 1 {
		return err
	}

	aliasName := *disasterRecoveryResponse.Values()[0].Name

	stateConf := &pluginsdk.StateChangeConf{
		Pending:    []string{string(servicebus.ProvisioningStateDRAccepted)},
		Target:     []string{string(servicebus.ProvisioningStateDRSucceeded)},
		MinTimeout: 30 * time.Second,
		Timeout:    timeout,
		Refresh: func() (interface{}, string, error) {
			read, err := disasterRecoveryClient.Get(ctx, resourceGroup, namespaceName, aliasName)
			if err != nil {
				return nil, "error", fmt.Errorf("wait read Service Bus Namespace Disaster Recovery Configs %q (Namespace %q / Resource Group %q): %v", aliasName, namespaceName, resourceGroup, err)
			}

			if props := read.ArmDisasterRecoveryProperties; props != nil {
				if props.ProvisioningState == servicebus.ProvisioningStateDRFailed {
					return read, "failed", fmt.Errorf("replication for Service Bus Namespace Disaster Recovery Configs %q (Namespace %q / Resource Group %q) failed", aliasName, namespaceName, resourceGroup)
				}
				return read, string(props.ProvisioningState), nil
			}

			return read, "nil", fmt.Errorf("waiting for replication error Service Bus Namespace Disaster Recovery Configs %q (Namespace %q / Resource Group %q): provisioning state is nil", aliasName, namespaceName, resourceGroup)
		},
	}

	_, waitErr := stateConf.WaitForStateContext(ctx)
	return waitErr
}
