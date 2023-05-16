package rules

import "fmt"

const defaultApiVersion = "2020-07-01"

func userAgent() string {
	return fmt.Sprintf("pandora/rules/%s", defaultApiVersion)
}
