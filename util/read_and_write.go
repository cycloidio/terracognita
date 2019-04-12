package util

import (
	"fmt"

	"github.com/cycloidio/terraforming/util/writer"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	tfaws "github.com/terraform-providers/terraform-provider-aws/aws"
)

func ReadIDsAndWrite(tfAWSClient interface{}, provider, resourceType string, tags []Tag, state bool, ids []string, w writer.Writer) error {
	for _, id := range ids {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap[resourceType]
		srd := resource.Data(nil)
		srd.SetId(id)
		srd.SetType(resourceType)

		// Some resources can not be filtered by tags,
		// so we have to do it manually
		// it's not all of them though
		for _, t := range tags {
			if srd.Get(fmt.Sprintf("tags.%s", t.Name)).(string) != t.Value {
				continue
			}
		}

		err := resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		// It can be imported
		if state {
			if importer := resource.Importer; importer != nil {
				resourceDatas, err := importer.State(srd, tfAWSClient)
				if err != nil {
					return err
				}
				// TODO: The multple return could potentially be the `depends_on` of the
				// terraform.ResourceState
				// Investigate on SG
				for i, rd := range resourceDatas {
					if i != 0 {
						// for now we scape all the other ones
						// the firs one is the one we need and
						// for what've see the others are
						// 'aws_security_group_rules' (in aws)
						// and are not imported by default by
						// Terraform
						continue
					}
					name := GetNameFromTag(rd, id)

					tis := rd.State()
					trs := &terraform.ResourceState{
						Type:     resourceType,
						Primary:  tis,
						Provider: provider,
					}
					if err := w.Write(fmt.Sprintf("%s.%s", tis.Ephemeral.Type, name), trs); err != nil && errors.Cause(err) == writer.ErrAlreadyExistsKey {
						err = w.Write(fmt.Sprintf("%s.%s", tis.Ephemeral.Type, tis.ID), trs)
						if err != nil {
							return err
						}
					}
				}
			} else {
				// If it can not be imported continue
				continue
			}
		} else {
			name := GetNameFromTag(srd, id)

			if err := w.Write(fmt.Sprintf("%s.%s", resourceType, name), mergeFullConfig(srd, resource.Schema, "")); err != nil && errors.Cause(err) == writer.ErrAlreadyExistsKey {
				err = w.Write(fmt.Sprintf("%s.%s", resourceType, id), mergeFullConfig(srd, resource.Schema, ""))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// mergeFullConfig creates the key to the map and if it had a value before set it, if
func mergeFullConfig(cfgr *schema.ResourceData, sch map[string]*schema.Schema, key string) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range sch {
		// If it's just a Computed value, do not add it to the output
		if v.Computed && !v.Optional && !v.Required {
			continue
		}

		// Basically calculates the needed
		// key to the current access
		kk := key
		if key != "" {
			kk = key + "." + k
		} else {
			kk = k
		}

		// schema.Resource means that it has nested fields
		if sr, ok := v.Elem.(*schema.Resource); ok {
			if v.Type == schema.TypeSet || v.Type == schema.TypeList {
				ar, ok := res[k]
				if !ok {
					ar = make([]interface{}, 0, 0)
				}

				list, ok := cfgr.GetOk(kk)
				if !ok {
					continue
				}
				if list != nil {
					// For the types that are a list, we have to set them in an array, and also
					// add the correct index for the number of setts (entries on the original config)
					// that there are on the provided configuration
					switch val := list.(type) {
					case []map[string]interface{}:
						for i := range val {
							ar = append(ar.([]interface{}), mergeFullConfig(cfgr, sr.Schema, fmt.Sprintf("%s.%d", kk, i)))
						}
					case []interface{}:
						for i := range val {
							ar = append(ar.([]interface{}), mergeFullConfig(cfgr, sr.Schema, fmt.Sprintf("%s.%d", kk, i)))
						}
					}
				} else {
					ar = append(ar.([]interface{}), mergeFullConfig(cfgr, sr.Schema, kk))
				}
				res[k] = ar
			} else {
				res[k] = mergeFullConfig(cfgr, sr.Schema, kk)
			}
			// As it's a nested element it does not require any of
			// the other code as it's for singel value schemas
			continue
		}

		vv, ok := cfgr.GetOk(kk)
		if !ok || vv == nil {
			continue
		}

		if s, ok := vv.(*schema.Set); ok {
			res[k] = s.List()
		} else {
			res[k] = vv
		}
	}
	return res
}
