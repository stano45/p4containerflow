sudo podman network create -d bridge -o isolate=true --interface-name h1-bridge --subnet 10.1.1.0/24 --gateway 10.1.1.10 h1-net
sudo podman pod create --name h1-pod --network h1-net:ip=10.1.1.1 --mac-address 08:00:00:00:01:01
sudo podman run -d --name server1 --pod h1-pod tcp-server

sudo ip link add veth-host type veth peer name s1-eth1
sudo ip link set veth-host master h1-bridge
sudo ip link set veth-host up
sudo ip link set s1-eth1 up
