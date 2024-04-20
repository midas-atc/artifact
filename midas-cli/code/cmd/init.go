package cmd

import (
	"github.com/spf13/cobra"
	"midas-sdk/cli/midascli"
)

func NewInitCommand(cli *midascli.midasCli) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "User init workspace and download the latest MIDAS config",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cli.XInit(args...)
		},
	}
}
