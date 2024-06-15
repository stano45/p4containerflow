#!/bin/bash

# Switch 1 port 2 <--> Switch 2 port 2
sudo ip link add s1-eth2 type veth peer name s2-eth2
sudo ip link set dev s1-eth2 address 00:00:00:01:02:00
sudo ip link set dev s2-eth2 address 00:00:00:02:02:00
# Switch 1 port 3 <--> Switch 3 port 2
sudo ip link add s1-eth3 type veth peer name s3-eth2
sudo ip link set dev s1-eth3 address 00:00:00:01:03:00
sudo ip link set dev s3-eth2 address 00:00:00:03:02:00


sudo podman network create --interface-name s1-eth1 --subnet 10.1.1.0/24 --gateway 10.1.1.10  h1-net
sudo podman pod create --name h1-pod --network h1-net:ip=10.1.1.1 --mac-address 08:00:00:00:01:01
# sudo podman run -d --name iperf-server1 --pod h1-pod docker.io/networkstatic/iperf3 -s
sudo podman run -d --name iperf-server1 --pod h1-pod docker.io/subfuzion/netcat -vl 12345
sudo ip link set dev s1-eth1 address 00:00:00:01:01:00
# sudo arp -i veth0 -s 10.1.1.10 00:00:00:01:01:00

sudo podman network create --interface-name s2-eth1 --subnet 10.2.2.0/24 --gateway 10.2.2.20  h2-net
sudo podman pod create --name h2-pod --network h2-net:ip=10.2.2.2 --mac-address 08:00:00:00:02:02
#sudo podman run -d --name iperf-server2 --pod h2-pod docker.io/networkstatic/iperf3 -s
sudo podman run -d --name iperf-server2 --pod h2-pod docker.io/subfuzion/netcat -vl 12345
sudo ip link set dev s2-eth1 address 00:00:00:02:01:00
# sudo arp -i veth1 -s 10.2.2.20 00:00:00:02:02:00

sudo podman network create --interface-name s3-eth1 --subnet 10.3.3.0/24 --gateway 10.3.3.30  h3-net
sudo podman pod create --name h3-pod --network h3-net:ip=10.3.3.3 --mac-address 08:00:00:00:03:03
# sudo podman run -d --name iperf-server3 --pod h3-pod docker.io/networkstatic/iperf3 -s
sudo podman run -d --name iperf-server3 --pod h3-pod docker.io/subfuzion/netcat -vl 12345
sudo ip link set dev s3-eth1 address 00:00:00:03:01:00
# sudo arp -i veth2 -s 10.3.3.30 00:00:00:03:03:00


sudo podman network create --interface-name s4-eth1 --subnet 10.4.4.0/24 --gateway 10.4.4.40  h4-net
sudo podman pod create --name h4-pod --network h4-net:ip=10.4.4.4 --mac-address 08:00:00:00:04:04
# sudo podman run -d --name iperf-server4 --pod h4-pod docker.io/networkstatic/iperf3 -s
sudo podman run -d --name iperf-server4 --pod h4-pod docker.io/subfuzion/netcat -vl 12345
sudo ip link set dev s4-eth1 address 00:00:00:04:01:00
# sudo arp -i veth3 -s 10.4.4.40 00:00:00:04:04:00


interfaces=(s1-eth1 s1-eth2 s1-eth3 s2-eth1 s2-eth2 s3-eth1 s3-eth2)
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


for iface in "${interfaces[@]}"; do
    sudo ip link set dev $iface up
done
