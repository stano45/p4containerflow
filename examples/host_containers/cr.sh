#!/bin/bash

# Check if exactly two arguments are provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <source_idx> <target_idx>"
    exit 1
fi

set -xe

# Assigning arguments to variables
SOURCE_IDX=$1
TARGET_IDX=$2

CHECKPOINT_DIR=/tmp/checkpoints
CHECKPOINT_PATH=$CHECKPOINT_DIR/checkpoint.tar

SOURCE_HOST=h$SOURCE_IDX
SOURCE_IP=10.0.$SOURCE_IDX.$SOURCE_IDX

TARGET_HOST=h$TARGET_IDX
TARGET_IP=10.0.$TARGET_IDX.$TARGET_IDX
TARGET_MAC=08:00:00:00:0$TARGET_IDX:0$TARGET_IDX

sudo mkdir -p $CHECKPOINT_DIR

sudo podman container checkpoint --export $CHECKPOINT_PATH --compress none --keep --tcp-established $SOURCE_HOST
sudo podman rm -f $SOURCE_HOST

sudo ../../scripts/edit_files_img.py $CHECKPOINT_PATH $SOURCE_IP $TARGET_IP

sudo podman container restore --import $CHECKPOINT_PATH --keep --tcp-established --ignore-static-ip --ignore-static-mac --pod ${TARGET_HOST}-pod
# --name cannot be used with --tcp-established on restore
sudo podman rename $SOURCE_HOST $TARGET_HOST

curl -X POST http://127.0.0.1:5000/update_node \
    -H "Content-Type: application/json" \
    -d "{\"old_ipv4\":\"${SOURCE_IP}\", \"new_ipv4\":\"${TARGET_IP}\", \"dmac\":\"${TARGET_MAC}\", \"eport\":\"${TARGET_IDX}\"}"