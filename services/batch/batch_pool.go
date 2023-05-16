package batch

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/batch/mgmt/2021-06-01/batch"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

// flattenBatchPoolAutoScaleSettings flattens the auto scale settings for a Batch pool
func flattenBatchPoolAutoScaleSettings(settings *batch.AutoScaleSettings) []interface{} {
	results := make([]interface{}, 0)

	if settings == nil {
		log.Printf("[DEBUG] settings is nil")
		return results
	}

	result := make(map[string]interface{})

	if settings.EvaluationInterval != nil {
		result["evaluation_interval"] = *settings.EvaluationInterval
	}

	if settings.Formula != nil {
		result["formula"] = *settings.Formula
	}

	return append(results, result)
}

// flattenBatchPoolFixedScaleSettings flattens the fixed scale settings for a Batch pool
func flattenBatchPoolFixedScaleSettings(settings *batch.FixedScaleSettings) []interface{} {
	results := make([]interface{}, 0)

	if settings == nil {
		log.Printf("[DEBUG] settings is nil")
		return results
	}

	result := make(map[string]interface{})

	if settings.TargetDedicatedNodes != nil {
		result["target_dedicated_nodes"] = *settings.TargetDedicatedNodes
	}

	if settings.TargetLowPriorityNodes != nil {
		result["target_low_priority_nodes"] = *settings.TargetLowPriorityNodes
	}

	if settings.ResizeTimeout != nil {
		result["resize_timeout"] = *settings.ResizeTimeout
	}

	return append(results, result)
}

// flattenBatchPoolImageReference flattens the Batch pool image reference
func flattenBatchPoolImageReference(image *batch.ImageReference) []interface{} {
	results := make([]interface{}, 0)
	if image == nil {
		log.Printf("[DEBUG] image is nil")
		return results
	}

	result := make(map[string]interface{})
	if image.Publisher != nil {
		result["publisher"] = *image.Publisher
	}
	if image.Offer != nil {
		result["offer"] = *image.Offer
	}
	if image.Sku != nil {
		result["sku"] = *image.Sku
	}
	if image.Version != nil {
		result["version"] = *image.Version
	}
	if image.ID != nil {
		result["id"] = *image.ID
	}

	return append(results, result)
}

// flattenBatchPoolStartTask flattens a Batch pool start task
func flattenBatchPoolStartTask(startTask *batch.StartTask) []interface{} {
	results := make([]interface{}, 0)

	if startTask == nil {
		log.Printf("[DEBUG] startTask is nil")
		return results
	}

	result := make(map[string]interface{})
	commandLine := ""
	if startTask.CommandLine != nil {
		commandLine = *startTask.CommandLine
	}
	result["command_line"] = commandLine

	waitForSuccess := false
	if startTask.WaitForSuccess != nil {
		waitForSuccess = *startTask.WaitForSuccess
	}
	result["wait_for_success"] = waitForSuccess

	maxTaskRetryCount := int32(0)
	if startTask.MaxTaskRetryCount != nil {
		maxTaskRetryCount = *startTask.MaxTaskRetryCount
	}

	result["task_retry_maximum"] = maxTaskRetryCount

	if startTask.UserIdentity != nil {
		userIdentity := make(map[string]interface{})
		if startTask.UserIdentity.AutoUser != nil {
			autoUser := make(map[string]interface{})

			elevationLevel := string(startTask.UserIdentity.AutoUser.ElevationLevel)
			scope := string(startTask.UserIdentity.AutoUser.Scope)

			autoUser["elevation_level"] = elevationLevel
			autoUser["scope"] = scope

			userIdentity["auto_user"] = []interface{}{autoUser}
		} else {
			userIdentity["user_name"] = *startTask.UserIdentity.UserName
		}

		result["user_identity"] = []interface{}{userIdentity}
	}

	resourceFiles := make([]interface{}, 0)
	if startTask.ResourceFiles != nil {
		for _, armResourceFile := range *startTask.ResourceFiles {
			resourceFile := make(map[string]interface{})
			if armResourceFile.AutoStorageContainerName != nil {
				resourceFile["auto_storage_container_name"] = *armResourceFile.AutoStorageContainerName
			}
			if armResourceFile.StorageContainerURL != nil {
				resourceFile["storage_container_url"] = *armResourceFile.StorageContainerURL
			}
			if armResourceFile.HTTPURL != nil {
				resourceFile["http_url"] = *armResourceFile.HTTPURL
			}
			if armResourceFile.BlobPrefix != nil {
				resourceFile["blob_prefix"] = *armResourceFile.BlobPrefix
			}
			if armResourceFile.FilePath != nil {
				resourceFile["file_path"] = *armResourceFile.FilePath
			}
			if armResourceFile.FileMode != nil {
				resourceFile["file_mode"] = *armResourceFile.FileMode
			}
			resourceFiles = append(resourceFiles, resourceFile)
		}
	}

	environment := make(map[string]interface{})
	if startTask.EnvironmentSettings != nil {
		for _, envSetting := range *startTask.EnvironmentSettings {
			environment[*envSetting.Name] = *envSetting.Value
		}
	}

	result["common_environment_properties"] = environment

	result["resource_file"] = resourceFiles

	return append(results, result)
}

// flattenBatchPoolCertificateReferences flattens a Batch pool certificate reference
func flattenBatchPoolCertificateReferences(armCertificates *[]batch.CertificateReference) []interface{} {
	if armCertificates == nil {
		return []interface{}{}
	}
	output := make([]interface{}, 0)

	for _, armCertificate := range *armCertificates {
		certificate := map[string]interface{}{}
		if armCertificate.ID != nil {
			certificate["id"] = *armCertificate.ID
		}
		certificate["store_location"] = string(armCertificate.StoreLocation)
		if armCertificate.StoreName != nil {
			certificate["store_name"] = *armCertificate.StoreName
		}
		visibility := &pluginsdk.Set{F: pluginsdk.HashString}
		if armCertificate.Visibility != nil {
			for _, armVisibility := range *armCertificate.Visibility {
				visibility.Add(string(armVisibility))
			}
		}
		certificate["visibility"] = visibility
		output = append(output, certificate)
	}
	return output
}

// flattenBatchPoolContainerConfiguration flattens a Batch pool container configuration
func flattenBatchPoolContainerConfiguration(d *pluginsdk.ResourceData, armContainerConfiguration *batch.ContainerConfiguration) interface{} {
	result := make(map[string]interface{})

	if armContainerConfiguration == nil {
		return nil
	}

	if armContainerConfiguration.Type != nil {
		result["type"] = *armContainerConfiguration.Type
	}

	names := &pluginsdk.Set{F: pluginsdk.HashString}
	if armContainerConfiguration.ContainerImageNames != nil {
		for _, armName := range *armContainerConfiguration.ContainerImageNames {
			names.Add(armName)
		}
	}
	result["container_image_names"] = names

	result["container_registries"] = flattenBatchPoolContainerRegistries(d, armContainerConfiguration.ContainerRegistries)

	return []interface{}{result}
}

func flattenBatchPoolContainerRegistries(d *pluginsdk.ResourceData, armContainerRegistries *[]batch.ContainerRegistry) []interface{} {
	results := make([]interface{}, 0)

	if armContainerRegistries == nil {
		return results
	}
	for _, armContainerRegistry := range *armContainerRegistries {
		result := flattenBatchPoolContainerRegistry(d, &armContainerRegistry)
		results = append(results, result)
	}
	return results
}

func flattenBatchPoolContainerRegistry(d *pluginsdk.ResourceData, armContainerRegistry *batch.ContainerRegistry) map[string]interface{} {
	result := make(map[string]interface{})

	if armContainerRegistry == nil {
		return result
	}
	if registryServer := armContainerRegistry.RegistryServer; registryServer != nil {
		result["registry_server"] = *registryServer
	}
	if userName := armContainerRegistry.UserName; userName != nil {
		result["user_name"] = *userName
	}

	// If we didn't specify a registry server and user name, just return what we have now rather than trying to locate the password
	if len(result) != 2 {
		return result
	}

	result["password"] = findBatchPoolContainerRegistryPassword(d, result["registry_server"].(string), result["user_name"].(string))

	return result
}

func findBatchPoolContainerRegistryPassword(d *pluginsdk.ResourceData, armServer string, armUsername string) interface{} {
	numContainerRegistries := 0
	if n, ok := d.GetOk("container_configuration.0.container_registries.#"); ok {
		numContainerRegistries = n.(int)
	} else {
		return ""
	}

	for i := 0; i < numContainerRegistries; i++ {
		if server, ok := d.GetOk(fmt.Sprintf("container_configuration.0.container_registries.%d.registry_server", i)); !ok || server != armServer {
			continue
		}
		if username, ok := d.GetOk(fmt.Sprintf("container_configuration.0.container_registries.%d.user_name", i)); !ok || username != armUsername {
			continue
		}
		return d.Get(fmt.Sprintf("container_configuration.0.container_registries.%d.password", i))
	}

	return ""
}

// ExpandBatchPoolImageReference expands Batch pool image reference
func ExpandBatchPoolImageReference(list []interface{}) (*batch.ImageReference, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("Error: storage image reference should be defined")
	}

	storageImageRef := list[0].(map[string]interface{})
	imageRef := &batch.ImageReference{}

	if storageImageRef["id"] != nil && storageImageRef["id"] != "" {
		storageImageRefID := storageImageRef["id"].(string)
		imageRef.ID = &storageImageRefID
	}

	if storageImageRef["offer"] != nil && storageImageRef["offer"] != "" {
		storageImageRefOffer := storageImageRef["offer"].(string)
		imageRef.Offer = &storageImageRefOffer
	}

	if storageImageRef["publisher"] != nil && storageImageRef["publisher"] != "" {
		storageImageRefPublisher := storageImageRef["publisher"].(string)
		imageRef.Publisher = &storageImageRefPublisher
	}

	if storageImageRef["sku"] != nil && storageImageRef["sku"] != "" {
		storageImageRefSku := storageImageRef["sku"].(string)
		imageRef.Sku = &storageImageRefSku
	}

	if storageImageRef["version"] != nil && storageImageRef["version"] != "" {
		storageImageRefVersion := storageImageRef["version"].(string)
		imageRef.Version = &storageImageRefVersion
	}

	return imageRef, nil
}

// ExpandBatchPoolContainerConfiguration expands the Batch pool container configuration
func ExpandBatchPoolContainerConfiguration(list []interface{}) (*batch.ContainerConfiguration, error) {
	if len(list) == 0 || list[0] == nil {
		return nil, nil
	}

	block := list[0].(map[string]interface{})

	containerRegistries, err := expandBatchPoolContainerRegistries(block["container_registries"].([]interface{}))
	if err != nil {
		return nil, err
	}

	obj := &batch.ContainerConfiguration{
		Type:                utils.String(block["type"].(string)),
		ContainerRegistries: containerRegistries,
		ContainerImageNames: utils.ExpandStringSlice(block["container_image_names"].(*pluginsdk.Set).List()),
	}

	return obj, nil
}

func expandBatchPoolContainerRegistries(list []interface{}) (*[]batch.ContainerRegistry, error) {
	result := []batch.ContainerRegistry{}

	for _, tempItem := range list {
		item := tempItem.(map[string]interface{})
		containerRegistry, err := expandBatchPoolContainerRegistry(item)
		if err != nil {
			return nil, err
		}
		result = append(result, *containerRegistry)
	}
	return &result, nil
}

func expandBatchPoolContainerRegistry(ref map[string]interface{}) (*batch.ContainerRegistry, error) {
	if len(ref) == 0 {
		return nil, fmt.Errorf("Error: container registry reference should be defined")
	}

	containerRegistry := batch.ContainerRegistry{
		RegistryServer: utils.String(ref["registry_server"].(string)),
		UserName:       utils.String(ref["user_name"].(string)),
		Password:       utils.String(ref["password"].(string)),
	}
	return &containerRegistry, nil
}

// ExpandBatchPoolCertificateReferences expands Batch pool certificate references
func ExpandBatchPoolCertificateReferences(list []interface{}) (*[]batch.CertificateReference, error) {
	var result []batch.CertificateReference

	for _, tempItem := range list {
		item := tempItem.(map[string]interface{})
		certificateReference, err := expandBatchPoolCertificateReference(item)
		if err != nil {
			return nil, err
		}
		result = append(result, *certificateReference)
	}
	return &result, nil
}

func expandBatchPoolCertificateReference(ref map[string]interface{}) (*batch.CertificateReference, error) {
	if len(ref) == 0 {
		return nil, fmt.Errorf("Error: storage image reference should be defined")
	}

	id := ref["id"].(string)
	storeLocation := ref["store_location"].(string)
	storeName := ref["store_name"].(string)
	visibilityRefs := ref["visibility"].(*pluginsdk.Set)
	var visibility []batch.CertificateVisibility
	if visibilityRefs != nil {
		for _, visibilityRef := range visibilityRefs.List() {
			visibility = append(visibility, batch.CertificateVisibility(visibilityRef.(string)))
		}
	}

	certificateReference := &batch.CertificateReference{
		ID:            &id,
		StoreLocation: batch.CertificateStoreLocation(storeLocation),
		StoreName:     &storeName,
		Visibility:    &visibility,
	}
	return certificateReference, nil
}

// ExpandBatchPoolStartTask expands Batch pool start task
func ExpandBatchPoolStartTask(list []interface{}) (*batch.StartTask, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("batch pool start task should be defined")
	}

	startTaskValue := list[0].(map[string]interface{})

	startTaskCmdLine := startTaskValue["command_line"].(string)

	maxTaskRetryCount := int32(1)

	if v := startTaskValue["task_retry_maximum"].(int); v > 0 {
		maxTaskRetryCount = int32(v)
	}

	waitForSuccess := startTaskValue["wait_for_success"].(bool)

	userIdentityList := startTaskValue["user_identity"].([]interface{})
	if len(userIdentityList) == 0 {
		return nil, fmt.Errorf("batch pool start task user identity should be defined")
	}

	userIdentityValue := userIdentityList[0].(map[string]interface{})
	userIdentity := batch.UserIdentity{}

	if autoUserValue, ok := userIdentityValue["auto_user"]; ok {
		autoUser := autoUserValue.([]interface{})
		if len(autoUser) != 0 {
			autoUserMap := autoUser[0].(map[string]interface{})
			userIdentity.AutoUser = &batch.AutoUserSpecification{
				ElevationLevel: batch.ElevationLevel(autoUserMap["elevation_level"].(string)),
				Scope:          batch.AutoUserScope(autoUserMap["scope"].(string)),
			}
		}
	}
	if userNameValue, ok := userIdentityValue["user_name"]; ok {
		userName := userNameValue.(string)
		if len(userName) != 0 {
			userIdentity.UserName = &userName
		}
	}

	resourceFileList := startTaskValue["resource_file"].([]interface{})
	resourceFiles := make([]batch.ResourceFile, 0)
	for _, resourceFileValueTemp := range resourceFileList {
		if resourceFileValueTemp == nil {
			continue
		}
		resourceFileValue := resourceFileValueTemp.(map[string]interface{})
		resourceFile := batch.ResourceFile{}
		if v, ok := resourceFileValue["auto_storage_container_name"]; ok {
			autoStorageContainerName := v.(string)
			if autoStorageContainerName != "" {
				resourceFile.AutoStorageContainerName = &autoStorageContainerName
			}
		}
		if v, ok := resourceFileValue["storage_container_url"]; ok {
			storageContainerURL := v.(string)
			if storageContainerURL != "" {
				resourceFile.StorageContainerURL = &storageContainerURL
			}
		}
		if v, ok := resourceFileValue["http_url"]; ok {
			httpURL := v.(string)
			if httpURL != "" {
				resourceFile.HTTPURL = &httpURL
			}
		}
		if v, ok := resourceFileValue["blob_prefix"]; ok {
			blobPrefix := v.(string)
			if blobPrefix != "" {
				resourceFile.BlobPrefix = &blobPrefix
			}
		}
		if v, ok := resourceFileValue["file_path"]; ok {
			filePath := v.(string)
			if filePath != "" {
				resourceFile.FilePath = &filePath
			}
		}
		if v, ok := resourceFileValue["file_mode"]; ok {
			fileMode := v.(string)
			if fileMode != "" {
				resourceFile.FileMode = &fileMode
			}
		}
		resourceFiles = append(resourceFiles, resourceFile)
	}

	startTask := &batch.StartTask{
		CommandLine:       &startTaskCmdLine,
		MaxTaskRetryCount: &maxTaskRetryCount,
		WaitForSuccess:    &waitForSuccess,
		UserIdentity:      &userIdentity,
		ResourceFiles:     &resourceFiles,
	}

	if v := startTaskValue["common_environment_properties"].(map[string]interface{}); len(v) > 0 {
		startTask.EnvironmentSettings = expandCommonEnvironmentProperties(v)
	}

	return startTask, nil
}

func expandCommonEnvironmentProperties(env map[string]interface{}) *[]batch.EnvironmentSetting {
	envSettings := make([]batch.EnvironmentSetting, 0)

	for k, v := range env {
		theValue := v.(string)
		theKey := k
		envSetting := batch.EnvironmentSetting{
			Name:  &theKey,
			Value: &theValue,
		}

		envSettings = append(envSettings, envSetting)
	}
	return &envSettings
}

// ExpandBatchMetaData expands Batch pool metadata
func ExpandBatchMetaData(input map[string]interface{}) *[]batch.MetadataItem {
	output := []batch.MetadataItem{}

	for k, v := range input {
		name := k
		value := v.(string)
		output = append(output, batch.MetadataItem{
			Name:  &name,
			Value: &value,
		})
	}

	return &output
}

// FlattenBatchMetaData flattens a Batch pool metadata
func FlattenBatchMetaData(metadatas *[]batch.MetadataItem) map[string]interface{} {
	output := make(map[string]interface{})

	if metadatas == nil {
		return output
	}

	for _, metadata := range *metadatas {
		if metadata.Name == nil || metadata.Value == nil {
			continue
		}

		output[*metadata.Name] = *metadata.Value
	}

	return output
}

// ExpandBatchPoolNetworkConfiguration expands Batch pool network configuration
func ExpandBatchPoolNetworkConfiguration(list []interface{}) (*batch.NetworkConfiguration, error) {
	if len(list) == 0 {
		return nil, nil
	}

	networkConfigValue := list[0].(map[string]interface{})
	networkConfiguration := &batch.NetworkConfiguration{}

	if v, ok := networkConfigValue["subnet_id"]; ok {
		if value := v.(string); value != "" {
			networkConfiguration.SubnetID = &value
		}
	}

	if v, ok := networkConfigValue["public_ips"]; ok {
		if networkConfiguration.PublicIPAddressConfiguration == nil {
			networkConfiguration.PublicIPAddressConfiguration = &batch.PublicIPAddressConfiguration{}
		}

		publicIPsRaw := v.(*pluginsdk.Set).List()
		networkConfiguration.PublicIPAddressConfiguration.IPAddressIds = utils.ExpandStringSlice(publicIPsRaw)
	}

	if v, ok := networkConfigValue["endpoint_configuration"]; ok {
		endpoint, err := expandPoolEndpointConfiguration(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		networkConfiguration.EndpointConfiguration = endpoint
	}

	if v, ok := networkConfigValue["public_address_provisioning_type"]; ok {
		if networkConfiguration.PublicIPAddressConfiguration == nil {
			networkConfiguration.PublicIPAddressConfiguration = &batch.PublicIPAddressConfiguration{}
		}

		if value := v.(string); value != "" {
			networkConfiguration.PublicIPAddressConfiguration.Provision = batch.IPAddressProvisioningType(value)
		}
	}

	return networkConfiguration, nil
}

func expandPoolEndpointConfiguration(list []interface{}) (*batch.PoolEndpointConfiguration, error) {
	if len(list) == 0 {
		return nil, nil
	}

	inboundNatPools := make([]batch.InboundNatPool, len(list))

	for i, inboundNatPoolsValue := range list {
		inboundNatPool := inboundNatPoolsValue.(map[string]interface{})

		name := inboundNatPool["name"].(string)
		protocol := batch.InboundEndpointProtocol(inboundNatPool["protocol"].(string))
		backendPort := int32(inboundNatPool["backend_port"].(int))
		frontendPortRange := inboundNatPool["frontend_port_range"].(string)
		parts := strings.Split(frontendPortRange, "-")
		frontendPortRangeStart, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		frontendPortRangeEnd, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}

		networkSecurityGroupRules := expandPoolNetworkSecurityGroupRule(inboundNatPool["network_security_group_rules"].([]interface{}))

		inboundNatPools[i] = batch.InboundNatPool{
			Name:                      &name,
			Protocol:                  protocol,
			BackendPort:               &backendPort,
			FrontendPortRangeStart:    utils.Int32(int32(frontendPortRangeStart)),
			FrontendPortRangeEnd:      utils.Int32(int32(frontendPortRangeEnd)),
			NetworkSecurityGroupRules: &networkSecurityGroupRules,
		}
	}

	return &batch.PoolEndpointConfiguration{
		InboundNatPools: &inboundNatPools,
	}, nil
}

func expandPoolNetworkSecurityGroupRule(list []interface{}) []batch.NetworkSecurityGroupRule {
	if len(list) == 0 {
		return []batch.NetworkSecurityGroupRule{}
	}

	networkSecurityGroupRule := make([]batch.NetworkSecurityGroupRule, 0)
	for _, groupRule := range list {
		groupRuleMap := groupRule.(map[string]interface{})

		priority := int32(groupRuleMap["priority"].(int))
		sourceAddressPrefix := groupRuleMap["source_address_prefix"].(string)
		access := batch.NetworkSecurityGroupRuleAccess(groupRuleMap["access"].(string))

		networkSecurityGroupRule = append(networkSecurityGroupRule, batch.NetworkSecurityGroupRule{
			Priority:            &priority,
			SourceAddressPrefix: &sourceAddressPrefix,
			Access:              access,
		})
	}

	return networkSecurityGroupRule
}

func flattenBatchPoolNetworkConfiguration(input *batch.NetworkConfiguration) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	subnetId := ""
	if input.SubnetID != nil {
		subnetId = *input.SubnetID
	}

	publicIPAddressIds := make([]interface{}, 0)
	publicAddressProvisioningType := ""
	if config := input.PublicIPAddressConfiguration; config != nil {
		publicIPAddressIds = utils.FlattenStringSlice(config.IPAddressIds)
		publicAddressProvisioningType = string(config.Provision)
	}

	endpointConfigs := make([]interface{}, 0)
	if config := input.EndpointConfiguration; config != nil && config.InboundNatPools != nil {
		for _, inboundNatPool := range *config.InboundNatPools {
			name := ""
			if inboundNatPool.Name != nil {
				name = *inboundNatPool.Name
			}

			backendPort := 0
			if inboundNatPool.BackendPort != nil {
				backendPort = int(*inboundNatPool.BackendPort)
			}

			frontendPortRange := ""
			if inboundNatPool.FrontendPortRangeStart != nil && inboundNatPool.FrontendPortRangeEnd != nil {
				frontendPortRange = fmt.Sprintf("%d-%d", *inboundNatPool.FrontendPortRangeStart, *inboundNatPool.FrontendPortRangeEnd)
			}

			networkSecurities := make([]interface{}, 0)
			if sgRules := inboundNatPool.NetworkSecurityGroupRules; sgRules != nil {
				for _, networkSecurity := range *sgRules {
					priority := 0
					if networkSecurity.Priority != nil {
						priority = int(*networkSecurity.Priority)
					}
					sourceAddressPrefix := ""
					if networkSecurity.SourceAddressPrefix != nil {
						sourceAddressPrefix = *networkSecurity.SourceAddressPrefix
					}
					networkSecurities = append(networkSecurities, map[string]interface{}{
						"access":                string(networkSecurity.Access),
						"priority":              priority,
						"source_address_prefix": sourceAddressPrefix,
					})
				}
			}

			endpointConfigs = append(endpointConfigs, map[string]interface{}{
				"backend_port":                 backendPort,
				"frontend_port_range":          frontendPortRange,
				"name":                         name,
				"network_security_group_rules": networkSecurities,
				"protocol":                     string(inboundNatPool.Protocol),
			})
		}
	}

	return []interface{}{
		map[string]interface{}{
			"endpoint_configuration":           endpointConfigs,
			"public_address_provisioning_type": publicAddressProvisioningType,
			"public_ips":                       pluginsdk.NewSet(pluginsdk.HashString, publicIPAddressIds),
			"subnet_id":                        subnetId,
		},
	}
}
