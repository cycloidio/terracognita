package writer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/fmtcmd"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/pkg/errors"
)

// HCLWriter is a Writer implementation that writes to
// a static map to then transform it to HCL
type HCLWriter struct {
	// TODO: Change it to "map[string]map[string]schema.ResourceData"
	Config map[string]interface{}
	w      io.Writer
}

// NewHCLWriter rerturns an HCLWriter initialization
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
	if len(keys) != 2 || (keys[0] == "" || keys[1] == "") {
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
