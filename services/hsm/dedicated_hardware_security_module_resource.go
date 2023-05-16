package hsm

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/zones"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Azure/azure-sdk-for-go/services/preview/hardwaresecuritymodules/mgmt/2018-10-31-preview/hardwaresecuritymodules"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	azValidate "github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/hsm/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/hsm/validate"
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceDedicatedHardwareSecurityModule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceDedicatedHardwareSecurityModuleCreate,
		Read:   resourceDedicatedHardwareSecurityModuleRead,
		Update: resourceDedicatedHardwareSecurityModuleUpdate,
		Delete: resourceDedicatedHardwareSecurityModuleDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.DedicatedHardwareSecurityModuleID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.DedicatedHardwareSecurityModuleName,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"location": commonschema.Location(),

			"sku_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(hardwaresecuritymodules.SafeNetLunaNetworkHSMA790),
				}, false),
			},

			"network_profile": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"network_interface_private_ip_addresses": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: azValidate.IPv4Address,
							},
						},

						"subnet_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: networkValidate.SubnetID,
						},
					},
				},
			},

			"stamp_id": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"stamp1",
					"stamp2",
				}, false),
			},

			"zones": commonschema.ZonesMultipleOptionalForceNew(),

			"tags": tags.Schema(),
		},
	}
}

func resourceDedicatedHardwareSecurityModuleCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).HSM.DedicatedHsmClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewDedicatedHardwareSecurityModuleID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	existing, err := client.Get(ctx, id.ResourceGroup, id.DedicatedHSMName)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_dedicated_hardware_security_module", id.ID())
	}

	parameters := hardwaresecuritymodules.DedicatedHsm{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		DedicatedHsmProperties: &hardwaresecuritymodules.DedicatedHsmProperties{
			NetworkProfile: expandDedicatedHsmNetworkProfile(d.Get("network_profile").([]interface{})),
		},
		Sku: &hardwaresecuritymodules.Sku{
			Name: hardwaresecuritymodules.Name(d.Get("sku_name").(string)),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if v, ok := d.GetOk("stamp_id"); ok {
		parameters.DedicatedHsmProperties.StampID = utils.String(v.(string))
	}

	if v, ok := d.GetOk("zones"); ok {
		zones := zones.Expand(v.(*schema.Set).List())
		if len(zones) > 0 {
			parameters.Zones = &zones
		}
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.DedicatedHSMName, parameters)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceDedicatedHardwareSecurityModuleRead(d, meta)
}

func resourceDedicatedHardwareSecurityModuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).HSM.DedicatedHsmClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DedicatedHardwareSecurityModuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.DedicatedHSMName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Dedicated Hardware Security Module %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Dedicate Hardware Security Module %q (Resource Group %q): %+v", id.DedicatedHSMName, id.ResourceGroup, err)
	}

	d.Set("name", id.DedicatedHSMName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	d.Set("zones", zones.Flatten(resp.Zones))

	if props := resp.DedicatedHsmProperties; props != nil {
		if err := d.Set("network_profile", flattenDedicatedHsmNetworkProfile(props.NetworkProfile)); err != nil {
			return fmt.Errorf("setting network_profile: %+v", err)
		}
		d.Set("stamp_id", props.StampID)
	}

	if sku := resp.Sku; sku != nil {
		d.Set("sku_name", sku.Name)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceDedicatedHardwareSecurityModuleUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).HSM.DedicatedHsmClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DedicatedHardwareSecurityModuleID(d.Id())
	if err != nil {
		return err
	}

	parameters := hardwaresecuritymodules.DedicatedHsmPatchParameters{}
	if d.HasChange("tags") {
		parameters.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.DedicatedHSMName, parameters)
	if err != nil {
		return fmt.Errorf("updating Dedicate Hardware Security Module %q (Resource Group %q): %+v", id.DedicatedHSMName, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on updating future for Dedicate Hardware Security Module %q (Resource Group %q): %+v", id.DedicatedHSMName, id.ResourceGroup, err)
	}

	return resourceDedicatedHardwareSecurityModuleRead(d, meta)
}

func resourceDedicatedHardwareSecurityModuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).HSM.DedicatedHsmClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DedicatedHardwareSecurityModuleID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.DedicatedHSMName)
	if err != nil {
		return fmt.Errorf("deleting Dedicated Hardware Security Module %q (Resource Group %q): %+v", id.DedicatedHSMName, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on deleting future for Dedicated Hardware Security Module %q (Resource Group %q): %+v", id.DedicatedHSMName, id.ResourceGroup, err)
	}

	return nil
}

func expandDedicatedHsmNetworkProfile(input []interface{}) *hardwaresecuritymodules.NetworkProfile {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	result := hardwaresecuritymodules.NetworkProfile{
		Subnet: &hardwaresecuritymodules.APIEntityReference{
			ID: utils.String(v["subnet_id"].(string)),
		},
		NetworkInterfaces: expandDedicatedHsmNetworkInterfacePrivateIPAddresses(v["network_interface_private_ip_addresses"].(*pluginsdk.Set).List()),
	}

	return &result
}

func expandDedicatedHsmNetworkInterfacePrivateIPAddresses(input []interface{}) *[]hardwaresecuritymodules.NetworkInterface {
	results := make([]hardwaresecuritymodules.NetworkInterface, 0)

	for _, item := range input {
		if item != nil {
			result := hardwaresecuritymodules.NetworkInterface{
				PrivateIPAddress: utils.String(item.(string)),
			}

			results = append(results, result)
		}
	}

	return &results
}

func flattenDedicatedHsmNetworkProfile(input *hardwaresecuritymodules.NetworkProfile) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var subnetId string
	if input.Subnet != nil && input.Subnet.ID != nil {
		subnetId = *input.Subnet.ID
	}

	return []interface{}{
		map[string]interface{}{
			"network_interface_private_ip_addresses": flattenDedicatedHsmNetworkInterfacePrivateIPAddresses(input.NetworkInterfaces),
			"subnet_id":                              subnetId,
		},
	}
}

func flattenDedicatedHsmNetworkInterfacePrivateIPAddresses(input *[]hardwaresecuritymodules.NetworkInterface) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		if item.PrivateIPAddress != nil {
			results = append(results, *item.PrivateIPAddress)
		}
	}

	return results
}
