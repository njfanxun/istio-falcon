package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "v0.0.1"

func InitVersionCommand() *cobra.Command {

	return &cobra.Command{
		Use:   "version",
		Short: "istio-falcon version and release information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Istio-Falcon Release Information")
			fmt.Printf("Version: %s", version)
		},
	}
}
