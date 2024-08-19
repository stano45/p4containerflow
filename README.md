# P4ContainerFlow

This is the repository for the Google Summer of Code project [P4-Enabled Container Migration in Kubernetes](https://summerofcode.withgoogle.com/programs/2024/projects/sYbpOJhD).

For more information about the project, please refer to the [final report](REPORT.md).

## Table of Contents
- [P4ContainerFlow](#p4containerflow)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
    - [Prerequisites](#prerequisites)
    - [Install Python Dependencies](#install-python-dependencies)
    - [Installing PI](#installing-pi)
    - [Build Custom Podman Images](#build-custom-podman-images)
  - [Running examples](#running-examples)


## Installation

### Prerequisites
- [Podman](https://podman.io/docs/installation)
- [Python 3](https://www.python.org/downloads/)
- [pip](https://pip.pypa.io/en/stable/installation/)
- [P4 Compiler (p4c)](https://github.com/p4lang/p4c)

### Install Python Dependencies
```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

### Installing PI
This project uses the P4Runtime API to communicate with the switch. The P4Runtime API is implemented in the [P4Runtime Interface (PI)](https://github.com/p4lang/PI).

When installing PI, make sure to configure with the `--with-proto` flag to compile proto files and `--with-python_prefix=/path/to/this/repo/.venv` to install the p4 library in your virtual environment. For example:
```bash
./configure --with-proto --with-python_prefix=/absolute/path/to/p4containerflow/.venv
```
After running:
```bash
sudo make install
```
The p4 library files in your `.venv` will be owned by the root user. Make sure to change owner by running:
```bash
sudo chown -R $USER .venv
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

Simply `cd` into the desired example directory and follow the instructions in the README.