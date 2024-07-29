SHELL := /bin/bash
EXAMPLES_PATH := ./examples

PROCESS_MIGRATION_PATH := $(EXAMPLES_PATH)/process_migration
CONTAINER_MIGRATION_PATH := $(EXAMPLES_PATH)/host_containers
SWITCH_CONTAINER_PATH := $(EXAMPLES_PATH)/switch_container

# Select the example to run
SELECTED_EXAMPLE := $(PROCESS_MIGRATION_PATH)

all: compile net control

compile:
	cd load_balancer && make

net: clean
	cd $(SELECTED_EXAMPLE) && ./build.sh

control:
	cd controller && sleep 2 && make

migrate:
	@if [ "${SOURCE}" = "" ] || [ "${TARGET}" = "" ]; then \
		echo "Usage: make migrate SOURCE=x TARGET=y"; \
	else \
		cd $(SELECTED_EXAMPLE) && ./cr.sh ${SOURCE} ${TARGET}; \
	fi

clean:
	cd $(SELECTED_EXAMPLE) && ./teardown.sh

images:
	cd tcp && make

tcp-client:
	sudo podman run -it --rm --replace --name tcp-client --pod h1-pod tcp-client

netcat-client:
	sudo podman run -it --rm --replace --name netcat-client --pod h1-pod docker.io/subfuzion/netcat -4 -v 10.1.1.10 12345

iperf-client:
	sudo podman run -it --rm --replace --name iperf-client --pod h1-pod docker.io/networkstatic/iperf3 -4 -c 10.1.1.10 -p 12345 -t 30

.PHONY: all net compile control clean build-images tcp-client netcat-client iperf-client
