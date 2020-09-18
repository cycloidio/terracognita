package hcl

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"

	kitlog "github.com/go-kit/kit/log"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/writer"
	"github.com/hashicorp/hcl2/hclwrite"
	"github.com/pkg/errors"
	cjson "github.com/zclconf/go-cty/cty/json"
)

// Writer is a Writer implementation that writes to
// a static map to then transform it to HCL
type Writer struct {
	// TODO: Change it to "map[string]map[string]schema.ResourceData"
	Config map[string]interface{}
	writer io.Writer
	opts   *writer.Options
	// File is used to build the HCL file
	File *hclwrite.File
}

// NewWriter rerturns an Writer initialization
func NewWriter(w io.Writer, opts *writer.Options) *Writer {
	cfg := make(map[string]interface{})
	cfg["resource"] = make(map[string]map[string]interface{})
	f := hclwrite.NewEmptyFile()
	return &Writer{
		Config: cfg,
		writer: w,
		opts:   opts,
		File:   f,
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

	name := strings.Join(keys[1:], "")

	if _, ok := w.Config["resource"].(map[string]map[string]interface{})[keys[0]][name]; ok {
		return errors.Wrapf(errcode.ErrWriterAlreadyExistsKey, "with key %q", key)
	}

	if _, ok := w.Config["resource"].(map[string]map[string]interface{})[keys[0]]; !ok {
		w.Config["resource"].(map[string]map[string]interface{})[keys[0]] = make(map[string]interface{})
	}

	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	log.Get().Log("func", "writer.Write(HCL)", "msg", "writing to internal config", "key", keys[0], "content", string(b))

	w.Config["resource"].(map[string]map[string]interface{})[keys[0]][name] = value

	return nil
}

// Has checks if the given key is already present or not
func (w *Writer) Has(key string) (bool, error) {
	keys := strings.Split(key, ".")
	if len(keys) != 2 || keys[0] == "" || keys[1] == "" {
		return false, errors.Wrapf(errcode.ErrWriterInvalidKey, "with key %q", key)
	}

	name := strings.Join(keys[1:], "")

	if _, ok := w.Config["resource"].(map[string]map[string]interface{})[keys[0]][name]; ok {
		return true, nil
	}

	return false, nil
}

// Sync writes the content of the Config to the
// internal w with the correct format
func (w *Writer) Sync() error {
	logger := log.Get()
	logger = kitlog.With(logger, "func", "writer.Write(HCL)")

	body := w.File.Body()

	src, err := json.Marshal(w.Config)
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

	// blockType can be "resource", "variable" or "output"
	for blockType, blockValue := range v.AsValueMap() {
		// resourceType is the type of the resource (e.g: `aws_security_groups`)
		for resourceType, resources := range blockValue.AsValueMap() {
			for name, resource := range resources.AsValueMap() {
				block := hclwrite.NewBlock(blockType, []string{resourceType, name})
				bbody := block.Body()
				for attr, value := range resource.AsValueMap() {
					// in JSON representation, we can have a list of object
					// e.g with ingress:[{ingress1}, {ingress2}, ... {ingressN}]
					// we need to add a dedicated block for each object instead of having
					// one block for the whole list
					if value.Type().IsTupleType() {
						iter := value.ElementIterator()
						for iter.Next() {
							_, val := iter.Element()
							if val.Type().IsPrimitiveType() {
								// the value is not an unconsistent object
								bbody.SetAttributeValue(attr, value)
								continue
							}
							obj := bbody.AppendNewBlock(attr, nil)
							bobj := obj.Body()
							ei := val.ElementIterator()
							for ei.Next() {
								kei, vei := ei.Element()
								bobj.SetAttributeValue(kei.AsString(), vei)
							}
						}
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
	formattedBytes := Format(w.File.Bytes())
	formattedBytes = hclwrite.Format(formattedBytes)
	w.writer.Write(formattedBytes)

	return nil
}

// Interpolate replaces the hardcoded resources link
// with TF interpolation
func (w *Writer) Interpolate(i map[string]string) {
	if !w.opts.Interpolate {
		return
	}
	resources := w.Config["resource"]
	// who's interpolated with who
	relations := make(map[string]struct{}, 0)
	// we need to isolate each resource
	// getting each resource is easier to avoid cycle
	// or interpolaception.
	// We first loop over resource type (e.g: aws_instance)
	for rt, resource := range resources.(map[string]map[string]interface{}) {
		// we loop over a resource (e.g: aws_instance.oDSOj)
		for name, block := range resource {
			src := reflect.ValueOf(block)

			// this will store the updated block
			dest := reflect.New(src.Type()).Elem()

			// walk through the resources to interpolate the good values
			walk(dest, src, i, name, rt, &relations)

			// remove reflect.Value wrapper from dest
			resources.(map[string]map[string]interface{})[rt][name] = dest.Interface()
		}
	}
}

// walk through a resource block. it's easier since we do not know how the block is made
// `dest` will be the new "block" with the values interpolated from `interpolate`
func walk(dest, src reflect.Value, interpolate map[string]string, name string, resourceType string, relations *map[string]struct{}) {
	switch src.Kind() {
	// it's an interface, so we basically need
	// to extract the elem and walk through it
	// as the initial call
	case reflect.Interface:
		srcValue := src.Elem()
		destValue := reflect.New(srcValue.Type()).Elem()
		walk(destValue, srcValue, interpolate, name, resourceType, relations)
		dest.Set(destValue)

	// if the current `src` is a slice
	// we iterate on each element.
	case reflect.Array, reflect.Slice:
		dest.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			walk(dest.Index(i), src.Index(i), interpolate, name, resourceType, relations)
		}

	// it's a map
	case reflect.Map:
		dest.Set(reflect.MakeMap(src.Type()))
		iter := src.MapRange()
		for iter.Next() {
			// New gives us a pointer, but again we want the value
			destValue := reflect.New(iter.Value().Type()).Elem()
			walk(destValue, iter.Value(), interpolate, name, resourceType, relations)
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
