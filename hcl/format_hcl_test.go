package hcl_test

import (
	"testing"

	"github.com/cycloidio/terracognita/hcl"
	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
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
				"2tag" = "2value"
				"t2tag" = "t2value"
			`),
			out: []byte(`
				role = value
				"en.v" = "value"
				"2tag" = "2value"
				t2tag = "t2value"
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
			name: "Remove`=`Form`= {` but not on tags",
			in: []byte(`
				"ebs_block_device" = {
					"volume_size" = 24
				}
				"=tc=tags" = {
					"some.thing" = "s"
				}
				"=tc=volume_tags" = {
					"some.thing" = "s"
				}
			`),
			out: []byte(`
				ebs_block_device {
					volume_size = 24
				}

				tags = {
					"some.thing" = "s"
				}

				volume_tags = {
					"some.thing" = "s"
				}

			`),
		},
		{
			name: "ReplaceEmptyLinesExceptBlocks",
			in: []byte(`
				"role" = value

				"env" = "value"

				"=tc=tags" = {
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

				"=tc=tags" = {
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
			out := hcl.Format(tt.in)
			assert.Equal(t, string(tt.out), string(out))
		})
	}
}
