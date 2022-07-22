package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cycloidio/terracognita/vsphere"
)

var (
	vsphereResourcesCmd = &cobra.Command{
		Use:   "resources",
		Short: "List of all the vSphere supported resources",
		Run: func(cmd *cobra.Command, args []string) {
			for _, r := range vsphere.ResourceTypeStrings() {
				fmt.Println(r)
			}
		},
	}
)
