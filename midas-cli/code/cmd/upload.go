package cmd

import (
	"github.com/spf13/cobra"
	"midas-sdk/cli/midascli"
)

func NewUploadCommand(cli *midascli.midasCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload file to MIDAS_USERDIR",
		Args:  cobra.MaximumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			IsCover, _ := cmd.Flags().GetBool("cover")
			cli.XUpload(IsCover, args...)
		},
	}
	var IsCover bool
	cmd.Flags().BoolVarP(&IsCover, "cover", "c", false, "Cover remote directory?")
	return cmd
}
