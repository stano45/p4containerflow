include ../../utils/Makefile
SHELL := /bin/bash

all: compile build controller

terminals: h1 h2 h3 h4

build: clean
	./build.sh

# Sleep for 2 seconds to allow the switches to start
controller:
	sleep 2 && ../../controller/controller.py --config config/switches.json

clean:
	./clean.sh

migrate:
	@if [ "${SOURCE}" = "" ] || [ "${TARGET}" = "" ]; then \
		echo "Usage: make migrate SOURCE=x TARGET=y"; \
	else \
		./cr.sh ${SOURCE} ${TARGET}; \
	fi

h1:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h1" -hold -e "sudo ip netns exec h1 /bin/bash" &

h2:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h2" -hold -e "sudo ip netns exec h2 /bin/bash" &

h3:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h3" -hold -e "sudo ip netns exec h3 /bin/bash" &

h4:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h4" -hold -e "sudo ip netns exec h4 /bin/bash" &

.PHONY: all terminals build controller clean migrate h1 h2 h3 h4
