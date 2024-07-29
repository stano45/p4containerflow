SHELL := /bin/bash
MIGRATION_SCRIPT_PATH := ./scripts/switch_container/cr.sh

all: compile net control

net: 
	cd scripts/switch_container && make

compile:
	cd load_balancer && make

control:
	cd controller && sleep 2 && make

migrate:
	@if [ "${SOURCE}" = "" ] || [ "${TARGET}" = "" ]; then \
		echo "Usage: make migrate SOURCE=x TARGET=y"; \
	else \
		$(MIGRATION_SCRIPT_PATH) ${SOURCE} ${TARGET}; \
	fi

clean:
	cd scripts/switch_container && make clean

build-images:
	cd tcp && make

tcp-client:
	sudo podman run -it --rm --replace --name tcp-client --pod h1-pod tcp-client

netcat-client:
	sudo podman run -it --rm --replace --name netcat-client --pod h1-pod docker.io/gophernet/netcat -v 10.1.1.10 12345

iperf-client:
	sudo podman run -it --rm --replace --name iperf-client --pod h1-pod docker.io/networkstatic/iperf3 -c 10.1.1.10 -p 12345 -t 30