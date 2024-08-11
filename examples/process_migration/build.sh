#!/bin/bash

set -xe

BUILD_DIR=../../load_balancer/build
PCAP_DIR=../../load_balancer/pcaps
LOG_DIR=../../load_balancer/logs

# Switch 1 port 1 <--> Host 1
sudo ip link add s1-eth1 type veth peer name h1-eth1
# Switch 1 port 2 <--> Switch 2 port 2
sudo ip link add s1-eth2 type veth peer name s2-eth2
# Switch 1 port 3 <--> Switch 3 port 2
sudo ip link add s1-eth3 type veth peer name s3-eth2
# Switch 2 port 1 <--> Host 2
sudo ip link add s2-eth1 type veth peer name h2-eth1
# Switch 3 port 1 <--> Host 3
sudo ip link add s3-eth1 type veth peer name h3-eth1
# Switch 4 port 1 <--> Host 4
sudo ip link add s4-eth1 type veth peer name h4-eth1

interfaces=(s1-eth1 s1-eth2 s1-eth3 s2-eth1 s2-eth2 s3-eth1 s3-eth2 h1-eth1 h2-eth1 h3-eth1 h4-eth1)
for iface in "${interfaces[@]}"; do
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

# Set MAC addresses
sudo ip link set dev s1-eth1 address 00:00:00:01:01:00
sudo ip addr add 10.0.1.10/24 dev s1-eth1
sudo ip link set dev s1-eth2 address 00:00:00:01:02:00
sudo ip link set dev s1-eth3 address 00:00:00:01:03:00
sudo ip link set dev s2-eth1 address 00:00:00:02:01:00
sudo ip addr add 10.0.2.20/24 dev s2-eth1
sudo ip link set dev s2-eth2 address 00:00:00:02:02:00
sudo ip link set dev s3-eth1 address 00:00:00:03:01:00
sudo ip addr add 10.0.3.30/24 dev s3-eth1
sudo ip link set dev s3-eth2 address 00:00:00:03:02:00
sudo ip link set dev s4-eth1 address 00:00:00:04:01:00
sudo ip addr add 10.0.4.40/24 dev s4-eth1

# Assign IP addresses to host interfaces (h1, h2, h3), and bring them up, add default gateway
sudo ip netns add h1
sudo ip link set h1-eth1 netns h1
sudo ip netns exec h1 ip link set lo up
sudo ip netns exec h1 ip link set dev h1-eth1 address 08:00:00:00:01:01
sudo ip netns exec h1 ip addr add 10.0.1.1/24 dev h1-eth1
sudo ip netns exec h1 ip link set dev h1-eth1 up
sudo ip netns exec h1 route add default gw 10.0.1.10 dev h1-eth1
sudo ip netns exec h1 arp -i h1-eth1 -s 10.0.1.10 00:00:00:01:01:00

sudo ip netns add h2
sudo ip link set h2-eth1 netns h2
sudo ip netns exec h2 ip link set lo up
sudo ip netns exec h2 ip link set dev h2-eth1 address 08:00:00:00:02:02
sudo ip netns exec h2 ip addr add 10.0.2.2/24 dev h2-eth1
sudo ip netns exec h2 ip link set dev h2-eth1 up
sudo ip netns exec h2 route add default gw 10.0.2.20 dev h2-eth1
sudo ip netns exec h2 arp -i h2-eth1 -s 10.0.2.20 00:00:00:02:01:00

sudo ip netns add h3
sudo ip link set h3-eth1 netns h3
sudo ip netns exec h3 ip link set lo up
sudo ip netns exec h3 ip link set dev h3-eth1 address 08:00:00:00:03:03
sudo ip netns exec h3 ip addr add 10.0.3.3/24 dev h3-eth1
sudo ip netns exec h3 ip link set dev h3-eth1 up
sudo ip netns exec h3 route add default gw 10.0.3.30 dev h3-eth1
sudo ip netns exec h3 arp -i h3-eth1 -s 10.0.3.30 00:00:00:03:01:00

sudo ip netns add h4
sudo ip link set h4-eth1 netns h4
sudo ip netns exec h4 ip link set lo up
sudo ip netns exec h4 ip link set dev h4-eth1 address 08:00:00:00:04:04
sudo ip netns exec h4 ip addr add 10.0.4.4/24 dev h4-eth1
sudo ip netns exec h4 ip link set dev h4-eth1 up
sudo ip netns exec h4 route add default gw 10.0.4.40 dev h4-eth1
sudo ip netns exec h4 arp -i h4-eth1 -s 10.0.4.40 00:00:00:04:01:00

# Bring up interfaces (except host interfaces, which were already brought up)
for iface in "${interfaces[@]:0:7}"; do
    sudo ip link set dev $iface up
done

# Run switches
sudo simple_switch_grpc \
    -i 1@s1-eth1 \
    -i 2@s1-eth2 \
    -i 3@s1-eth3 \
    -i 4@s1-eth4 \
    --pcap ${PCAP_DIR} \
    --device-id 0 \
    ${BUILD_DIR}/load_balance.json \
    --log-console \
    --thrift-port 9090 \
    -- \
    --grpc-server-addr 0.0.0.0:50051 > ${LOG_DIR}/s1.log \
    > ${LOG_DIR}/s1.log \
    2>&1 & \
    echo $!

sudo simple_switch_grpc \
    -i 1@s2-eth1 \
    -i 2@s2-eth2 \
    --pcap ${PCAP_DIR} \
    --device-id 1 \
    ${BUILD_DIR}/load_balance.json \
    --log-console \
    --thrift-port 9091 \
    -- \
    --grpc-server-addr 0.0.0.0:50052 > ${LOG_DIR}/s2.log \
    > ${LOG_DIR}/s2.log \
    2>&1 & \
    echo $!

sudo simple_switch_grpc \
    -i 1@s3-eth1 \
    -i 2@s3-eth2 \
    --pcap ${PCAP_DIR} \
    --device-id 2 \
    ${BUILD_DIR}/load_balance.json \
    --log-console \
    --thrift-port 9092 \
    -- \
    --grpc-server-addr 0.0.0.0:50053 \
    > ${LOG_DIR}/s3.log \
    2>&1 & \
    echo $!

sudo simple_switch_grpc \
    -i 1@s4-eth1 \
    -i 2@s4-eth2 \
    --pcap ${PCAP_DIR} \
    --device-id 3 \
    ${BUILD_DIR}/load_balance.json \
    --log-console \
    --thrift-port 9093 \
    -- \
    --grpc-server-addr 0.0.0.0:50054 \
    > ${LOG_DIR}/s4.log \
    2>&1 & \
    echo $!