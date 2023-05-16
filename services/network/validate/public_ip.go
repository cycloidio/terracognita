package validate

import (
	"fmt"
	"regexp"
)

func PublicIpDomainNameLabel(v interface{}, k string) (warnings []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[a-z][a-z0-9-]{1,61}[a-z0-9]$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%s must contain only lowercase alphanumeric characters, numbers and hyphens. It must start with a letter, end only with a number or letter and not exceed 63 characters in length", k))
	}
	return warnings, errors
}
