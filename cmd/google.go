package cmd

import (
	"context"
	"fmt"
	"strings"

	kitlog "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/google"
	"github.com/cycloidio/terracognita/hcl"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/state"
	"github.com/cycloidio/terracognita/tag"
	"github.com/cycloidio/terracognita/writer"
)

var (
	googleCmd = &cobra.Command{
		Use:   "google",
		Short: "Terracognita reads from GCP and generates hcl resources and/or terraform state",
		Long:  "Terracognita reads from GCP and generates hcl resources and/or terraform state",
		PreRun: func(cmd *cobra.Command, args []string) {
			preRunEOutput(cmd, args)
			viper.BindPFlag("credentials", cmd.Flags().Lookup("credentials"))
			viper.BindPFlag("project", cmd.Flags().Lookup("project"))
			viper.BindPFlag("region", cmd.Flags().Lookup("region"))
			viper.BindPFlag("labels", cmd.Flags().Lookup("labels"))
			viper.BindPFlag("max-results", cmd.Flags().Lookup("max-results"))
		},
		PostRunE: postRunEOutput,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Get()
			logger = kitlog.With(logger, "func", "cmd.google.RunE")
			// Validate required flags
			if err := requiredStringFlags("region", "project", "credentials"); err != nil {
				return err
			}

			// Initialize the tags
			tags := make([]tag.Tag, 0, len(viper.GetStringSlice("tags")))
			for _, t := range viper.GetStringSlice("tags") {
				values := strings.Split(t, ":")
				if len(values) != 2 {
					return errors.New("invalid format for --tags, the expected format is 'NAME:VALUE'")
				}
				tags = append(tags, tag.Tag{Name: values[0], Value: values[1]})
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

			f := &filter.Filter{
				Tags:    tags,
				Include: include,
				Exclude: exclude,
				Targets: targets,
			}

			var hclW, stateW writer.Writer

			if hclOut != nil {
				logger.Log("msg", "initialzing HCL writer")
				hclW = hcl.NewWriter(hclOut)
			}

			if stateOut != nil {
				logger.Log("msg", "initialzing TFState writer")
				stateW = state.NewWriter(stateOut)
			}

			logger.Log("msg", "importing")

			fmt.Fprintf(logsOut, "Starting Terracognita with version %s\n", Version)
			logger.Log("msg", "starting terracognita", "version", Version)
			err = provider.Import(ctx, googleP, hclW, stateW, f, logsOut)
			if err != nil {
				return errors.Wrap(err, "could not import from google")
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
	googleCmd.Flags().StringSliceVarP(&tags, "labels", "t", []string{}, "List of labels to filter with format 'NAME:VALUE'")

	// Optional flags
	googleCmd.Flags().Uint64("max-results", 500, "max results to fetch when pagination is used")
}
