include ../../utils/Makefile

all: build-app build-producer build-client compile build controller

build: clean
	./build.sh

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

build-app:
	sudo podman build -t redis-app -f app/Containerfile app

build-producer:
	sudo podman build -t redis-producer -f producer/Containerfile producer

build-client:
	sudo podman build -t redis-client -f client/Containerfile client

client:
	sudo podman run --name redis-client --pod h1-pod --rm --replace -it redis-client	

.PHONY: all terminals build controller clean migrate h1 h2 h3 h4 build-app build-producer build-client client
