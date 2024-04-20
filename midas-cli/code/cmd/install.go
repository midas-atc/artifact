package cmd

import (
	"github.com/spf13/cobra"
	"midas-sdk/cli/midascli"
)

func NewInstallCommand(cli *midascli.midasCli) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install environment at localhost",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli.XInstall(args...)
		},
	}
}
