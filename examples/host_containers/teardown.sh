#!/bin/bash
NUM_HOSTS=4


for i in $(seq 1 $NUM_HOSTS); do
    sudo podman container kill h${i}
    sudo podman rm -f h${i}
    sudo podman pod rm -f h${i}-pod
    sudo podman network rm -f h${i}-net
done

interfaces=(s1-eth1 s1-eth2 s1-eth3 s2-eth1 s2-eth2 s3-eth1 s3-eth2 s4-eth1 s4-eth2 h1-veth h2-veth h3-veth h4-veth h1-br h2-br h3-br h4-br)
for iface in "${interfaces[@]}"; do
    if ip link show $iface &> /dev/null; then
        sudo ip link set dev $iface down
        sudo ip link del $iface
    fi
done

