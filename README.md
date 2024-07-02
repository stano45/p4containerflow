# P4ContainerFlow

This is the repository for the Google Summer of Code project [P4-Enabled Container Migration in Kubernetes](https://summerofcode.withgoogle.com/programs/2024/projects/sYbpOJhD). The project is actively being worked on.


## Running Hosts and Switches in Containers
### Prerequisites
- [Podman](https://podman.io/docs/installation)
- [P4 Compiler (p4c)](https://github.com/p4lang/p4c)
- [Python 3](https://www.python.org/downloads/)
- [pip](https://pip.pypa.io/en/stable/installation/)

### Install Python Dependencies
```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

### Build Custom Podman Images
```bash
make build-images
```
This will build the following images:
- `tcp-client`: A simple TCP client that sends a message to a server (this will run in h1-pod)
- `tcp-server`: A simple TCP server that listens for a message from a client (this will run in all other pods)

You can configure the target IP of the client and the port of the server in the [tcp/Containerfile.server](tcp/Containerfile.server) and [tcp/Containerfile.client](tcp/Containerfile.client) files respectively.

Furthermore, you can specify which image to run in the hosts by changing the `IMG` and `ARGS` variables in [scripts/switch_container/build.sh](scripts/switch_container/build.sh).

### Creating the Network Topology, Hosts and Switches
In the root directory of the repo, run:
```bash
make
```
This will create compile the p4 code, create 4 networks (h1-net, h2-net, h3-net, h4-net), a pod in each network (h1-pod, h2-pod, h3-pod, h4-pod) and a host container in each network (h1, h2, h3, h4). A switch (s1) will be created in the host network, connected to all the host networks. For details on the network topology, refer to [scripts/switch_container/build.sh](scripts/switch_container/build.sh).
Finally, the script will run the controller, which will programm the switch with the p4 code.

### Run the TCP Client
In a new terminal, run:
```bash
make tcp-client
```
This will run the `tcp-client` image in the `h1-pod`. The client will continuously send messages to the switch (load balancer), at address `10.1.1.11`, which should be load-balanced between h2 and h3.