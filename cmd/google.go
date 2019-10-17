package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	googleCmd = &cobra.Command{
		Use:      "google",
		Short:    "Terracognita reads from GCP and generates hcl resources and/or terraform state",
		Long:     "Terracognita reads from GCP and generates hcl resources and/or terraform state",
		PreRunE:  preRunEOutput,
		PostRunE: postRunEOutput,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Not implemented yet")
			return nil
		},
	}
)

func init() {
	googleCmd.AddCommand(googleResourcesCmd)
}
