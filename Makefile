all: run

run:
	cd load_balancer && make && sleep 3 && cd ../controller && make

net: 
	cd load_balancer && make net

stop:
	cd load_balancer && make stop

clean:
	cd load_balancer && make clean

h1:
	sudo ip netns exec h1 /bin/bash -c 'cd ../scripts && /bin/bash'

h2:
	sudo ip netns exec h2 /bin/bash -c 'cd ../scripts && /bin/bash'

h3:
	sudo ip netns exec h3 /bin/bash -c 'cd ./scripts && /bin/bash'

h4:
	sudo ip netns exec h4 /bin/bash -c 'cd ./scripts && /bin/bash'
