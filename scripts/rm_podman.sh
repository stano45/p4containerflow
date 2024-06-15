sudo podman rm -f redis1
sudo podman pod rm -f h1-pod
sudo podman network rm -f h1-net

sudo podman rm -f redis2
sudo podman pod rm -f h2-pod
sudo podman network rm -f h2-net

sudo podman rm -f redis3
sudo podman pod rm -f h3-pod
sudo podman network rm -f h3-net

sudo podman rm -f redis4
sudo podman pod rm -f h4-pod
sudo podman network rm -f h4-net