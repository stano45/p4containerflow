#!/bin/bash
# Install p4lang/PI on Ubuntu 22.04+

# Print script commands and exit on errors.
set -xe

# If you want to use a specific venv,
# change this variable to its absolute path
PYTHON_VENV="$VIRTUAL_ENV"
source "${PYTHON_VENV}/bin/activate"

sudo apt-get install -y \
    libreadline-dev \
    valgrind \
    libtool-bin \
    libboost-dev \
    libboost-system-dev \
    libboost-thread-dev \
    libprotobuf-dev \
    protobuf-compiler \
    protobuf-compiler-grpc \
    libgrpc-dev \
    libgrpc++-dev

pip3 install protobuf==3.20.0

git clone https://github.com/p4lang/PI.git
cd PI
git checkout 05cb92564af77ae4826565cbde84e3fd4960c6bd
git submodule update --init --recursive

./autogen.sh
configure_python_prefix="--with-python_prefix=${PYTHON_VENV}"
./configure --with-proto --without-internal-rpc --without-cli --without-bmv2 ${configure_python_prefix}
make -j$(nproc)
sudo make install
sudo ldconfig
sudo chown -R $USER $PYTHON_VENV

cd ..
# rm -rf PI
