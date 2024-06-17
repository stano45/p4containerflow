sudo podman network create --interface-name s1-eth1 --subnet 10.1.1.0/24 --gateway 10.1.1.10 h1-net
sudo podman pod create --name h1-pod --network h1-net:ip=10.1.1.1 --mac-address 08:00:00:00:01:01
sudo podman run -d --pod h1-pod --name redis1 redis

sudo podman network create --interface-name s2-eth1 --subnet 10.2.2.0/24 --gateway 10.2.2.20  h2-net
sudo podman pod create --name h2-pod --network h2-net:ip=10.2.2.2 --mac-address 08:00:00:00:02:02
sudo podman run -d --pod h2-pod --name redis2 redis

sudo podman network create --interface-name s3-eth1 --subnet 10.3.3.0/24 --gateway 10.3.3.30  h3-net
sudo podman pod create --name h3-pod --network h3-net:ip=10.3.3.3 --mac-address 08:00:00:00:03:03
sudo podman run -d --pod h3-pod --name redis3 redis

sudo podman network create --interface-name s4-eth1 --subnet 10.4.4.0/24 --gateway 10.4.4.40  h4-net
sudo podman pod create --name h4-pod --network h4-net:ip=10.4.4.4 --mac-address 08:00:00:00:04:04
sudo podman run -d --pod h4-pod --name redis4 redis
