#!/bin/bash

sudo criu dump -t $(pgrep server) --images-dir /home/p4/images -v4 -o dump.log --shell-job --tcp-established && echo "OK" || echo "Dump failed"
