package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the value of the current verion, this
	// is set via -ldflags
	Version string

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints the current build version",
		Run: func(cmd *cobra.Command, args []string) {
			if Version != "" {
				fmt.Printf("The current version is: %s\n", Version)
			} else {
				fmt.Printf("No version defined\n")
			}
		},
	}
)
