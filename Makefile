.PHONY: default load_balancer controller update_node clean

default:
	@echo "Error: No target specified. Use 'make load_balancer' or 'make controller' or 'make update_node' or 'make clean'"
	@exit 1


load_balancer: 
	cd load_balancer && make clean && make

controller: 
	cd controller && ./controller.py

update_node:
	cd scripts && ./update_node.sh

clean:
	cd load_balancer && make clean

