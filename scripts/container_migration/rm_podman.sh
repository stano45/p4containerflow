
sudo ip link del s1-eth1
sudo ip link del s1-eth0
sudo ip link del veth-host

sudo podman container kill server1
sudo podman rm -f server1
sudo podman pod rm -f h1-pod
sudo podman network rm -f h1-net
