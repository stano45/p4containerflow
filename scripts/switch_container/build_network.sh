#!/bin/bash

NUM_HOSTS=4

# Container image and arguments
IMG="tcp-server"
ARGS=""

# IMG="docker.io/networkstatic/iperf3"
# ARGS="-s -p 12345"

# IMG="docker.io/subfuzion/netcat"
# ARGS="-vl 12345"

# Host 1
sudo podman network create --driver bridge --interface-name h1-br --gateway 10.1.1.10 --route 10.1.1.0/24,10.1.1.10 --subnet 10.1.1.0/24 h1-net
sudo podman pod create --name h1-pod --network h1-net --mac-address 08:00:00:00:01:01 --ip 10.1.1.1
sudo podman run --detach --privileged --name h1 --pod h1-pod --cap-add NET_ADMIN $IMG $ARGS

# Host 2
sudo podman network create --driver bridge  --interface-name h2-br --gateway 10.2.2.20 --route 10.1.1.0/24,10.2.2.20 --subnet 10.2.2.0/24 h2-net
sudo podman pod create --name h2-pod --network h2-net --mac-address 08:00:00:00:02:02 --ip 10.2.2.2
sudo podman run --detach --privileged --name h2 --pod h2-pod --cap-add NET_ADMIN $IMG $ARGS

# Host 3
sudo podman network create --driver bridge --interface-name h3-br --gateway 10.3.3.30 --route 10.1.1.0/24,10.3.3.30 --subnet 10.3.3.0/24 h3-net
sudo podman pod create --name h3-pod --network h3-net --mac-address 08:00:00:00:03:03 --ip 10.3.3.3
sudo podman run --detach --privileged --name h3 --pod h3-pod --cap-add NET_ADMIN $IMG $ARGS

# Host 4
sudo podman network create --driver bridge  --interface-name h4-br --gateway 10.4.4.40 --route 10.1.1.0/24,10.4.4.40 --subnet 10.4.4.0/24 h4-net
sudo podman pod create --name h4-pod --network h4-net --mac-address 08:00:00:00:04:04 --ip 10.4.4.4
sudo podman run --detach --privileged --name h4 --pod h4-pod --cap-add NET_ADMIN $IMG $ARGS