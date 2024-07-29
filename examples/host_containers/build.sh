#!/bin/bash

NUM_HOSTS=4

# Container image and arguments
IMG="tcp-server"
ARGS=""

# IMG="docker.io/networkstatic/iperf3"
# ARGS="-s -p 12345"

# IMG="docker.io/subfuzion/netcat"
# ARGS="-vl 12345"

for i in $(seq 1 $NUM_HOSTS); do
    NET="h${i}-net"
    POD="h${i}-pod"
    CONTAINER="h${i}"
    SUBNET="10.${i}.${i}.0/24"
    GATEWAY="10.${i}.${i}.${i}0"
    CONTAINER_IP="10.${i}.${i}.${i}"
    CONTAINER_MAC="08:00:00:00:0${i}:0${i}"
    BRIDGE="h${i}-br"
    VETH="h${i}-veth"
    SW_PORT_IFACE="s${i}-eth1"

    # Podman networks (netavark BE)
    # Creates bridges h1-br, h2-br, h3-br, h4-br
    sudo podman network create --driver bridge --interface-name $BRIDGE --subnet $SUBNET $NET

    # Creates pods h1-pod, h2-pod, h3-pod, h4-pod
    # Each pod is connected to a network and has a static IP and MAC address
    sudo podman pod create --name $POD --network $NET --mac-address $CONTAINER_MAC --ip $CONTAINER_IP

    # Creates containers h1, h2, h3, h4
    sudo podman run --detach --privileged --name $CONTAINER --pod $POD --cap-add NET_ADMIN $IMG $ARGS

    # MAC of switch port
    sudo podman exec $CONTAINER arp -s $GATEWAY 00:00:00:0${i}:01:00

    # Set the default gateway of the container
    sudo podman exec $CONTAINER ip route add default via $GATEWAY

    # Add veth pair from bridge to switch port (host always on port 1)
    sudo ip link add $VETH type veth peer name $SW_PORT_IFACE

    # Set the master of the veth interfaces to the corresponding bridge
    sudo ip link set $VETH master $BRIDGE
done

# Connect load balancer (s1) to other switches
for i in $(seq 2 $NUM_HOSTS); do
    sudo ip link add s1-eth${i} type veth peer name s${i}-eth2
done

# All interfaces
interfaces=()

# Assign MAC addresses to load balancer ports (s1)
for i in $(seq 1 $NUM_HOSTS); do
    iface="s1-eth${i}"
    sudo ip link set dev $iface address 00:00:00:01:0${i}:00
    interfaces+=($iface)
done

# Assign MAC addresses to (non load-balancer) switch ports (s2, s3...)
for i in $(seq 2 $NUM_HOSTS); do
    iface1="s${i}-eth1"
    iface2="s${i}-eth2"
    sudo ip link set dev $iface1 address 00:00:00:0${i}:01:00
    sudo ip link set dev $iface2 address 00:00:00:0${i}:02:00
    interfaces+=($iface1 $iface2)
done

# Add host-side veth interfaces to the list
for i in $(seq 1 $NUM_HOSTS); do
    interfaces+=("h${i}-veth")
done


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

    # Bring the interfaces up
    sudo ip link set dev $iface up
done


# Enable IP forwarding
sudo sysctl -w net.ipv4.ip_forward=1

sudo podman kill h4
sudo podman rm -f h4
