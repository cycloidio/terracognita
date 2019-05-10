package writer

import (
	"strings"

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
}

// HCLWriter it's a Writer implementation that writes to
// a static map to then use to traform to HCL
type HCLWriter struct {
	// TODO: Change it to "map[string]map[string]schema.ResourceData"
	Config map[string]interface{}
}

// NewHCLWriter rerturns a HCLWriter initializatioon
func NewHCLWriter() *HCLWriter {
	cfg := make(map[string]interface{})
	cfg["resource"] = make(map[string]map[string]interface{})
	return &HCLWriter{
		Config: cfg,
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

	//keys := strings.Split(key, ".")
	//if len(keys) != 2 {
	//return errors.Wrapf(ErrInvalidKey, "with key %q", key)
	//}
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return errors.Wrapf(ErrInvalidKey, "with key %q", key)
	}

	name := strings.Join(keys[1:len(keys)], "")

	if _, ok := hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]][name]; ok {
		return errors.Wrapf(ErrAlreadyExistsKey, "with key %q", key)
	}

	if _, ok := hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]]; !ok {
		hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]] = make(map[string]interface{})
	}
	hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]][name] = value

	return nil
}

// TFStateWriter it's a Writer implementation that it's ment to
// then generate a TFState
type TFStateWriter struct {
	Config map[string]*terraform.ResourceState
}

// NewTFStateWriter returns a TFStateWriter initialization
func NewTFStateWriter() *TFStateWriter {
	return &TFStateWriter{
		Config: make(map[string]*terraform.ResourceState),
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
