package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceVirtualHubSecurityPartnerProvider() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVirtualHubSecurityPartnerProviderCreate,
		Read:   resourceVirtualHubSecurityPartnerProviderRead,
		Update: resourceVirtualHubSecurityPartnerProviderUpdate,
		Delete: resourceVirtualHubSecurityPartnerProviderDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SecurityPartnerProviderID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"security_provider_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.SecurityProviderNameZScaler),
					string(network.SecurityProviderNameIBoss),
					string(network.SecurityProviderNameCheckpoint),
				}, false),
			},

			"virtual_hub_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: networkValidate.VirtualHubID,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceVirtualHubSecurityPartnerProviderCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityPartnerProviderClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewSecurityPartnerProviderID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing %s: %+v", id, err)
		}
	}

	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_virtual_hub_security_partner_provider", id.ID())
	}

	parameters := network.SecurityPartnerProvider{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		SecurityPartnerProviderPropertiesFormat: &network.SecurityPartnerProviderPropertiesFormat{
			SecurityProviderName: network.SecurityProviderName(d.Get("security_provider_name").(string)),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if v, ok := d.GetOk("virtual_hub_id"); ok {
		parameters.SecurityPartnerProviderPropertiesFormat.VirtualHub = &network.SubResource{
			ID: utils.String(v.(string)),
		}
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on creating future for %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceVirtualHubSecurityPartnerProviderRead(d, meta)
}

func resourceVirtualHubSecurityPartnerProviderRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityPartnerProviderClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SecurityPartnerProviderID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] security partner provider %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Security Partner Provider %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := resp.SecurityPartnerProviderPropertiesFormat; props != nil {
		d.Set("security_provider_name", props.SecurityProviderName)

		if props.VirtualHub != nil && props.VirtualHub.ID != nil {
			d.Set("virtual_hub_id", props.VirtualHub.ID)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceVirtualHubSecurityPartnerProviderUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityPartnerProviderClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SecurityPartnerProviderID(d.Id())
	if err != nil {
		return err
	}

	parameters := network.TagsObject{}

	if d.HasChange("tags") {
		parameters.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	if _, err := client.UpdateTags(ctx, id.ResourceGroup, id.Name, parameters); err != nil {
		return fmt.Errorf("updating Security Partner Provider %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return resourceVirtualHubSecurityPartnerProviderRead(d, meta)
}

func resourceVirtualHubSecurityPartnerProviderDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityPartnerProviderClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SecurityPartnerProviderID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Security Partner Provider %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on deleting future for Security Partner Provider %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}
