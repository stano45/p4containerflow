#!/bin/bash
# Install CRIU (Checkpoint/Restore In Userspace) on Ubuntu 22.04+

# Print script commands and exit on errors.
set -xe

# Install dependencies
sudo apt-get install -y \
    build-essential \
    libprotobuf-dev \
    libprotobuf-c-dev \
    protobuf-c-compiler \
    protobuf-compiler \
    python3-protobuf \
    libbsd-dev \
    pkg-config \
    libbsd-dev \
    iproute2 \
    libnftables-dev \
    libcap-dev \
    libnl-3-dev \
    libnl-3-200 \
    libnet1-dev \
    libnet1 \
    libnl-3-dev \
    libnet-dev \
    libaio-dev \
    libgnutls28-dev \
    python3-future \
    libdrm-dev \
    asciidoc \
    xmlto

git clone --depth 1 --branch v3.19 https://github.com/checkpoint-restore/criu.git
cd criu

# NOTE: On Ubuntu 24+, you might experience an error similar to:
# error: ‘net/unix/’ directive output truncated writing 9 bytes into a region of size 0 [-Werror=format-truncation=]
# The following issue suggests a fix:
# https://github.com/checkpoint-restore/criu/issues/2398#issuecomment-2079198028
make -j$(nproc)
# Install criu globally
sudo make install-criu
# Install crit in a virtual environment,
# make sure to activate it before running the script.
# make BINDIR=... crit never worked for me,
# see https://github.com/checkpoint-restore/criu/blob/criu-dev/INSTALL.md
pip3 install ./lib
pip3 install ./crit

criu --version
crit --version

cd ..
# rm -rf criu
