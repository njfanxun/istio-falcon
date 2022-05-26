package cmd

import (
	"github.com/njfanxun/istio-falcon/pkg/version"

	"github.com/spf13/cobra"
)

func InitVersionCommand() *cobra.Command {

	return &cobra.Command{
		Use:   "version",
		Short: "istio-falcon version and release information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Get())
		},
	}
}
