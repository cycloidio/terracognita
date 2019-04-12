package writer

import (
	"strings"

	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
)

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

type HCLWriter struct {
	// TODO: Change it to "map[string]map[string]schema.ResourceData"
	Config map[string]interface{}
}

func NewHCLWriter() *HCLWriter {
	cfg := make(map[string]interface{})
	cfg["resource"] = make(map[string]map[string]interface{})
	return &HCLWriter{
		Config: cfg,
	}
}

// Write expects a key similar to "aws_instance.your_name"
func (hclw *HCLWriter) Write(key string, value interface{}) error {
	if key == "" {
		return ErrRequiredKey
	}

	if value == nil {
		return ErrRequiredValue
	}

	keys := strings.Split(key, ".")
	if len(keys) != 2 {
		return errors.Wrapf(ErrInvalidKey, "with key %q", key)
	}

	if _, ok := hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]][keys[1]]; ok {
		return errors.Wrapf(ErrAlreadyExistsKey, "with key %q", key)
	}

	hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]] = make(map[string]interface{})
	hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]][keys[1]] = value

	return nil
}

type TFStateWriter struct {
	Config map[string]*terraform.ResourceState
}

func NewTFStateWriter() *TFStateWriter {
	return &TFStateWriter{
		Config: make(map[string]*terraform.ResourceState),
	}
}

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
