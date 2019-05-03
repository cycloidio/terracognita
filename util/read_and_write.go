package util

import (
	"fmt"
	"reflect"

	"github.com/cycloidio/terraforming/util/writer"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	tfaws "github.com/terraform-providers/terraform-provider-aws/aws"
)

type ResourceDataFn func(*schema.ResourceData) error

// ReadIDsAndWrite reades the config from the provider+resourceType ids and writes to w the state or the HCL
func ReadIDsAndWrite(tfAWSClient interface{}, provider, resourceType string, tags []Tag, state bool, ids []string, rdfn ResourceDataFn, w writer.Writer) error {
	for _, id := range ids {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap[resourceType]
		srd := resource.Data(nil)
		srd.SetId(id)
		srd.SetType(resourceType)

		if rdfn != nil {
			err := rdfn(srd)
			if err != nil {
				return err
			}
		}

		// Some resources can not be filtered by tags,
		// so we have to do it manually
		// it's not all of them though
		for _, t := range tags {
			if v, ok := srd.GetOk(fmt.Sprintf("tags.%s", t.Name)); ok && v.(string) != t.Value {
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
					if tis == nil {
						// IDK why some times it does not have
						// the ID (the only case in tis it's nil)
						// so if nil we don't need it
						continue
					}
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
			// Example would be aws_security_group
			if v.Type == schema.TypeSet {
				s, ok := cfgr.GetOk(kk)
				if !ok {
					continue
				}

				// Was the default from Set/List
				//ar = append(ar.([]interface{}), mergeFullConfig(cfgr, sr.Schema, kk))

				res[k] = normalizSet(s.(*schema.Set))
			} else if v.Type == schema.TypeList {
				var ar interface{}
				ar = make([]interface{}, 0, 0)

				l, ok := cfgr.GetOk(kk)
				if !ok {
					continue
				}

				list := l.([]interface{})
				for i := range list {
					ar = append(ar.([]interface{}), mergeFullConfig(cfgr, sr.Schema, fmt.Sprintf("%s.%d", kk, i)))
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

// normalizSet returns the normalization of a schema.Set
// it could be a simple list or a embedded structure
func normalizSet(ss *schema.Set) interface{} {
	var ar interface{}
	ar = make([]interface{}, 0, 0)

	setList := ss.List()
	for _, set := range setList {
		switch val := set.(type) {
		case map[string]interface{}:
			// This case it's when a TypeSet has
			// a nested structure,
			// example: aws_security_group.ingress
			res := make(map[string]interface{})
			for k, v := range val {
				if s, ok := v.(*schema.Set); ok {
					ns := normalizSet(s)
					if !isDefault(ns) {
						res[k] = ns
					}
				} else {
					if !isDefault(v) {
						res[k] = v
					}
				}
			}
			ar = append(ar.([]interface{}), res)
		case interface{}:
			// This case is normally for the
			// "Type: schema.TypeSet, Elm: schema.Schema{Type: schema.TypeString}"
			// definitions on TF,
			// example: aws_security_group.ingress.security_groups
			ar = append(ar.([]interface{}), val)
		}
	}

	return ar
}

var (
	// Ideally this could be generated using "enumer", it
	// would be a better idea as then we do not have
	// to maintain this list
	tfTypes = []schema.ValueType{
		schema.TypeBool,
		schema.TypeInt,
		schema.TypeFloat,
		schema.TypeString,
		schema.TypeList,
		schema.TypeMap,
		schema.TypeSet,
	}
)

// isDefault is used on normalizSet as the Sets do not use the normal
// TF strucure (access by key) and are stored as raw maps with some
// default values that we don't want on the HCL output.
// example: [], false, "", 0 ...
func isDefault(v interface{}) bool {
	for _, t := range tfTypes {
		if reflect.DeepEqual(t.Zero(), v) {
			return true
		}
	}
	return false
}
