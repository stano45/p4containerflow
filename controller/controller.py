#!/usr/bin/env python3
import argparse
import os
import sys

from flask import Flask, request, jsonify
import grpc

from switch_controller import SwitchController

app = Flask(__name__)

global switchController
 

def main(p4info_file_path, bmv2_file_path):
    try:
        global switchController
        switchController = SwitchController(p4info_file_path, bmv2_file_path )
    except KeyboardInterrupt:
        print("Shutting down.")
    except grpc.RpcError as e:
        printGrpcError(e)

    # Run Flask app
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


@app.route("/insert_hop", methods=["POST"])
def insert_hop():
    data = request.get_json()

    ecmp_select = data.get("ecmp_select")
    dmac = data.get("dmac")
    ipv4 = data.get("ipv4")
    port = data.get("port")

    if not all([ecmp_select, dmac, ipv4, port]):
        return jsonify({"error": "Missing parameters"}), 400

    try:
        switchController.upsertEcmpNhopEntry(
            ecmp_select, dmac, ipv4, port, update_type="INSERT"
        )
        return jsonify({"status": "success"}), 200
    except grpc.RpcError as e:
        printGrpcError(e)
        return jsonify({"error": str(e)}), 500


@app.route("/update_hop", methods=["POST"])
def update_hop():
    data = request.get_json()

    ecmp_select = data.get("ecmp_select")
    dmac = data.get("dmac")
    ipv4 = data.get("ipv4")
    port = data.get("port")

    if not all([ecmp_select, dmac, ipv4, port]):
        return jsonify({"error": "Missing parameters"}), 400

    try:
        switchController.upsertEcmpNhopEntry(
            ecmp_select, dmac, ipv4, port, update_type="MODIFY"
        )
        return jsonify({"status": "success"}), 200
    except grpc.RpcError as e:
        printGrpcError(e)
        return jsonify({"error": str(e)}), 500


@app.route("/delete_hop", methods=["POST"])
def delete_hop():
    data = request.get_json()

    ecmp_select = data.get("ecmp_select")

    if ecmp_select is None:
        return jsonify({"error": "Missing ecmp_select parameter"}), 400

    try:
        switchController.deleteEcmpNhopEntry(ecmp_select)
        return jsonify({"status": "success"}), 200
    except grpc.RpcError as e:
        printGrpcError(e)
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
