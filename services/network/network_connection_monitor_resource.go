package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	computeValidate "github.com/hashicorp/terraform-provider-azurerm/services/compute/validate"
	logAnalyticsValidate "github.com/hashicorp/terraform-provider-azurerm/services/loganalytics/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceNetworkConnectionMonitor() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceNetworkConnectionMonitorCreateUpdate,
		Read:   resourceNetworkConnectionMonitorRead,
		Update: resourceNetworkConnectionMonitorCreateUpdate,
		Delete: resourceNetworkConnectionMonitorDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ConnectionMonitorID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: resourceNetworkConnectionMonitorSchema(),
	}
}

func resourceNetworkConnectionMonitorSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"network_watcher_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: networkValidate.NetworkWatcherID,
		},

		"location": azure.SchemaLocation(),

		"endpoint": {
			Type:     pluginsdk.TypeSet,
			Required: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"address": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						ValidateFunc: validation.Any(
							validation.IsIPv4Address,
							networkValidate.NetworkConnectionMonitorEndpointAddress,
						),
					},

					"coverage_level": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							string(network.CoverageLevelAboveAverage),
							string(network.CoverageLevelAverage),
							string(network.CoverageLevelBelowAverage),
							string(network.CoverageLevelDefault),
							string(network.CoverageLevelFull),
							string(network.CoverageLevelLow),
						}, false),
					},

					"excluded_ip_addresses": {
						Type:     pluginsdk.TypeSet,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
							ValidateFunc: validation.Any(
								validation.IsIPv4Address,
								validation.IsIPv6Address,
								validation.IsCIDR,
							),
						},
					},

					"filter": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"item": {
									Type:     pluginsdk.TypeSet,
									Optional: true,
									Elem: &pluginsdk.Resource{
										Schema: map[string]*pluginsdk.Schema{
											"address": {
												Type:         pluginsdk.TypeString,
												Optional:     true,
												ValidateFunc: azure.ValidateResourceID,
											},

											"type": {
												Type:     pluginsdk.TypeString,
												Optional: true,
												Default:  string(network.ConnectionMonitorEndpointFilterItemTypeAgentAddress),
												ValidateFunc: validation.StringInSlice([]string{
													string(network.ConnectionMonitorEndpointFilterItemTypeAgentAddress),
												}, false),
											},
										},
									},
								},

								"type": {
									Type:     pluginsdk.TypeString,
									Optional: true,
									Default:  string(network.ConnectionMonitorEndpointFilterTypeInclude),
									ValidateFunc: validation.StringInSlice([]string{
										string(network.ConnectionMonitorEndpointFilterTypeInclude),
									}, false),
								},
							},
						},
					},

					"included_ip_addresses": {
						Type:     pluginsdk.TypeSet,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
							ValidateFunc: validation.Any(
								validation.IsIPv4Address,
								validation.IsIPv6Address,
								validation.IsCIDR,
							),
						},
					},

					"target_resource_id": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						Computed: true,
						ValidateFunc: validation.Any(
							computeValidate.VirtualMachineID,
							logAnalyticsValidate.LogAnalyticsWorkspaceID,
							networkValidate.SubnetID,
							networkValidate.VirtualNetworkID,
						),
					},

					"target_resource_type": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							string(network.EndpointTypeAzureSubnet),
							string(network.EndpointTypeAzureVM),
							string(network.EndpointTypeAzureVNet),
							string(network.EndpointTypeExternalAddress),
							string(network.EndpointTypeMMAWorkspaceMachine),
							string(network.EndpointTypeMMAWorkspaceNetwork),
						}, false),
					},
				},
			},
		},

		"test_configuration": {
			Type:     pluginsdk.TypeSet,
			Required: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"protocol": {
						Type:     pluginsdk.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							string(network.ConnectionMonitorTestConfigurationProtocolTCP),
							string(network.ConnectionMonitorTestConfigurationProtocolHTTP),
							string(network.ConnectionMonitorTestConfigurationProtocolIcmp),
						}, false),
					},

					"http_configuration": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"method": {
									Type:     pluginsdk.TypeString,
									Optional: true,
									Default:  string(network.HTTPConfigurationMethodGet),
									ValidateFunc: validation.StringInSlice([]string{
										string(network.HTTPConfigurationMethodGet),
										string(network.HTTPConfigurationMethodPost),
									}, false),
								},

								"path": {
									Type:         pluginsdk.TypeString,
									Optional:     true,
									ValidateFunc: networkValidate.NetworkConnectionMonitorHttpPath,
								},

								"port": {
									Type:         pluginsdk.TypeInt,
									Optional:     true,
									ValidateFunc: validate.PortNumber,
								},

								"prefer_https": {
									Type:     pluginsdk.TypeBool,
									Optional: true,
									Default:  false,
								},

								"request_header": {
									Type:     pluginsdk.TypeSet,
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

								"valid_status_code_ranges": {
									Type:     pluginsdk.TypeSet,
									Optional: true,
									Elem: &pluginsdk.Schema{
										Type:         pluginsdk.TypeString,
										ValidateFunc: networkValidate.NetworkConnectionMonitorValidStatusCodeRanges,
									},
								},
							},
						},
					},

					"icmp_configuration": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"trace_route_enabled": {
									Type:     pluginsdk.TypeBool,
									Optional: true,
									Default:  true,
								},
							},
						},
					},

					"preferred_ip_version": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							string(network.PreferredIPVersionIPv4),
							string(network.PreferredIPVersionIPv6),
						}, false),
					},

					//lintignore:XS003
					"success_threshold": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"checks_failed_percent": {
									Type:         pluginsdk.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IntBetween(0, 100),
								},

								"round_trip_time_ms": {
									Type:         pluginsdk.TypeFloat,
									Optional:     true,
									ValidateFunc: validation.FloatAtLeast(0),
								},
							},
						},
					},

					"tcp_configuration": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"port": {
									Type:         pluginsdk.TypeInt,
									Required:     true,
									ValidateFunc: validate.PortNumber,
								},

								"trace_route_enabled": {
									Type:     pluginsdk.TypeBool,
									Optional: true,
									Default:  true,
								},

								"destination_port_behavior": {
									Type:     pluginsdk.TypeString,
									Optional: true,
									ValidateFunc: validation.StringInSlice([]string{
										string(network.DestinationPortBehaviorNone),
										string(network.DestinationPortBehaviorListenIfAvailable),
									}, false),
								},
							},
						},
					},

					"test_frequency_in_seconds": {
						Type:         pluginsdk.TypeInt,
						Optional:     true,
						Default:      60,
						ValidateFunc: validation.IntBetween(30, 1800),
					},
				},
			},
		},

		"test_group": {
			Type:     pluginsdk.TypeSet,
			Required: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"destination_endpoints": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Elem: &pluginsdk.Schema{
							Type:         pluginsdk.TypeString,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},

					"source_endpoints": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Elem: &pluginsdk.Schema{
							Type:         pluginsdk.TypeString,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},

					"test_configuration_names": {
						Type:     pluginsdk.TypeSet,
						Required: true,
						Elem: &pluginsdk.Schema{
							Type:         pluginsdk.TypeString,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},

					"enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  true,
					},
				},
			},
		},

		// API accepts any value including empty string.
		"notes": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"output_workspace_resource_ids": {
			Type:       pluginsdk.TypeSet,
			Optional:   true,
			Computed:   true,
			ConfigMode: pluginsdk.SchemaConfigModeAttr,
			Elem: &pluginsdk.Schema{
				Type:         pluginsdk.TypeString,
				ValidateFunc: logAnalyticsValidate.LogAnalyticsWorkspaceID,
			},
		},

		"tags": tags.Schema(),
	}
}

func resourceNetworkConnectionMonitorCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.ConnectionMonitorsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	location := azure.NormalizeLocation(d.Get("location").(string))

	watcherId, err := parse.NetworkWatcherID(d.Get("network_watcher_id").(string))
	if err != nil {
		return err
	}

	connectionMonitorId := parse.NewConnectionMonitorID(subscriptionId, watcherId.ResourceGroup, watcherId.Name, d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, watcherId.ResourceGroup, watcherId.Name, connectionMonitorId.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", connectionMonitorId, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_network_connection_monitor", *existing.ID)
		}
	}

	properties := network.ConnectionMonitor{
		Location: utils.String(location),
		Tags:     tags.Expand(d.Get("tags").(map[string]interface{})),
		ConnectionMonitorParameters: &network.ConnectionMonitorParameters{
			Outputs:            expandNetworkConnectionMonitorOutput(d.Get("output_workspace_resource_ids").(*pluginsdk.Set).List()),
			TestConfigurations: expandNetworkConnectionMonitorTestConfiguration(d.Get("test_configuration").(*pluginsdk.Set).List()),
			TestGroups:         expandNetworkConnectionMonitorTestGroup(d.Get("test_group").(*pluginsdk.Set).List()),
		},
	}

	if v, err := expandNetworkConnectionMonitorEndpoint(d.Get("endpoint").(*pluginsdk.Set).List()); err == nil {
		properties.ConnectionMonitorParameters.Endpoints = v
	} else {
		return err
	}

	if notes, ok := d.GetOk("notes"); ok {
		properties.Notes = utils.String(notes.(string))
	}

	future, err := client.CreateOrUpdate(ctx, watcherId.ResourceGroup, watcherId.Name, connectionMonitorId.Name, properties, "") // empty string indicating we are not migrating V1 to V2
	if err != nil {
		return fmt.Errorf("creating %s: %+v", connectionMonitorId, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of %s: %+v", connectionMonitorId, err)
	}

	d.SetId(connectionMonitorId.ID())

	return resourceNetworkConnectionMonitorRead(d, meta)
}

func resourceNetworkConnectionMonitorRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.ConnectionMonitorsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ConnectionMonitorID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.NetworkWatcherName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading %s: %+v", *id, err)
	}

	if resp.ConnectionMonitorType == network.ConnectionMonitorTypeSingleSourceDestination {
		return fmt.Errorf("the resource created via API version 2019-06-01 or before (a.k.a v1) isn't compatible to this version of provider. Please migrate to v2 pluginsdk.")
	}

	d.Set("name", id.Name)

	networkWatcherId := parse.NewNetworkWatcherID(id.SubscriptionId, id.ResourceGroup, id.NetworkWatcherName)
	d.Set("network_watcher_id", networkWatcherId.ID())

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.ConnectionMonitorResultProperties; props != nil {
		d.Set("notes", props.Notes)

		if err := d.Set("endpoint", flattenNetworkConnectionMonitorEndpoint(props.Endpoints)); err != nil {
			return fmt.Errorf("setting `endpoint`: %+v", err)
		}

		if err := d.Set("output_workspace_resource_ids", flattenNetworkConnectionMonitorOutput(props.Outputs)); err != nil {
			return fmt.Errorf("setting `output`: %+v", err)
		}

		if err := d.Set("test_configuration", flattenNetworkConnectionMonitorTestConfiguration(props.TestConfigurations)); err != nil {
			return fmt.Errorf("setting `test_configuration`: %+v", err)
		}

		if err := d.Set("test_group", flattenNetworkConnectionMonitorTestGroup(props.TestGroups)); err != nil {
			return fmt.Errorf("setting `test_group`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceNetworkConnectionMonitorDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.ConnectionMonitorsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ConnectionMonitorID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.NetworkWatcherName, id.Name)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("deleting %s: %+v", *id, err)
		}
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of %s: %+v", *id, err)
	}

	return nil
}

func expandNetworkConnectionMonitorEndpoint(input []interface{}) (*[]network.ConnectionMonitorEndpoint, error) {
	results := make([]network.ConnectionMonitorEndpoint, 0)

	for _, item := range input {
		v := item.(map[string]interface{})

		result := network.ConnectionMonitorEndpoint{
			Name:   utils.String(v["name"].(string)),
			Filter: expandNetworkConnectionMonitorEndpointFilter(v["filter"].([]interface{})),
		}

		if address := v["address"]; address != "" {
			result.Address = utils.String(address.(string))
		}

		if coverageLevel := v["coverage_level"]; coverageLevel != "" {
			result.CoverageLevel = network.CoverageLevel(coverageLevel.(string))
		}

		excludedItems := v["excluded_ip_addresses"].(*pluginsdk.Set).List()
		includedItems := v["included_ip_addresses"].(*pluginsdk.Set).List()
		if len(excludedItems) != 0 || len(includedItems) != 0 {
			result.Scope = &network.ConnectionMonitorEndpointScope{}

			if len(excludedItems) != 0 {
				var excludedAddresses []network.ConnectionMonitorEndpointScopeItem
				for _, v := range excludedItems {
					excludedAddresses = append(excludedAddresses, network.ConnectionMonitorEndpointScopeItem{
						Address: utils.String(v.(string)),
					})
				}
				result.Scope.Exclude = &excludedAddresses
			}

			if len(includedItems) != 0 {
				var includedAddresses []network.ConnectionMonitorEndpointScopeItem
				for _, v := range includedItems {
					includedAddresses = append(includedAddresses, network.ConnectionMonitorEndpointScopeItem{
						Address: utils.String(v.(string)),
					})
				}
				result.Scope.Include = &includedAddresses
			}
		}

		if resourceId := v["target_resource_id"]; resourceId != "" {
			result.ResourceID = utils.String(resourceId.(string))
		}

		if endpointType := v["target_resource_type"]; endpointType != "" {
			result.Type = network.EndpointType(endpointType.(string))
		}

		results = append(results, result)
	}

	return &results, nil
}

func expandNetworkConnectionMonitorEndpointFilter(input []interface{}) *network.ConnectionMonitorEndpointFilter {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	return &network.ConnectionMonitorEndpointFilter{
		Type:  network.ConnectionMonitorEndpointFilterType(v["type"].(string)),
		Items: expandNetworkConnectionMonitorEndpointFilterItem(v["item"].(*pluginsdk.Set).List()),
	}
}

func expandNetworkConnectionMonitorEndpointFilterItem(input []interface{}) *[]network.ConnectionMonitorEndpointFilterItem {
	if len(input) == 0 {
		return nil
	}

	results := make([]network.ConnectionMonitorEndpointFilterItem, 0)

	for _, item := range input {
		v := item.(map[string]interface{})

		result := network.ConnectionMonitorEndpointFilterItem{
			Type: network.ConnectionMonitorEndpointFilterItemType(v["type"].(string)),
		}

		if address := v["address"]; address != "" {
			result.Address = utils.String(address.(string))
		}

		results = append(results, result)
	}

	return &results
}

func expandNetworkConnectionMonitorTestConfiguration(input []interface{}) *[]network.ConnectionMonitorTestConfiguration {
	results := make([]network.ConnectionMonitorTestConfiguration, 0)

	for _, item := range input {
		v := item.(map[string]interface{})

		result := network.ConnectionMonitorTestConfiguration{
			Name:              utils.String(v["name"].(string)),
			HTTPConfiguration: expandNetworkConnectionMonitorHTTPConfiguration(v["http_configuration"].([]interface{})),
			IcmpConfiguration: expandNetworkConnectionMonitorIcmpConfiguration(v["icmp_configuration"].([]interface{})),
			Protocol:          network.ConnectionMonitorTestConfigurationProtocol(v["protocol"].(string)),
			SuccessThreshold:  expandNetworkConnectionMonitorSuccessThreshold(v["success_threshold"].([]interface{})),
			TCPConfiguration:  expandNetworkConnectionMonitorTCPConfiguration(v["tcp_configuration"].([]interface{})),
			TestFrequencySec:  utils.Int32(int32(v["test_frequency_in_seconds"].(int))),
		}

		if preferredIPVersion := v["preferred_ip_version"]; preferredIPVersion != "" {
			result.PreferredIPVersion = network.PreferredIPVersion(preferredIPVersion.(string))
		}

		results = append(results, result)
	}

	return &results
}

func expandNetworkConnectionMonitorHTTPConfiguration(input []interface{}) *network.ConnectionMonitorHTTPConfiguration {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	props := &network.ConnectionMonitorHTTPConfiguration{
		Method:         network.HTTPConfigurationMethod(v["method"].(string)),
		PreferHTTPS:    utils.Bool(v["prefer_https"].(bool)),
		RequestHeaders: expandNetworkConnectionMonitorHTTPHeader(v["request_header"].(*pluginsdk.Set).List()),
	}

	if path := v["path"]; path != "" {
		props.Path = utils.String(path.(string))
	}

	if port := v["port"]; port != 0 {
		props.Port = utils.Int32(int32(port.(int)))
	}

	if ranges := v["valid_status_code_ranges"].(*pluginsdk.Set).List(); len(ranges) != 0 {
		props.ValidStatusCodeRanges = utils.ExpandStringSlice(ranges)
	}

	return props
}

func expandNetworkConnectionMonitorTCPConfiguration(input []interface{}) *network.ConnectionMonitorTCPConfiguration {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	result := &network.ConnectionMonitorTCPConfiguration{
		Port:              utils.Int32(int32(v["port"].(int))),
		DisableTraceRoute: utils.Bool(!v["trace_route_enabled"].(bool)),
	}

	if destinationPortBehavior := v["destination_port_behavior"].(string); destinationPortBehavior != "" {
		result.DestinationPortBehavior = network.DestinationPortBehavior(destinationPortBehavior)
	}

	return result
}

func expandNetworkConnectionMonitorIcmpConfiguration(input []interface{}) *network.ConnectionMonitorIcmpConfiguration {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	return &network.ConnectionMonitorIcmpConfiguration{
		DisableTraceRoute: utils.Bool(!v["trace_route_enabled"].(bool)),
	}
}

func expandNetworkConnectionMonitorSuccessThreshold(input []interface{}) *network.ConnectionMonitorSuccessThreshold {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	v := input[0].(map[string]interface{})

	return &network.ConnectionMonitorSuccessThreshold{
		ChecksFailedPercent: utils.Int32(int32(v["checks_failed_percent"].(int))),
		RoundTripTimeMs:     utils.Float(v["round_trip_time_ms"].(float64)),
	}
}

func expandNetworkConnectionMonitorHTTPHeader(input []interface{}) *[]network.HTTPHeader {
	if len(input) == 0 {
		return nil
	}

	results := make([]network.HTTPHeader, 0)

	for _, item := range input {
		v := item.(map[string]interface{})

		result := network.HTTPHeader{
			Name:  utils.String(v["name"].(string)),
			Value: utils.String(v["value"].(string)),
		}

		results = append(results, result)
	}

	return &results
}

func expandNetworkConnectionMonitorTestGroup(input []interface{}) *[]network.ConnectionMonitorTestGroup {
	results := make([]network.ConnectionMonitorTestGroup, 0)

	for _, item := range input {
		v := item.(map[string]interface{})

		result := network.ConnectionMonitorTestGroup{
			Name:               utils.String(v["name"].(string)),
			Destinations:       utils.ExpandStringSlice(v["destination_endpoints"].(*pluginsdk.Set).List()),
			Disable:            utils.Bool(!v["enabled"].(bool)),
			Sources:            utils.ExpandStringSlice(v["source_endpoints"].(*pluginsdk.Set).List()),
			TestConfigurations: utils.ExpandStringSlice(v["test_configuration_names"].(*pluginsdk.Set).List()),
		}

		results = append(results, result)
	}

	return &results
}

func expandNetworkConnectionMonitorOutput(input []interface{}) *[]network.ConnectionMonitorOutput {
	results := make([]network.ConnectionMonitorOutput, 0)

	for _, item := range input {
		result := network.ConnectionMonitorOutput{
			Type: network.OutputTypeWorkspace,
			WorkspaceSettings: &network.ConnectionMonitorWorkspaceSettings{
				WorkspaceResourceID: utils.String(item.(string)),
			},
		}

		results = append(results, result)
	}

	return &results
}

func flattenNetworkConnectionMonitorEndpoint(input *[]network.ConnectionMonitorEndpoint) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var name string
		if item.Name != nil {
			name = *item.Name
		}

		var address string
		if item.Address != nil {
			address = *item.Address
		}

		var coverageLevel string
		if item.CoverageLevel != "" {
			coverageLevel = string(item.CoverageLevel)
		}

		var endpointType string
		if item.Type != "" {
			endpointType = string(item.Type)
		}

		var resourceId string
		if item.ResourceID != nil {
			resourceId = *item.ResourceID
		}

		v := map[string]interface{}{
			"name":                 name,
			"address":              address,
			"coverage_level":       coverageLevel,
			"target_resource_id":   resourceId,
			"target_resource_type": endpointType,
			"filter":               flattenNetworkConnectionMonitorEndpointFilter(item.Filter),
		}

		if scope := item.Scope; scope != nil {
			if includeScope := scope.Include; includeScope != nil {
				includedAddresses := make([]interface{}, 0)

				for _, includedItem := range *includeScope {
					if includedAddress := includedItem.Address; includedAddress != nil {
						includedAddresses = append(includedAddresses, includedAddress)
					}
				}

				v["included_ip_addresses"] = includedAddresses
			}

			if excludeScope := scope.Exclude; excludeScope != nil {
				excludedAddresses := make([]interface{}, 0)

				for _, excludedItem := range *excludeScope {
					if excludedAddress := excludedItem.Address; excludedAddress != nil {
						excludedAddresses = append(excludedAddresses, excludedAddress)
					}
				}

				v["excluded_ip_addresses"] = excludedAddresses
			}
		}

		results = append(results, v)
	}
	return results
}

func flattenNetworkConnectionMonitorEndpointFilter(input *network.ConnectionMonitorEndpointFilter) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var t network.ConnectionMonitorEndpointFilterType
	if input.Type != "" {
		t = input.Type
	}
	return []interface{}{
		map[string]interface{}{
			"item": flattenNetworkConnectionMonitorEndpointFilterItem(input.Items),
			"type": t,
		},
	}
}

func flattenNetworkConnectionMonitorEndpointFilterItem(input *[]network.ConnectionMonitorEndpointFilterItem) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var address string
		if item.Address != nil {
			address = *item.Address
		}

		var t network.ConnectionMonitorEndpointFilterItemType
		if item.Type != "" {
			t = item.Type
		}

		v := map[string]interface{}{
			"address": address,
			"type":    t,
		}

		results = append(results, v)
	}

	return results
}

func flattenNetworkConnectionMonitorTestConfiguration(input *[]network.ConnectionMonitorTestConfiguration) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var name string
		if item.Name != nil {
			name = *item.Name
		}

		var protocol network.ConnectionMonitorTestConfigurationProtocol
		if item.Protocol != "" {
			protocol = item.Protocol
		}

		var preferredIpVersion network.PreferredIPVersion
		if item.PreferredIPVersion != "" {
			preferredIpVersion = item.PreferredIPVersion
		}

		var testFrequencySec int32
		if item.TestFrequencySec != nil {
			testFrequencySec = *item.TestFrequencySec
		}

		v := map[string]interface{}{
			"name":                      name,
			"protocol":                  protocol,
			"http_configuration":        flattenNetworkConnectionMonitorHTTPConfiguration(item.HTTPConfiguration),
			"icmp_configuration":        flattenNetworkConnectionMonitorIcmpConfiguration(item.IcmpConfiguration),
			"preferred_ip_version":      preferredIpVersion,
			"success_threshold":         flattenNetworkConnectionMonitorSuccessThreshold(item.SuccessThreshold),
			"tcp_configuration":         flattenNetworkConnectionMonitorTCPConfiguration(item.TCPConfiguration),
			"test_frequency_in_seconds": testFrequencySec,
		}

		results = append(results, v)
	}

	return results
}

func flattenNetworkConnectionMonitorHTTPConfiguration(input *network.ConnectionMonitorHTTPConfiguration) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var method network.HTTPConfigurationMethod
	if input.Method != "" {
		method = input.Method
	}

	var p string
	if input.Path != nil {
		p = *input.Path
	}

	var port int32
	if input.Port != nil {
		port = *input.Port
	}

	var preferHttps bool
	if input.PreferHTTPS != nil {
		preferHttps = *input.PreferHTTPS
	}

	return []interface{}{
		map[string]interface{}{
			"method":                   method,
			"path":                     p,
			"port":                     port,
			"prefer_https":             preferHttps,
			"request_header":           flattenNetworkConnectionMonitorHTTPHeader(input.RequestHeaders),
			"valid_status_code_ranges": utils.FlattenStringSlice(input.ValidStatusCodeRanges),
		},
	}
}

func flattenNetworkConnectionMonitorIcmpConfiguration(input *network.ConnectionMonitorIcmpConfiguration) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var enableTraceRoute bool
	if input.DisableTraceRoute != nil {
		enableTraceRoute = !*input.DisableTraceRoute
	}

	return []interface{}{
		map[string]interface{}{
			"trace_route_enabled": enableTraceRoute,
		},
	}
}

func flattenNetworkConnectionMonitorSuccessThreshold(input *network.ConnectionMonitorSuccessThreshold) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var checksFailedPercent int32
	if input.ChecksFailedPercent != nil {
		checksFailedPercent = *input.ChecksFailedPercent
	}

	var roundTripTimeMs float64
	if input.RoundTripTimeMs != nil {
		roundTripTimeMs = *input.RoundTripTimeMs
	}

	return []interface{}{
		map[string]interface{}{
			"checks_failed_percent": checksFailedPercent,
			"round_trip_time_ms":    roundTripTimeMs,
		},
	}
}

func flattenNetworkConnectionMonitorTCPConfiguration(input *network.ConnectionMonitorTCPConfiguration) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var enableTraceRoute bool
	if input.DisableTraceRoute != nil {
		enableTraceRoute = !*input.DisableTraceRoute
	}

	var port int32
	if input.Port != nil {
		port = *input.Port
	}

	var destinationPortBehavior network.DestinationPortBehavior
	if input.DestinationPortBehavior != "" {
		destinationPortBehavior = input.DestinationPortBehavior
	}

	return []interface{}{
		map[string]interface{}{
			"trace_route_enabled":       enableTraceRoute,
			"port":                      port,
			"destination_port_behavior": string(destinationPortBehavior),
		},
	}
}

func flattenNetworkConnectionMonitorHTTPHeader(input *[]network.HTTPHeader) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var name string
		if item.Name != nil {
			name = *item.Name
		}

		var value string
		if item.Value != nil {
			value = *item.Value
		}

		v := map[string]interface{}{
			"name":  name,
			"value": value,
		}

		results = append(results, v)
	}

	return results
}

func flattenNetworkConnectionMonitorTestGroup(input *[]network.ConnectionMonitorTestGroup) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var name string
		if item.Name != nil {
			name = *item.Name
		}

		var disable bool
		if item.Disable != nil {
			disable = *item.Disable
		}

		v := map[string]interface{}{
			"name":                     name,
			"destination_endpoints":    utils.FlattenStringSlice(item.Destinations),
			"source_endpoints":         utils.FlattenStringSlice(item.Sources),
			"test_configuration_names": utils.FlattenStringSlice(item.TestConfigurations),
			"enabled":                  !disable,
		}

		results = append(results, v)
	}
	return results
}

func flattenNetworkConnectionMonitorOutput(input *[]network.ConnectionMonitorOutput) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var workspaceResourceId string
		if item.WorkspaceSettings != nil && item.WorkspaceSettings.WorkspaceResourceID != nil {
			workspaceResourceId = *item.WorkspaceSettings.WorkspaceResourceID
		}

		results = append(results, workspaceResourceId)
	}

	return results
}
