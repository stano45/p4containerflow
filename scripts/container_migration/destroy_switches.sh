#!/bin/bash
sudo podman network rm -f s1-net
sudo podman kill s1
sudo podman rm -f s1