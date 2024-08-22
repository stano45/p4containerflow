#!/bin/bash
# Install the p4 compiler on Ubuntu 22.04+

# Print script commands and exit on errors.
set -xe

sudo apt-get install -y \
    cmake \
    g++ \
    git \
    automake \
    libtool \
    libgc-dev \
    bison \
    flex \
    libfl-dev \
    libboost-dev \
    libboost-iostreams-dev \
    libboost-graph-dev \
    llvm \
    pkg-config \
    python3 \
    python3-pip \
    tcpdump

git clone --depth 1 --branch v1.2.4.14 https://github.com/p4lang/p4c.git
cd p4c
git submodule update --init --recursive

pip3 install --user -r requirements.txt
mkdir -p build
cd build
cmake .. -DENABLE_TEST_TOOLS=ON
# Note: This might take minutes to hours depending on your machine.
# If you are on a distro which provides the p4lang-p4c package,
# it's recommended to install it instead of building from source.
make -j$(nproc)
sudo make install/strip
sudo ldconfig

p4c --version

cd ../..
# rm -rf p4c
