package logic

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/logic/mgmt/2019-05-01/logic"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/logic/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/logic/validate"
	networkParse "github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceIntegrationServiceEnvironment() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceIntegrationServiceEnvironmentCreateUpdate,
		Read:   resourceIntegrationServiceEnvironmentRead,
		Update: resourceIntegrationServiceEnvironmentCreateUpdate,
		Delete: resourceIntegrationServiceEnvironmentDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.IntegrationServiceEnvironmentID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(5 * time.Hour),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(5 * time.Hour),
			Delete: pluginsdk.DefaultTimeout(5 * time.Hour),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.IntegrationServiceEnvironmentName(),
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			// Maximum scale units that you can add	10 - https://docs.microsoft.com/en-US/azure/logic-apps/logic-apps-limits-and-config#integration-service-environment-ise
			// Developer Always 0 capacity
			"sku_name": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  "Developer_0",
				ValidateFunc: validation.StringInSlice([]string{
					"Developer_0",
					"Premium_0",
					"Premium_1",
					"Premium_2",
					"Premium_3",
					"Premium_4",
					"Premium_5",
					"Premium_6",
					"Premium_7",
					"Premium_8",
					"Premium_9",
					"Premium_10",
				}, false),
			},

			"access_endpoint_type": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true, // The access end point type cannot be changed once the integration service environment is provisioned.
				ValidateFunc: validation.StringInSlice([]string{
					string(logic.IntegrationServiceEnvironmentAccessEndpointTypeInternal),
					string(logic.IntegrationServiceEnvironmentAccessEndpointTypeExternal),
				}, false),
			},

			"virtual_network_subnet_ids": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				ForceNew: true, // The network configuration subnets cannot be updated after integration service environment is created.
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: networkValidate.SubnetID,
				},
				MinItems: 4,
				MaxItems: 4,
			},

			"connector_endpoint_ip_addresses": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
			},

			"connector_outbound_ip_addresses": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
			},

			"workflow_endpoint_ip_addresses": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
			},

			"workflow_outbound_ip_addresses": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
			},

			"tags": tags.Schema(),
		},

		CustomizeDiff: pluginsdk.CustomDiffWithAll(
			pluginsdk.ForceNewIfChange("sku_name", func(ctx context.Context, old, new, meta interface{}) bool {
				oldSku := strings.Split(old.(string), "_")
				newSku := strings.Split(new.(string), "_")
				// The SKU cannot be changed once integration service environment has been provisioned. -> we need ForceNew
				return oldSku[0] != newSku[0]
			}),
		),
	}
}

func resourceIntegrationServiceEnvironmentCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Logic.IntegrationServiceEnvironmentClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure ARM Integration Service Environment creation.")

	id := parse.NewIntegrationServiceEnvironmentID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_integration_service_environment", id.ID())
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	accessEndpointType := d.Get("access_endpoint_type").(string)
	virtualNetworkSubnetIds := d.Get("virtual_network_subnet_ids").(*pluginsdk.Set).List()
	t := d.Get("tags").(map[string]interface{})

	sku, err := expandIntegrationServiceEnvironmentSkuName(d.Get("sku_name").(string))
	if err != nil {
		return fmt.Errorf("expanding `sku_name` for %s: %v", id, err)
	}

	integrationServiceEnvironment := logic.IntegrationServiceEnvironment{
		Name:     &id.Name,
		Location: &location,
		Properties: &logic.IntegrationServiceEnvironmentProperties{
			NetworkConfiguration: &logic.NetworkConfiguration{
				AccessEndpoint: &logic.IntegrationServiceEnvironmentAccessEndpoint{
					Type: logic.IntegrationServiceEnvironmentAccessEndpointType(accessEndpointType),
				},
				Subnets: expandSubnetResourceID(virtualNetworkSubnetIds),
			},
		},
		Sku:  sku,
		Tags: tags.Expand(t),
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, integrationServiceEnvironment)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceIntegrationServiceEnvironmentRead(d, meta)
}

func resourceIntegrationServiceEnvironmentRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Logic.IntegrationServiceEnvironmentClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.IntegrationServiceEnvironmentID(d.Id())
	if err != nil {
		return err
	}

	name := id.Name
	resourceGroup := id.ResourceGroup

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Integration Service Environment %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if err := d.Set("sku_name", flattenIntegrationServiceEnvironmentSkuName(resp.Sku)); err != nil {
		return fmt.Errorf("setting `sku_name`: %+v", err)
	}

	if props := resp.Properties; props != nil {
		if netCfg := props.NetworkConfiguration; netCfg != nil {
			if accessEndpoint := netCfg.AccessEndpoint; accessEndpoint != nil {
				d.Set("access_endpoint_type", accessEndpoint.Type)
			}

			d.Set("virtual_network_subnet_ids", flattenSubnetResourceID(netCfg.Subnets))
		}

		if props.EndpointsConfiguration == nil || props.EndpointsConfiguration.Connector == nil {
			d.Set("connector_endpoint_ip_addresses", []interface{}{})
			d.Set("connector_outbound_ip_addresses", []interface{}{})
		} else {
			d.Set("connector_endpoint_ip_addresses", flattenIPAddresses(props.EndpointsConfiguration.Connector.AccessEndpointIPAddresses))
			d.Set("connector_outbound_ip_addresses", flattenIPAddresses(props.EndpointsConfiguration.Connector.OutgoingIPAddresses))
		}

		if props.EndpointsConfiguration == nil || props.EndpointsConfiguration.Workflow == nil {
			d.Set("workflow_endpoint_ip_addresses", []interface{}{})
			d.Set("workflow_outbound_ip_addresses", []interface{}{})
		} else {
			d.Set("workflow_endpoint_ip_addresses", flattenIPAddresses(props.EndpointsConfiguration.Workflow.AccessEndpointIPAddresses))
			d.Set("workflow_outbound_ip_addresses", flattenIPAddresses(props.EndpointsConfiguration.Workflow.OutgoingIPAddresses))
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceIntegrationServiceEnvironmentDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Logic.IntegrationServiceEnvironmentClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.IntegrationServiceEnvironmentID(d.Id())
	if err != nil {
		return fmt.Errorf("parsing Integration Service Environment ID `%q`: %+v", d.Id(), err)
	}

	name := id.Name
	resourceGroup := id.ResourceGroup

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil
		}
		return fmt.Errorf("retrieving Integration Service Environment %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	// Get subnet IDs before delete
	subnetIDs := getSubnetIDs(&resp)

	// Not optimal behaviour for now
	// It deletes synchronously and resource is not available anymore after return from delete operation
	// Next, after return - delete operation is still in progress in the background and is still occupying subnets.
	// As workaround we are checking on all involved subnets presence of serviceAssociationLink and resourceNavigationLink
	// If the operation fails we are lost. We do not have original resource and we cannot resume delete operation.
	// User has to wait for completion of delete operation in the background.
	// It would be great to have async call with future struct
	if resp, err := client.Delete(ctx, resourceGroup, name); err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return fmt.Errorf("deleting Integration Service Environment %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	stateConf := &pluginsdk.StateChangeConf{
		Pending:                   []string{string(logic.WorkflowProvisioningStateDeleting)},
		Target:                    []string{string(logic.WorkflowProvisioningStateDeleted)},
		MinTimeout:                5 * time.Minute,
		Refresh:                   integrationServiceEnvironmentDeleteStateRefreshFunc(ctx, meta.(*clients.Client), d.Id(), subnetIDs),
		Timeout:                   d.Timeout(pluginsdk.TimeoutDelete),
		ContinuousTargetOccurence: 1,
		NotFoundChecks:            1,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for deletion of Integration Service Environment %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	return nil
}

func flattenIntegrationServiceEnvironmentSkuName(input *logic.IntegrationServiceEnvironmentSku) string {
	if input == nil {
		return ""
	}

	return fmt.Sprintf("%s_%d", string(input.Name), *input.Capacity)
}

func expandIntegrationServiceEnvironmentSkuName(skuName string) (*logic.IntegrationServiceEnvironmentSku, error) {
	parts := strings.Split(skuName, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("sku_name (%s) has the wrong number of parts (%d) after splitting on _", skuName, len(parts))
	}

	var sku logic.IntegrationServiceEnvironmentSkuName
	switch parts[0] {
	case string(logic.IntegrationServiceEnvironmentSkuNameDeveloper):
		sku = logic.IntegrationServiceEnvironmentSkuNameDeveloper
	case string(logic.IntegrationServiceEnvironmentSkuNamePremium):
		sku = logic.IntegrationServiceEnvironmentSkuNamePremium
	default:
		return nil, fmt.Errorf("sku_name %s has unknown sku %s", skuName, parts[0])
	}

	capacity, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cannot convert sku_name %s capacity %s to int", skuName, parts[1])
	}

	if sku != logic.IntegrationServiceEnvironmentSkuNamePremium && capacity > 0 {
		return nil, fmt.Errorf("`capacity` can only be greater than zero for `sku_name` `Premium`")
	}

	return &logic.IntegrationServiceEnvironmentSku{
		Name:     sku,
		Capacity: utils.Int32(int32(capacity)),
	}, nil
}

func expandSubnetResourceID(input []interface{}) *[]logic.ResourceReference {
	results := make([]logic.ResourceReference, 0)
	for _, item := range input {
		results = append(results, logic.ResourceReference{
			ID: utils.String(item.(string)),
		})
	}
	return &results
}

func flattenSubnetResourceID(input *[]logic.ResourceReference) []interface{} {
	subnetIDs := make([]interface{}, 0)
	if input == nil {
		return subnetIDs
	}

	for _, resourceRef := range *input {
		if resourceRef.ID == nil || *resourceRef.ID == "" {
			continue
		}

		subnetIDs = append(subnetIDs, resourceRef.ID)
	}

	return subnetIDs
}

func getSubnetIDs(input *logic.IntegrationServiceEnvironment) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	if props := input.Properties; props != nil {
		if netCfg := props.NetworkConfiguration; netCfg != nil {
			return flattenSubnetResourceID(netCfg.Subnets)
		}
	}

	return results
}

func integrationServiceEnvironmentDeleteStateRefreshFunc(ctx context.Context, client *clients.Client, iseID string, subnetIDs []interface{}) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		linkExists, err := linkExists(ctx, client, iseID, subnetIDs)
		if err != nil {
			return string(logic.WorkflowProvisioningStateDeleting), string(logic.WorkflowProvisioningStateDeleting), err
		}

		if linkExists {
			return string(logic.WorkflowProvisioningStateDeleting), string(logic.WorkflowProvisioningStateDeleting), nil
		}

		return string(logic.WorkflowProvisioningStateDeleted), string(logic.WorkflowProvisioningStateDeleted), nil
	}
}

func linkExists(ctx context.Context, client *clients.Client, iseID string, subnetIDs []interface{}) (bool, error) {
	for _, subnetID := range subnetIDs {
		if subnetID == nil {
			continue
		}

		id := *(subnetID.(*string))
		log.Printf("Checking links on subnetID: %q\n", id)

		hasLink, err := serviceAssociationLinkExists(ctx, client.Network.ServiceAssociationLinkClient, iseID, id)
		if err != nil {
			return false, err
		}

		if hasLink {
			return true, nil
		} else {
			hasLink, err := resourceNavigationLinkExists(ctx, client.Network.ResourceNavigationLinkClient, id)
			if err != nil {
				return false, err
			}

			if hasLink {
				return true, nil
			}
		}
	}

	return false, nil
}

func serviceAssociationLinkExists(ctx context.Context, client *network.ServiceAssociationLinksClient, iseID string, subnetID string) (bool, error) {
	id, err := networkParse.SubnetID(subnetID)
	if err != nil {
		return false, err
	}

	resp, err := client.List(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return false, nil
		}
		return false, fmt.Errorf("retrieving Service Association Links from Virtual Network %q, subnet %q (Resource Group %q): %+v", id.VirtualNetworkName, id.Name, id.ResourceGroup, err)
	}

	if resp.Value != nil {
		for _, link := range *resp.Value {
			if link.ServiceAssociationLinkPropertiesFormat != nil && link.ServiceAssociationLinkPropertiesFormat.Link != nil {
				if strings.EqualFold(iseID, *link.ServiceAssociationLinkPropertiesFormat.Link) {
					log.Printf("Has Service Association Link: %q\n", *link.ID)
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func resourceNavigationLinkExists(ctx context.Context, client *network.ResourceNavigationLinksClient, subnetID string) (bool, error) {
	id, err := networkParse.SubnetID(subnetID)
	if err != nil {
		return false, err
	}

	resp, err := client.List(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return false, nil
		}
		return false, fmt.Errorf("retrieving Resource Navigation Links from Virtual Network %q, subnet %q (Resource Group %q): %+v", id.VirtualNetworkName, id.Name, id.ResourceGroup, err)
	}

	if resp.Value != nil {
		for _, link := range *resp.Value {
			log.Printf("Has Resource Navigation Link: %q\n", *link.ID)
			return true, nil
		}
	}

	return false, nil
}
