package cmd

import (
	"midas-sdk/cli/midascli"

	"github.com/spf13/cobra"
)

func NewENVLSCommand(cli *midascli.midasCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Check environment",
	}
	cmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
	cmd.PersistentFlags().Lookup("help").Hidden = true
	cmd.AddCommand(sub(cli))
	return cmd
}

func sub(cli *midascli.midasCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List environment name",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			IsEnv, _ := cmd.Flags().GetBool("env")
			cli.XENVLS(IsEnv, args...)
		},
	}
	var IsEnv bool
	cmd.Flags().BoolVarP(&IsEnv, "env", "n", false, "List environment packages.")
	return cmd
}
