#!/usr/bin/env python3
import argparse
import json
import os
import sys

import grpc
from flask import Flask, jsonify, request
from node_manager import NodeManager
from switch_controller import SwitchController

app = Flask(__name__)

global nodeManager


def main(config_file_path):
    try:
        with open(config_file_path, "r") as config_file:
            switch_configs = json.load(config_file)

        switch_controllers = []
        master_config = None

        for config in switch_configs:
            # Master needs to be initialized last,
            # otherwise performing master arbitration update will fail
            if config.get("master", False):
                if master_config is not None:
                    raise Exception(
                        "Multiple master switches specified "
                        "in the configuration file."
                    )
                master_config = config
                continue

            switch_controller = SwitchController(
                p4info_file_path=config["p4info_file_path"],
                bmv2_file_path=config["bmv2_file_path"],
                sw_name=config["name"],
                sw_addr=config["addr"],
                sw_id=config["id"],
                proto_dump_file=config["proto_dump_file"],
                initial_table_rules_file=config["runtime_file"],
            )
            switch_controllers.append(switch_controller)

        if master_config is None:
            raise Exception(
                "No master switch specified in the configuration file."
            )
        lb_nodes = master_config.get("lb_nodes", None)

        master_controller = SwitchController(
            p4info_file_path=master_config["p4info_file_path"],
            bmv2_file_path=master_config["bmv2_file_path"],
            sw_name=master_config["name"],
            sw_addr=master_config["addr"],
            sw_id=master_config["id"],
            proto_dump_file=master_config["proto_dump_file"],
            initial_table_rules_file=master_config["runtime_file"],
        )

        global nodeManager
        nodeManager = NodeManager(
            switch_controller=master_controller, lb_nodes=lb_nodes
        )

    except KeyboardInterrupt:
        print("Shutting down.")
    except grpc.RpcError as e:
        printGrpcError(e)
        exit(1)
    except Exception as e:
        print(f"Error: {e}")
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


@app.route("/add_node", methods=["POST"])
def add_node():
    data = request.get_json()

    ipv4 = data.get("ipv4")
    dest_mac = data.get("dmac")
    src_mac = data.get("smac")
    isClient = data.get("isClient")

    try:
        egress_port = int(data.get("eport"))
    except ValueError:
        return jsonify({"error": "Invalid eport parameter"}), 400

    if ipv4 is None or dest_mac is None or src_mac is None or isClient is None:
        return jsonify({"error": "Missing parameters"}), 400

    try:
        nodeManager.addNode(ipv4, src_mac, dest_mac, egress_port, isClient)
        return jsonify({"status": "success"}), 200
    except grpc.RpcError as e:
        printGrpcError(e)
        return jsonify({"error": str(e)}), 500
    except Exception as e:
        return jsonify({"error": str(e)}), 500



if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="P4Runtime Controller")
    parser.add_argument(
        "--config",
        help="JSON configuration file for switches",
        type=str,
        action="store",
        required=True,
    )
    args = parser.parse_args()

    if not os.path.exists(args.config):
        parser.print_help()
        print(f"\nConfiguration file not found: {args.config}")
        parser.exit(1)
    main(args.config)
