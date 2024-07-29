BUILD_DIR = build
PCAP_DIR = pcaps
LOG_DIR = logs

all: compile

compile: dirs
	p4c-bm2-ss --p4v 16 --p4runtime-files $(BUILD_DIR)/load_balance.p4.p4info.txt -o $(BUILD_DIR)/load_balance.json load_balance.p4

dirs:
	mkdir -p $(BUILD_DIR) $(PCAP_DIR) $(LOG_DIR)

clean:
	rm -f *.pcap
	rm -rf $(BUILD_DIR) $(PCAP_DIR) $(LOG_DIR)

.PHONY: all run run_switches compile dirs stop clean
