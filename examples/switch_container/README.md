# Example: Containerized switch

This is the fullly containerized version of the project. All hosts and the switch are containerized. In this version, there is only one switch, instead of four, to simplify the setup.

The switch uses the network bridges of the four host networks as ports.

## Running the example

To build the topology, start containers and run the controller, run:
```bash
make
```

To start a tcp client, run:
```bash
make tcp-client
```
Alternatively, you can choose the iperf3 or netcat container image in the `build.sh` file, by uncommenting the corresponding `IMAGE` and `ARGS` variables.
Then, to run the corresponding client:
```bash
make iperf-client
# OR
make netcat-client
```

To migrate the container from the source host to the target host, and update the switch accordingly, run:
```bash
make migrate SOURCE=<> TARGET=<>
``` 

To cleanup the topology, networks, pods, and containers, run:
```bash
make teardown
```

To show logs of a given container, run:
```bash
sudo podman logs -f <host_id>
```

To show logs of the switch, run:
```bash
sudo podman logs -f s1
```

