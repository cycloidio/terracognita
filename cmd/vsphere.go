package cmd

import (
	"context"

	kitlog "github.com/go-kit/kit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/vsphere"
)

var (
	vsphereCmd = &cobra.Command{
		Use:   "vsphere",
		Short: "Terracognita reads from vSphere and generates hcl resources and/or terraform state",
		Long:  "Terracognita reads from vSphere and generates hcl resources and/or terraform state",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := preRunEOutput(cmd, args)
			if err != nil {
				return err
			}
			viper.BindPFlag("soap-url", cmd.Flags().Lookup("soap-url"))
			viper.BindPFlag("username", cmd.Flags().Lookup("username"))
			viper.BindPFlag("password", cmd.Flags().Lookup("password"))
			viper.BindPFlag("vsphereserver", cmd.Flags().Lookup("vsphereserver"))
			viper.BindPFlag("insecure", cmd.Flags().Lookup("insecure"))

			return nil
		},
		PostRunE: postRunEOutput,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Get()
			logger = kitlog.With(logger, "func", "cmd.vsphere.RunE")
			// Validate required flags
			if err := requiredStringFlags("soap-url", "username", "password"); err != nil {
				return err
			}

			ctx := context.Background()

			vsphereProvider, err := vsphere.NewProvider(
				ctx,
				viper.GetString("soap-url"),
				viper.GetString("username"),
				viper.GetString("password"),
				viper.GetString("vsphereserver"),
				viper.GetBool("insecure"),
			)
			if err != nil {
				return err
			}

			err = importProvider(ctx, logger, vsphereProvider)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	vsphereCmd.AddCommand(vsphereResourcesCmd)

	// Required flags
	vsphereCmd.Flags().String("soap-url", "", "URL of a vCenter or ESXi instance (required)")
	vsphereCmd.Flags().String("username", "", "Username (required)")
	vsphereCmd.Flags().String("password", "", "Password (required)")
	vsphereCmd.Flags().String("vsphereserver", "", "This is the vCenter Server FQDN or IP Address for vSphere API operations (required)")
	vsphereCmd.Flags().Bool("insecure", true, "Insecure")
}
