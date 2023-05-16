package validate

import (
	"fmt"

	networkParse "github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
)

func FirewallSubnetName(v interface{}, k string) (warnings []string, errors []error) {
	parsed, err := networkParse.SubnetID(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("parsing %q: %+v", v.(string), err))
		return warnings, errors
	}

	if parsed.Name != "AzureFirewallSubnet" {
		errors = append(errors, fmt.Errorf("The name of the Subnet for %q must be exactly 'AzureFirewallSubnet' to be used for the Azure Firewall resource", k))
	}

	return warnings, errors
}
