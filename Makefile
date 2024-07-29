all: compile net control

net: 
	cd scripts/switch_container && make

compile:
	cd load_balancer && make

control:
	cd controller && sleep 2 && make

clean:
	cd scripts/switch_container && make clean

build-images:
	cd tcp && make

tcp-client:
	sudo podman rm -f tcp-client && sudo podman run --name tcp-client --pod h1-pod tcp-client