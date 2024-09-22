#!/bin/bash
# Install the p4 compiler on Ubuntu 22.04+

git clone https://github.com/p4lang/behavioral-model.git
sudo apt-get install -y automake cmake libgmp-dev \
    libpcap-dev libboost-dev libboost-test-dev libboost-program-options-dev \
    libboost-system-dev libboost-filesystem-dev libboost-thread-dev \
    libevent-dev libtool flex bison pkg-config g++ libssl-dev

cd behavioral-model
cd ci
chmod +x install-nanomsg.sh
chmod +x install-thrift.sh
./install-nanomsg.sh
./install-thrift.sh
cd ..
./autogen.sh
./configure
make -j$(nproc)
sudo make install  # if you need to install bmv2
cd ..
# rm -rf behavioral-model