.PHONY: default load_balancer controller

default:
	@echo "Error: No target specified. Use 'make load_balancer' or 'make controller'"
	@exit 1


load_balancer: 
	cd load_balancer && make clean && make

controller: 
	cd controller && ./controller.py

clean:
	cd load_balancer && make clean

