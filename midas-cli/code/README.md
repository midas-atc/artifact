## Code Structure of midas-SDK

#### Directory description

+ Directory tree

  ~~~
  .
  ├── Makefile
  ├── README.md
  ├── cli
  │   ├── Makefile
  │   ├── cmd
  │   │   ├── add.go
  │   │   ├── config.go
  │   │   ├── download.go
  │   │   ├── init.go
  │   │   ├── install.go
  │   │   ├── log.go
  │   │   ├── ps.go
  │   │   └── submit.go
  │   ├── main.go
  │   └── midascli
  │       ├── midascli.go
  │       ├── MidasJobconfig.go
  │       └── userconfig.go
  ├── doc
  └── example
  ~~~

+ `cli` directory

  + `cmd`
    + This folder stores the description of all midas commands.
    + Each file corresponds to a unique midas command. E.g., `submit.go` is related to command `midas submit`.
    + The function in each file defines the information of the corresponding command,
    and is called when user runs that command. The information of the command may include:  command description, requirement of args, and callable subcommand function.
  + `main.go`
    + In this file, we define the main function of CLI. 
    + New command can be added in the main function as subcommand of `midas`.
  + `midascli`
    + In this folder, we define all operations of `midas command-line`.
    + `midascli.go` defines the concrete operations of CLI subcommands. It packages each operation into `X<func>` functions for calling.
    + `MidasJobconfig.go` defines the parsing and env operations upon `MidasJob.conf`, The operations vary in: MidasJob.conf parsing, `<conf_file>` generating and `MIDAS_env` configuring.
    + `userconfig.go` defines the user configuration functions for CLI, which are called when initializing CLI. `userconfig` includes: __UserName__, __SSHpath__ to MIDAS cluster, __Authfile__ for user to authenticate MIDAS cluster, __Dir__ defines user's directory in MIDAS. (By default, Dir[0]=`RepoDir`, Dir[1]=`UserDir`), __path__ is where the file locally stored.

#### How to add a new command(top down view)

Here we set `midas submit` as an example to illustrate the process for adding a new command.

+ Register new command in `main.go`:

  ~~~go
  func newmidasCommand(cli *midascli.midasCli) *cobra.Command {
  	var midasCmd = &cobra.Command{
  		Use:     "midas",
  		Short:   "MIDAS Command-line Interface v" + VERSION,
  		Version: VERSION,
  	}
  	midasCmd.AddCommand(cmd.NewSubmitCommand(cli)) // register submit command
  	...
  	return midasCmd
  }
  ~~~

+ Add callable command in `cmd` directory

  ~~~go
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
  ~~~

+ Add operations of `midas submit` in `midascli/midascli.go`

  ~~~go
  func (midascli *midasCli) XSubmit(args ...string) bool {
  	...
  	return false
  }
  ~~~

  
