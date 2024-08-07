# Example: Redis Migration 

This example demonstrates the migration of a Redis instance across different hosts. We will use three hosts and several containers to illustrate this process.

## Setup

### Host 1
- **Containers**:
  1. **redis-app**: Exposes the value of the key `counter` via an HTTP GET endpoint.
  2. **redis-producer**: Fetches the value of the key `counter` from a Redis instance (accessed via a load balancer), increments it by 1, and updates the key `counter` in the Redis instance every second.
  3. **redis-client**: Simulates a frontend client that fetches data (the counter) from the backend (redis-app) every second.

### Host 2
- **Container**:
  1. **redis-instance**: Runs the initial Redis instance.

### Host 3
- Host 3 initially has no containers running. We will migrate the Redis instance from Host 2 to Host 3.

## Running the example

To build the topology, start containers and run the controller, run:
```bash
make
```

To start the redis-client, run:
```bash
make client
```
You should see the counter value being fetched every second.

To migrate the container from the source host to the target host, and update the switch accordingly, run:
```bash
make migrate SOURCE=2 TARGET=3
``` 
You should see no change in the counter value being fetched every second. However, the Redis instance is now running on Host 3. You can verify this by running:
```bash
sudo podman ps
```
and 
```bash
sudo podman logs -f h2 # Nothing should be printed
sudo podman logs -f h3 # You should see the Redis logs
```

To cleanup the topology, networks, pods, and containers, run:
```bash
make clean
```

To show logs of the switch, run:
```bash
sudo podman logs -f s1
```

