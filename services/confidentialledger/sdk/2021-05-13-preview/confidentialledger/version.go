package confidentialledger

import "fmt"

const defaultApiVersion = "2021-05-13-preview"

func userAgent() string {
	return fmt.Sprintf("pandora/confidentialledger/%s", defaultApiVersion)
}
