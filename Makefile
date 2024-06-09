all: run

run:
	cd load_balancer && make && sleep 3 && cd ../controller && make

net: 
	cd load_balancer && make net

clean:
	cd load_balancer && make clean

