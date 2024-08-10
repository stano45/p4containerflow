#!/bin/sh
mkdir /pcaps
simple_switch_grpc \
    -i 1@h1-br \
    -i 2@h2-br \
    -i 3@h3-br \
    -i 4@h4-br \
    --pcap /pcaps \
    --device-id 0 \
    /load_balance.json \
    --log-console \
    -- \
    --grpc-server-addr 0.0.0.0:50051