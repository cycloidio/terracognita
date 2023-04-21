package hcl

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strings"

	kitlog "github.com/go-kit/kit/log"

	"github.com/cycloidio/mxwriter"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/interpolator"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/util"
	"github.com/cycloidio/terracognita/writer"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	cjson "github.com/zclconf/go-cty/cty/json"
)

const (
	defaultCategory      = "hcl"
	variablesCategoryKey = "variables"
	isMap                = true
)

// Writer is a Writer implementation that writes to
// a static map to then transform it to HCL
type Writer struct {
	// The keys are Config[<category>]["resources"][<resource_type>]
	Config map[string]map[string]interface{}
	// Keeps the order of the Config keys
	// so we can always expect the same order
	categories []string
	writer     io.Writer
	opts       *writer.Options
	provider   provider.Provider
}

// NewWriter rerturns an Writer initialization
func NewWriter(w io.Writer, pv provider.Provider, opts *writer.Options) *Writer {
	cfg := make(map[string]map[string]interface{})

	wr := &Writer{
		Config:   cfg,
		writer:   w,
		opts:     opts,
		provider: pv,
	}

	tfcfg := map[string]interface{}{
		"required_version": ">= 1.0",
		"required_providers": map[string]interface{}{
			// We use the =tc= prefix as we want this to be an
			// object attribute and not a block. By default
			// on the formater we have we replace all the '= {` for
			// just '{' so this would be included too and it would
			// be invalid configuration
			fmt.Sprintf("=tc=%s", pv.String()): map[string]interface{}{
				"source": pv.Source(),
			},
		},
	}
	var cat string
	if opts.HasModule() {
		cat = writer.ModuleCategoryKey
		wr.Config[cat] = map[string]interface{}{
			"module": map[string]interface{}{
				opts.Module: map[string]interface{}{
					"source": fmt.Sprintf("./module-%s", opts.Module),
				},
			},
		}
	} else {
		cat = defaultCategory
		wr.Config[cat] = make(map[string]interface{})
		wr.Config[cat]["resource"] = make(map[string]map[string]interface{})
		wr.categories = append(wr.categories, cat)
	}

	// If no option is given to write Terraform block elsewhere
	// write it into the file of the module
	tfKey := cat
	if opts.TerraformCategoryKey != "" {
		tfKey = opts.TerraformCategoryKey
		wr.Config[tfKey] = make(map[string]interface{})
		wr.categories = append(wr.categories, tfKey)
	}
	wr.Config[tfKey]["terraform"] = tfcfg

	if opts.HCLProviderBlock {
		pvcfg := map[string]interface{}{
			pv.String(): make(map[string]interface{}),
		}
		wr.Config[tfKey]["provider"] = pvcfg
		wr.setProviderConfig(tfKey)
	}

	return wr
}

// Write expects a key similar to "aws_instance.your_name"
// repeated keys will report an error
func (w *Writer) Write(key string, value interface{}) error {
	if key == "" {
		return errcode.ErrWriterRequiredKey
	}

	if value == nil {
		return errcode.ErrWriterRequiredValue
	}

	keys := strings.Split(key, ".")
	if len(keys) != 2 || (keys[0] == "" || keys[1] == "") {
		return errors.Wrapf(errcode.ErrWriterInvalidKey, "with key %q", key)
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return errors.Wrap(errcode.ErrWriterInvalidTypeValue, "we expect the value to be a map[string]interface{}")
	}

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	var category string
	ic, ok := m[writer.ResourceCategoryKey]
	if !ok {
		category = defaultCategory
	} else {
		category = ic.(string)
	}

	if _, ok := w.Config[category]; !ok {
		w.Config[category] = make(map[string]interface{})
		w.Config[category]["resource"] = make(map[string]map[string]interface{})
		w.categories = append(w.categories, category)
	}

	name := strings.Join(keys[1:], "")

	if _, ok := w.Config[category]["resource"].(map[string]map[string]interface{})[keys[0]][name]; ok {
		return errors.Wrapf(errcode.ErrWriterAlreadyExistsKey, "with key %q", key)
	}

	if _, ok := w.Config[category]["resource"].(map[string]map[string]interface{})[keys[0]]; !ok {
		w.Config[category]["resource"].(map[string]map[string]interface{})[keys[0]] = make(map[string]interface{})
	}
	log.Get().Log("func", "writer.Write(HCL)", "msg", "writing to internal config", "key", keys[0], "content", string(b))

	w.Config[category]["resource"].(map[string]map[string]interface{})[keys[0]][name] = value

	return nil
}

// Has checks if the given key is already present or not
func (w *Writer) Has(key string) (bool, error) {
	keys := strings.Split(key, ".")
	if len(keys) != 2 || keys[0] == "" || keys[1] == "" {
		return false, errors.Wrapf(errcode.ErrWriterInvalidKey, "with key %q", key)
	}

	name := strings.Join(keys[1:], "")

	for k, v := range w.Config {
		if k == writer.ModuleCategoryKey || k == variablesCategoryKey || k == w.opts.TerraformCategoryKey {
			continue
		}
		if _, ok := v["resource"].(map[string]map[string]interface{})[keys[0]][name]; ok {
			return true, nil
		}
	}

	return false, nil
}

// Sync writes the content of the Config to the
// internal w with the correct format
func (w *Writer) Sync() error {
	logger := log.Get()
	logger = kitlog.With(logger, "func", "writer.Write(HCL)")

	categories := w.categories
	if w.opts.HasModule() {
		categories = append(categories, []string{writer.ModuleCategoryKey, variablesCategoryKey}...)
		w.setVariables()
	}

	for _, category := range categories {
		f := hclwrite.NewEmptyFile()
		body := f.Body()
		cfg, ok := w.Config[category]
		if !ok {
			continue
		}

		src, err := json.Marshal(cfg)
		if err != nil {
			return errors.Wrap(err, "unable to marshal JSON config")
		}

		t, err := cjson.ImpliedType(src)
		if err != nil {
			return errors.Wrap(err, "unable to get cty.Type from config")
		}
		v, err := cjson.Unmarshal(src, t)
		if err != nil {
			return errors.Wrap(err, "unable to get cty.Value from cty.Type and JSON config")
		}

		blockKeys := getValueKeys(v)
		if len(blockKeys) == 0 {
			continue
		}
		blockMap := v.AsValueMap()
		// blockType can be "resource", "variable", "output", "provider" etc
		for _, blockType := range blockKeys {
			blockValue := blockMap[blockType]
			if blockType == "terraform" {
				block := hclwrite.NewBlock(blockType, nil)
				bbody := block.Body()

				attrKeys := getValueKeys(blockValue)
				if len(attrKeys) == 0 {
					// This will allow to declare empty blocks like
					// empty variable definitions
					body.AppendBlock(block)
					body.AppendNewline()
					continue
				}
				attrMap := blockValue.AsValueMap()
				for _, attr := range attrKeys {
					bbody.SetAttributeValue(attr, attrMap[attr])
				}

				body.AppendBlock(block)
				body.AppendNewline()
				continue
			}

			resourceKeys := getValueKeys(blockValue)
			if len(resourceKeys) == 0 {
				continue
			}
			resourceMap := blockValue.AsValueMap()
			// resourceType is the type of the resource (e.g: `aws_security_groups`)
			for _, resourceType := range resourceKeys {
				resources := resourceMap[resourceType]
				if blockType == "variable" || blockType == "module" || blockType == "provider" {
					block := hclwrite.NewBlock(blockType, []string{resourceType})
					bbody := block.Body()

					attrKeys := getValueKeys(resources)
					if len(attrKeys) == 0 {
						// This will allow to declare empty blocks like
						// empty variable definitions
						body.AppendBlock(block)
						body.AppendNewline()
						continue
					}
					attrMap := resources.AsValueMap()
					for _, attr := range attrKeys {
						bbody.SetAttributeValue(attr, attrMap[attr])
					}

					body.AppendBlock(block)
					body.AppendNewline()
					continue
				}
				resourceKeys := getValueKeys(resources)
				if len(resourceKeys) == 0 {
					continue
				}
				resourceMap := resources.AsValueMap()
				for _, name := range resourceKeys {
					resource := resourceMap[name]
					block := hclwrite.NewBlock(blockType, []string{resourceType, name})
					bbody := block.Body()

					attrKeys := getValueKeys(resource)
					if len(attrKeys) == 0 {
						continue
					}
					attrMap := resource.AsValueMap()
					for _, attr := range attrKeys {
						// We do not want to print on the HCL the
						// resource category as it's just for
						// internal usage
						if attr == writer.ResourceCategoryKey {
							continue
						}
						value := attrMap[attr]
						// in JSON representation, we can have a list of object
						// e.g with ingress:[{ingress1}, {ingress2}, ... {ingressN}]
						// we need to add a dedicated block for each object instead of having
						// one block for the whole list
						if value.Type().IsTupleType() {
							writeTuple(body, bbody, attr, value)
						} else {
							bbody.SetAttributeValue(attr, value)
						}
					}
					body.AppendBlock(block)
					body.AppendNewline()
				}
			}
		}

		// we don't use the file.WriteTo method because we need to use
		// our own Format method before writing to the writer
		formattedBytes := Format(f.Bytes())
		formattedBytes = hclwrite.Format(formattedBytes)
		mxwriter.Write(w.writer, category, formattedBytes)
	}

	return nil
}

// getValueKeys will return a sorted list of the keys the val has
// so then they can be used to access maps without having to deal
// with random order which messes the output and would generate
// diffs between generations that are not true
func getValueKeys(val cty.Value) []string {
	if !val.CanIterateElements() {
		return nil
	}
	keys := make([]string, 0, val.LengthInt())
	for it := val.ElementIterator(); it.Next(); {
		k, _ := it.Element()
		keys = append(keys, k.AsString())
	}

	sort.Strings(keys)

	return keys
}

// setVariables will replace all the values for variables or just the ones ModuleVariables
// if it has been defined
func (w *Writer) setVariables() {
	variables := make(map[string]interface{})
	for c, cfg := range w.Config {
		if c == writer.ModuleCategoryKey || c == variablesCategoryKey || c == w.opts.TerraformCategoryKey {
			continue
		}
		for k, v := range cfg["resource"].(map[string]map[string]interface{}) {
			cfg["resource"].(map[string]map[string]interface{})[k] = walkVariables(v, w.opts.ModuleVariables, k, variables)
		}
	}
	w.Config[variablesCategoryKey] = map[string]interface{}{
		"variable": variables,
	}

	for k, v := range variables {
		msi := v.(map[string]interface{})
		if d, ok := msi["default"]; ok {
			w.Config[writer.ModuleCategoryKey]["module"].(map[string]interface{})[w.opts.Module].(map[string]interface{})[k] = d
		} else if d, ok := msi["=tc=default"]; ok {
			w.Config[writer.ModuleCategoryKey]["module"].(map[string]interface{})[w.opts.Module].(map[string]interface{})[fmt.Sprintf("=tc=%s", k)] = d
		}
	}
}

// walkVariables will walk the cfg until it reached the last elements, the k is the current key (as it's recursive can be aws_lb.ingress.from_port)
// variables is the map of all the variables assigned. It returns the new cfg with the variable interpolation.
// If the validVariables is not empty only those will be used as variables, if not all the attributes will be converted in variables
func walkVariables(cfg map[string]interface{}, validVariables map[string]struct{}, k string, variables map[string]interface{}) map[string]interface{} {
	for key, value := range cfg {
		currentKey := fmt.Sprintf("%s.%s", k, key)
		switch v := value.(type) {
		case map[string]interface{}:
			if ok, nk := hasKey(validVariables, currentKey, isMap); ok {
				varName := util.NormalizeName(strings.ReplaceAll(nk, ".", "_"))
				prefix := ""
				if strings.Contains(currentKey, "=tc=") {
					prefix = "=tc="
				}
				variables[varName] = map[string]interface{}{
					fmt.Sprintf("%sdefault", prefix): cfg[key],
				}
				cfg[key] = fmt.Sprintf("${var.%s}", varName)
			} else {
				cfg[key] = walkVariables(v, validVariables, currentKey, variables)
			}
		case []interface{}:
			if len(v) == 0 {
				if ok, nk := hasKey(validVariables, currentKey, !isMap); ok {
					varName := util.NormalizeName(strings.ReplaceAll(nk, ".", "_"))
					prefix := ""
					if strings.Contains(currentKey, "=tc=") {
						prefix = "=tc="
					}
					variables[varName] = map[string]interface{}{
						fmt.Sprintf("%sdefault", prefix): cfg[key],
					}
					cfg[key] = fmt.Sprintf("${var.%s}", varName)
				}
				continue
			}
			// For slices we need to check the first element, if it's a map then
			// it has complex data, if not it's a "simple" slice of values
			if _, ok := v[0].(map[string]interface{}); ok {
				for i, vvv := range v {
					v[i] = walkVariables(vvv.(map[string]interface{}), validVariables, fmt.Sprintf("%s.%d", currentKey, i), variables)
				}
			} else {
				if ok, nk := hasKey(validVariables, currentKey, !isMap); ok {
					varName := util.NormalizeName(strings.ReplaceAll(nk, ".", "_"))
					prefix := ""
					if strings.Contains(currentKey, "=tc=") {
						prefix = "=tc="
					}
					variables[varName] = map[string]interface{}{
						fmt.Sprintf("%sdefault", prefix): cfg[key],
					}
					cfg[key] = fmt.Sprintf("${var.%s}", varName)
				}
			}
		default:
			// This means is a "simple" value so we can
			// directly replace it with the variable
			if ok, nk := hasKey(validVariables, currentKey, !isMap); ok {
				varName := util.NormalizeName(strings.ReplaceAll(nk, ".", "_"))
				prefix := ""
				if strings.Contains(currentKey, "=tc=") {
					prefix = "=tc="
				}
				variables[varName] = map[string]interface{}{
					fmt.Sprintf("%sdefault", prefix): cfg[key],
				}
				cfg[key] = fmt.Sprintf("${var.%s}", varName)
			}
		}
	}
	return cfg
}

var (
	reIndexKey = regexp.MustCompile(`\.[\d]+\.`)
)

// hasKey will validate that the key is present on the map.
// The key will have the format: aws_instance.front.attr1.attr2...
// and the validVariables will not have the `front` interpolation, also
// it'll validate that if validVariables is empty it's always true
// It also returns the key to use in case is a special =tc= key
// The isMap is used to not return true if validVariables is empty, because then
// all the resources would not be written as all would be variables
func hasKey(validVariables map[string]struct{}, key string, isMap bool) (bool, string) {
	if len(validVariables) == 0 {
		return !isMap, key
	}

	sk := strings.Split(key, ".")
	// This remove the 'front' from 'aws_instance.front.attr1'
	k := append(sk[0:1], sk[2:]...)

	// If the key comes from an array it'll have the position on it
	// so it'll look something like `aws_instance.ebs_block_device.1.volume_size`
	// but the variable was defined as `aw_instance.ebs_block_device.volume_size`
	// so we have to strip all the indices if any
	nk := reIndexKey.ReplaceAllString(strings.Join(k, "."), ".")
	_, ok := validVariables[nk]

	if !ok && strings.Contains(key, "=tc=") {
		key = strings.Replace(key, "=tc=", "", -1)
		ok, key = hasKey(validVariables, key, isMap)
	}

	return ok, key
}

// writeTuple will write anything that's already been identified as IsTupleType
func writeTuple(pbody, body *hclwrite.Body, attr string, value cty.Value) {
	// When it's empty it'll not be printed
	// on the ElementIterator so it would be ignored
	// and if it's on this stage it's required to be
	// printed
	if value.LengthInt() == 0 {
		body.SetAttributeValue(attr, value)
	}
	iter := value.ElementIterator()
	for iter.Next() {
		_, val := iter.Element()
		if val.Type().IsPrimitiveType() {
			// the value is not an unconsistent object
			// it's basically [1,2,3]
			body.SetAttributeValue(attr, value)
			continue
		}

		newBody := body

		// only create a new body if we have a name for it
		if attr != "" {
			obj := body.AppendNewBlock(attr, nil)
			newBody = obj.Body()
		}

		iterator := val.ElementIterator()
		for iterator.Next() {
			kei, vei := iterator.Element()

			if vei.Type().IsTupleType() {
				// If it was an object it has to be
				// appended to the same one
				if val.Type().IsObjectType() {
					writeTuple(pbody, newBody, kei.AsString(), vei)
				} else {
					tobj := newBody.AppendNewBlock(kei.AsString(), nil)
					btobj := tobj.Body()
					writeTuple(newBody, btobj, "", vei)
				}
			} else {
				newBody.SetAttributeValue(kei.AsString(), vei)
			}
		}
	}
}

// Interpolate replaces the hardcoded resources link
// with TF interpolation.
func (w *Writer) Interpolate(i *interpolator.Interpolator) {
	// If Interpolation is disabled or
	// a module without specific attributes to
	// create as variables (so all are variables),
	// we ignore the Interpolation as this may happened and is invalid
	// variable "aws_ses_identity_notification_topic_XoqJb_identity" {
	//   default = aws_ses_domain_mail_from.SSWXE.id
	// }
	if !w.opts.Interpolate ||
		(w.opts.HasModule() && len(w.opts.ModuleVariables) == 0) {
		return
	}
	// who's interpolated with who
	relations := make(map[string]struct{}, 0)
	for k, v := range w.Config {
		if k == writer.ModuleCategoryKey || k == variablesCategoryKey || k == w.opts.TerraformCategoryKey {
			continue
		}
		resources := v["resource"]
		// we need to isolate each resource
		// getting each resource is easier to avoid cycle
		// or interpolation.
		// We first loop over resource type (e.g: aws_instance)
		for rt, resource := range resources.(map[string]map[string]interface{}) {
			// we loop over a resource (e.g: aws_instance.oDSOj)
			for name, block := range resource {
				src := reflect.ValueOf(block)

				// this will store the updated block
				dest := reflect.New(src.Type()).Elem()

				// walk through the resources to interpolate the good values
				w.walkInterpolation(dest, src, i, name, "", rt, relations)

				// remove reflect.Value wrapper from dest
				resources.(map[string]map[string]interface{})[rt][name] = dest.Interface()
			}
		}
	}
}

// walkInterpolation through a resource block. it's easier since we do not know how the block is made
// `dest` will be the new "block" with the values interpolated from `interpolate`
func (w *Writer) walkInterpolation(dest, src reflect.Value, interpolate *interpolator.Interpolator, name, key string, resourceType string, relations map[string]struct{}) {
	switch src.Kind() {
	// it's an interface, so we basically need
	// to extract the elem and walk through it
	// as the initial call
	case reflect.Interface:
		srcValue := src.Elem()
		destValue := reflect.New(srcValue.Type()).Elem()
		w.walkInterpolation(destValue, srcValue, interpolate, name, key, resourceType, relations)
		dest.Set(destValue)

	// if the current `src` is a slice
	// we iterate on each element.
	case reflect.Array, reflect.Slice:
		dest.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			w.walkInterpolation(dest.Index(i), src.Index(i), interpolate, name, key, resourceType, relations)
		}

	// it's a map
	case reflect.Map:
		dest.Set(reflect.MakeMap(src.Type()))
		iter := src.MapRange()
		for iter.Next() {
			// New gives us a pointer, but again we want the value
			destValue := reflect.New(iter.Value().Type()).Elem()
			nk := iter.Key().String()
			if key != "" {
				nk = fmt.Sprintf("%s.%s", key, iter.Key())
			}
			w.walkInterpolation(destValue, iter.Value(), interpolate, name, nk, resourceType, relations)
			dest.SetMapIndex(iter.Key(), destValue)
		}

	// what we want to interpolate is a string
	// we do not interpolate a custom tag (like cycloid.io) since it's key.
	// for now, only "strings" are interpolated since it's not that easy to interpolate
	// a bool / an int without more context.
	case reflect.String:
		// we check if there is a value to interpolate
		skey := strings.Split(key, ".")
		ak := skey[len(skey)-1]
		if interpolatedValue, ok := interpolate.Interpolate(ak, src.Interface().(string)); ok {
			irt, in := extractResourceTypeAndName(interpolatedValue)
			source := fmt.Sprintf("%s.%s", resourceType, name)
			target := fmt.Sprintf("%s.%s", irt, in)
			if w.opts.HasModule() && len(w.opts.ModuleVariables) != 0 {
				// If the current value is part of the ModulesVariables do not try to interpolate it
				if _, ok := w.opts.ModuleVariables[fmt.Sprintf("%s.%s", resourceType, key)]; ok {
					dest.Set(src)
					return
				}
			}
			// avoid to interpolate a resource by "itself" (interpolaception) and avoid to interpolate a resource type with resource
			// of the same type (cyclic interpolation)
			// we also check for mutual interpolation
			if !(strings.Contains(interpolatedValue, name) || strings.Contains(interpolatedValue, resourceType) || isMutualInterpolation(target, source, relations)) {
				dest.SetString(interpolatedValue)
				// we store this new relationship
				relations[fmt.Sprintf("%s+%s", source, target)] = struct{}{}
			} else {
				dest.SetString(src.Interface().(string))
			}
		} else {
			dest.SetString(src.Interface().(string))
		}
	default:
		dest.Set(src)
	}
}

// isMutualInterpolation will simply go through the list of relations to find out
// if a relation is already present between the two resources in one direction
// or the other
func isMutualInterpolation(target, source string, relations map[string]struct{}) bool {
	if _, ok := relations[fmt.Sprintf("%s+%s", source, target)]; ok {
		return true
	}
	if _, ok := relations[fmt.Sprintf("%s+%s", target, source)]; ok {
		return true
	}
	return false
}

// extractResourceTypeAndName will parse a TF variable to return
// the resource type and the name of the resource
func extractResourceTypeAndName(value string) (string, string) {
	res := regexp.MustCompile(`\${(.+)\.(.+)\.(.+)}`)
	match := res.FindStringSubmatch(value)
	return match[1], match[2]

}

// setProviderConfig will set the required fields to the provider configuration
// under the given category
func (w *Writer) setProviderConfig(cat string) {
	pcfg := w.provider.Configuration()
	for k, s := range w.provider.TFProvider().Schema {
		if s.Required {
			if _, ok := w.Config[cat]["variable"]; !ok {
				w.Config[cat]["variable"] = make(map[string]interface{})
			}
			varVal := map[string]interface{}{}
			if v, ok := pcfg[k]; ok {
				varVal["default"] = v
			} else if s.Default != nil {
				varVal["default"] = s.Default
			} else if s.DefaultFunc != nil {
				v, err := s.DefaultFunc()
				// If we have an error we ignore it
				if err == nil {
					varVal["default"] = v
				}
			}
			w.Config[cat]["variable"].(map[string]interface{})[k] = varVal
			w.Config[cat]["provider"].(map[string]interface{})[w.provider.String()].(map[string]interface{})[k] = fmt.Sprintf("${var.%s}", k)
		}
	}
}
