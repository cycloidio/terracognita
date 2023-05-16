package iothub

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/iothub/mgmt/2021-07-02/devices"
	"github.com/Azure/azure-sdk-for-go/services/provisioningservices/mgmt/2021-10-15/iothub"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/iothub/parse"
	iothubValidate "github.com/hashicorp/terraform-provider-azurerm/services/iothub/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceIotHubDPS() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceIotHubDPSCreateUpdate,
		Read:   resourceIotHubDPSRead,
		Update: resourceIotHubDPSCreateUpdate,
		Delete: resourceIotHubDPSDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.IotHubDpsID(id)
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
				ValidateFunc: iothubValidate.IoTHubName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(), // azure.SchemaResourceGroupNameDiffSuppress(),

			"location": azure.SchemaLocation(),

			"sku": {
				Type:     pluginsdk.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(devices.IotHubSkuS1),
							}, false),
						},

						"capacity": {
							Type:         pluginsdk.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 200),
						},
					},
				},
			},

			"linked_hub": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"connection_string": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							// Azure returns the key as ****. We'll suppress that here.
							DiffSuppressFunc: func(k, old, new string, d *pluginsdk.ResourceData) bool {
								secretKeyRegex := regexp.MustCompile("(SharedAccessKey)=[^;]+")
								maskedNew := secretKeyRegex.ReplaceAllString(new, "$1=****")
								return (new == d.Get(k).(string)) && (maskedNew == old)
							},
							Sensitive: true,
						},
						"location": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							StateFunc:    azure.NormalizeLocation,
						},
						"apply_allocation_policy": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  true,
						},
						"allocation_weight": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
						"hostname": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"ip_filter_rule": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"ip_mask": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.CIDR,
						},
						"action": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(devices.IPFilterActionTypeAccept),
								string(devices.IPFilterActionTypeReject),
							}, false),
						},
						"target": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								azure.TitleCase(string(iothub.IPFilterTargetTypeAll)),
								azure.TitleCase(string(iothub.IPFilterTargetTypeServiceAPI)),
								azure.TitleCase(string(iothub.IPFilterTargetTypeDeviceAPI)),
							}, false),
						},
					},
				},
			},

			"allocation_policy": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(iothub.AllocationPolicyHashed),
				ValidateFunc: validation.StringInSlice([]string{
					string(iothub.AllocationPolicyHashed),
					string(iothub.AllocationPolicyGeoLatency),
					string(iothub.AllocationPolicyStatic),
				}, false),
			},

			"public_network_access_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"device_provisioning_host_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"id_scope": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"service_operations_host_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceIotHubDPSCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTHub.DPSResourceClient
	subscriptionId := meta.(*clients.Client).IoTHub.DPSResourceClient.SubscriptionID
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewIotHubDpsID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ProvisioningServiceName, id.ResourceGroup)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing IoT Device Provisioning Service %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_iothub_dps", id.ID())
		}
	}

	publicNetworkAccess := iothub.PublicNetworkAccessEnabled
	if !d.Get("public_network_access_enabled").(bool) {
		publicNetworkAccess = iothub.PublicNetworkAccessDisabled
	}

	iotdps := iothub.ProvisioningServiceDescription{
		Location: utils.String(azure.NormalizeLocation(d.Get("location").(string))),
		Name:     utils.String(id.ProvisioningServiceName),
		Sku:      expandIoTHubDPSSku(d),
		Properties: &iothub.IotDpsPropertiesDescription{
			IotHubs:             expandIoTHubDPSIoTHubs(d.Get("linked_hub").([]interface{})),
			AllocationPolicy:    iothub.AllocationPolicy(d.Get("allocation_policy").(string)),
			IPFilterRules:       expandDpsIPFilterRules(d),
			PublicNetworkAccess: publicNetworkAccess,
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.ProvisioningServiceName, iotdps)
	if err != nil {
		return fmt.Errorf("creating/updating IoT Device Provisioning Service %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the completion of the creating/updating of IoT Device Provisioning Service %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceIotHubDPSRead(d, meta)
}

func resourceIotHubDPSRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTHub.DPSResourceClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.IotHubDpsID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ProvisioningServiceName, id.ResourceGroup)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.ProvisioningServiceName)
	d.Set("resource_group_name", id.ResourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}
	sku := flattenIoTHubDPSSku(resp.Sku)
	if err := d.Set("sku", sku); err != nil {
		return fmt.Errorf("setting `sku`: %+v", err)
	}

	if props := resp.Properties; props != nil {
		if err := d.Set("linked_hub", flattenIoTHubDPSLinkedHub(props.IotHubs)); err != nil {
			return fmt.Errorf("setting `linked_hub`: %+v", err)
		}

		ipFilterRules := flattenDpsIPFilterRules(props.IPFilterRules)
		if err := d.Set("ip_filter_rule", ipFilterRules); err != nil {
			return fmt.Errorf("setting `ip_filter_rule` in IoTHub DPS %q: %+v", id.ProvisioningServiceName, err)
		}

		d.Set("service_operations_host_name", props.ServiceOperationsHostName)
		d.Set("device_provisioning_host_name", props.DeviceProvisioningHostName)
		d.Set("id_scope", props.IDScope)
		d.Set("allocation_policy", string(props.AllocationPolicy))
		publicNetworkAccess := true
		if props.PublicNetworkAccess != "" {
			publicNetworkAccess = strings.EqualFold("Enabled", string(props.PublicNetworkAccess))
		}
		d.Set("public_network_access_enabled", publicNetworkAccess)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceIotHubDPSDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTHub.DPSResourceClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.IotHubDpsID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ProvisioningServiceName, id.ResourceGroup)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("deleting %s: %+v", id, err)
		}
	}

	return waitForIotHubDPSToBeDeleted(ctx, client, id.ResourceGroup, id.ProvisioningServiceName, d)
}

func waitForIotHubDPSToBeDeleted(ctx context.Context, client *iothub.IotDpsResourceClient, resourceGroup, name string, d *pluginsdk.ResourceData) error {
	// we can't use the Waiter here since the API returns a 404 once it's deleted which is considered a polling status code..
	log.Printf("[DEBUG] Waiting for IoT Device Provisioning Service %q (Resource Group %q) to be deleted", name, resourceGroup)
	stateConf := &pluginsdk.StateChangeConf{
		Pending: []string{"200"},
		Target:  []string{"404"},
		Refresh: iothubdpsStateStatusCodeRefreshFunc(ctx, client, resourceGroup, name),
		Timeout: d.Timeout(pluginsdk.TimeoutDelete),
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for IoT Device Provisioning Service %q (Resource Group %q) to be deleted: %+v", name, resourceGroup, err)
	}

	return nil
}

func iothubdpsStateStatusCodeRefreshFunc(ctx context.Context, client *iothub.IotDpsResourceClient, resourceGroup, name string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, name, resourceGroup)

		log.Printf("Retrieving IoT Device Provisioning Service %q (Resource Group %q) returned Status %d", resourceGroup, name, res.StatusCode)

		if err != nil {
			if utils.ResponseWasNotFound(res.Response) {
				return res, strconv.Itoa(res.StatusCode), nil
			}
			return nil, "", fmt.Errorf("polling for the status of the IoT Device Provisioning Service %q (Resource Group %q): %+v", name, resourceGroup, err)
		}

		return res, strconv.Itoa(res.StatusCode), nil
	}
}

func expandIoTHubDPSSku(d *pluginsdk.ResourceData) *iothub.IotDpsSkuInfo {
	skuList := d.Get("sku").([]interface{})
	skuMap := skuList[0].(map[string]interface{})

	return &iothub.IotDpsSkuInfo{
		Name:     iothub.IotDpsSku(skuMap["name"].(string)),
		Capacity: utils.Int64(int64(skuMap["capacity"].(int))),
	}
}

func expandIoTHubDPSIoTHubs(input []interface{}) *[]iothub.DefinitionDescription {
	linkedHubs := make([]iothub.DefinitionDescription, 0)

	for _, attr := range input {
		linkedHubConfig := attr.(map[string]interface{})
		linkedHub := iothub.DefinitionDescription{
			ConnectionString:      utils.String(linkedHubConfig["connection_string"].(string)),
			AllocationWeight:      utils.Int32(int32(linkedHubConfig["allocation_weight"].(int))),
			ApplyAllocationPolicy: utils.Bool(linkedHubConfig["apply_allocation_policy"].(bool)),
			Location:              utils.String(azure.NormalizeLocation(linkedHubConfig["location"].(string))),
		}

		linkedHubs = append(linkedHubs, linkedHub)
	}

	return &linkedHubs
}

func flattenIoTHubDPSSku(input *iothub.IotDpsSkuInfo) []interface{} {
	output := make(map[string]interface{})

	output["name"] = string(input.Name)
	if capacity := input.Capacity; capacity != nil {
		output["capacity"] = int(*capacity)
	}

	return []interface{}{output}
}

func flattenIoTHubDPSLinkedHub(input *[]iothub.DefinitionDescription) []interface{} {
	linkedHubs := make([]interface{}, 0)
	if input == nil {
		return linkedHubs
	}

	for _, attr := range *input {
		linkedHub := make(map[string]interface{})

		if attr.Name != nil {
			linkedHub["hostname"] = *attr.Name
		}
		if attr.ApplyAllocationPolicy != nil {
			linkedHub["apply_allocation_policy"] = *attr.ApplyAllocationPolicy
		}
		if attr.AllocationWeight != nil {
			linkedHub["allocation_weight"] = *attr.AllocationWeight
		}
		if attr.ConnectionString != nil {
			linkedHub["connection_string"] = *attr.ConnectionString
		}
		if attr.Location != nil {
			linkedHub["location"] = azure.NormalizeLocation(*attr.Location)
		}

		linkedHubs = append(linkedHubs, linkedHub)
	}

	return linkedHubs
}

func expandDpsIPFilterRules(d *pluginsdk.ResourceData) *[]iothub.IPFilterRule {
	ipFilterRuleList := d.Get("ip_filter_rule").([]interface{})
	if len(ipFilterRuleList) == 0 {
		return nil
	}

	rules := make([]iothub.IPFilterRule, 0)

	for _, r := range ipFilterRuleList {
		rawRule := r.(map[string]interface{})
		rule := &iothub.IPFilterRule{
			FilterName: utils.String(rawRule["name"].(string)),
			Action:     iothub.IPFilterActionType(rawRule["action"].(string)),
			IPMask:     utils.String(rawRule["ip_mask"].(string)),
			Target:     iothub.IPFilterTargetType(azure.TitleCase(rawRule["target"].(string))),
		}

		rules = append(rules, *rule)
	}
	return &rules
}

func flattenDpsIPFilterRules(in *[]iothub.IPFilterRule) []interface{} {
	rules := make([]interface{}, 0)
	if in == nil {
		return rules
	}

	for _, r := range *in {
		rawRule := make(map[string]interface{})

		if r.FilterName != nil {
			rawRule["name"] = *r.FilterName
		}

		rawRule["action"] = string(r.Action)

		if r.IPMask != nil {
			rawRule["ip_mask"] = *r.IPMask
		}

		if r.Target != "" {
			rawRule["target"] = r.Target
		}

		rules = append(rules, rawRule)
	}
	return rules
}
