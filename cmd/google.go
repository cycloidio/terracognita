package cmd

import (
	"context"

	kitlog "github.com/go-kit/kit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cycloidio/terracognita/google"
	"github.com/cycloidio/terracognita/log"
)

var (
	googleCmd = &cobra.Command{
		Use:   "google",
		Short: "Terracognita reads from GCP and generates hcl resources and/or terraform state",
		Long:  "Terracognita reads from GCP and generates hcl resources and/or terraform state",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := preRunEOutput(cmd, args)
			if err != nil {
				return err
			}
			viper.BindPFlag("credentials", cmd.Flags().Lookup("credentials"))
			viper.BindPFlag("project", cmd.Flags().Lookup("project"))
			viper.BindPFlag("region", cmd.Flags().Lookup("region"))
			viper.BindPFlag("labels", cmd.Flags().Lookup("labels"))
			viper.BindPFlag("max-results", cmd.Flags().Lookup("max-results"))

			return nil
		},
		PostRunE: postRunEOutput,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Get()
			logger = kitlog.With(logger, "func", "cmd.google.RunE")
			// Validate required flags
			if err := requiredStringFlags("region", "project", "credentials"); err != nil {
				return err
			}

			tags, err := initializeTags("labels")
			if err != nil {
				return err
			}

			ctx := context.Background()

			googleP, err := google.NewProvider(
				ctx,
				viper.GetUint64("max-results"),
				viper.GetString("project"),
				viper.GetString("region"),
				viper.GetString("credentials"),
			)
			if err != nil {
				return err
			}

			err = importProvider(ctx, logger, googleP, tags)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	googleCmd.AddCommand(googleResourcesCmd)

	// Required flags
	googleCmd.Flags().String("credentials", "", "path to the JSON credential (required)")
	googleCmd.Flags().String("project", "", "project (required)")
	googleCmd.Flags().String("region", "", "region (required)")

	// Filter flags
	googleCmd.Flags().StringSliceVarP(&tags, "labels", "l", []string{}, "List of labels to filter with format 'NAME:VALUE'")

	// Optional flags
	googleCmd.Flags().Uint64("max-results", 500, "max results to fetch when pagination is used")
}
