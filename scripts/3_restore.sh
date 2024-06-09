#!/bin/bash

sudo criu restore -D /home/p4/images -vvv --shell-job --tcp-established -d -o restore.log
