BUILD_DIR = build
PCAP_DIR = pcaps
LOG_DIR = logs

all: run

run: compile
#	sudo simple_switch_grpc \
#		-i 0@veth0-0 \
#		-i 1@veth1-0 \
#		--log-console \
#		--thrift-port 9090 \
#		$(BUILD_DIR)/main.json \
#		-- \
#		--grpc-server-addr 0.0.0.0:9551

	sudo simple_switch_grpc \
		-i 1@s1-eth1 \
		-i 2@s1-eth2 \
		-i 3@s1-eth3 \
		--pcap /home/p4/p4containerflow/load_balancer/pcaps \
		--nanolog ipc:///tmp/bm-0-log.ipc \
		--device-id 0 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9090 \
		-- \
		--grpc-server-addr 0.0.0.0:50051

	simple_switch_grpc \
		-i 1@s2-eth1 \
		-i 2@s2-eth2 \
		--pcap /home/p4/p4containerflow/load_balancer/pcaps \
		--nanolog ipc:///tmp/bm-1-log.ipc \
		--device-id 1 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9091 \
		-- \
		--grpc-server-addr 0.0.0.0:50052

	simple_switch_grpc \
		-i 1@s3-eth1 \
		-i 2@s3-eth2 \
		--pcap /home/p4/p4containerflow/load_balancer/pcaps \
		--nanolog ipc:///tmp/bm-2-log.ipc \
		--device-id 2 \
		$(BUILD_DIR)/load_balance.json \
		--log-console \
		--thrift-port 9092 \
		-- --grpc-server-addr 0.0.0.0:50053

compile: dirs
	p4c-bm2-ss --p4v 16 --p4runtime-files $(BUILD_DIR)/load_balance.p4.p4info.txt -o $(BUILD_DIR)/load_balance.json load_balance.p4

dirs:
	mkdir -p $(BUILD_DIR) $(PCAP_DIR) $(LOG_DIR)

veth:
	sudo ip link add s1-eth1 type veth peer name h1-eth1
	sudo ip link add s1-eth2 type veth peer name s2-eth1
	sudo ip link add s1-eth3 type veth peer name s3-eth1
	sudo ip link add s2-eth2 type veth peer name h2-eth1
	sudo ip link add s3-eth2 type veth peer name h3-eth1

	sudo ip netns add h1
	sudo ip link set h1-eth1 netns h1
	sudo ip netns exec h1 ip link set dev h1-eth1 address 08:00:00:00:01:01
	sudo ip netns exec h1 ip addr add 10.0.1.1/24 dev h1-eth1
	sudo ip netns exec h1 ip link set dev h1-eth1 up
	sudo ip netns exec h1 route add default gw 10.0.1.10 dev h1-eth1
	sudo ip netns exec h1 arp -i h1-eth1 -s 10.0.1.10 08:00:00:00:01:00

	sudo ip netns add h2
	sudo ip link set h2-eth1 netns h2
	sudo ip netns exec h2 ip link set dev h2-eth1 address 08:00:00:00:02:02
	sudo ip netns exec h2 ip addr add 10.0.2.2/24 dev h2-eth1
	sudo ip netns exec h2 ip link set dev h2-eth1 up
	sudo ip netns exec h2 route add default gw 10.0.2.20 dev h2-eth1
	sudo ip netns exec h2 arp -i h2-eth1 -s 10.0.2.20 08:00:00:00:02:00

	sudo ip netns add h3
	sudo ip link set h3-eth1 netns h3
	sudo ip netns exec h3 ip link set dev h3-eth1 address 08:00:00:00:03:03
	sudo ip netns exec h3 ip addr add 10.0.3.3/24 dev h3-eth1
	sudo ip netns exec h3 ip link set dev h3-eth1 up
	sudo ip netns exec h3 route add default gw 10.0.3.30 dev h3-eth1
	sudo ip netns exec h3 arp -i h3-eth1 -s 10.0.3.30 08:00:00:00:03:00

	sudo ip link set dev s1-eth1 up
	sudo ip link set dev s1-eth2 up
	sudo ip link set dev s1-eth3 up
	sudo ip link set dev s2-eth1 up
	sudo ip link set dev s2-eth2 up
	sudo ip link set dev s3-eth1 up
	sudo ip link set dev s3-eth2 up

#	sudo ethtool -K veth0-0 tx off
#	sudo ethtool -K veth0-1 tx off
#	sudo ethtool -K veth1-0 tx off
#	sudo ethtool -K veth1-1 tx off

clean:
	rm -f *.pcap
	rm -rf $(BUILD_DIR) $(PCAP_DIR) $(LOG_DIR)
	sudo ip link del s1-eth1
	sudo ip link del s1-eth2
	sudo ip link del s1-eth3
	sudo ip link del s2-eth2
	sudo ip link del s3-eth2

	sudo ip netns del h1
	sudo ip netns del h2
	sudo ip netns del h3
	