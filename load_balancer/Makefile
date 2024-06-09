BUILD_DIR = build
PCAP_DIR = pcaps
LOG_DIR = logs

all: run

run: stop compile
	xterm -T "s1" -hold -e "sudo simple_switch_grpc \
		-i 1@s1-eth1 \
		-i 2@s1-eth2 \
		-i 3@s1-eth3 \
		--pcap $(PCAP_DIR) \
		--device-id 0 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9090 \
		-- \
		--grpc-server-addr 0.0.0.0:50051" &

	xterm -T "s2" -hold -e "sudo simple_switch_grpc \
		-i 1@s2-eth1 \
		-i 2@s2-eth2 \
		--pcap $(PCAP_DIR) \
		--device-id 1 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9091 \
		-- \
		--grpc-server-addr 0.0.0.0:50052" &

	xterm -T "s3" -hold -e "sudo simple_switch_grpc \
		-i 1@s3-eth1 \
		-i 2@s3-eth2 \
		--pcap $(PCAP_DIR) \
		--device-id 2 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9092 \
		-- \
		--grpc-server-addr 0.0.0.0:50053" &
	
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h1" -hold -e "sudo ip netns exec h1 /bin/bash" &
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h2" -hold -e "sudo ip netns exec h2 /bin/bash" &
	xterm -xrm 'XTerm.vt100.allowTitleOps: false' -T "h3" -hold -e "sudo ip netns exec h3 /bin/bash" &


compile: dirs
	p4c-bm2-ss --p4v 16 --p4runtime-files $(BUILD_DIR)/load_balance.p4.p4info.txt -o $(BUILD_DIR)/load_balance.json load_balance.p4

dirs:
	mkdir -p $(BUILD_DIR) $(PCAP_DIR) $(LOG_DIR)

net: clean
	../scripts/build_network.sh

stop:
	sudo killall -s 9 xterm || true

clean: stop
	rm -f *.pcap
	rm -rf $(BUILD_DIR) $(PCAP_DIR) $(LOG_DIR)
	../scripts/destroy_network.sh
	