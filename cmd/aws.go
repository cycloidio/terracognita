package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cycloidio/terraforming/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tags []string

	awsCmd = &cobra.Command{
		Use:   "aws",
		Short: "Terraforming reads from AWS and generates TF",
		Long:  "Terraforming reads from AWS and generates TF",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate required flags
			if err := requiredStringFlags("access-key", "secret-key", "region"); err != nil {
				return err
			}

			// Initialize the tags
			tags := make([]aws.Tag, 0, len(viper.GetStringSlice("tags")))
			for _, t := range viper.GetStringSlice("tags") {
				values := strings.Split(t, ":")
				if len(values) != 2 {
					return errors.New("invalid format for --tags, the expected format is 'NAME:VALUE'")
				}
				tags = append(tags, aws.Tag{Name: values[0], Value: values[1]})
			}

			ctx := context.Background()

			err := aws.Import(
				ctx, viper.GetString("access-key"), viper.GetString("secret-key"), viper.GetString("region"),
				tags, viper.GetStringSlice("include"), viper.GetStringSlice("exclude"), out,
			)
			if err != nil {
				return fmt.Errorf("could not import from AWS: %s", err)
			}

			return nil
		},
	}
)

func init() {

	// Required flags

	awsCmd.Flags().String("access-key", "", "Access Key (required)")
	viper.BindPFlag("access-key", awsCmd.Flags().Lookup("access-key"))

	awsCmd.Flags().String("secret-key", "", "Secret Key (required)")
	viper.BindPFlag("secret-key", awsCmd.Flags().Lookup("secret-key"))

	awsCmd.Flags().String("region", "", "Region to search in, for now * it's not supported (required)")
	viper.BindPFlag("region", awsCmd.Flags().Lookup("region"))

	// Filter flags

	awsCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "List of tags to filter with format 'NAME:VALUE'")
	viper.BindPFlag("tags", awsCmd.Flags().Lookup("tags"))

}

func requiredStringFlags(names ...string) error {
	for _, n := range names {
		if viper.GetString(n) == "" {
			return fmt.Errorf("the flag %q is required", n)
		}
	}

	return nil
}
