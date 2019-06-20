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
	hcl              io.Writer
	tfstate          io.Writer
	closeOut         []io.Closer
	include, exclude []string

	// RootCmd it's the entry command for the cmd on terracognita
	RootCmd = &cobra.Command{
		Use:   "terracognita",
		Short: "Reads from Providers and generates a Terraform configuration",
		Long:  "Reads from Providers and generates a Terraform configuration, all the flags can be used also with ENV (ex: --access-key == ACCESS_KEY)",
	}
)

func preRunEOutput(cmd *cobra.Command, args []string) error {
	closeOut = make([]io.Closer, 0)
	if viper.GetString("hcl") != "" {
		f, err := os.OpenFile(viper.GetString("hcl"), os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("could not OpenFile %s because: %s", viper.GetString("hcl"), err)
		}
		hcl = f
		closeOut = append(closeOut, f)
	}
	if viper.GetString("tfstate") != "" {
		f, err := os.OpenFile(viper.GetString("tfstate"), os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("could not OpenFile %s because: %s", viper.GetString("tfstate"), err)
		}
		tfstate = f
		closeOut = append(closeOut, f)
	}

	if len(closeOut) == 0 {
		return fmt.Errorf("one of --hcl or --tfstate are required")
	}
	return nil
}

func postRunEOutput(cmd *cobra.Command, args []string) error {
	for _, c := range closeOut {
		if err := c.Close(); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	cobra.OnInitialize(initViper)
	RootCmd.AddCommand(awsCmd)
	RootCmd.AddCommand(versionCmd)

	RootCmd.PersistentFlags().String("hcl", "", "HCL output file")
	_ = viper.BindPFlag("hcl", RootCmd.PersistentFlags().Lookup("hcl"))

	RootCmd.PersistentFlags().String("tfstate", "", "TFState output file")
	_ = viper.BindPFlag("tfstate", RootCmd.PersistentFlags().Lookup("tfstate"))

	RootCmd.PersistentFlags().StringSliceVarP(&include, "include", "i", []string{}, "List of resources to import, this names are the ones on TF (ex: aws_instance). If not set then means that all the resources will be imported")
	_ = viper.BindPFlag("include", RootCmd.PersistentFlags().Lookup("include"))

	RootCmd.PersistentFlags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "List of resources to not import, this names are the ones on TF (ex: aws_instance). If not set then means that none the resources will be excluded")
	_ = viper.BindPFlag("exclude", RootCmd.PersistentFlags().Lookup("exclude"))
}

func initViper() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
