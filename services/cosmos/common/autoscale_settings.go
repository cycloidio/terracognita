package common

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2021-10-15/documentdb"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func ExpandCosmosDbAutoscaleSettings(d *pluginsdk.ResourceData) *documentdb.AutoscaleSettings {
	i := d.Get("autoscale_settings").([]interface{})
	if len(i) == 0 || i[0] == nil {
		log.Printf("[DEBUG] Cosmos DB autoscale settings are not set on the resource")
		return nil
	}
	input := i[0].(map[string]interface{})

	autoscaleSettings := documentdb.AutoscaleSettings{}

	if maxThroughput, ok := input["max_throughput"].(int); ok {
		autoscaleSettings.MaxThroughput = utils.Int32(int32(maxThroughput))
	}

	return &autoscaleSettings
}

func FlattenCosmosDbAutoscaleSettings(throughputResponse documentdb.ThroughputSettingsGetResults) []interface{} {
	results := make([]interface{}, 0)

	props := throughputResponse.ThroughputSettingsGetProperties
	if props == nil {
		return results
	}

	res := props.Resource
	if res == nil {
		return results
	}

	autoscaleSettings := res.AutoscaleSettings
	if autoscaleSettings == nil {
		log.Printf("[DEBUG] Cosmos DB autoscale settings are not set on the throughput response")
		return results
	}

	result := make(map[string]interface{})

	if autoscaleSettings.MaxThroughput != nil {
		result["max_throughput"] = autoscaleSettings.MaxThroughput
	}

	return append(results, result)
}

func ExpandCosmosDbAutoscaleSettingsResource(d *pluginsdk.ResourceData) *documentdb.AutoscaleSettingsResource {
	autoscaleSettings := ExpandCosmosDbAutoscaleSettings(d)
	autoscaleSettingResource := documentdb.AutoscaleSettingsResource{}

	autoscaleSettingResource.MaxThroughput = autoscaleSettings.MaxThroughput
	return &autoscaleSettingResource
}
