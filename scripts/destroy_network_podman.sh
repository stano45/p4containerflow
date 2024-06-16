#!/bin/bash


sudo podman container kill server1
sudo podman rm -f server1
sudo podman container kill tcp-client
sudo podman pod rm -f h1-pod
sudo podman network rm -f h1-net

sudo podman container kill server2
sudo podman rm -f server2
sudo podman pod rm -f h2-pod
sudo podman network rm -f h2-net

sudo podman container kill server3
sudo podman rm -f server3
sudo podman pod rm -f h3-pod
sudo podman network rm -f h3-net

sudo podman container kill server4
sudo podman rm -f server4
sudo podman pod rm -f h4-pod
sudo podman network rm -f h4-net


interfaces=(s1-eth1 s1-eth2 s1-eth3 s2-eth1 s3-eth1 s4-eth1)
for iface in "${interfaces[@]}"; do
    if ip link show $iface &> /dev/null; then
        sudo ip link set dev $iface down
        sudo ip link del $iface
    fi
done
