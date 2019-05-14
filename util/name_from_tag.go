package util

import (
	"github.com/chr4/pwgen"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/helper/schema"
)

// GetNameFromTag returns the 'tags.Name' from the src or the fallback
// if it's not defined.
// Also validates that the 'tags.Name' and fallback are valid, if not it
// generates a random one
func GetNameFromTag(srd *schema.ResourceData, fallback string) string {
	var n string
	if name, ok := srd.GetOk("tags.Name"); ok {
		n = name.(string)
	}

	if isValidResourceName(n) {
		return n
	} else if isValidResourceName(fallback) {
		return fallback
	} else {
		return pwgen.Alpha(5)
	}
}

// isValidResourceName checks with the TF regex
// for names to validate if it's valid
func isValidResourceName(name string) bool {
	return config.NameRegexp.Match([]byte(name))
}
