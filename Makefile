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

h1:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h1" -hold -e "sudo podman rm -f tcp-client && sudo podman run --name tcp-client --pod h1-pod tcp-client" &

h2:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h2" -hold -e "sudo podman logs -f server2" &

h3:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h3" -hold -e "sudo podman logs -f server3" &

h4:
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h4" -hold -e "sudo podman logs -f server4" &

