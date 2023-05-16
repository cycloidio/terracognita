package network

import "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"

func FindNetworkInterfaceIPConfiguration(input *[]network.InterfaceIPConfiguration, name string) *network.InterfaceIPConfiguration {
	if input == nil {
		return nil
	}

	for _, v := range *input {
		if v.Name == nil {
			continue
		}

		if *v.Name == name {
			return &v
		}
	}

	return nil
}

func updateNetworkInterfaceIPConfiguration(config network.InterfaceIPConfiguration, configs *[]network.InterfaceIPConfiguration) *[]network.InterfaceIPConfiguration {
	output := make([]network.InterfaceIPConfiguration, 0)
	if configs == nil {
		return &output
	}

	for _, v := range *configs {
		if v.Name == nil {
			continue
		}

		if *v.Name != *config.Name {
			output = append(output, v)
		} else {
			output = append(output, config)
		}
	}

	return &output
}
