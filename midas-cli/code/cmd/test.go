package cmd

import (
	"midas-sdk/cli/midascli"

	"github.com/spf13/cobra"
)

func NewTestCommand(cli *midascli.midasCli) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "For test only",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli.XTest(args...)
		},
	}
}
