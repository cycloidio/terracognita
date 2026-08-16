package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestOverwriteFlag(t *testing.T) {
	flag := RootCmd.PersistentFlags().Lookup("overwrite")
	if flag == nil {
		t.Fatal("overwrite flag is not registered")
	}
	if flag.DefValue != "false" {
		t.Fatalf("overwrite default = %q, want false", flag.DefValue)
	}
}

func TestPreRunEOutputOverwriteExistingDirectory(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "module", key: "module"},
		{name: "HCL directory", key: "hcl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := filepath.Join(t.TempDir(), "output")
			if err := os.MkdirAll(filepath.Join(outputDir, "nested"), 0755); err != nil {
				t.Fatal(err)
			}
			existingFile := filepath.Join(outputDir, "nested", "existing.tf")
			if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
				t.Fatal(err)
			}

			viper.Set("module", "")
			viper.Set("hcl", "")
			viper.Set("tfstate", "")
			viper.Set(tt.key, outputDir)
			viper.Set("overwrite", true)
			t.Cleanup(func() {
				viper.Set("module", "")
				viper.Set("hcl", "")
				viper.Set("tfstate", "")
				viper.Set("overwrite", false)
			})

			if err := preRunEOutput(nil, nil); err != nil {
				t.Fatal(err)
			}

			info, err := os.Stat(outputDir)
			if err != nil {
				t.Fatal(err)
			}
			if !info.IsDir() {
				t.Fatalf("output path %q is not a directory", outputDir)
			}
			if _, err := os.Stat(existingFile); !os.IsNotExist(err) {
				t.Fatalf("existing file was not removed: %v", err)
			}
		})
	}
}
