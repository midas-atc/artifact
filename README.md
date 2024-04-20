Welcome to the MIDAS Project Source repository!
=====

## midas-cli directory
The `midas-cli` directory houses the source code for the CLI that interfaces with the MIDAS cluster. This includes the Makefile for building and xcompile.sh for cross-compilation across Linux, MacOS, and Windows platforms. 

The source code itself can be found in `midas-cli/code`. Refer to `midas-cli/README.me` for usage instructions and `midas-cli/code/README.md` for detailed developer documentation on the CLI.

## server-side directory
Configurations for server setups used in the MIDAS project, include configurations for Docker, GlusterFS, and Slurm and its scheduler plugin. 

We are still in the process of anonymizing and adding more system scripts and config files.

## example directory
Example code and projects for machine learning frameworks, including MXNet, PyTorch, TensorFlow. A simple "helloworld" project is also included for quickstart.

## trace directory
The trace directory includes the schema for the workload trace. However, please note that the actual trace data is pending approval from the university and will be released once approved.

## To be added
* cluster specification and topology, as suggested by ATC reviewers