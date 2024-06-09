#!/bin/bash

curl -X POST http://127.0.0.1:5000/update_node \
    -H "Content-Type: application/json" \
    -d "{\"old_ipv4\":\"10.0.3.3\", \"new_ipv4\":\"10.0.4.4\", \"dmac\":\"08:00:00:00:04:04\", \"eport\":\"4\"}"