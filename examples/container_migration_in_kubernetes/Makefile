include ../../utils/Makefile

all: clean load_balancer bmv2-cni build 

load_balancer: 
	make -C ../../load_balancer && cp -r ../../load_balancer/build/* config

build:
	./build.sh

clean:
	./clean.sh

bmv2-cni:
	make -C ../../bmv2-cni

.PHONY: all clean load_balancer bmv2-cni build
