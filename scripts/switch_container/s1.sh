#!/bin/sh
mkdir /pcaps
simple_switch_grpc \
    -i 1@eth0 \
    -i 2@eth1 \
    -i 3@eth2 \
    -i 4@eth3 \
    --pcap /pcaps \
    --device-id 0 \
    /load_balance.json \
    --log-console \
    -- \
    --grpc-server-addr 0.0.0.0:50051