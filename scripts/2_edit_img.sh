#!/bin/bash

sudo chown -R p4:p4 /home/p4/images

python3 /home/p4/p4containerflow/tcp/edit_files_img.py /home/p4/images/files.img "10.0.3.3" "10.0.4.4"