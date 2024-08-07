include ../../utils/Makefile

all: compile build control

terminals: h1 h2 h3 h4

build: clean
	./build.sh

clean:
	./clean.sh

migrate:
	@if [ "${SOURCE}" = "" ] || [ "${TARGET}" = "" ]; then \
		echo "Usage: make migrate SOURCE=x TARGET=y"; \
	else \
		./cr.sh ${SOURCE} ${TARGET}; \
	fi

redis-app:
	sudo podman build -t redis-app -f app/Containerfile app

producer:
	sudo podman build -t redis-producer -f producer/Containerfile producer

h1:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h1" -hold -e "sudo podman rm -f tcp-client && sudo podman run --name tcp-client --pod h1-pod tcp-client" &

h2:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h2" -hold -e "sudo podman logs -f h2" &

h3:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h3" -hold -e "sudo podman logs -f h3" &

h4:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h4" -hold -e "sudo podman logs -f h4" &

.PHONY: all terminals build clean migrate h1 h2 h3 h4 redis-app producer
