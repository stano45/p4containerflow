#!/bin/bash

interfaces=(s1-eth1 s1-eth2 s1-eth3 s2-eth1 s3-eth1 s4-eth1)
for iface in "${interfaces[@]}"; do
    if ip link show $iface &> /dev/null; then
        sudo ip link set dev $iface down
        sudo ip link del $iface
    fi
done


namespaces=(h1 h2 h3 h4)
for ns in "${namespaces[@]}"; do
    if ip netns list | grep -qw "$ns"; then
        sudo ip netns exec "$ns" ip link set lo down
        sudo ip netns del "$ns"
    fi
done
