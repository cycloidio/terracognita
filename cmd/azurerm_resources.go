package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cycloidio/terracognita/azurerm"
)

var (
	azurermResourcesCmd = &cobra.Command{
		Use:   "resources",
		Short: "List of all the AzureRM supported resources",
		Run: func(cmd *cobra.Command, args []string) {
			for _, r := range azurerm.ResourceTypeStrings() {
				fmt.Println(r)
			}
		},
	}
)
