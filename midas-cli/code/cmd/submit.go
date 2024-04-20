package cmd

import (
	"github.com/spf13/cobra"
	"midas-sdk/cli/midascli"
)

func NewSubmitCommand(cli *midascli.midasCli) *cobra.Command {
	return &cobra.Command{
		Use:   "submit",
		Short: "Submit a job to MIDAS",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli.XSubmit(args...)
		},
	}
}
