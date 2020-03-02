package cmd

import (
	"github.com/cycloidio/terracognita/log"
	kitlog "github.com/go-kit/kit/log"
	"github.com/spf13/cobra"
)

var (
	azurermCmd = &cobra.Command{
		Use:   "azurerm",
		Short: "Terracognita reads from Azure and generates hcl resources and/or terraform state",
		Long:  "Terracognita reads from Azure and generates hcl resources and/or terraform state",
		PreRun: func(cmd *cobra.Command, args []string) {
			preRunEOutput(cmd, args)
		},
		PostRunE: postRunEOutput,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Get()
			logger = kitlog.With(logger, "func", "cmd.azure.RunE")
			return nil
		},
	}
)

func init() {
	azurermCmd.AddCommand(azurermResourcesCmd)
}
