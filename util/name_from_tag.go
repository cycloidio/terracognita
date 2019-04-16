package util

import "github.com/hashicorp/terraform/helper/schema"

// GetNameFromTag returns the 'tags.Name' from the src or the fallback
// if it's not defined
func GetNameFromTag(srd *schema.ResourceData, fallback string) string {
	if name, ok := srd.GetOk("tags.Name"); ok {
		return name.(string)
	}

	return fallback
}
