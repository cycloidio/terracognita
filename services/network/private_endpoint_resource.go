package network

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/services/privatedns/sdk/2018-09-01/privatezones"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	cosmosParse "github.com/hashicorp/terraform-provider-azurerm/services/cosmos/parse"
	mariaDBParse "github.com/hashicorp/terraform-provider-azurerm/services/mariadb/parse"
	mysqlParse "github.com/hashicorp/terraform-provider-azurerm/services/mysql/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	postgresqlParse "github.com/hashicorp/terraform-provider-azurerm/services/postgres/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/signalr/sdk/2020-05-01/signalr"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourcePrivateEndpoint() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourcePrivateEndpointCreate,
		Read:   resourcePrivateEndpointRead,
		Update: resourcePrivateEndpointUpdate,
		Delete: resourcePrivateEndpointDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.PrivateEndpointID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(60 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(60 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.PrivateLinkName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"subnet_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"network_interface": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"private_dns_zone_group": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.PrivateLinkName,
						},
						"private_dns_zone_ids": {
							Type:     pluginsdk.TypeList,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: privatezones.ValidatePrivateDnsZoneID,
							},
						},
					},
				},
			},

			"private_service_connection": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.PrivateLinkName,
						},
						"is_manual_connection": {
							Type:     pluginsdk.TypeBool,
							Required: true,
							ForceNew: true,
						},
						"private_connection_resource_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: azure.ValidateResourceID,
							ExactlyOneOf: []string{"private_service_connection.0.private_connection_resource_alias", "private_service_connection.0.private_connection_resource_id"},
						},
						"private_connection_resource_alias": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validate.PrivateConnectionResourceAlias,
							ExactlyOneOf: []string{"private_service_connection.0.private_connection_resource_alias", "private_service_connection.0.private_connection_resource_id"},
						},
						"subresource_names": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validate.PrivateLinkSubResourceName,
							},
						},
						"request_message": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 140),
						},
						"private_ip_address": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"custom_dns_configs": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"fqdn": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"ip_addresses": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},
					},
				},
			},

			"private_dns_zone_configs": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"private_dns_zone_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"record_sets": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"type": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"fqdn": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"ttl": {
										Type:     pluginsdk.TypeInt,
										Computed: true,
									},
									"ip_addresses": {
										Type:     pluginsdk.TypeList,
										Computed: true,
										Elem: &pluginsdk.Schema{
											Type: pluginsdk.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func resourcePrivateEndpointCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PrivateEndpointClient
	dnsClient := meta.(*clients.Client).Network.PrivateDnsZoneGroupClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewPrivateEndpointID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if err := validatePrivateEndpointSettings(d); err != nil {
		return fmt.Errorf("validating the configuration for the Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	if existing.PrivateEndpointProperties != nil {
		return tf.ImportAsExistsError("azurerm_private_endpoint", id.ID())
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	privateDnsZoneGroup := d.Get("private_dns_zone_group").([]interface{})
	privateServiceConnections := d.Get("private_service_connection").([]interface{})
	subnetId := d.Get("subnet_id").(string)

	parameters := network.PrivateEndpoint{
		Location: utils.String(location),
		PrivateEndpointProperties: &network.PrivateEndpointProperties{
			PrivateLinkServiceConnections:       expandPrivateLinkEndpointServiceConnection(privateServiceConnections, false),
			ManualPrivateLinkServiceConnections: expandPrivateLinkEndpointServiceConnection(privateServiceConnections, true),
			Subnet: &network.Subnet{
				ID: utils.String(subnetId),
			},
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	cosmosDbResIds := getCosmosDbResIdInPrivateServiceConnections(parameters.PrivateEndpointProperties)
	for _, cosmosDbResId := range cosmosDbResIds {
		log.Printf("[DEBUG] Add Lock For Private Endpoint %q, lock name: %q", id.Name, cosmosDbResId)
		locks.ByName(cosmosDbResId, "azurerm_private_endpoint")
		//goland:noinspection GoDeferInLoop
		defer locks.UnlockByName(cosmosDbResId, "azurerm_private_endpoint")
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		if strings.EqualFold(err.Error(), "is missing required parameter 'group Id'") {
			return fmt.Errorf("creating Private Endpoint %q (Resource Group %q) due to missing 'group Id', ensure that the 'subresource_names' type is populated: %+v", id.Name, id.ResourceGroup, err)
		} else {
			return fmt.Errorf("creating Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.SetId(id.ID())

	// 1 Private Endpoint can have 1 Private DNS Zone Group
	// since this is a new resource, there shouldn't be an existing one - so there's no need to delete it
	if len(privateDnsZoneGroup) > 0 {
		log.Printf("[DEBUG] Creating Private DNS Zone Group associated with Private Endpoint %q / Resource Group %q..", id.Name, id.ResourceGroup)
		if err := createPrivateDnsZoneGroupForPrivateEndpoint(ctx, dnsClient, id, privateDnsZoneGroup); err != nil {
			return err
		}
		log.Printf("[DEBUG] Created the Existing Private DNS Zone Group associated with Private Endpoint %q / Resource Group %q.", id.Name, id.ResourceGroup)
	}

	return resourcePrivateEndpointRead(d, meta)
}

func getCosmosDbResIdInPrivateServiceConnections(p *network.PrivateEndpointProperties) []string {
	var ids []string
	exists := make(map[string]struct{})

	for _, l := range *p.PrivateLinkServiceConnections {
		id := *l.PrivateLinkServiceID
		if _, err := cosmosParse.DatabaseAccountID(id); err == nil {
			_, ok := exists[id]
			if !ok {
				ids = append(ids, id)
				exists[id] = struct{}{}
			}
		}
	}
	for _, l := range *p.ManualPrivateLinkServiceConnections {
		id := *l.PrivateLinkServiceID
		if _, err := cosmosParse.DatabaseAccountID(id); err == nil {
			_, ok := exists[id]
			if !ok {
				ids = append(ids, id)
				exists[id] = struct{}{}
			}
		}
	}
	// Sort ids, force adding lock in consistent order to avoid potential deadlock
	sort.Strings(ids)
	return ids
}

func resourcePrivateEndpointUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PrivateEndpointClient
	dnsClient := meta.(*clients.Client).Network.PrivateDnsZoneGroupClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.PrivateEndpointID(d.Id())
	if err != nil {
		return err
	}

	if err := validatePrivateEndpointSettings(d); err != nil {
		return fmt.Errorf("validating the configuration for the Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	privateDnsZoneGroup := d.Get("private_dns_zone_group").([]interface{})
	privateServiceConnections := d.Get("private_service_connection").([]interface{})
	subnetId := d.Get("subnet_id").(string)

	// TODO: in future it'd be nice to support conditional updates here, but one problem at a time
	parameters := network.PrivateEndpoint{
		Location: utils.String(location),
		PrivateEndpointProperties: &network.PrivateEndpointProperties{
			PrivateLinkServiceConnections:       expandPrivateLinkEndpointServiceConnection(privateServiceConnections, false),
			ManualPrivateLinkServiceConnections: expandPrivateLinkEndpointServiceConnection(privateServiceConnections, true),
			Subnet: &network.Subnet{
				ID: utils.String(subnetId),
			},
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		if strings.EqualFold(err.Error(), "is missing required parameter 'group Id'") {
			return fmt.Errorf("updating Private Endpoint %q (Resource Group %q) due to missing 'group Id', ensure that the 'subresource_names' type is populated: %+v", id.Name, id.ResourceGroup, err)
		} else {
			return fmt.Errorf("updating Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	// 1 Private Endpoint can have 1 Private DNS Zone Group - so to update we need to Delete & Recreate
	if d.HasChange("private_dns_zone_group") {
		existingDnsZoneGroups, err := retrievePrivateDnsZoneGroupsForPrivateEndpoint(ctx, dnsClient, *id)
		if err != nil {
			return err
		}

		newDnsZoneGroups := d.Get("private_dns_zone_group").([]interface{})
		newDnsZoneName := ""
		if len(newDnsZoneGroups) > 0 {
			groupRaw := newDnsZoneGroups[0].(map[string]interface{})
			newDnsZoneName = groupRaw["name"].(string)
		}

		needToRemove := newDnsZoneName == ""
		nameHasChanged := false
		if existingDnsZoneGroups != nil && newDnsZoneName != "" {
			needToRemove = len(*existingDnsZoneGroups) > 0 && len(newDnsZoneGroups) == 0

			// there should only be a single one, but there's no harm checking all returned
			for _, existing := range *existingDnsZoneGroups {
				if existing.Name != newDnsZoneName {
					nameHasChanged = true
					break
				}
			}
		}

		if needToRemove || nameHasChanged {
			log.Printf("[DEBUG] Deleting the Existing Private DNS Zone Group associated with Private Endpoint %q / Resource Group %q..", id.Name, id.ResourceGroup)
			if err := deletePrivateDnsZoneGroupForPrivateEndpoint(ctx, dnsClient, *id); err != nil {
				return err
			}
			log.Printf("[DEBUG] Deleted the Existing Private DNS Zone Group associated with Private Endpoint %q / Resource Group %q.", id.Name, id.ResourceGroup)
		}

		if len(privateDnsZoneGroup) > 0 {
			log.Printf("[DEBUG] Creating Private DNS Zone Group associated with Private Endpoint %q / Resource Group %q..", id.Name, id.ResourceGroup)
			if err := createPrivateDnsZoneGroupForPrivateEndpoint(ctx, dnsClient, *id, privateDnsZoneGroup); err != nil {
				return err
			}
			log.Printf("[DEBUG] Created the Existing Private DNS Zone Group associated with Private Endpoint %q / Resource Group %q.", id.Name, id.ResourceGroup)
		}
	}

	return resourcePrivateEndpointRead(d, meta)
}

func resourcePrivateEndpointRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PrivateEndpointClient
	nicsClient := meta.(*clients.Client).Network.InterfacesClient
	dnsClient := meta.(*clients.Client).Network.PrivateDnsZoneGroupClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.PrivateEndpointID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Private Endpoint %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	privateDnsZoneIds, err := retrievePrivateDnsZoneGroupsForPrivateEndpoint(ctx, dnsClient, *id)
	if err != nil {
		return err
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := resp.PrivateEndpointProperties; props != nil {
		if err := d.Set("custom_dns_configs", flattenCustomDnsConfigs(props.CustomDNSConfigs)); err != nil {
			return fmt.Errorf("setting `custom_dns_configs`: %+v", err)
		}

		networkInterfaceId := ""
		privateIpAddress := ""
		if nics := props.NetworkInterfaces; nics != nil && len(*nics) > 0 {
			nic := (*nics)[0]
			if nic.ID != nil && *nic.ID != "" {
				networkInterfaceId = *nic.ID
				privateIpAddress = getPrivateIpAddress(ctx, nicsClient, networkInterfaceId)
			}
		}

		networkInterface := flattenNetworkInterface(networkInterfaceId)
		if err := d.Set("network_interface", networkInterface); err != nil {
			return fmt.Errorf("setting `network_interface`: %+v", err)
		}

		flattenedConnection := flattenPrivateLinkEndpointServiceConnection(props.PrivateLinkServiceConnections, props.ManualPrivateLinkServiceConnections, privateIpAddress)
		if err := d.Set("private_service_connection", flattenedConnection); err != nil {
			return fmt.Errorf("setting `private_service_connection`: %+v", err)
		}

		subnetId := ""
		if props.Subnet != nil && props.Subnet.ID != nil {
			subnetId = *props.Subnet.ID
		}
		d.Set("subnet_id", subnetId)
	}

	privateDnsZoneConfigs := make([]interface{}, 0)
	privateDnsZoneGroups := make([]interface{}, 0)
	if privateDnsZoneIds != nil {
		for _, dnsZoneId := range *privateDnsZoneIds {
			flattened, err := retrieveAndFlattenPrivateDnsZone(ctx, dnsClient, dnsZoneId)
			if err != nil {
				return nil
			}

			// an exceptional case but no harm in handling
			if flattened == nil {
				continue
			}

			privateDnsZoneConfigs = append(privateDnsZoneConfigs, flattened.DnsZoneConfig...)
			privateDnsZoneGroups = append(privateDnsZoneGroups, flattened.DnsZoneGroup)
		}
	}
	if err := d.Set("private_dns_zone_configs", privateDnsZoneConfigs); err != nil {
		return fmt.Errorf("setting `private_dns_zone_configs`: %+v", err)
	}
	if err := d.Set("private_dns_zone_group", privateDnsZoneGroups); err != nil {
		return fmt.Errorf("setting `private_dns_zone_group`: %+v", err)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourcePrivateEndpointDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PrivateEndpointClient
	dnsZoneGroupsClient := meta.(*clients.Client).Network.PrivateDnsZoneGroupClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.PrivateEndpointID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting the Private DNS Zone Group associated with Private Endpoint %q / Resource Group %q..", id.Name, id.ResourceGroup)
	if err := deletePrivateDnsZoneGroupForPrivateEndpoint(ctx, dnsZoneGroupsClient, *id); err != nil {
		return err
	}
	log.Printf("[DEBUG] Deleted the Private DNS Zone Group associated with Private Endpoint %q / Resource Group %q.", id.Name, id.ResourceGroup)

	privateServiceConnections := d.Get("private_service_connection").([]interface{})
	parameters := network.PrivateEndpoint{
		PrivateEndpointProperties: &network.PrivateEndpointProperties{
			PrivateLinkServiceConnections:       expandPrivateLinkEndpointServiceConnection(privateServiceConnections, false),
			ManualPrivateLinkServiceConnections: expandPrivateLinkEndpointServiceConnection(privateServiceConnections, true),
		},
	}
	cosmosDbResIds := getCosmosDbResIdInPrivateServiceConnections(parameters.PrivateEndpointProperties)
	for _, cosmosDbResId := range cosmosDbResIds {
		locks.ByName(cosmosDbResId, "azurerm_private_endpoint")
		//goland:noinspection GoDeferInLoop
		defer locks.UnlockByName(cosmosDbResId, "azurerm_private_endpoint")
	}

	log.Printf("[DEBUG] Deleting the Private Endpoint %q / Resource Group %q..", id.Name, id.ResourceGroup)
	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deletion of Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}
	log.Printf("[DEBUG] Deleted the Private Endpoint %q / Resource Group %q.", id.Name, id.ResourceGroup)

	return nil
}

func expandPrivateLinkEndpointServiceConnection(input []interface{}, parseManual bool) *[]network.PrivateLinkServiceConnection {
	results := make([]network.PrivateLinkServiceConnection, 0)

	for _, item := range input {
		v := item.(map[string]interface{})
		privateConnectionResourceId := v["private_connection_resource_id"].(string)
		if privateConnectionResourceId == "" {
			privateConnectionResourceId = v["private_connection_resource_alias"].(string)
		}
		subresourceNames := v["subresource_names"].([]interface{})
		requestMessage := v["request_message"].(string)
		isManual := v["is_manual_connection"].(bool)
		name := v["name"].(string)

		if isManual == parseManual {
			result := network.PrivateLinkServiceConnection{
				Name: utils.String(name),
				PrivateLinkServiceConnectionProperties: &network.PrivateLinkServiceConnectionProperties{
					GroupIds:             utils.ExpandStringSlice(subresourceNames),
					PrivateLinkServiceID: utils.String(privateConnectionResourceId),
				},
			}

			if requestMessage != "" {
				result.PrivateLinkServiceConnectionProperties.RequestMessage = utils.String(requestMessage)
			}

			results = append(results, result)
		}
	}

	return &results
}

func flattenCustomDnsConfigs(customDnsConfigs *[]network.CustomDNSConfigPropertiesFormat) []interface{} {
	results := make([]interface{}, 0)
	if customDnsConfigs == nil {
		return results
	}

	for _, item := range *customDnsConfigs {
		results = append(results, map[string]interface{}{
			"fqdn":         item.Fqdn,
			"ip_addresses": utils.FlattenStringSlice(item.IPAddresses),
		})
	}

	return results
}

func flattenPrivateLinkEndpointServiceConnection(serviceConnections *[]network.PrivateLinkServiceConnection, manualServiceConnections *[]network.PrivateLinkServiceConnection, privateIPAddress string) []interface{} {
	results := make([]interface{}, 0)
	if serviceConnections == nil && manualServiceConnections == nil {
		return results
	}

	if serviceConnections != nil {
		for _, item := range *serviceConnections {
			name := ""
			if item.Name != nil {
				name = *item.Name
			}

			privateConnectionId := ""
			subResourceNames := make([]interface{}, 0)

			if props := item.PrivateLinkServiceConnectionProperties; props != nil {
				if v := props.GroupIds; v != nil {
					subResourceNames = utils.FlattenStringSlice(v)
				}
				if props.PrivateLinkServiceID != nil {
					privateConnectionId = *props.PrivateLinkServiceID
				}
			}
			attrs := map[string]interface{}{
				"name":                 name,
				"is_manual_connection": false,
				"private_ip_address":   privateIPAddress,
				"subresource_names":    subResourceNames,
			}
			if strings.HasSuffix(privateConnectionId, ".azure.privatelinkservice") {
				attrs["private_connection_resource_alias"] = privateConnectionId
			} else {
				// There is a bug from service, the PE created from portal could be with the connection id for postgresql server "Microsoft.DBForPostgreSQL" instead of "Microsoft.DBforPostgreSQL"
				// and for Mysql and MariaDB
				if strings.Contains(strings.ToLower(privateConnectionId), "microsoft.dbforpostgresql") {
					if serverId, err := postgresqlParse.ServerID(privateConnectionId); err == nil {
						privateConnectionId = serverId.ID()
					}
				}
				if strings.Contains(strings.ToLower(privateConnectionId), "microsoft.dbformysql") {
					if serverId, err := mysqlParse.ServerID(privateConnectionId); err == nil {
						privateConnectionId = serverId.ID()
					}
				}
				if strings.Contains(strings.ToLower(privateConnectionId), "microsoft.dbformariadb") {
					if serverId, err := mariaDBParse.ServerID(privateConnectionId); err == nil {
						privateConnectionId = serverId.ID()
					}
				}
				if strings.Contains(strings.ToLower(privateConnectionId), "microsoft.signalrservice") {
					if serviceId, err := signalr.ParseSignalRIDInsensitively(privateConnectionId); err == nil {
						privateConnectionId = serviceId.ID()
					}
				}
				attrs["private_connection_resource_id"] = privateConnectionId
			}

			results = append(results, attrs)
		}
	}

	if manualServiceConnections != nil {
		for _, item := range *manualServiceConnections {
			name := ""
			if item.Name != nil {
				name = *item.Name
			}

			privateConnectionId := ""
			requestMessage := ""
			subResourceNames := make([]interface{}, 0)

			if props := item.PrivateLinkServiceConnectionProperties; props != nil {
				if v := props.GroupIds; v != nil {
					subResourceNames = utils.FlattenStringSlice(v)
				}
				if props.PrivateLinkServiceID != nil {
					privateConnectionId = *props.PrivateLinkServiceID
				}
				if props.RequestMessage != nil {
					requestMessage = *props.RequestMessage
				}
			}

			attrs := map[string]interface{}{
				"name":                 name,
				"is_manual_connection": true,
				"private_ip_address":   privateIPAddress,
				"request_message":      requestMessage,
				"subresource_names":    subResourceNames,
			}
			if strings.HasSuffix(privateConnectionId, ".azure.privatelinkservice") {
				attrs["private_connection_resource_alias"] = privateConnectionId
			} else {
				// There is a bug from service, the PE created from portal could be with the connection id for postgresql server "Microsoft.DBForPostgreSQL" instead of "Microsoft.DBforPostgreSQL"
				// and for Mysql and MariaDB
				if strings.Contains(strings.ToLower(privateConnectionId), "microsoft.dbforpostgresql") {
					if serverId, err := postgresqlParse.ServerID(privateConnectionId); err == nil {
						privateConnectionId = serverId.ID()
					}
				}
				if strings.Contains(strings.ToLower(privateConnectionId), "microsoft.dbformysql") {
					if serverId, err := mysqlParse.ServerID(privateConnectionId); err == nil {
						privateConnectionId = serverId.ID()
					}
				}
				if strings.Contains(strings.ToLower(privateConnectionId), "microsoft.dbformariadb") {
					if serverId, err := mariaDBParse.ServerID(privateConnectionId); err == nil {
						privateConnectionId = serverId.ID()
					}
				}
				attrs["private_connection_resource_id"] = privateConnectionId
			}

			results = append(results, attrs)
		}
	}

	return results
}

func createPrivateDnsZoneGroupForPrivateEndpoint(ctx context.Context, client *network.PrivateDNSZoneGroupsClient, id parse.PrivateEndpointId, inputRaw []interface{}) error {
	if len(inputRaw) != 1 {
		return fmt.Errorf("expected a single Private DNS Zone Groups but got %d", len(inputRaw))
	}
	item := inputRaw[0].(map[string]interface{})

	dnsGroupName := item["name"].(string)
	privateDnsZoneIdsRaw := item["private_dns_zone_ids"].([]interface{})
	privateDnsZoneConfigs := make([]network.PrivateDNSZoneConfig, 0)
	for _, item := range privateDnsZoneIdsRaw {
		v := item.(string)

		privateDnsZone, err := privatezones.ParsePrivateDnsZoneID(v)
		if err != nil {
			return err
		}

		privateDnsZoneConfigs = append(privateDnsZoneConfigs, network.PrivateDNSZoneConfig{
			Name: utils.String(privateDnsZone.PrivateZoneName),
			PrivateDNSZonePropertiesFormat: &network.PrivateDNSZonePropertiesFormat{
				PrivateDNSZoneID: utils.String(privateDnsZone.ID()),
			},
		})
	}

	parameters := network.PrivateDNSZoneGroup{
		Name: utils.String(id.Name),
		PrivateDNSZoneGroupPropertiesFormat: &network.PrivateDNSZoneGroupPropertiesFormat{
			PrivateDNSZoneConfigs: &privateDnsZoneConfigs,
		},
	}
	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, dnsGroupName, parameters)
	if err != nil {
		return fmt.Errorf("creating Private DNS Zone Group %q for Private Endpoint %q (Resource Group %q): %+v", dnsGroupName, id.Name, id.ResourceGroup, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of Private DNS Zone Group %q for Private Endpoint %q (Resource Group %q): %+v", dnsGroupName, id.Name, id.ResourceGroup, err)
	}

	return nil
}

func deletePrivateDnsZoneGroupForPrivateEndpoint(ctx context.Context, client *network.PrivateDNSZoneGroupsClient, id parse.PrivateEndpointId) error {
	// lookup and delete the (should be, Single) Private DNS Zone Group associated with this Private Endpoint
	privateDnsZoneIds, err := retrievePrivateDnsZoneGroupsForPrivateEndpoint(ctx, client, id)
	if err != nil {
		return err
	}

	for _, privateDnsZoneId := range *privateDnsZoneIds {
		log.Printf("[DEBUG] Deleting Private DNS Zone Group %q (Private Endpoint %q / Resource Group %q)..", privateDnsZoneId.Name, privateDnsZoneId.PrivateEndpointName, privateDnsZoneId.ResourceGroup)
		future, err := client.Delete(ctx, privateDnsZoneId.ResourceGroup, privateDnsZoneId.PrivateEndpointName, privateDnsZoneId.Name)
		if err != nil {
			if !response.WasNotFound(future.Response()) {
				return fmt.Errorf("deleting Private DNS Zone Group %q (Private Endpoint %q / Resource Group %q): %+v", privateDnsZoneId.Name, privateDnsZoneId.PrivateEndpointName, privateDnsZoneId.ResourceGroup, err)
			}
		}

		if !response.WasNotFound(future.Response()) {
			log.Printf("[DEBUG] Waiting for deletion of Private DNS Zone Group %q (Private Endpoint %q / Resource Group %q)..", privateDnsZoneId.Name, privateDnsZoneId.PrivateEndpointName, privateDnsZoneId.ResourceGroup)
			if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
				if !response.WasNotFound(future.Response()) {
					return fmt.Errorf("waiting for deletion of Private DNS Zone Group %q (Private Endpoint %q / Resource Group %q): %+v", privateDnsZoneId.Name, privateDnsZoneId.PrivateEndpointName, privateDnsZoneId.ResourceGroup, err)
				}
			}
			log.Printf("[DEBUG] Deleted Private DNS Zone Group %q (Private Endpoint %q / Resource Group %q).", privateDnsZoneId.Name, privateDnsZoneId.PrivateEndpointName, privateDnsZoneId.ResourceGroup)
		}
	}

	return nil
}

func retrievePrivateDnsZoneGroupsForPrivateEndpoint(ctx context.Context, client *network.PrivateDNSZoneGroupsClient, id parse.PrivateEndpointId) (*[]parse.PrivateDnsZoneGroupId, error) {
	output := make([]parse.PrivateDnsZoneGroupId, 0)

	dnsZones, err := client.ListComplete(ctx, id.Name, id.ResourceGroup) // looks odd.. matches the SDK method
	if err != nil {
		if utils.ResponseWasNotFound(dnsZones.Response().Response) {
			return &output, nil
		}

		return nil, fmt.Errorf("retrieving Private DNS Zone Groups for Private Endpoint %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	for dnsZones.NotDone() {
		privateDnsZoneGroup := dnsZones.Value()
		if privateDnsZoneGroup.ID != nil {
			groupId, err := parse.PrivateDnsZoneGroupID(*privateDnsZoneGroup.ID)
			if err != nil {
				return nil, err
			}

			output = append(output, *groupId)
		}

		if err := dnsZones.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return &output, nil
}

type flattenedPrivateDnsZoneGroup struct {
	DnsZoneConfig []interface{}
	DnsZoneGroup  map[string]interface{}
}

func retrieveAndFlattenPrivateDnsZone(ctx context.Context, client *network.PrivateDNSZoneGroupsClient, id parse.PrivateDnsZoneGroupId) (*flattenedPrivateDnsZoneGroup, error) {
	resp, err := client.Get(ctx, id.ResourceGroup, id.PrivateEndpointName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil, nil
		}

		return nil, fmt.Errorf("retrieving Private DNS Zone Group %q (Private Endpoint %q / Resource Group %q): %+v", id.Name, id.PrivateEndpointName, id.ResourceGroup, err)
	}

	privateDnsZoneIds := make([]string, 0)
	dnsZoneConfigs := make([]interface{}, 0)

	if resp.PrivateDNSZoneGroupPropertiesFormat != nil && resp.PrivateDNSZoneGroupPropertiesFormat.PrivateDNSZoneConfigs != nil {
		for _, config := range *resp.PrivateDNSZoneGroupPropertiesFormat.PrivateDNSZoneConfigs {
			if config.Name == nil {
				// necessary to build up the ID
				continue
			}
			if config.PrivateDNSZonePropertiesFormat == nil || config.PrivateDNSZonePropertiesFormat.PrivateDNSZoneID == nil {
				// necessary for a bunch of other things
				continue
			}
			props := *config.PrivateDNSZonePropertiesFormat
			name := *config.Name
			privateDnsZoneId := *props.PrivateDNSZoneID

			privateDnsZoneIds = append(privateDnsZoneIds, privateDnsZoneId)

			recordSets := flattenPrivateDnsZoneGroupRecordSets(props.RecordSets)
			dnsZoneConfigs = append(dnsZoneConfigs, map[string]interface{}{
				"id":                  parse.NewPrivateDnsZoneConfigID(id.SubscriptionId, id.ResourceGroup, id.PrivateEndpointName, id.Name, name).ID(),
				"name":                name,
				"private_dns_zone_id": privateDnsZoneId,
				"record_sets":         recordSets,
			})
		}
	}

	return &flattenedPrivateDnsZoneGroup{
		DnsZoneConfig: dnsZoneConfigs,
		DnsZoneGroup: map[string]interface{}{
			"id":                   id.ID(),
			"name":                 id.Name,
			"private_dns_zone_ids": privateDnsZoneIds,
		},
	}, nil
}

func flattenPrivateDnsZoneGroupRecordSets(input *[]network.RecordSet) []interface{} {
	output := make([]interface{}, 0)
	if input == nil {
		return output
	}

	for _, v := range *input {
		fqdn := ""
		if v.Fqdn != nil {
			fqdn = *v.Fqdn
		}

		name := ""
		if v.RecordSetName != nil {
			name = *v.RecordSetName
		}

		recordType := ""
		if v.RecordType != nil {
			recordType = *v.RecordType
		}

		ttl := 0
		if v.TTL != nil {
			ttl = int(*v.TTL)
		}

		ipAddresses := make([]string, 0)
		if v.IPAddresses != nil {
			ipAddresses = *v.IPAddresses
		}

		output = append(output, map[string]interface{}{
			"fqdn":         fqdn,
			"ip_addresses": ipAddresses,
			"name":         name,
			"ttl":          ttl,
			"type":         recordType,
		})
	}

	return output
}

func validatePrivateEndpointSettings(d *pluginsdk.ResourceData) error {
	privateServiceConnections := d.Get("private_service_connection").([]interface{})

	for _, psc := range privateServiceConnections {
		privateServiceConnection := psc.(map[string]interface{})
		name := privateServiceConnection["name"].(string)

		// If this is not a manual connection and the message is set return an error since this does not make sense.
		if !privateServiceConnection["is_manual_connection"].(bool) && privateServiceConnection["request_message"].(string) != "" {
			return fmt.Errorf(`"private_service_connection":%q is invalid, the "request_message" attribute cannot be set if the "is_manual_connection" attribute is "false"`, name)
		}

		// If this is a manual connection and the message isn't set return an error.
		if privateServiceConnection["is_manual_connection"].(bool) && strings.TrimSpace(privateServiceConnection["request_message"].(string)) == "" {
			return fmt.Errorf(`"private_service_connection":%q is invalid, the "request_message" attribute must not be empty`, name)
		}
	}

	return nil
}
