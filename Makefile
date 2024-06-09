all: run

run:
	cd load_balancer && make && sleep 3 && cd ../controller && make

clean:
	cd load_balancer && make clean

