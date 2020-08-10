package cmd

import (
	"context"
	"fmt"

	kitlog "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cycloidio/terracognita/azurerm"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/hcl"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/state"
	"github.com/cycloidio/terracognita/writer"
)

var (
	azurermCmd = &cobra.Command{
		Use:   "azurerm",
		Short: "Terracognita reads from Azure and generates hcl resources and/or terraform state",
		Long:  "Terracognita reads from Azure and generates hcl resources and/or terraform state",
		PreRun: func(cmd *cobra.Command, args []string) {
			preRunEOutput(cmd, args)
			viper.BindPFlag("client-id", cmd.Flags().Lookup("client-id"))
			viper.BindPFlag("client-secret", cmd.Flags().Lookup("client-secret"))
			viper.BindPFlag("environment", cmd.Flags().Lookup("environment"))
			viper.BindPFlag("resource-group-name", cmd.Flags().Lookup("resource-group-name"))
			viper.BindPFlag("subscription-id", cmd.Flags().Lookup("subscription-id"))
			viper.BindPFlag("tenant-id", cmd.Flags().Lookup("tenant-id"))
		},
		PostRunE: postRunEOutput,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Get()
			logger = kitlog.With(logger, "func", "cmd.azure.RunE")
			// Validate required flags
			if err := requiredStringFlags(
				"client-id", "client-secret", "resource-group-name", "subscription-id", "tenant-id",
			); err != nil {
				return err
			}

			ctx := context.Background()

			azureRMP, err := azurerm.NewProvider(
				ctx,
				viper.GetString("client-id"),
				viper.GetString("client-secret"),
				viper.GetString("environment"),
				viper.GetString("resource-group-name"),
				viper.GetString("subscription-id"),
				viper.GetString("tenant-id"),
			)
			if err != nil {
				return err
			}

			f := &filter.Filter{
				Include: include,
				Exclude: exclude,
				Targets: targets,
			}

			var hclW, stateW writer.Writer
			options := &writer.Options{Interpolate: viper.GetBool("interpolate")}

			if hclOut != nil {
				logger.Log("msg", "initializing HCL writer")
				hclW = hcl.NewWriter(hclOut, options)
			}

			if stateOut != nil {
				logger.Log("msg", "initializing TFState writer")
				stateW = state.NewWriter(stateOut, options)
			}

			logger.Log("msg", "importing")

			fmt.Fprintf(logsOut, "Starting Terracognita with version %s\n", Version)
			logger.Log("msg", "starting terracognita", "version", Version)
			err = provider.Import(ctx, azureRMP, hclW, stateW, f, logsOut)
			if err != nil {
				return errors.Wrap(err, "could not import from Azure")
			}

			return nil
		},
	}
)

func init() {
	azurermCmd.AddCommand(azurermResourcesCmd)

	// Required flags
	azurermCmd.Flags().String("client-id", "", "Client ID (required)")
	azurermCmd.Flags().String("client-secret", "", "Client Secret (required)")
	azurermCmd.Flags().String("resource-group-name", "", "Resource Group Name (required)")
	azurermCmd.Flags().String("subscription-id", "", "Subscription ID (required)")
	azurermCmd.Flags().String("tenant-id", "", "Tenant ID (required)")

	// Optional flags
	azurermCmd.Flags().String("environment", "public", "Environment")
}
