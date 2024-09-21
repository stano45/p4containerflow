all: build install

build:
	GOOS=linux GOARCH=amd64 go build -o bmv2-cni .

install:
	sudo chmod +x bmv2-cni
	sudo cp bmv2-cni /opt/cni/bin/
	sudo cp 10-bmv2-cni.conf /etc/cni/net.d/

PHONY: all build install