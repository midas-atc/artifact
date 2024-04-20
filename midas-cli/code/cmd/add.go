package cmd

import (
	"github.com/spf13/cobra"
	"midas-sdk/cli/midascli"
)

func NewAddCommand(cli *midascli.midasCli) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add dependency to MidasJob.conf file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli.XAdd(args...)
		},
	}
}
