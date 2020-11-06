package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
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

			viper.BindPFlag("aws-access-key", cmd.Flags().Lookup("aws-access-key"))
			viper.BindPFlag("aws-secret-access-key", cmd.Flags().Lookup("aws-secret-access-key"))
			viper.BindPFlag("aws-default-region", cmd.Flags().Lookup("aws-default-region"))

			viper.BindPFlag("aws-shared-credentials-file", cmd.Flags().Lookup("aws-shared-credentials-file"))
			viper.BindPFlag("aws-profile", cmd.Flags().Lookup("aws-profile"))

			viper.BindPFlag("tags", cmd.Flags().Lookup("tags"))

			// We define aliases so we have an easier access on the code
			viper.RegisterAlias("access-key", "aws-access-key")
			viper.RegisterAlias("secret-key", "aws-secret-key")
			viper.RegisterAlias("region", "aws-default-region")
		},
		PostRunE: postRunEOutput,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Get()
			logger = kitlog.With(logger, "func", "cmd.aws.RunE")

			loadAWSCredentials()

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
	awsCmd.Flags().String("aws-access-key", "", "Access Key (required)")
	awsCmd.Flags().String("aws-secret-access-key", "", "Secret Key (required)")
	awsCmd.Flags().String("aws-default-region", "", "Region to search in, for now * is not supported (required)")
	awsCmd.Flags().String("aws-shared-credentials-file", "", "Path to the AWS credential path")
	awsCmd.Flags().String("aws-profile", "", "Name of the Profile to use with the Credentials")

	// Filter flags
	awsCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "List of tags to filter with format 'NAME:VALUE'")
}

// loadAWSCredentials will first read from ENV and if AccessKey and SecretAccessKey are not found (both of them)
// will fallback to the SharedCredentials with the profile
func loadAWSCredentials() error {
	creds := credentials.NewCredentials(&credentials.ChainProvider{
		Providers: []credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{Filename: viper.GetString("aws-shared-credentials-file"), Profile: viper.GetString("aws-profile")},
		},
	})

	value, err := creds.Get()
	if err != nil {
		// The NoCredentialProviders is an error returned by Get to identify that none
		// of the Providers (credentials.EnvProvider and credentials.SharedCredentialsProvider)
		// did find any information.
		// So we escape it means nothing was found by AWS
		if awsE, ok := err.(awserr.Error); ok && awsE.Code() == "NoCredentialProviders" {
			return nil
		}
		return err
	}

	// If the values are already set
	// it'll not be override as they
	// are more relevant
	if !viper.IsSet("access-key") {
		viper.Set("access-key", value.AccessKeyID)
	}

	if !viper.IsSet("secret-key") {
		viper.Set("secret-key", value.SecretAccessKey)
	}

	return nil
}
