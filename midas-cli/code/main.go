package main

import (
	"log"

	"github.com/spf13/cobra"

	"os"
	"os/exec"
	"path/filepath"
	"midas-sdk/cli/cmd"
	"midas-sdk/cli/midascli"
)

var VERSION = "0.4.1"

func main() {
	home := homeDIR()
	midasInit(home)
	userConfig := midascli.NewUserConfig(filepath.Join(home, ".midas", ".userconfig"))
	clusterConfig := midascli.NewClusterConfig(filepath.Join(home, ".midas", ".clusterconfig"))
	cli := midascli.NewmidasCli(userConfig, clusterConfig)
	midasCmd := newmidasCommand(cli)
	if err := midasCmd.Execute(); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

}

func newmidasCommand(cli *midascli.midasCli) *cobra.Command {
	var midasCmd = &cobra.Command{
		Use:   "midas",
		Short: "MIDAS Command-line Interface " + VERSION,
	}
	midasCmd.AddCommand(cmd.NewSubmitCommand(cli))
	midasCmd.AddCommand(cmd.NewConfigCommand(cli))
	midasCmd.AddCommand(cmd.NewPSCommand(cli))
	midasCmd.AddCommand(cmd.NewCancelCommand(cli))
	midasCmd.AddCommand(cmd.NewInitCommand(cli))
	midasCmd.AddCommand(cmd.NewUploadCommand(cli))
	midasCmd.AddCommand(cmd.NewDownloadCommand(cli))
	midasCmd.AddCommand(cmd.NewAddCommand(cli))
	midasCmd.AddCommand(cmd.NewInstallCommand(cli))
	// midasCmd.AddCommand(cmd.NewDatasetCommand(cli))
	midasCmd.AddCommand(cmd.NewLSCommand(cli))
	midasCmd.AddCommand(cmd.NewCatCommand(cli))
	midasCmd.AddCommand(cmd.NewENVLSCommand(cli))

	// midasCmd.AddCommand(cmd.NewTestCommand(cli))

	var Verbose bool
	midasCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	return midasCmd
}

func homeDIR() string {
	return os.Getenv("HOME")
}

func midasInit(home string) bool {
	log.SetPrefix("[midas Error] ")
	log.SetFlags(log.Ldate | log.Lshortfile)

	midasDIR := filepath.Join(home, ".midas")
	init_cmd := exec.Command("mkdir", "-p", midasDIR)
	if _, err := init_cmd.CombinedOutput(); err != nil {
		log.Println("Failed to obtain midas metadata. Error message:", err.Error())
		return false
	}
	file1, err := os.Open(filepath.Join(home, ".midas", ".userconfig"))
	if err != nil && os.IsNotExist(err) {
		os.Create(filepath.Join(home, ".midas", ".userconfig"))
	}
	file2, err := os.Open(filepath.Join(home, ".midas", ".clusterconfig"))
	if err != nil && os.IsNotExist(err) {
		os.Create(filepath.Join(home, ".midas", ".clusterconfig"))
	}

	defer file1.Close()
	defer file2.Close()

	return true
}
