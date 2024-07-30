#!/bin/bash
NUM_HOSTS=4


for i in $(seq 1 $NUM_HOSTS); do
    printf "\n-----Removing host ${i}-----\n"
    sudo podman container kill h${i}
    sudo podman rm -f h${i}
    sudo podman pod rm -f h${i}-pod
    sudo podman network rm -f h${i}-net
done

printf "\n-----Removing switch-----\n"
sudo podman kill s1
sudo podman rm -f s1
