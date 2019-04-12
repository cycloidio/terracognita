package util

import "github.com/hashicorp/terraform/helper/schema"

func GetNameFromTag(srd *schema.ResourceData, fallback string) string {
	if name, ok := srd.GetOk("tags.Name"); ok {
		return name.(string)
	}

	return fallback
}
