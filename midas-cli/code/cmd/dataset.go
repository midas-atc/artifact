package cmd

import (
	"github.com/spf13/cobra"
	"midas-sdk/cli/midascli"
)

func NewDatasetCommand(cli *midascli.midasCli) *cobra.Command {
	return &cobra.Command{
		Use:   "dataset",
		Short: "Allow access to dataset",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli.XDataset(args...)
		},
	}
}
