include ../../utils/Makefile

all: clean load_balancer bmv2-cni install build

load_balancer: 
	make -C ../../load_balancer && cp -r ../../load_balancer/build/* config

build:
	./build.sh

clean:
	./clean.sh

bmv2-cni:
	make -C ../../bmv2-cni

install:
	sudo mkdir -p /etc/p4containerflow/bmv2-switch
	sudo mkdir -p /etc/p4containerflow/bmv2-controller
	sudo cp config/load_balance.json /etc/p4containerflow/bmv2-switch/
	sudo cp config/* /etc/p4containerflow/bmv2-controller/

.PHONY: all clean load_balancer bmv2-cni build install
