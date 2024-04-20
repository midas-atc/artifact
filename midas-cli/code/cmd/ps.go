package cmd

import (
	"github.com/spf13/cobra"
	"midas-sdk/cli/midascli"
)

func NewPSCommand(cli *midascli.midasCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "Check submitted jobs' status",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			job, err := cmd.Flags().GetString("job")
			if err != nil {
				cli.XPS("", args...)
			} else {
				cli.XPS(job, args...)
			}
		},
	}

	var job string
	cmd.Flags().StringVarP(&job, "job", "j", "", "Show <JOB_ID> status")
	return cmd
}
