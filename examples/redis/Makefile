include ../../utils/Makefile

all: app producer compile build controller

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

app:
	sudo podman build -t redis-app -f app/Containerfile app

producer:
	sudo podman build -t redis-producer -f producer/Containerfile producer

client:
	watch -n 1 sudo podman run --pod h1-pod -it --rm --replace --name redis-client --pod h1-pod docker.io/curlimages/curl localhost:8080/get?key=counter

.PHONY: all terminals build controller clean migrate h1 h2 h3 h4 app producer client
