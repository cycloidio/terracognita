package devtestlabs

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/devtestlabs/mgmt/2018-09-15/dtl"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/devtestlabs/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/devtestlabs/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceArmDevTestVirtualNetwork() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceArmDevTestVnetRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"lab_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.DevTestLabName(),
			},

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"unique_identifier": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"allowed_subnets": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"allow_public_ip": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"lab_subnet_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"resource_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"subnet_overrides": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"lab_subnet_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"resource_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"use_in_vm_creation_permission": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"use_public_ip_address_permission": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"virtual_network_pool_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceArmDevTestVnetRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DevTestLabs.VirtualNetworksClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewDevTestVirtualNetworkID(subscriptionId, d.Get("resource_group_name").(string), d.Get("lab_name").(string), d.Get("name").(string))

	resp, err := client.Get(ctx, id.ResourceGroup, id.LabName, id.VirtualNetworkName, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}

		return fmt.Errorf("making Read request on %s: %+v", id, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("API returns a nil/empty id on %s: %+v", id, err)
	}
	d.SetId(id.ID())

	if props := resp.VirtualNetworkProperties; props != nil {
		if as := props.AllowedSubnets; as != nil {
			if err := d.Set("allowed_subnets", flattenDevTestVirtualNetworkAllowedSubnets(as)); err != nil {
				return fmt.Errorf("setting `allowed_subnets`: %v", err)
			}
		}
		if so := props.SubnetOverrides; so != nil {
			if err := d.Set("subnet_overrides", flattenDevTestVirtualNetworkSubnetOverrides(so)); err != nil {
				return fmt.Errorf("setting `subnet_overrides`: %v", err)
			}
		}
		d.Set("unique_identifier", props.UniqueIdentifier)
	}
	return nil
}

func flattenDevTestVirtualNetworkAllowedSubnets(input *[]dtl.Subnet) []interface{} {
	result := make([]interface{}, 0)

	if input == nil {
		return result
	}

	for _, v := range *input {
		allowedSubnet := make(map[string]interface{})

		allowedSubnet["allow_public_ip"] = string(v.AllowPublicIP)

		if resourceID := v.ResourceID; resourceID != nil {
			allowedSubnet["resource_id"] = *resourceID
		}

		if labSubnetName := v.LabSubnetName; labSubnetName != nil {
			allowedSubnet["lab_subnet_name"] = *labSubnetName
		}

		result = append(result, allowedSubnet)
	}

	return result
}

func flattenDevTestVirtualNetworkSubnetOverrides(input *[]dtl.SubnetOverride) []interface{} {
	result := make([]interface{}, 0)

	if input == nil {
		return result
	}

	for _, v := range *input {
		subnetOverride := make(map[string]interface{})
		if v.LabSubnetName != nil {
			subnetOverride["lab_subnet_name"] = *v.LabSubnetName
		}
		if v.ResourceID != nil {
			subnetOverride["resource_id"] = *v.ResourceID
		}

		subnetOverride["use_public_ip_address_permission"] = string(v.UsePublicIPAddressPermission)
		subnetOverride["use_in_vm_creation_permission"] = string(v.UseInVMCreationPermission)

		if v.VirtualNetworkPoolName != nil {
			subnetOverride["virtual_network_pool_name"] = *v.VirtualNetworkPoolName
		}

		result = append(result, subnetOverride)
	}

	return result
}
