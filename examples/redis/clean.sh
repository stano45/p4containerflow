#!/bin/bash
NUM_HOSTS=3

sudo podman kill redis-app
sudo podman rm -f redis-app
sudo podman kill redis-producer
sudo podman rm -f redis-producer
sudo podman kill redis-client
sudo podman rm -f redis-client

for i in $(seq 1 $NUM_HOSTS); do
    printf "\n-----Removing host ${i}-----\n"
    sudo podman kill h${i}
    sudo podman rm -f h${i}
    sudo podman pod rm -f h${i}-pod
    sudo podman network rm -f h${i}-net
done


printf "\n-----Removing switch-----\n"
sudo podman kill s1
sudo podman rm -f s1
