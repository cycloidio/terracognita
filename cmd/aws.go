package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	kitlog "github.com/go-kit/kit/log"

	"github.com/cycloidio/terracognita/aws"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/hcl"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/state"
	"github.com/cycloidio/terracognita/tag"
	"github.com/cycloidio/terracognita/writer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tags []string

	awsCmd = &cobra.Command{
		Use:   "aws",
		Short: "Terracognita reads from AWS and generates hcl resources and/or terraform state",
		Long:  "Terracognita reads from AWS and generates hcl resources and/or terraform state",
		PreRun: func(cmd *cobra.Command, args []string) {
			preRunEOutput(cmd, args)
			viper.BindPFlag("access-key", cmd.Flags().Lookup("access-key"))
			viper.BindPFlag("secret-key", cmd.Flags().Lookup("secret-key"))
			viper.BindPFlag("region", cmd.Flags().Lookup("region"))
			viper.BindPFlag("tags", cmd.Flags().Lookup("tags"))
		},
		PostRunE: postRunEOutput,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Get()
			logger = kitlog.With(logger, "func", "cmd.aws.RunE")
			// Validate required flags
			if err := requiredStringFlags("access-key", "secret-key", "region"); err != nil {
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

			awsP, err := aws.NewProvider(ctx, viper.GetString("access-key"), viper.GetString("secret-key"), viper.GetString("region"))
			if err != nil {
				return err
			}

			f := &filter.Filter{
				Tags:    tags,
				Include: include,
				Exclude: exclude,
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
			err = provider.Import(ctx, awsP, hclW, stateW, f, logsOut)
			if err != nil {
				return fmt.Errorf("could not import from AWS: %+v", err)
			}

			return nil
		},
	}
)

func init() {
	awsCmd.AddCommand(awsResourcesCmd)

	// Required flags
	awsCmd.Flags().String("access-key", "", "Access Key (required)")
	awsCmd.Flags().String("secret-key", "", "Secret Key (required)")
	awsCmd.Flags().String("region", "", "Region to search in, for now * it's not supported (required)")

	// Filter flags
	awsCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "List of tags to filter with format 'NAME:VALUE'")
}
