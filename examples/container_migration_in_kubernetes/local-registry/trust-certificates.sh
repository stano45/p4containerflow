#!/bin/bash

if [ -f /etc/os-release ]; then
	. /etc/os-release
else
	echo "Cannot find /etc/os-release"
	exit 1
fi

case "$ID" in
	fedora|rhel)
		sudo mkdir -p /etc/pki/ca-trust/source/anchors/
		sudo cp ./certs/container-registry-local.crt /etc/pki/ca-trust/source/anchors/
		sudo update-ca-trust
		;;

	ubuntu)
		sudo mkdir -p /usr/local/share/ca-certificates
		sudo cp ./certs/container-registry-local.crt /usr/local/share/ca-certificates/
		sudo update-ca-certificates
		;;

	*)
		echo "Unsupported or unrecognized distribution: $ID"
		exit 1
		;;
esac

exit 0
