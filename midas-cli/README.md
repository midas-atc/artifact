# midas-SDK
## Command-line Interface used for MIDAS job submission.
```
midas Command-line Interface v0.4.1

Usage:
midas [command] [flags] [args]

Available Commands:
    midas init
    midas config [-u/-f] [args]
    midas upload [-c] <local_dirpath> [<remote_dirpath>]
    midas download [<filepath>]
    midas add [<dependency_name>]
    midas submit [<path_to_repo>]
    midas ps [-j] [<JOB_ID>]
    midas install [<path_to_repo>]
    midas cancel [-j] [<JOB_ID>]
    midas ls [<dirpath>]

Use "midas [command] --help" for more information about commands.
```

## Installation
You can try out the latest features by directly install from master branch:

```
git clone https://github.com/MIDAS/midas-sdk.git
cd midas-sdk
echo 'export GOPATH=$PWD' >> ~/.bash_profile
make install
```

## Configuration
### CLI Configuration
1. Before using the midas CLI to submit ML jobs, you need to configure your MIDAS credentials. You can do this by running the `midas config` command:
```
$ midas config [-u/--username] MYUSERNAME
$ midas config [-f/--file] MYPRIVATEFILEPATH
```
2. You need to run `midas init` command to obtain the latest cluster hardware information from MIDAS cluster.

### Job Configuration
#### TUXIV.CONF

You can use `midas init` to pull the latest cluster configuration from MIDAS. There are four parts in `tuxiv.conf` that configure different parts of job submission. Noted that `tuxiv.conf` follows **yaml format**.

+ Entrypoint

  In this section, you should input you shell commands to run your code line-by-line. The midas CLI will help run the job according to your commands.

  ~~~yaml
  entrypoint:
      - python ${MIDAS_WORKDIR}/mnist.py --epoch=3
  ~~~

+ Environment

  In this section, you can specify your software  requirements, including the environment name, dependencies, source channels and so on. The midas CLI will help build your environment with *miniconda*.

  ~~~yaml
  environment:
      name: torch-env
      dependencies:
          - pytorch=1.6.0
          - torchvision=0.7.0
      channels: pytorch
  ~~~

+ Job

  In this section, you can specify your slurm configurations for slurm cluster resources, including number of nodes, CPUs, GPUs, output file and so on. All the slurm cluster configuration should be set in the general part.

  ~~~yaml
  job:
      name: test
      general:
          - nodes=2
          - output=${MIDAS_SLURM_USERLOG}/output.log
  ~~~

  **Note:** You can modify the output log path in Job section. For debugging purpose, we recommend you set the `output` value under `${MIDAS_USERDIR}` directory and check it using `midas ls` and `midas download`.

+ Datasets

  In this section, you can specify your required CityNet dataset name, and midas will help place the dataset access in `MIDAS_USERDIR`. You can view the table of CityNet datasets at [CityNet Dataset Info](https://docs.google.com/spreadsheets/d/18qi2YpYvuXkWns7KY9pHYQclhS1Yyt5ysqgZ4plYcTg/edit#gid=0).

  ~~~yaml
  datasets:
    - OpenRoadMap
  ~~~

#### MIDAS VARIABLES

+ `MIDAS_WORKDIR`: MIDAS job workspace directory. Each job has a different workspace directory.
+ `MIDAS_USERDIR`: MIDAS User directory.
+ `MIDAS_SLURM_USERLOG`: Slurm log directory. The default value is `${MIDAS_USERDIR}/slurm_log`.

## Example

Basic examples are provided under the [example](example) folder. These examples include: [HelloWorld](example/helloworld), [TensorFlow](example/TensorFlow), [PyTorch](example/PyTorch) and [MXNet](example/MXNet).

