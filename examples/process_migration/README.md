# Example:  Process migration

This is the simplest initial version of the system. The hosts are run in separate network namespaces, and connected to the switch via virtual ethernet (veth) pairs.

## Running the example

To build the topology, start containers and run the controller, run:
```bash
make
```

To run a xterm session in each network namespace, run:
```bash
make terminals
```

In hosts h2 and h3, run:
```
./server.sh
```
to start the server.

In host h1, run:
```
./client.sh
```
to start the client.

To migrate a process from the source host to the target host, and update the switch accordingly, run:
```bash
make migrate SOURCE=<> TARGET=<>
``` 

To cleanup the topology, networks, pods, and containers, run:
```bash
make clean
```

To show logs of e.g. the s1 switch, run:
```bash
tail -f ../../load_balancer/logs/s1.log
```
