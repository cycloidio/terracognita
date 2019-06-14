package writer

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
)

// TFStateWriter is a Writer implementation
// that is meant to generate a TFState
type TFStateWriter struct {
	Config map[string]*terraform.ResourceState
	w      io.Writer
}

// NewTFStateWriter returns a TFStateWriter initialization
func NewTFStateWriter(w io.Writer) *TFStateWriter {
	return &TFStateWriter{
		Config: make(map[string]*terraform.ResourceState),
		w:      w,
	}
}

// Write expects a key similar to "aws_instance" and the value to be *terraform.ResourceState
// repeated keys will report an error
func (tfsw *TFStateWriter) Write(key string, value interface{}) error {
	if key == "" {
		return errcode.ErrWriterRequiredKey
	}

	if value == nil {
		return errcode.ErrWriterRequiredValue
	}

	if _, ok := tfsw.Config[key]; ok {
		return errors.Wrapf(errcode.ErrWriterAlreadyExistsKey, "with key %q", key)
	}

	trs, ok := value.(*terraform.ResourceState)
	if !ok {
		return errors.Wrapf(errcode.ErrWriterInvalidTypeValue, "expected *terraform.ResourceState, found %T", value)
	}

	tfsw.Config[key] = trs

	return nil
}

// Sync writes the content of the Config to the
// internal w with the correct format
func (tfsw *TFStateWriter) Sync() error {
	// Write to root because then the NewState is called
	// it creates by default a 'root' one and then on the
	// AddModuleState we replace that empty module for this one
	ms := &terraform.ModuleState{
		Path: []string{"root"},
	}

	ms.Resources = tfsw.Config

	state := terraform.NewState()
	state.AddModuleState(ms)

	enc := json.NewEncoder(tfsw.w)
	enc.SetIndent("", "  ")
	err := enc.Encode(state)
	if err != nil {
		return fmt.Errorf("could not encode state due to: %s", err)
	}

	return nil
}
