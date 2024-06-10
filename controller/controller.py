#!/usr/bin/env python3
import argparse
import os
import sys

from flask import Flask, request, jsonify
import grpc

from switch_controller import SwitchController
from node_manager import NodeManager

app = Flask(__name__)

global nodeManager


def main(p4info_file_path, bmv2_file_path):
    try:
        # Program switches
        # s1
        switch_controller = SwitchController(
            p4info_file_path=p4info_file_path,
            bmv2_file_path=bmv2_file_path,
            sw_name="s1",
            sw_addr="127.0.0.1:50051",
            sw_id=0,
            proto_dump_file="../load_balancer/logs/s1-p4runtime-requests.txt",
            initial_table_rules_file="../load_balancer/s1-runtime.json",
        )
        global nodeManager
        nodeManager = NodeManager(switch_controller)

        # s2
        SwitchController(
            p4info_file_path=p4info_file_path,
            bmv2_file_path=bmv2_file_path,
            sw_name="s2",
            sw_addr="127.0.0.1:50052",
            sw_id=1,
            proto_dump_file="../load_balancer/logs/s2-p4runtime-requests.txt",
            initial_table_rules_file="../load_balancer/s2-runtime.json",
        )

        # s3
        SwitchController(
            p4info_file_path=p4info_file_path,
            bmv2_file_path=bmv2_file_path,
            sw_name="s3",
            sw_addr="127.0.0.1:50053",
            sw_id=2,
            proto_dump_file="../load_balancer/logs/s3-p4runtime-requests.txt",
            initial_table_rules_file="../load_balancer/s3-runtime.json",
        )

        # s4 (no path to this switch from s1)
        SwitchController(
            p4info_file_path=p4info_file_path,
            bmv2_file_path=bmv2_file_path,
            sw_name="s4",
            sw_addr="127.0.0.1:50054",
            sw_id=3,
            proto_dump_file="../load_balancer/logs/s4-p4runtime-requests.txt",
            initial_table_rules_file="../load_balancer/s4-runtime.json",
        )
    except KeyboardInterrupt:
        print("Shutting down.")
    except grpc.RpcError as e:
        printGrpcError(e)
        exit(1)

    app.run(port=5000)


def printGrpcError(e):
    print("gRPC Error:", e.details(), end=" ")
    status_code = e.code()
    print("(%s)" % status_code.name, end=" ")
    traceback = sys.exc_info()[2]
    if traceback:
        print(
            "[%s:%d]"
            % (traceback.tb_frame.f_code.co_filename, traceback.tb_lineno)
        )


@app.route("/update_node", methods=["POST"])
def update_node():
    data = request.get_json()

    old_ipv4 = data.get("old_ipv4")
    new_ipv4 = data.get("new_ipv4")
    dest_mac = data.get("dmac")

    try:
        egress_port = int(data.get("eport"))
    except ValueError:
        return jsonify({"error": "Invalid eport parameter"}), 400

    if not all([old_ipv4, new_ipv4, dest_mac, egress_port]):
        return jsonify({"error": "Missing parameters"}), 400

    try:
        nodeManager.updateNode(old_ipv4, new_ipv4, dest_mac, egress_port)
        return jsonify({"status": "success"}), 200
    except grpc.RpcError as e:
        printGrpcError(e)
        return jsonify({"error": str(e)}), 500
    except Exception as e:
        return jsonify({"error": str(e)}), 500


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="P4Runtime Controller")
    parser.add_argument(
        "--p4info",
        help="p4info proto in text format from p4c",
        type=str,
        action="store",
        required=False,
        default="../load_balancer/build/load_balance.p4.p4info.txt",
    )
    parser.add_argument(
        "--bmv2-json",
        help="BMv2 JSON file from p4c",
        type=str,
        action="store",
        required=False,
        default="../load_balancer/build/load_balance.json",
    )
    args = parser.parse_args()

    if not os.path.exists(args.p4info):
        parser.print_help()
        print(
            "\np4info file not found: %s\nHave you run 'make'?" % args.p4info
        )
        parser.exit(1)
    if not os.path.exists(args.bmv2_json):
        parser.print_help()
        print(
            "\nBMv2 JSON file not found: %s\nHave you run 'make'?"
            % args.bmv2_json
        )
        parser.exit(1)
    main(args.p4info, args.bmv2_json)
