#!/bin/bash

read -p "Enter old IPv4 address: " old_ipv4
read -p "Enter new IPv4 address: " new_ipv4
read -p "Enter destination MAC address (dmac): " dmac
read -p "Enter eport: " eport

curl -X POST http://127.0.0.1:5000/update_node \
    -H "Content-Type: application/json" \
    -d "{\"old_ipv4\":\"$old_ipv4\", \"new_ipv4\":\"$new_ipv4\", \"dmac\":\"$dmac\", \"eport\":\"$eport\"}"
