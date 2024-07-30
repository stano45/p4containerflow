#!/bin/bash

# Check if exactly two arguments are provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <source_idx> <target_idx>"
    exit 1
fi

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

# Creating checkpoint directory
sudo mkdir -p $CHECKPOINT_DIR

# Dump the process
sudo criu dump -t $(pgrep server) --images-dir $CHECKPOINT_PATH -v4 -o ${CHECKPOINT_PATH}/dump.log --shell-job --tcp-established && echo "OK" || echo "Dump failed"

# Edit the checkpoint files with new IP
sudo /home/p4/p4containerflow/scripts/edit_files_img.py $CHECKPOINT_PATH $SOURCE_IP $TARGET_IP

# Restore the process
sudo criu restore -D $CHECKPOINT_PATH -vvv --shell-job --tcp-established -d -o ${CHECKPOINT_PATH}/restore.log

# Update the node information
curl -X POST http://127.0.0.1:5000/update_node \
    -H "Content-Type: application/json" \
    -d "{\"old_ipv4\":\"${SOURCE_IP}\", \"new_ipv4\":\"${TARGET_IP}\", \"dmac\":\"${TARGET_MAC}\", \"eport\":\"${TARGET_IDX}\"}"