all: net run 

run:
	cd load_balancer && make && sleep 2 && cd ../controller && make

net: 
	cd load_balancer && make net

stop:
	sudo killall -s 9 xterm || true
	cd load_balancer && make stop

clean: stop
	cd load_balancer && make clean

tcp-client:
	sudo podman rm -f tcp-client && sudo podman run --name tcp-client --pod h1-pod tcp-client