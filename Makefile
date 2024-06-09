all: run

run:
	cd load_balancer && make clean && make veth && make && sleep 3 && cd ../controller && ./controller.py

clean:
	cd load_balancer && make clean

