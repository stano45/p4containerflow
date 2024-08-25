# Scripts

This directory contains scripts to install CRIU, the P4 compiler, PI and Podman. The scripts have been tested on Ubuntu 22.04 and 24.04. If you encounter any issues with other Linux distributions, please refer to the official documentation of the respective projects.

The `edit_files_img.py` script is used to edit the `files.img` file in a container checkpoint by rewriting the socket IP address. This is used by some of the examples to successfully restore containers with a changing IP address.