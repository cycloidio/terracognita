package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cycloidio/terracognita/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	hclOut                    io.Writer
	stateOut                  io.Writer
	closeOut                  []io.Closer
	include, exclude, targets []string
	logsOut                   io.Writer

	// RootCmd it's the entry command for the cmd on terracognita
	RootCmd = &cobra.Command{
		Use:   "terracognita",
		Short: "Reads from Providers and generates a Terraform configuration",
		Long:  "Reads from Providers and generates a Terraform configuration, all the flags can be used also with ENV (ex: --access-key == ACCESS_KEY)",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize the logs by setting by default the logs
			// to Stdout, but if 'v' or 'd' is defined the logger
			// will be initialized and structured logs will be used
			// and if 'd' it's defined TF_LOG will be used too
			if viper.GetBool("verbose") || viper.GetBool("debug") {
				logsOut = ioutil.Discard
				log.Init(os.Stdout, viper.GetBool("debug"))
			} else {
				logsOut = os.Stdout
				log.Init(ioutil.Discard, false)
			}
		},
	}
)

func requiredStringFlags(names ...string) error {
	for _, n := range names {
		if viper.GetString(n) == "" {
			return fmt.Errorf("the flag %q is required", n)
		}
	}

	return nil
}

func preRunEOutput(cmd *cobra.Command, args []string) error {
	// Initializes/Validates the HCL and TFSTATE flags
	closeOut = make([]io.Closer, 0)
	if viper.GetString("hcl") != "" {
		f, err := os.OpenFile(viper.GetString("hcl"), os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("could not OpenFile %s because: %s", viper.GetString("hcl"), err)
		}
		hclOut = f
		closeOut = append(closeOut, f)
	}
	if viper.GetString("tfstate") != "" {
		f, err := os.OpenFile(viper.GetString("tfstate"), os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("could not OpenFile %s because: %s", viper.GetString("tfstate"), err)
		}
		stateOut = f
		closeOut = append(closeOut, f)
	}

	if len(closeOut) == 0 {
		return fmt.Errorf("one of --hcl or --tfstate are required")
	}
	return nil
}

func postRunEOutput(cmd *cobra.Command, args []string) error {
	// Closes all the opened files
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
	RootCmd.AddCommand(googleCmd)
	RootCmd.AddCommand(versionCmd)

	RootCmd.PersistentFlags().String("hcl", "", "HCL output file")
	_ = viper.BindPFlag("hcl", RootCmd.PersistentFlags().Lookup("hcl"))

	RootCmd.PersistentFlags().String("tfstate", "", "TFState output file")
	_ = viper.BindPFlag("tfstate", RootCmd.PersistentFlags().Lookup("tfstate"))

	RootCmd.PersistentFlags().StringSliceVarP(&include, "include", "i", []string{}, "List of resources to import, this names are the ones on TF (ex: aws_instance). If not set then means that all the resources will be imported")
	_ = viper.BindPFlag("include", RootCmd.PersistentFlags().Lookup("include"))

	RootCmd.PersistentFlags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "List of resources to not import, this names are the ones on TF (ex: aws_instance). If not set then means that none the resources will be excluded")
	_ = viper.BindPFlag("exclude", RootCmd.PersistentFlags().Lookup("exclude"))

	RootCmd.PersistentFlags().StringSliceVar(&targets, "target", []string{}, "List of resources to import via ID, those IDs are the ones documented on Terraform that are needed to Import. The format is 'aws_instance.ID'")
	_ = viper.BindPFlag("target", RootCmd.PersistentFlags().Lookup("target"))

	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Activate the verbose mode")
	_ = viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))

	RootCmd.PersistentFlags().BoolP("debug", "d", false, "Activate the debug mode wich includes TF logs via TF_LOG=TRACE|DEBUG|INFO|WARN|ERROR configuration https://www.terraform.io/docs/internals/debugging.html")
	_ = viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
}

func initViper() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
