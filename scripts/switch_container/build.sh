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
printf "\n-----Creating host 1-----\n"
sudo podman network create --driver bridge --interface-name h1-br --gateway 10.1.1.10 --route 10.1.1.0/24,10.1.1.10 --subnet 10.1.1.0/24 h1-net
sudo podman pod create --name h1-pod --network h1-net --mac-address 08:00:00:00:01:01 --ip 10.1.1.1
sudo podman run --detach --privileged --name h1 --pod h1-pod --cap-add NET_ADMIN $IMG $ARGS

# Host 2
printf "\n-----Creating host 2-----\n"
sudo podman network create --driver bridge  --interface-name h2-br --gateway 10.2.2.20 --route 10.1.1.0/24,10.2.2.20 --subnet 10.2.2.0/24 h2-net
sudo podman pod create --name h2-pod --network h2-net --mac-address 08:00:00:00:02:02 --ip 10.2.2.2
sudo podman run --detach --privileged --name h2 --pod h2-pod --cap-add NET_ADMIN $IMG $ARGS

# Host 3
printf "\n-----Creating host 3-----\n"
sudo podman network create --driver bridge --interface-name h3-br --gateway 10.3.3.30 --route 10.1.1.0/24,10.3.3.30 --subnet 10.3.3.0/24 h3-net
sudo podman pod create --name h3-pod --network h3-net --mac-address 08:00:00:00:03:03 --ip 10.3.3.3
sudo podman run --detach --privileged --name h3 --pod h3-pod --cap-add NET_ADMIN $IMG $ARGS

# Host 4
printf "\n-----Creating host 4-----\n"
sudo podman network create --driver bridge  --interface-name h4-br --gateway 10.4.4.40 --route 10.1.1.0/24,10.4.4.40 --subnet 10.4.4.0/24 h4-net
sudo podman pod create --name h4-pod --network h4-net --mac-address 08:00:00:00:04:04 --ip 10.4.4.4
sudo podman run --detach --privileged --name h4 --pod h4-pod --cap-add NET_ADMIN $IMG $ARGS


# Switch
printf "\n-----Creating switch-----\n"
sudo podman run -d \
    --name s1 \
    --privileged \
    --replace \
    --publish 50051:50051 \
    --network h1-net:ip=10.1.1.11,mac=00:00:00:01:01:00,interface_name=eth0 \
    --network h2-net:ip=10.2.2.22,mac=00:00:00:01:02:00,interface_name=eth1 \
    --network h3-net:ip=10.3.3.33,mac=00:00:00:01:03:00,interface_name=eth2 \
    --network h4-net:ip=10.4.4.44,mac=00:00:00:01:04:00,interface_name=eth3 \
    -v /home/p4/p4containerflow/scripts/switch_container/s1.sh:/s1.sh \
    -v /home/p4/p4containerflow/load_balancer/build/load_balance.json:/load_balance.json \
    --entrypoint /s1.sh \
    p4c

# sudo podman exec s1 iptables -A OUTPUT -p tcp --tcp-flags RST RST -s 10.1.1.11 -o eth0 -j DROP
# sudo ip route del 10.1.1.0/24
# sudo ip route del 10.2.2.0/24

# sudo podman exec s1 ip link set eth0 promisc on
# sudo podman exec s1 ip link set eth1 promisc on
# sudo podman exec s1 ip link set eth2 promisc on
# sudo podman exec s1 ip link set eth3 promisc on