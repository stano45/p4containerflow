#!/bin/bash
# Install podman on Ubuntu 22.04+

# Print script commands and exit on errors.
set -xe

sudo apt-get install -y \
    btrfs-progs \
    crun \
    git \
    go-md2man \
    iptables \
    libassuan-dev \
    libbtrfs-dev \
    libc6-dev \
    libdevmapper-dev \
    libglib2.0-dev \
    libgpgme-dev \
    libgpg-error-dev \
    libprotobuf-dev \
    libprotobuf-c-dev \
    libseccomp-dev \
    libselinux1-dev \
    libsystemd-dev \
    make \
    pkg-config \
    uidmap \
    conmon \
    runc

# Create default configuration files
sudo mkdir -p /etc/containers
sudo curl -L -o /etc/containers/registries.conf https://raw.githubusercontent.com/containers/image/main/registries.conf
sudo curl -L -o /etc/containers/policy.json https://raw.githubusercontent.com/containers/image/main/default-policy.json

sudo mkdir -p /usr/share/containers
# runtime = "runc", infra_image="k8s.gcr.io/pause:3.8", infra_command = "/pause", otherwise default
sudo cp "$(dirname "$(realpath "$0")")/containers.conf" /usr/share/containers/containers.conf

# Requires go v1.21.0 (strict)
git clone --depth 1 --branch v5.2.1 https://github.com/containers/podman/
cd podman
make -j$(nproc) BUILDTAGS="selinux seccomp" PREFIX=/usr
sudo GOPATH=/usr/bin/go PATH=$GOPATH/bin:$PATH make install PREFIX=/usr

podman --version

cd ..
# rm -rf podman
