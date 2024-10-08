# Example: Containerized hosts, but not switch

In this example, we have containerized all hosts. The switches are not containerized, but connected to network bridges via virtual ethernet (veth) pairs.

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
make clean
```

To show logs of a given container, run:
```bash
sudo podman logs -f <host_id>
```

To show logs of e.g. the s1 switch, run:
```bash
tail -f ../../load_balancer/logs/s1.log
```

## Troubleshooting

This example requires Podman version 5.2.0 or newer, which includes a [patch](https://github.com/containers/podman/pull/23056)
that enables container restore into a Pod.

```
Error: cannot add container f96670b26e53e70f7f451191ea39a093c940c6c48b47218aeeef1396cb860042 to pod h2-pod: no such pod
```