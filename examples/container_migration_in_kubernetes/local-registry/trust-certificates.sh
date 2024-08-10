#!/bin/bash

sudo mkdir -p /etc/pki/ca-trust/source/anchors/

sudo cp ./certs/container-registry-local.crt /etc/pki/ca-trust/source/anchors/
sudo update-ca-trust
sudo trust list
