include ../../Makefile

all: compile build control_multi_switch

terminals: h1 h2 h3 h4

build: teardown
	./build.sh

teardown:
	./teardown.sh

migrate:
	@if [ "${SOURCE}" = "" ] || [ "${TARGET}" = "" ]; then \
		echo "Usage: make migrate SOURCE=x TARGET=y"; \
	else \
		./cr.sh ${SOURCE} ${TARGET}; \
	fi

h1:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h1" -hold -e "sudo podman rm -f tcp-client && sudo podman run --name tcp-client --pod h1-pod tcp-client" &

h2:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h2" -hold -e "sudo podman logs -f h2" &

h3:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h3" -hold -e "sudo podman logs -f h3" &

h4:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h4" -hold -e "sudo podman logs -f h4" &

.PHONY: all terminals build teardown migrate h1 h2 h3 h4
