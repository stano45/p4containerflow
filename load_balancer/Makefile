BUILD_DIR = build
PCAP_DIR = pcaps
LOG_DIR = logs

all: run

run: compile

run_switches:	
	sudo simple_switch_grpc \
		-i 1@s1-eth1 \
		-i 2@s1-eth2 \
		-i 3@s1-eth3 \
		-i 4@s1-eth4 \
		--pcap $(PCAP_DIR) \
		--device-id 0 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9090 \
		-- \
		--grpc-server-addr 0.0.0.0:50051 > $(LOG_DIR)/s1.log \
		> $(LOG_DIR)/s1.log \
		2>&1 & \
		echo $!

	sudo simple_switch_grpc \
		-i 1@s2-eth1 \
		-i 2@s2-eth2 \
		--pcap $(PCAP_DIR) \
		--device-id 1 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9091 \
		-- \
		--grpc-server-addr 0.0.0.0:50052 > $(LOG_DIR)/s2.log \
		> $(LOG_DIR)/s2.log \
		2>&1 & \
		echo $!

	sudo simple_switch_grpc \
		-i 1@s3-eth1 \
		-i 2@s3-eth2 \
		--pcap $(PCAP_DIR) \
		--device-id 2 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9092 \
		-- \
		--grpc-server-addr 0.0.0.0:50053 \
		> $(LOG_DIR)/s3.log \
		2>&1 & \
		echo $!

	sudo simple_switch_grpc \
		-i 1@s4-eth1 \
		-i 2@s4-eth2 \
		--pcap $(PCAP_DIR) \
		--device-id 3 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9093 \
		-- \
		--grpc-server-addr 0.0.0.0:50054 \
		> $(LOG_DIR)/s4.log \
		2>&1 & \
		echo $!

compile: dirs
	p4c-bm2-ss --p4v 16 --p4runtime-files $(BUILD_DIR)/load_balance.p4.p4info.txt -o $(BUILD_DIR)/load_balance.json load_balance.p4

dirs:
	mkdir -p $(BUILD_DIR) $(PCAP_DIR) $(LOG_DIR)

stop:
	sudo killall -s 9 simple_switch_grpc || true
	sudo killall -s 9 xterm || true

clean: stop
	rm -f *.pcap
	rm -rf $(BUILD_DIR) $(PCAP_DIR) $(LOG_DIR)
	