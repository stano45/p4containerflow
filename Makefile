all:

compile:
	cd /home/p4/p4containerflow/load_balancer && make

control_multi_switch:
	cd /home/p4/p4containerflow/controller && sleep 2 && ./controller.py --multi_switch=True

control:
	cd /home/p4/p4containerflow/controller && sleep 2 && ./controller.py

images:
	cd /home/p4/p4containerflow/tcp && make

tcp-client:
	sudo podman run -it --rm --replace --name tcp-client --pod h1-pod tcp-client

netcat-client:
	sudo podman run -it --rm --replace --name netcat-client --pod h1-pod docker.io/subfuzion/netcat -4 -v 10.0.1.10 12345

iperf-client:
	sudo podman run -it --rm --replace --name iperf-client --pod h1-pod docker.io/networkstatic/iperf3 -4 -c 10.0.1.10 -p 12345 -t 30

.PHONY: all net compile control clean build-images tcp-client netcat-client iperf-client
