package cmd

import (
	"fmt"

	"github.com/cycloidio/terracognita/google"
	"github.com/spf13/cobra"
)

var (
	googleResourcesCmd = &cobra.Command{
		Use:   "resources",
		Short: "List of all the Google supported Resources",
		Run: func(cmd *cobra.Command, args []string) {
			for _, r := range google.ResourceTypeStrings() {
				fmt.Println(r)
			}
		},
	}
)
