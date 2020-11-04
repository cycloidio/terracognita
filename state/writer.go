package state

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/writer"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/hashicorp/terraform/states/statemgr"
	"github.com/pkg/errors"
)

// used to match a TF resource ${aws_instance.my-instance.id}
var regexResource = regexp.MustCompile(`\${(.+)\.(.+)\.(.+)}`)

// Writer is a Writer implementation
// that is meant to generate a TFState
type Writer struct {
	Config map[string]provider.Resource
	writer io.Writer
	state  *states.SyncState
	opts   *writer.Options
}

// NewWriter returns a TFStateWriter initialization
func NewWriter(w io.Writer, opts *writer.Options) *Writer {
	return &Writer{
		Config: make(map[string]provider.Resource),
		writer: w,
		state:  states.NewState().SyncWrapper(),
		opts:   opts,
	}
}

// Write expects a key similar to "aws_instance.your_name" and
// the value to be *terraform.ResourceState repeated keys will report an error
func (w *Writer) Write(key string, value interface{}) error {
	if key == "" {
		return errcode.ErrWriterRequiredKey
	}

	if value == nil {
		return errcode.ErrWriterRequiredValue
	}

	if _, ok := w.Config[key]; ok {
		return errors.Wrapf(errcode.ErrWriterAlreadyExistsKey, "with key %q", key)
	}

	if len(strings.Split(key, ".")) != 2 {
		return errors.Wrapf(errcode.ErrWriterInvalidKey, "with key %q", key)
	}

	r, ok := value.(provider.Resource)
	if !ok {
		return errors.Wrapf(errcode.ErrWriterInvalidTypeValue, "expected provider.Resource, found %T", value)
	}

	absAddr := addrs.AbsResourceInstance{
		Module: nil,
		Resource: addrs.ResourceInstance{
			Resource: addrs.Resource{
				Mode: addrs.ManagedResourceMode,
				Type: r.Type(),
				Name: strings.Split(key, ".")[1],
			},
			Key: nil,
		},
	}

	absProviderConf := addrs.AbsProviderConfig{
		Module: nil,
		ProviderConfig: addrs.ProviderConfig{
			Type: addrs.NewLegacyProvider(r.Provider().String()),
		},
	}

	src, err := r.ResourceInstanceObject().Encode(r.ImpliedType(), uint64(r.TFResource().SchemaVersion))
	if err != nil {
		return err
	}

	w.state.SetResourceInstanceCurrent(absAddr, src, absProviderConf)

	log.Get().Log("func", "state.Write(State)", "msg", "writing to internal config", "key", key, "content", r)
	w.Config[key] = r

	return nil
}

// Has checks if the given key it's already present or not
func (w *Writer) Has(key string) (bool, error) {
	_, ok := w.Config[key]
	return ok, nil
}

// Sync writes the content of the Config to the
// internal w with the correct format
func (w *Writer) Sync() error {

	lstate := w.state.Lock()
	defer w.state.Unlock()

	log.Get().Log("func", "state.Sync(State)", "msg", "writting state to state file")
	file := statemgr.NewStateFile()
	file.State = lstate

	err := statefile.Write(file, w.writer)
	if err != nil {
		return err
	}

	return nil
}

// Interpolate will defined dependencies for each component using
// the `i` map built in the import.
func (w *Writer) Interpolate(i map[string]string) {
	if !w.opts.Interpolate {
		return
	}
	// keep the existing relations in order to avoid cyclic
	// dependencies
	relations := make(map[string]struct{})

	// acquire the actual terraform.State
	lstate := w.state.Lock()
	defer w.state.Unlock()

	// loop over the whole state to write the dependencies
	// for resources having deps
	for _, module := range lstate.Modules {
		for name, resource := range module.Resources {
			// keep the existing dependencies in order to avoid
			// duplicated
			deps := make(map[string]struct{}, 0)
			// fetch the Terracognita resource representation
			// to access the attributes later
			res, ok := w.Config[name]
			if !ok {
				continue
			}
			for _, attribute := range res.InstanceState().Attributes {
				// if we find any relevant link between the instance attribute and the interpolation map,
				// we flag a dependency
				if dependency, ok := i[attribute]; ok {
					rt, rn := extractResourceTypeAndName(dependency)
					rsc := fmt.Sprintf("%s.%s", rt, rn)
					// avoid mutual dependencies
					if rt == res.Type() || name == rsc || isMutualInterpolation(name, rsc, relations) {
						continue
					}
					// avoid adding the same dependency for a resource
					if _, ok := deps[rsc]; ok {
						continue
					}
					// save the resource as a dependency
					deps[rsc] = struct{}{}
				}
			}
			for _, instance := range resource.Instances {
				for dependency := range deps {
					// dependency is like google_compute_instance.instance-name
					s := strings.Split(dependency, ".")
					rt := s[0]
					rn := s[1]
					instance.Current.Dependencies = append(instance.Current.Dependencies, addrs.AbsResource{
						Module: nil,
						Resource: addrs.Resource{
							Mode: addrs.ManagedResourceMode,
							Type: rt,
							Name: rn,
						},
					})
					// save the relationship
					relations[fmt.Sprintf("%s+%s", dependency, name)] = struct{}{}
				}
			}
		}
	}

}

// extractResourceTypeAndName will parse a TF variable to return
// the resource type and the name of the resource
func extractResourceTypeAndName(value string) (string, string) {
	match := regexResource.FindStringSubmatch(value)
	return match[1], match[2]
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
