package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceBastionHost() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceBastionHostCreateUpdate,
		Read:   resourceBastionHostRead,
		Update: resourceBastionHostCreateUpdate,
		Delete: resourceBastionHostDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.BastionHostID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.BastionHostName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"copy_paste_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"file_copy_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"ip_configuration": {
				Type:     pluginsdk.TypeList,
				ForceNew: true,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.BastionIPConfigName,
						},
						"subnet_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.BastionSubnetName,
						},
						"public_ip_address_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
					},
				},
			},

			"ip_connect_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"scale_units": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(2, 50),
				Default:      2,
			},

			"shareable_link_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"sku": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.BastionHostSkuNameBasic),
					string(network.BastionHostSkuNameStandard),
				}, false),
				Default: string(network.BastionHostSkuNameBasic),
			},

			"tunneling_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"dns_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceBastionHostCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.BastionHostsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Println("[INFO] preparing arguments for Azure Bastion Host creation.")

	id := parse.NewBastionHostID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})
	scaleUnits := d.Get("scale_units").(int)
	sku := d.Get("sku").(string)
	fileCopyEnabled := d.Get("file_copy_enabled").(bool)
	ipConnectEnabled := d.Get("ip_connect_enabled").(bool)
	shareableLinkEnabled := d.Get("shareable_link_enabled").(bool)
	tunnelingEnabled := d.Get("tunneling_enabled").(bool)

	if scaleUnits > 2 && sku == string(network.BastionHostSkuNameBasic) {
		return fmt.Errorf("`scale_units` only can be changed when `sku` is `Standard`. `scale_units` is always `2` when `sku` is `Basic`")
	}

	if fileCopyEnabled && sku == string(network.BastionHostSkuNameBasic) {
		return fmt.Errorf("`file_copy_enabled` is only supported when `sku` is `Standard`")
	}

	if ipConnectEnabled && sku == string(network.BastionHostSkuNameBasic) {
		return fmt.Errorf("`ip_connect_enabled` is only supported when `sku` is `Standard`")
	}

	if shareableLinkEnabled && sku == string(network.BastionHostSkuNameBasic) {
		return fmt.Errorf("`shareable_link_enabled` is only supported when `sku` is `Standard`")
	}

	if tunnelingEnabled && sku == string(network.BastionHostSkuNameBasic) {
		return fmt.Errorf("`tunneling_enabled` is only supported when `sku` is `Standard`")
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_bastion_host", id.ID())
		}
	}

	parameters := network.BastionHost{
		Location: &location,
		BastionHostPropertiesFormat: &network.BastionHostPropertiesFormat{
			DisableCopyPaste:    utils.Bool(!d.Get("copy_paste_enabled").(bool)),
			EnableFileCopy:      utils.Bool(fileCopyEnabled),
			EnableIPConnect:     utils.Bool(ipConnectEnabled),
			EnableShareableLink: utils.Bool(shareableLinkEnabled),
			EnableTunneling:     utils.Bool(tunnelingEnabled),
			IPConfigurations:    expandBastionHostIPConfiguration(d.Get("ip_configuration").([]interface{})),
			ScaleUnits:          utils.Int32(int32(d.Get("scale_units").(int))),
		},
		Sku: &network.Sku{
			Name: network.BastionHostSkuName(sku),
		},
		Tags: tags.Expand(t),
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation/update of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceBastionHostRead(d, meta)
}

func resourceBastionHostRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.BastionHostsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BastionHostID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			log.Printf("[DEBUG] %s was not found - removing from state!", *id)
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if sku := resp.Sku; sku != nil {
		d.Set("sku", sku.Name)
	}

	if props := resp.BastionHostPropertiesFormat; props != nil {
		d.Set("dns_name", props.DNSName)
		d.Set("scale_units", props.ScaleUnits)
		d.Set("file_copy_enabled", props.EnableFileCopy)
		d.Set("ip_connect_enabled", props.EnableIPConnect)
		d.Set("shareable_link_enabled", props.EnableShareableLink)
		d.Set("tunneling_enabled", props.EnableTunneling)

		copyPasteEnabled := true
		if props.DisableCopyPaste != nil {
			copyPasteEnabled = !*props.DisableCopyPaste
		}
		d.Set("copy_paste_enabled", copyPasteEnabled)

		if ipConfigs := props.IPConfigurations; ipConfigs != nil {
			if err := d.Set("ip_configuration", flattenBastionHostIPConfiguration(ipConfigs)); err != nil {
				return fmt.Errorf("flattening `ip_configuration`: %+v", err)
			}
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceBastionHostDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.BastionHostsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BastionHostID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
		}
	}

	return nil
}

func expandBastionHostIPConfiguration(input []interface{}) (ipConfigs *[]network.BastionHostIPConfiguration) {
	if len(input) == 0 {
		return nil
	}

	property := input[0].(map[string]interface{})
	ipConfName := property["name"].(string)
	subID := property["subnet_id"].(string)
	pipID := property["public_ip_address_id"].(string)

	return &[]network.BastionHostIPConfiguration{
		{
			Name: &ipConfName,
			BastionHostIPConfigurationPropertiesFormat: &network.BastionHostIPConfigurationPropertiesFormat{
				Subnet: &network.SubResource{
					ID: &subID,
				},
				PublicIPAddress: &network.SubResource{
					ID: &pipID,
				},
			},
		},
	}
}

func flattenBastionHostIPConfiguration(ipConfigs *[]network.BastionHostIPConfiguration) []interface{} {
	result := make([]interface{}, 0)
	if ipConfigs == nil {
		return result
	}

	for _, config := range *ipConfigs {
		ipConfig := make(map[string]interface{})

		if config.Name != nil {
			ipConfig["name"] = *config.Name
		}

		if props := config.BastionHostIPConfigurationPropertiesFormat; props != nil {
			if subnet := props.Subnet; subnet != nil {
				ipConfig["subnet_id"] = *subnet.ID
			}

			if pip := props.PublicIPAddress; pip != nil {
				ipConfig["public_ip_address_id"] = *pip.ID
			}
		}

		result = append(result, ipConfig)
	}
	return result
}
