package hcl

import (
	"bytes"
	"fmt"
	"regexp"
)

var (
	// transformations are all the steps that
	// we'll apply to the HCL and in the order
	// those will be applied
	transformations = []struct {
		match     *regexp.Regexp
		replace   []byte
		replaceFn func([]byte) []byte
	}{
		{
			// Replace all the `"key" = "value"` for `key = "value"` except
			// if it has a `.` on the key
			match:   regexp.MustCompile(`"([^\d][\w\-_=]+)"\s=`),
			replace: []byte(`$1 =`),
		},
		{
			// Replace all the `key = {` for `key {` except
			// on the detected blocks.
			// It'll ignore the '=tc=' ones as those are a key word
			// to refer internally as attributes
			match: regexp.MustCompile(`([\w\-_=]+)\s=\s{`),
			replaceFn: func(m []byte) []byte {
				if bytes.Contains(m, []byte(`=tc=`)) {
					return []byte(fmt.Sprintf("%s", m))
				}

				return bytes.Replace(m, []byte(`= `), nil, 1)

			},
		},
		{
			match:   regexp.MustCompile(`=tc=`),
			replace: []byte{},
		},
		{
			// Add new lines before blocks
			match:   regexp.MustCompile("\n(\t*)(?:([\\w\\-_\\.]+\\s{)|([\\w\\-_\\.]+\\s=\\s{))"),
			replace: []byte("\n\n$1$2$3"),
		},
		{
			// Replace all the empty lines
			match:   regexp.MustCompile("\n\n"),
			replace: []byte("\n"),
		},
		{
			// Add new lines after block
			match:   regexp.MustCompile("}\n"),
			replace: []byte("}\n\n"),
		},
		{
			// Remove "" from resources definition like
			// '"resource" "aws_lb" "front {' -> 'resource "aws_lb" "front" {'
			match:   regexp.MustCompile(`"([\w\-_\.]+)"\s("(?:[\w\-_\.]+)")\s("(?:[\w\-_\.]+)")\s{`),
			replace: []byte(`$1 $2 $3 {`),
		},
	}
)

// Format formats the hcl to have a better formatter that the default one
// returned from HCL printer.Fprint
func Format(hcl []byte) []byte {
	for _, m := range transformations {
		if m.replace != nil {
			hcl = m.match.ReplaceAll(hcl, m.replace)
		} else if m.replaceFn != nil {
			hcl = m.match.ReplaceAllFunc(hcl, m.replaceFn)
		}
	}

	return hcl
}
