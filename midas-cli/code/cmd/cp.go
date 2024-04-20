package cmd

import (
	"github.com/spf13/cobra"
	"midas-sdk/cli/midascli"
)

func NewCopyCommand(cli *midascli.midasCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cp",
		Short: "Copy file/directory to user's directory",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			IsDir, _ := cmd.Flags().GetBool("recursive")
			cli.XCP(IsDir, args...)
		},
	}
	var IsDir bool
	cmd.Flags().BoolVarP(&IsDir, "recursive", "r", false, "recursively copy")

	return cmd
}
