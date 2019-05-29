package writer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl/hcl/printer"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/fmtcmd"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
)

// The list of possible errors
var (
	ErrRequiredKey      = errors.New("the key it's required")
	ErrRequiredValue    = errors.New("the value it's required")
	ErrInvalidKey       = errors.New("invalid key")
	ErrInvalidTypeValue = errors.New("invalid type value")
	ErrAlreadyExistsKey = errors.New("already exists key")
)

// Writer it's an interface used to abstract the logic
// of writing results to a Key Value without having to
// deal with types or internal structures
type Writer interface {
	// Write sets the value with the key to the internal structure,
	// the value will be casted to the correct type of each
	// implementation and an error can be returned normally for
	// repeated keys
	Write(key string, value interface{}) error

	// Sync writes the content of the writer
	// to the internal system. Each Writter may have
	// a different implementation of it with different
	// output formats
	Sync() error
}

// HCLWriter it's a Writer implementation that writes to
// a static map to then use to traform to HCL
type HCLWriter struct {
	// TODO: Change it to "map[string]map[string]schema.ResourceData"
	Config map[string]interface{}
	w      io.Writer
}

// NewHCLWriter rerturns a HCLWriter initializatioon
func NewHCLWriter(w io.Writer) *HCLWriter {
	cfg := make(map[string]interface{})
	cfg["resource"] = make(map[string]map[string]interface{})
	return &HCLWriter{
		Config: cfg,
		w:      w,
	}
}

// Write expects a key similar to "aws_instance.your_name"
// repeated keys will report an error
func (hclw *HCLWriter) Write(key string, value interface{}) error {
	if key == "" {
		return ErrRequiredKey
	}

	if value == nil {
		return ErrRequiredValue
	}

	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return errors.Wrapf(ErrInvalidKey, "with key %q", key)
	}

	name := strings.Join(keys[1:], "")

	if _, ok := hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]][name]; ok {
		return errors.Wrapf(ErrAlreadyExistsKey, "with key %q", key)
	}

	if _, ok := hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]]; !ok {
		hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]] = make(map[string]interface{})
	}
	hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]][name] = value

	return nil
}

// Sync writes the content of the Config to the
// internal w with the correct format
func (hclw *HCLWriter) Sync() error {
	b, err := json.Marshal(hclw.Config)
	if err != nil {
		return err
	}

	f, err := hcl.ParseBytes(b)
	if err != nil {
		return fmt.Errorf("error while 'hcl.ParseBytes': %s", err)
	}

	buff := &bytes.Buffer{}
	err = printer.Fprint(buff, f.Node)
	if err != nil {
		return fmt.Errorf("error while pretty printing HCL: %s", err)
	}

	buff = bytes.NewBuffer(FormatHCL(buff.Bytes()))

	err = fmtcmd.Run(nil, nil, buff, hclw.w, fmtcmd.Options{})
	if err != nil {
		return fmt.Errorf("error while fmt HCL: %s", err)
	}
	return nil
}

// TFStateWriter it's a Writer implementation that it's ment to
// then generate a TFState
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
		return ErrRequiredKey
	}

	if value == nil {
		return ErrRequiredValue
	}

	if _, ok := tfsw.Config[key]; ok {
		return errors.Wrapf(ErrAlreadyExistsKey, "with key %q", key)
	}

	trs, ok := value.(*terraform.ResourceState)
	if !ok {
		return errors.Wrapf(ErrInvalidTypeValue, "expected *terraform.ResourceState, found %T", value)
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
