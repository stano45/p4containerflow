BUILD_DIR=../../load_balancer/build
PCAP_DIR=../../load_balancer/pcaps
LOG_DIR=../../load_balancer/logs

# -i 1@s1-eth1 \
# -i 2@s1-eth2 \
# -i 3@s1-eth3 \
# -i 4@s1-eth4 \
sudo simple_switch_grpc \
    --pcap ${PCAP_DIR} \
    --device-id 0 \
    ${BUILD_DIR}/load_balance.json \
    --log-console \
    --thrift-port 9090 \
    -- \
    --grpc-server-addr 0.0.0.0:50051 >${LOG_DIR}/s1.log \
    >${LOG_DIR}/s1.log
