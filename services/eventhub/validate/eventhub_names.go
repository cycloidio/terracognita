package validate

import (
	"regexp"

	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
)

// validation
func ValidateEventHubNamespaceName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile("^[a-zA-Z][-a-zA-Z0-9]{4,48}[a-zA-Z0-9]$"),
		"The namespace name can contain only letters, numbers and hyphens. The namespace must start with a letter, and it must end with a letter or number and be between 6 and 50 characters long.",
	)
}

func ValidateEventHubName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile("^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,254}[a-zA-Z0-9])?$"),
		"The event hub name can contain only letters, numbers, periods (.), hyphens (-),and underscores (_), up to 256 characters, and it must begin and end with a letter or number.",
	)
}

func ValidateEventHubConsumerName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile("^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,48}[a-zA-Z0-9])?$"),
		"The consumer group name can contain only letters, numbers, periods (.), hyphens (-),and underscores (_), up to 50 characters, and it must begin and end with a letter or number.",
	)
}

func ValidateEventHubAuthorizationRuleName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile("^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,58}[a-zA-Z0-9])?$"),
		"The authorization rule name can contain only letters, numbers, periods, hyphens and underscores. The name must start and end with a letter or number and be up to 60 characters long.",
	)
}
