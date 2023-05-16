package loadbalancer

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/services/loadbalancer/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/loadbalancer/validate"
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

var (
	_ sdk.Resource           = BackendAddressPoolAddressResource{}
	_ sdk.ResourceWithUpdate = BackendAddressPoolAddressResource{}
)

type BackendAddressPoolAddressResource struct{}

type BackendAddressPoolAddressModel struct {
	Name                 string `tfschema:"name"`
	BackendAddressPoolId string `tfschema:"backend_address_pool_id"`
	VirtualNetworkId     string `tfschema:"virtual_network_id"`
	IPAddress            string `tfschema:"ip_address"`
}

func (r BackendAddressPoolAddressResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"backend_address_pool_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.LoadBalancerBackendAddressPoolID,
		},

		"virtual_network_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: networkValidate.VirtualNetworkID,
		},

		"ip_address": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validation.IsIPAddress,
		},
	}
}

func (r BackendAddressPoolAddressResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r BackendAddressPoolAddressResource) ModelObject() interface{} {
	return &BackendAddressPoolAddressModel{}
}

func (r BackendAddressPoolAddressResource) ResourceType() string {
	return "azurerm_lb_backend_address_pool_address"
}

func (r BackendAddressPoolAddressResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.LoadBalancers.LoadBalancerBackendAddressPoolsClient
			subscriptionId := metadata.Client.Account.SubscriptionId

			var model BackendAddressPoolAddressModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			poolId, err := parse.LoadBalancerBackendAddressPoolID(model.BackendAddressPoolId)
			if err != nil {
				return err
			}

			locks.ByName(poolId.BackendAddressPoolName, backendAddressPoolResourceName)
			defer locks.UnlockByName(poolId.BackendAddressPoolName, backendAddressPoolResourceName)

			// Backend Addresses can not be created for Basic sku, so we have to check
			lb, err := metadata.Client.LoadBalancers.LoadBalancersClient.Get(ctx, poolId.ResourceGroup, poolId.LoadBalancerName, "")
			if err != nil {
				return fmt.Errorf("retrieving Load Balancer %q (Resource Group %q): %+v", poolId.LoadBalancerName, poolId.ResourceGroup, err)
			}
			isBasicSku := true
			if lb.Sku != nil && lb.Sku.Name != network.LoadBalancerSkuNameBasic {
				isBasicSku = false
			}
			if isBasicSku {
				return fmt.Errorf("Backend Addresses are not supported on Basic SKU Load Balancers")
			}

			id := parse.NewBackendAddressPoolAddressID(subscriptionId, poolId.ResourceGroup, poolId.LoadBalancerName, poolId.BackendAddressPoolName, model.Name)
			pool, err := client.Get(ctx, poolId.ResourceGroup, poolId.LoadBalancerName, poolId.BackendAddressPoolName)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", *poolId, err)
			}
			if pool.BackendAddressPoolPropertiesFormat == nil {
				return fmt.Errorf("retrieving %s: `properties` was nil", *poolId)
			}

			addresses := make([]network.LoadBalancerBackendAddress, 0)
			if pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses != nil {
				addresses = *pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses
			}

			metadata.Logger.Infof("checking for existing %s..", id)
			for _, address := range addresses {
				if address.Name == nil {
					continue
				}

				if *address.Name == id.AddressName {
					return metadata.ResourceRequiresImport(r.ResourceType(), id)
				}
			}

			addresses = append(addresses, network.LoadBalancerBackendAddress{
				LoadBalancerBackendAddressPropertiesFormat: &network.LoadBalancerBackendAddressPropertiesFormat{
					IPAddress: utils.String(model.IPAddress),
					VirtualNetwork: &network.SubResource{
						ID: utils.String(model.VirtualNetworkId),
					},
				},
				Name: utils.String(id.AddressName),
			})
			pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses = &addresses

			metadata.Logger.Infof("adding %s..", id)
			future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName, pool)
			if err != nil {
				return fmt.Errorf("updating %s: %+v", id, err)
			}
			metadata.Logger.Infof("waiting for update %s..", id)
			if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for update of %s: %+v", id, err)
			}
			metadata.SetID(id)
			return nil
		},
		Timeout: 30 * time.Minute,
	}
}

func (r BackendAddressPoolAddressResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.LoadBalancers.LoadBalancerBackendAddressPoolsClient
			id, err := parse.BackendAddressPoolAddressID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			pool, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}
			if pool.BackendAddressPoolPropertiesFormat == nil {
				return fmt.Errorf("retrieving %s: `properties` was nil", *id)
			}

			var backendAddress *network.LoadBalancerBackendAddress
			if pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses != nil {
				for _, address := range *pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses {
					if address.Name == nil {
						continue
					}

					if *address.Name == id.AddressName {
						backendAddress = &address
						break
					}
				}
			}
			if backendAddress == nil {
				return metadata.MarkAsGone(id)
			}

			backendAddressPoolId := parse.NewLoadBalancerBackendAddressPoolID(id.SubscriptionId, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName)
			model := BackendAddressPoolAddressModel{
				Name:                 id.AddressName,
				BackendAddressPoolId: backendAddressPoolId.ID(),
			}

			if props := backendAddress.LoadBalancerBackendAddressPropertiesFormat; props != nil {
				if props.IPAddress != nil {
					model.IPAddress = *props.IPAddress
				}

				if props.VirtualNetwork != nil && props.VirtualNetwork.ID != nil {
					model.VirtualNetworkId = *props.VirtualNetwork.ID
				}
			}

			return metadata.Encode(&model)
		},
		Timeout: 5 * time.Minute,
	}
}

func (r BackendAddressPoolAddressResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.LoadBalancers.LoadBalancerBackendAddressPoolsClient
			id, err := parse.BackendAddressPoolAddressID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			locks.ByName(id.BackendAddressPoolName, backendAddressPoolResourceName)
			defer locks.UnlockByName(id.BackendAddressPoolName, backendAddressPoolResourceName)

			pool, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}
			if pool.BackendAddressPoolPropertiesFormat == nil {
				return fmt.Errorf("retrieving %s: `properties` was nil", *id)
			}

			addresses := make([]network.LoadBalancerBackendAddress, 0)
			if pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses != nil {
				addresses = *pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses
			}

			newAddresses := make([]network.LoadBalancerBackendAddress, 0)
			for _, address := range addresses {
				if address.Name == nil {
					continue
				}

				if *address.Name != id.AddressName {
					newAddresses = append(newAddresses, address)
				}
			}
			pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses = &newAddresses

			metadata.Logger.Infof("removing %s..", *id)
			future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName, pool)
			if err != nil {
				return fmt.Errorf("removing %s: %+v", *id, err)
			}
			if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for removal of %s: %+v", *id, err)
			}
			return nil
		},
		Timeout: 30 * time.Minute,
	}
}

func (r BackendAddressPoolAddressResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return validate.BackendAddressPoolAddressID
}

func (r BackendAddressPoolAddressResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.LoadBalancers.LoadBalancerBackendAddressPoolsClient
			id, err := parse.BackendAddressPoolAddressID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			locks.ByName(id.BackendAddressPoolName, backendAddressPoolResourceName)
			defer locks.UnlockByName(id.BackendAddressPoolName, backendAddressPoolResourceName)

			var model BackendAddressPoolAddressModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			pool, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}
			if pool.BackendAddressPoolPropertiesFormat == nil {
				return fmt.Errorf("retrieving %s: `properties` was nil", *id)
			}

			addresses := make([]network.LoadBalancerBackendAddress, 0)
			if pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses != nil {
				addresses = *pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses
			}
			index := -1
			for i, address := range addresses {
				if address.Name == nil {
					continue
				}

				if *address.Name == id.AddressName {
					index = i
					break
				}
			}
			if index == -1 {
				return fmt.Errorf("%s was not found", *id)
			}

			addresses[index] = network.LoadBalancerBackendAddress{
				LoadBalancerBackendAddressPropertiesFormat: &network.LoadBalancerBackendAddressPropertiesFormat{
					IPAddress: utils.String(model.IPAddress),
					VirtualNetwork: &network.SubResource{
						ID: utils.String(model.VirtualNetworkId),
					},
				},
				Name: utils.String(id.AddressName),
			}
			pool.BackendAddressPoolPropertiesFormat.LoadBalancerBackendAddresses = &addresses

			future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName, pool)
			if err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}
			if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for update of %s: %+v", *id, err)
			}
			return nil
		},
		Timeout: 30 * time.Minute,
	}
}
