package state

import (
	"io"
	"strings"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/hashicorp/terraform/states/statemgr"
	"github.com/pkg/errors"
)

// Writer is a Writer implementation
// that is meant to generate a TFState
type Writer struct {
	Config map[string]provider.Resource
	writer io.Writer
	state  *states.SyncState
}

// NewWriter returns a TFStateWriter initialization
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Config: make(map[string]provider.Resource),
		writer: w,
		state:  states.NewState().SyncWrapper(),
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
			Type: r.Provider().String(),
		},
	}

	src, err := r.ResourceInstanceObject().Encode(r.CoreConfigSchema().ImpliedType(), uint64(r.TFResource().SchemaVersion))
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
