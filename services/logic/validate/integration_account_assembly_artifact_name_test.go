package validate

import (
	"strings"
	"testing"
)

func TestIntegrationAccountAssemblyArtifactName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			input: "",
			valid: false,
		},
		{
			input: "test1",
			valid: true,
		},
		{
			input: "a2-.()b",
			valid: true,
		},
		{
			input: "a2&b",
			valid: false,
		},
		{
			input: strings.Repeat("s", 79),
			valid: true,
		},
		{
			input: strings.Repeat("s", 80),
			valid: true,
		},
		{
			input: strings.Repeat("s", 81),
			valid: false,
		},
	}

	validationFunction := IntegrationAccountAssemblyArtifactName()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validationFunction(tt.input, "")
			valid := err == nil
			if valid != tt.valid {
				t.Errorf("Expected valid status %t but got %t for input %s", tt.valid, valid, tt.input)
			}
		})
	}
}
