#!/bin/bash

NUM_HOSTS=3

# Container image and arguments
REDIS_IMG="docker.io/redis:latest"
REDIS_ARGS=""


# Host 1
printf "\n-----Creating host 1-----\n"
sudo podman network create --driver bridge --opt isolate=1 --disable-dns --interface-name h1-br --gateway 10.0.1.10 --subnet 10.0.1.0/24 h1-net
sudo podman pod create --name h1-pod --network h1-net --mac-address 08:00:00:00:01:01 --ip 10.0.1.1
sudo podman run --replace --detach --privileged --name redis-app --pod h1-pod --cap-add NET_ADMIN redis-app
sudo podman run --replace --detach --privileged --name redis-producer --pod h1-pod --cap-add NET_ADMIN redis-producer

# Host 2
printf "\n-----Creating host 2-----\n"
sudo podman network create --driver bridge --opt isolate=1 --disable-dns --interface-name h2-br --gateway 10.0.2.20 --route 10.0.1.0/24,10.0.2.20 --subnet 10.0.2.0/24 h2-net
sudo podman pod create --name h2-pod --network h2-net --mac-address 08:00:00:00:02:02 --ip 10.0.2.2
sudo podman run --replace --detach --privileged --name h2 --pod h2-pod --cap-add NET_ADMIN $REDIS_IMG $REDIS_ARGS

# Host 3
printf "\n-----Creating host 3-----\n"
sudo podman network create --driver bridge  --opt isolate=1 --disable-dns --interface-name h3-br --gateway 10.0.3.30 --route 10.0.1.0/24,10.0.3.30 --subnet 10.0.3.0/24 h3-net
sudo podman pod create --name h3-pod --network h3-net --mac-address 08:00:00:00:03:03 --ip 10.0.3.3
sudo podman run --replace --detach --privileged --name h3 --pod h3-pod --cap-add NET_ADMIN $REDIS_IMG $REDIS_ARGS


# Configure interfaces
printf "\n-----Configuring interfaces-----\n"
interfaces=(h1-br h2-br h3-br)
for iface in "${interfaces[@]}"; do
printf "Interface: %s\n" $iface
    # Disable IPv6 on the interfaces, so that the Linux kernel
    # will not automatically send IPv6 MDNS, Router Solicitation,
    # and Multicast Listener Report packets on the interface,
    # which can make P4 program debugging more confusing.
    sudo sysctl net.ipv6.conf.$iface.disable_ipv6=1

    # Disable tx/rx/sg offloading
    sudo ethtool -K $iface tx off
    sudo ethtool -K $iface rx off
    sudo ethtool -K $iface sg off
    
    # Set the MTU of these interfaces to be larger than default of
    # 1500 bytes, so that P4 behavioral-model testing can be done
    # on jumbo frames.
    sudo ip link set $iface mtu 9500
done

sudo podman kill h3
sudo podman rm -f h3

# Switch
printf "\n-----Creating switch-----\n"
sudo podman run -d \
    --name s1 \
    --privileged \
    --replace \
    --network host \
    -v ../../examples/switch_container/s1.sh:/s1.sh \
    -v ../../load_balancer/build/load_balance.json:/load_balance.json \
    --entrypoint /s1.sh \
    p4c