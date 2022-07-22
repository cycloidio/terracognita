package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	kitlog "github.com/go-kit/kit/log"

	"github.com/cycloidio/terracognita/aws"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/tag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tags []string

	awsCmd = &cobra.Command{
		Use:   "aws",
		Short: "Terracognita reads from AWS and generates hcl resources and/or terraform state",
		Long:  "Terracognita reads from AWS and generates hcl resources and/or terraform state",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := preRunEOutput(cmd, args)
			if err != nil {
				return err
			}

			viper.BindPFlag("aws-access-key", cmd.Flags().Lookup("aws-access-key"))
			viper.BindPFlag("aws-secret-access-key", cmd.Flags().Lookup("aws-secret-access-key"))
			viper.BindPFlag("aws-default-region", cmd.Flags().Lookup("aws-default-region"))
			viper.BindPFlag("aws-session-token", cmd.Flags().Lookup("aws-session-token"))

			viper.BindPFlag("aws-shared-credentials-file", cmd.Flags().Lookup("aws-shared-credentials-file"))
			viper.BindPFlag("aws-profile", cmd.Flags().Lookup("aws-profile"))

			viper.BindPFlag("tags", cmd.Flags().Lookup("tags"))

			// We define aliases so we have an easier access on the code
			viper.RegisterAlias("access-key", "aws-access-key")
			viper.RegisterAlias("secret-key", "aws-secret-access-key")
			viper.RegisterAlias("session-token", "aws-session-token")
			viper.RegisterAlias("region", "aws-default-region")

			return nil
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
				tg, err := tag.New(t)
				if err != nil {
					return fmt.Errorf("invalid format for --tags with value %q: %w", t, err)
				}
				tags = append(tags, tg)
			}

			ctx := context.Background()

			awsP, err := aws.NewProvider(ctx, viper.GetString("access-key"), viper.GetString("secret-key"), viper.GetString("region"), viper.GetString("session-token"))
			if err != nil {
				return err
			}

			err = importProvider(ctx, logger, awsP)
			if err != nil {
				return err
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
	awsCmd.Flags().String("aws-session-token", "", "Use to validate the temporary security credentials")
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

	if !viper.IsSet("session-token") {
		viper.Set("session-token", value.SessionToken)
	}

	return nil
}
