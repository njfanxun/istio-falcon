package cmd

import "github.com/spf13/cobra"

func InitCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "istio-falcon",
		Short: "Monitor istio Gateway CRD, auto expose external port",
	}
	command.AddCommand(InitVersionCommand())
	command.AddCommand(InitManagerCommand())
	return command
}
