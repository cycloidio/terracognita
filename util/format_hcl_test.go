package util_test

import (
	"testing"

	"github.com/cycloidio/terraforming/util"
	"github.com/stretchr/testify/assert"
)

func TestFormatHCL(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		out  []byte
	}{
		{
			name: "Replace\"\"OnKeys",
			in: []byte(`
				"role" = value
				"en.v" = "value"
			`),
			out: []byte(`
				role = value
				en.v = "value"
			`),
		},
		{
			name: "ReplaceEmptyLines",
			in: []byte(`
				"role" = value

				"env" = "value"
			`),
			out: []byte(`
				role = value
				env = "value"
			`),
		},
		{
			name: "ReplaceEmptyLinesExceptBlocks",
			in: []byte(`
				"role" = value

				"env" = "value"

				"tags" = {
					"something" = "s"

					"another" = "a"
				}

				"env" = "value"
				"role" = value
			`),
			out: []byte(`
				role = value
				env = "value"

				tags = {
					something = "s"
					another = "a"
				}

				env = "value"
				role = value
			`),
		},
		{
			name: "ReplaceResourceDefinitions",
			in: []byte(`
			"resource" "aws_instance" "name" {
				"role" = value

				"env" = "value"

				"tags" = {
					"something" = "s"

					"another" = "a"
				}

				"env" = "value"
				"role" = value
			}`),
			out: []byte(`
			resource "aws_instance" "name" {
				role = value
				env = "value"

				tags = {
					something = "s"
					another = "a"
				}

				env = "value"
				role = value
			}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := util.FormatHCL(tt.in)
			assert.Equal(t, string(tt.out), string(out))
		})
	}
}
