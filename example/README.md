# midas Examples

MIDAS supports multiple ML frameworks such as TensorFlow, PyTorch and MXNet. We will later support some specialized ML framework like FATE, etc. Here we list several job examples of different frameworks.

## HelloWorld

+ CityNet Dataset: OpenRoadMap
+ Task: basic usage of midas
+ Code: [main.py](helloworld/main.py)

### Getting started

+ Install midas CLI, and run `midas init` to pull the latest cluster configurations from remote.

+ Configuration

  + Configure user information using `midas config`.

  + MIDAS ENV

    ~~~shell
    MIDAS_WORKDIR # default repo directory
    MIDAS_USERDIR # user directory
    MIDAS_SLURM_USERLOG # slurm log directory default: ${MIDAS_USERDIR}/slurm_log
    ~~~

  + MidasJob configuration

    ~~~yaml
    # MidasJob.conf
    entrypoint:
    - python ${MIDAS_WORKDIR}/main.py
    environment:
        name: hello 
        dependencies:
            - python=3.6.9
    job:
        name: test
        general:
            - output=${MIDAS_SLURM_USERLOG}/hello.out
            - nodes=1
            - ntasks=1
            - cpus-per-task=1
    datasets:
      - OpenRoadMap
    ~~~

  + Model code modification

    ~~~python
    import os
    import shutil
    # get variables from env
    WORKDIR = os.environ.get('MIDAS_WORKDIR')
    USERDIR = os.environ.get('MIDAS_USERDIR')
    # show the directory tree
    os.system('tree -L 2 {}'.format(USERDIR))
    # basic copy operation
    shutil.copytree(WORKDIR, "{}/helloworld".format(USERDIR))
    ~~~

### Submit job

+ Enter the `helloworld` directory and follow the following steps.
+ Build environment and submit job: `midas submit`
+ Monitor job: `midas ps [-j] [<JOB_ID>]`
+ Obtain log: `midas download helloworld/slurm_log/hello.out`
+ Cancel job: `midas cancel [-j] [<JOB_ID>]`
+ View UserDir: `midas ls <PATH>`



## TensorFlow

+ Dataset: mnist
+ Task: image classification
+ Code: [mnist.py](TensorFlow/mnist.py)

### Getting started

+ Install midas CLI, and run `midas init` to pull cluster configurations from remote.

+ Configuration

  + Config user informations using `midas config`.

  + MIDAS ENV

    ~~~shell
    MIDAS_WORKDIR # default repo directory
    MIDAS_USERDIR # user directory
    MIDAS_SLURM_USERLOG # slurm log directory default: ${MIDAS_USERDIR}/slurm_log
    ~~~

  + MidasJob configuration
  
    ~~~yaml
    # MidasJob.conf
    entrypoint:
        - python ${MIDAS_WORKDIR}/mnist.py 
        - --task_index=0
        - --data_dir=${MIDAS_WORKDIR}/datasets/mnist_data
        - --batch_size=1
    environment:
        name: tf 
        dependencies:
            - tensorflow=1.15
    job:
        name: mnist
        general:
          - nodes=2
    ~~~

  + Model code modification
  
    Use ` tf.distribute.cluster_resolver.SlurmClusterResolver`  instead of other resolvers.

### Training

+ Enter the `TensorFlow` directory and follow the following steps.
+ Build environment and submit job: `midas submit`
+ Monitor job: `midas ps [-j] [<JOB_ID>]`
+ Cancel job: `midas cancel [-j] [<JOB_ID>]`
+ View UserDir: `midas ls <PATH>`



## PyTorch

+ Dataset: mnist
+ Task: image classification
+ Code: [mnist.py](PyTorch/mnist.py)

### Getting started

+ Install midas CLI, and run `midas init` to pull cluster configurations from remote.

+ Configuration

  + Config user informations using `midas config`.

  + MIDAS ENV

    ~~~shell
    MIDAS_WORKDIR # default repo directory
    MIDAS_USERDIR # user directory
    MIDAS_SLURM_USERLOG # slurm log directory default: ${MIDAS_USERDIR}/slurm_log
    ~~~

  + MidasJob configuration
  
    ~~~yaml
    # MidasJob.conf
    entrypoint:
        - python ${MIDAS_WORKDIR}/mnist.py --epoch=3
    environment:
        name: torch-env
        dependencies:
            - pytorch=1.6.0
            - torchvision=0.7.0
        channels: pytorch
    job:
        name: test
        general:
          - nodes=2
    ~~~

  + Model code modification

    Obtain environment variables from slurm cluster, and set the parameters for initialize the cluster.
  
    ~~~python
    # example
    def dist_init(host_addr, rank, local_rank, world_size, port=23456):
        host_addr_full = 'tcp://' + host_addr + ':' + str(port)
        torch.distributed.init_process_group("gloo", init_method=host_addr_full,
                                             rank=rank, world_size=world_size)
      assert torch.distributed.is_initialized()
    
    def get_ip(iplist):
        ip = iplist.split('[')[0] + iplist.split('[')[1].split('-')[0]
        
    rank = int(os.environ['SLURM_PROCID'])
    local_rank = int(os.environ['SLURM_LOCALID'])
    world_size = int(os.environ['SLURM_NTASKS'])
    iplist = os.environ['SLURM_STEP_NODELIST']
    ip = get_ip(iplist) # function get_ip() is depends on the format of nodelist 
    dist_init(ip, rank, local_rank, world_size)
    ~~~

### Training

+ Enter the `PyTorch` directory and follow the following steps.
+ Build environment and submit job: `midas submit`
+ Monitor job: `midas ps [-j] [<JOB_ID>]`
+ Cancel job: `midas cancel [-j] [<JOB_ID>]`
+ View UserDir: `midas ls <PATH>`



## MXNet

+ Dataset: mnist
+ Task: image classification
+ Code: [mnist.py](MXNet/mnist.py)

### Getting started

+ Install midas CLI, and run `midas init` to pull cluster configurations from remote.

+ Configuration

  + Config user informations using `midas config`.

  + MIDAS ENV

    ~~~shell
    MIDAS_WORKDIR # default repo directory
    MIDAS_USERDIR # user directory
    MIDAS_SLURM_USERLOG # slurm log directory default: ${MIDAS_USERDIR}/slurm_log
    ~~~

  + MidasJob configuration
  
    ~~~yaml
    # MidasJob.conf
    entrypoint:
    - python ${MIDAS_WORKDIR}/mnist.py
    environment:
        name: mxnet-env 
        dependencies:
            - mxnet=1.5.0
    job:
        name: test
        general:
          - nodes=2
    ~~~

  + Model code modification
  
    Obtain environment variables from slurm cluster, and set the parameters for initialize the cluster.

### Training

+ Enter the `MXNet` directory and follow the following steps.
+ Build environment and submit job: `midas submit`
+ Monitor job: `midas ps [-j] [<JOB_ID>]`
+ Cancel job: `midas cancel [-j] [<JOB_ID>]`
+ View UserDir: `midas ls <PATH>`

