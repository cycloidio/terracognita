package privatelinkresources

import "fmt"

const defaultApiVersion = "2020-06-01"

func userAgent() string {
	return fmt.Sprintf("pandora/privatelinkresources/%s", defaultApiVersion)
}
