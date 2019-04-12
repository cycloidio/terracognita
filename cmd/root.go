package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	out              io.Writer
	closeOut         io.Closer
	include, exclude []string

	// RootCmd it's the entry command for the cmd on terraforming
	RootCmd = &cobra.Command{
		Use:   "terraforming",
		Short: "Reads from Providers and generates a Terraform configuration",
		Long:  "Reads from Providers and generates a Terraform configuration, all the flags can be used also with ENV (ex: --access-key == ACCESS_KEY)",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetString("out") == "" {
				out = os.Stdout
			} else {
				f, err := os.OpenFile(viper.GetString("out"), os.O_RDWR|os.O_CREATE, 0755)
				if err != nil {
					return fmt.Errorf("could not OpenFile %s because: %s", viper.GetString("out"), err)
				}
				out = f
				closeOut = f
			}
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if closeOut != nil {
				if err := closeOut.Close(); err != nil {
					return fmt.Errorf("could not close the output file %s because: %s", viper.GetString("out"), err)
				}
			}

			return nil
		},
	}
)

func init() {
	cobra.OnInitialize(initViper)
	RootCmd.AddCommand(awsCmd)

	RootCmd.PersistentFlags().StringP("out", "o", "", "Output file of the config, by default STDOUT")
	viper.BindPFlag("out", RootCmd.PersistentFlags().Lookup("out"))

	RootCmd.PersistentFlags().StringSliceVarP(&include, "include", "i", []string{}, "List of resources to import, this names are the ones on TF (ex: aws_instance). If not set then means that all the resources will be imported")
	viper.BindPFlag("include", RootCmd.PersistentFlags().Lookup("include"))

	RootCmd.PersistentFlags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "List of resources to not import, this names are the ones on TF (ex: aws_instance). If not set then means that none the resources will be excluded")
	viper.BindPFlag("exclude", RootCmd.PersistentFlags().Lookup("exclude"))

	RootCmd.PersistentFlags().BoolP("tf-state", "s", false, "Import the Terraform State. If activated no TF files will be generated, only TF or TFState can work at the same time")
	viper.BindPFlag("tf-state", RootCmd.PersistentFlags().Lookup("tf-state"))
}

func initViper() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
