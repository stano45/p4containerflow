sudo podman run -d \
    --name s1 \
    --privileged \
    --replace \
    --publish 50051:50051 \
    --network h1-net:ip=10.1.1.11,mac=00:00:00:01:01:00,interface_name=eth0 \
    --network h2-net:ip=10.2.2.22,mac=00:00:00:01:02:00,interface_name=eth1 \
    --network h3-net:ip=10.3.3.33,mac=00:00:00:01:03:00,interface_name=eth2 \
    --network h4-net:ip=10.4.4.44,mac=00:00:00:01:04:00,interface_name=eth3 \
    -v /home/p4/p4containerflow/scripts/container_migration/s1.sh:/s1.sh \
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