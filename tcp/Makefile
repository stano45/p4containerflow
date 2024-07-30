all: server-bin client-bin server-container client-container

client-bin:: client.c
	gcc -Wall -o client client.c

server-bin:: server.c
	gcc -Wall -o server server.c

server-container:: Containerfile.server
	sudo podman build -t tcp-server -f Containerfile.server .

client-container:: Containerfile.client
	sudo podman build -t tcp-client -f Containerfile.client .

clean:
	rm -rf client server
	sudo podman rmi tcp-server tcp-client

.PHONY: all client-bin server-bin server-container clean