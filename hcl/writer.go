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
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/writer"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	cjson "github.com/zclconf/go-cty/cty/json"
)

const (
	// ResourceCategoryKey is an internal key used to specify the category of
	// a resource when writing, it'll be used to select in which file
	// will be written
	ResourceCategoryKey = "tc_category"

	// ModuleCategoryKey is the category used to identify
	// the Module
	ModuleCategoryKey = "tc_module"

	defaultCategory      = "hcl"
	variablesCategoryKey = "variables"
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
}

// NewWriter rerturns an Writer initialization
func NewWriter(w io.Writer, opts *writer.Options) *Writer {
	cfg := make(map[string]map[string]interface{})
	if opts.HasModule() {
		cfg[ModuleCategoryKey] = map[string]interface{}{
			"module": map[string]interface{}{
				opts.Module: map[string]interface{}{
					"source": fmt.Sprintf("module-%s", opts.Module),
				},
			},
		}
	}

	return &Writer{
		Config: cfg,
		writer: w,
		opts:   opts,
	}
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
	ic, ok := m[ResourceCategoryKey]
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
		if k == ModuleCategoryKey || k == variablesCategoryKey {
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
		categories = append(categories, []string{ModuleCategoryKey, variablesCategoryKey}...)
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
		blockMap := v.AsValueMap()
		// blockType can be "resource", "variable" or "output"
		for _, blockType := range blockKeys {
			blockValue := blockMap[blockType]

			resourceKeys := getValueKeys(blockValue)
			resourceMap := blockValue.AsValueMap()
			// resourceType is the type of the resource (e.g: `aws_security_groups`)
			for _, resourceType := range resourceKeys {
				resources := resourceMap[resourceType]
				if blockType == "variable" || blockType == "module" {
					block := hclwrite.NewBlock(blockType, []string{resourceType})
					bbody := block.Body()

					attrKeys := getValueKeys(resources)
					attrMap := resources.AsValueMap()
					for _, attr := range attrKeys {
						bbody.SetAttributeValue(attr, attrMap[attr])
					}

					body.AppendBlock(block)
					body.AppendNewline()
				} else {
					resourceKeys := getValueKeys(resources)
					resourceMap := resources.AsValueMap()
					for _, name := range resourceKeys {
						resource := resourceMap[name]
						block := hclwrite.NewBlock(blockType, []string{resourceType, name})
						bbody := block.Body()

						attrKeys := getValueKeys(resource)
						attrMap := resource.AsValueMap()
						for _, attr := range attrKeys {
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
	keys := make([]string, 0, val.LengthInt())
	for it := val.ElementIterator(); it.Next(); {
		k, _ := it.Element()
		keys = append(keys, k.AsString())
	}

	sort.Strings(keys)

	return keys
}

func (w *Writer) setVariables() {
	variables := make(map[string]interface{})
	for c, cfg := range w.Config {
		if c == ModuleCategoryKey || c == variablesCategoryKey {
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
		w.Config[ModuleCategoryKey]["module"].(map[string]interface{})[w.opts.Module].(map[string]interface{})[fmt.Sprintf("# %s", k)] = v.(map[string]interface{})["default"]
	}
	// TODO: Also add those to the Module
}

// walkVariables will walk the cfg until it reached the last elements, the k is the current key (as it's recursive can be aws_lb.ingress.from_port)
// variables is the map of all the variables assigned. It returns the new cfg with the variable interpolation.
// If the validVariables is not empty only those will be used as variables, if not all the attributes will be converted in variables
func walkVariables(cfg map[string]interface{}, validVariables map[string]struct{}, k string, variables map[string]interface{}) map[string]interface{} {
	for key, value := range cfg {
		currentKey := fmt.Sprintf("%s.%s", k, key)
		switch v := value.(type) {
		case map[string]interface{}:
			cfg[key] = walkVariables(v, validVariables, currentKey, variables)
		case []interface{}:
			// For slices we need to check the first element, if it's a map then
			// it has complex data, if not it's a "simple" slice of values
			if _, ok := v[0].(map[string]interface{}); ok {
				for i, vvv := range v {
					v[i] = walkVariables(vvv.(map[string]interface{}), validVariables, currentKey, variables)
				}
			} else {
				if hasKey(validVariables, currentKey) {
					varName := strings.ReplaceAll(currentKey, ".", "_")
					variables[varName] = map[string]interface{}{
						"default": cfg[key],
					}
					cfg[key] = fmt.Sprintf("${var.%s}", varName)
				}
			}
		default:
			// This means is a "simple" value so we can
			// directly replace it with the variable
			if hasKey(validVariables, currentKey) {
				varName := strings.ReplaceAll(currentKey, ".", "_")
				variables[varName] = map[string]interface{}{
					"default": cfg[key],
				}
				cfg[key] = fmt.Sprintf("${var.%s}", varName)
			}
		}
	}
	return cfg
}

// hasKey will validate that the key is present on the map.
// The key will have the format: aws_instance.front.attr1.attr2...
// and the validVariables will not have the `front` interpolation, also
// it'll validate that if validVariables is empty it's always true
func hasKey(validVariables map[string]struct{}, key string) bool {
	if len(validVariables) == 0 {
		return true
	}

	sk := strings.Split(key, ".")
	k := append(sk[0:1], sk[2:]...)
	_, ok := validVariables[strings.Join(k, ".")]

	return ok
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
// with TF interpolation
func (w *Writer) Interpolate(i map[string]string) {
	if !w.opts.Interpolate {
		return
	}
	for k, v := range w.Config {
		if k == ModuleCategoryKey || k == variablesCategoryKey {
			continue
		}
		resources := v["resource"]
		// who's interpolated with who
		relations := make(map[string]struct{}, 0)
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
				walkInterpolation(dest, src, i, name, rt, &relations)

				// remove reflect.Value wrapper from dest
				resources.(map[string]map[string]interface{})[rt][name] = dest.Interface()
			}
		}
	}
}

// walkInterpolation through a resource block. it's easier since we do not know how the block is made
// `dest` will be the new "block" with the values interpolated from `interpolate`
func walkInterpolation(dest, src reflect.Value, interpolate map[string]string, name string, resourceType string, relations *map[string]struct{}) {
	switch src.Kind() {
	// it's an interface, so we basically need
	// to extract the elem and walk through it
	// as the initial call
	case reflect.Interface:
		srcValue := src.Elem()
		destValue := reflect.New(srcValue.Type()).Elem()
		walkInterpolation(destValue, srcValue, interpolate, name, resourceType, relations)
		dest.Set(destValue)

	// if the current `src` is a slice
	// we iterate on each element.
	case reflect.Array, reflect.Slice:
		dest.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			walkInterpolation(dest.Index(i), src.Index(i), interpolate, name, resourceType, relations)
		}

	// it's a map
	case reflect.Map:
		dest.Set(reflect.MakeMap(src.Type()))
		iter := src.MapRange()
		for iter.Next() {
			// New gives us a pointer, but again we want the value
			destValue := reflect.New(iter.Value().Type()).Elem()
			walkInterpolation(destValue, iter.Value(), interpolate, name, resourceType, relations)
			dest.SetMapIndex(iter.Key(), destValue)
		}

	// what we want to interpolate is a string
	// we do not interpolate a custom tag (like cycloid.io) since it's key.
	// for now, only "strings" are interpolated since it's not that easy to interpolate
	// a bool / an int without more context.
	case reflect.String:
		// we check if there is a value to interpolate
		if interpolatedValue, ok := interpolate[src.Interface().(string)]; ok {
			irt, in := extractResourceTypeAndName(interpolatedValue)
			target := fmt.Sprintf("%s.%s", irt, in)
			source := fmt.Sprintf("%s.%s", resourceType, name)
			// avoid to interpolate a resource by "itself" (interpolaception) and avoid to interpolate a resource type with resource
			// of the same type (cyclic interpolation)
			// we also check for mutual interpolation
			if !(strings.Contains(interpolatedValue, name) || strings.Contains(interpolatedValue, resourceType) || isMutualInterpolation(target, source, relations)) {
				dest.SetString(interpolatedValue)
				// we store this new relationship
				(*relations)[fmt.Sprintf("%s+%s", source, target)] = struct{}{}
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
func isMutualInterpolation(target, source string, relations *map[string]struct{}) bool {
	if _, ok := (*relations)[fmt.Sprintf("%s+%s", source, target)]; ok {
		return true
	}
	if _, ok := (*relations)[fmt.Sprintf("%s+%s", target, source)]; ok {
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
