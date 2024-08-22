# P4ContainerFlow

This is the repository for the Google Summer of Code project [P4-Enabled Container Migration in Kubernetes](https://summerofcode.withgoogle.com/programs/2024/projects/sYbpOJhD).

For more information about the project, please refer to the [final report](REPORT.md).

## Table of Contents
- [P4ContainerFlow](#p4containerflow)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
    - [Prerequisites](#prerequisites)
    - [Install Python Dependencies](#install-python-dependencies)
    - [Build Custom Podman Images](#build-custom-podman-images)
  - [Running examples](#running-examples)


## Installation

### Prerequisites
- [Python 3](https://www.python.org/downloads/) (3.10+)
- [pip3](https://pip.pypa.io/en/stable/installation/)
- [CRIU](https://criu.org/Main_Page) (v3.19)
- [crit](https://criu.org/CRIT) (v3.19)
- [P4 Compiler](https://github.com/p4lang/p4c) (v1.2.4.14)
- [PI](https://github.com/p4lang/PI)
- [Podman](https://podman.io/docs/installation) (v5.2.1)

We have provided [scripts](scripts) to install CRIU (with crit), the P4 compiler, PI, and Podman. The scripts have been tested on Ubuntu 22.04 and 24.04 and are not guaranteed to work on all machines. If you encounter any issues, please refer to the official documentation of the respective projects.

### Install Python Dependencies
```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

### Build Custom Podman Images
```bash
make images
```
This will build the following images:
- `tcp-client`: A simple TCP client that sends a message to a server (this will run in h1-pod)
- `tcp-server`: A simple TCP server that listens for a message from a client (this will run in all other pods)

You can configure the target IP of the client and the port of the server in the [tcp/Containerfile.server](tcp/Containerfile.server) and [tcp/Containerfile.client](tcp/Containerfile.client) files respectively.

Furthermore, you can specify which image to run in the hosts by changing the `IMG` and `ARGS` variables in [scripts/switch_container/build.sh](scripts/switch_container/build.sh).

## Running examples
There are three examples in the `examples` directory:
- [process_migration](examples/process_migration): Process migration demo using network namespaces
- [host_containers](examples/host_containers): Container migration demo using containerized hosts, but not switch
- [switch_container](examples/switch_container): Container migration demo with all hosts and the switch containerized
- [redis](examples/redis): Redis container migration demo using the [Redis](https://redis.io/) in-memory database
- [container_migration_in_kubernetes](examples/container_migration_in_kubernetes): Container migration demo in Kubernetes

Simply `cd` into the desired example directory and follow the instructions in the README.