package cmd

import (
	"fmt"

	"github.com/cycloidio/terracognita/aws"
	"github.com/spf13/cobra"
)

var (
	awsResourcesCmd = &cobra.Command{
		Use:   "resources",
		Short: "List of all the AWS supported Resources",
		Run: func(cmd *cobra.Command, args []string) {
			for _, r := range aws.ResourceTypeStrings() {
				fmt.Println(r)
			}
		},
	}
)
