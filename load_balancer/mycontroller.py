#!/usr/bin/env python3
import argparse
import os
import sys

from flask import Flask, request, jsonify
import grpc

# Import P4Runtime lib from parent utils dir
# Probably there's a better way of doing this.
sys.path.append(
    os.path.join(os.path.dirname(os.path.abspath(__file__)),
                 '../utils/'))
import p4runtime_lib.bmv2
import p4runtime_lib.helper
from p4runtime_lib.switch import ShutdownAllSwitchConnections

app = Flask(__name__)

# Global variables for switch and P4 info helper
s1 = None
p4info_helper = None


def writeTableEntriesS1(p4info_helper, sw):
    # Group IP (for load balancing)
    table_entry = p4info_helper.buildTableEntry(
        table_name="MyIngress.ecmp_group",
        default_action=True,
        action_name="MyIngress.drop",
        action_params={})
    sw.WriteTableEntry(table_entry, update_type="MODIFY")
    print("Installed ecmp_group drop rule on %s" % sw.name)
    
    table_entry = p4info_helper.buildTableEntry(
        table_name="MyIngress.ecmp_group",
        match_fields={
            "hdr.ipv4.dstAddr": ("10.0.0.1", 32)
        },
        action_name="MyIngress.set_ecmp_select",
        action_params={
            "ecmp_base": 0,
            "ecmp_count": 2
        })
    sw.WriteTableEntry(table_entry)
    print("Installed ecmp_group set_ecmp_select rule on %s" % sw.name)
    
    # Hops
    updateEcmpNhopTable(p4info_helper=p4info_helper, sw=sw, ecmp_select=0, dmac="00:00:00:00:01:02", ipv4="10.0.2.2", port=2)
    updateEcmpNhopTable(p4info_helper=p4info_helper, sw=sw, ecmp_select=1, dmac="00:00:00:00:01:03", ipv4="10.0.3.3", port=3)
    
    # Egress
    updateSendFrameTable(p4info_helper=p4info_helper, sw=sw, egress_port=2, smac="00:00:00:01:02:00")
    updateSendFrameTable(p4info_helper=p4info_helper, sw=sw, egress_port=3, smac="00:00:00:01:03:00")


def updateSendFrameTable(p4info_helper, sw, egress_port, smac):
    table_entry = p4info_helper.buildTableEntry(
        table_name="MyEgress.send_frame",
        match_fields={
            "standard_metadata.egress_port": egress_port
        },
        action_name="MyEgress.rewrite_mac",
        action_params={
            "smac": smac
        })
    sw.WriteTableEntry(table_entry)
    print(f"Updated the 'send_frame' table on {sw.name=} with {egress_port=}, {smac=}")

def updateEcmpNhopTable(p4info_helper, sw, ecmp_select, dmac, ipv4, port, update_type="INSERT"):
    table_entry = p4info_helper.buildTableEntry(
        table_name="MyIngress.ecmp_nhop",
        match_fields={
            "meta.ecmp_select": ecmp_select
        },
        action_name="MyIngress.set_nhop",
        action_params={
            "nhop_dmac": dmac,
            "nhop_ipv4": ipv4,
            "port": port
        })
    sw.WriteTableEntry(table_entry, update_type=update_type)
    print(f"Updated the 'ecmp_nhop' table on {sw.name=} with {ecmp_select=}, {ipv4=}, {dmac=}, {port=}")
    
def deleteEcmpNhopTable(p4info_helper, sw, ecmp_select):
    table_entry = p4info_helper.buildTableEntry(
        table_name="MyIngress.ecmp_nhop",
        match_fields={
            "meta.ecmp_select": ecmp_select
        },
        )
    sw.WriteTableEntry(table_entry, update_type="DELETE")
    print(f"Deleted a 'ecmp_nhop' table entry on {sw.name=} with {ecmp_select=}")

def readTableRules(p4info_helper, sw):
    """
    Reads the table entries from all tables on the switch.

    :param p4info_helper: the P4Info helper
    :param sw: the switch connection
    """
    print('\n----- Reading tables rules for %s -----' % sw.name)
    for response in sw.ReadTableEntries():
        for entity in response.entities:
            entry = entity.table_entry
            table_name = p4info_helper.get_tables_name(entry.table_id)
            print('%s: ' % table_name, end=' ')
            for m in entry.match:
                print(p4info_helper.get_match_field_name(table_name, m.field_id), end=' ')
                print('%r' % (p4info_helper.get_match_field_value(m),), end=' ')
            action = entry.action.action
            action_name = p4info_helper.get_actions_name(action.action_id)
            print('->', action_name, end=' ')
            for p in action.params:
                print(p4info_helper.get_action_param_name(action_name, p.param_id), end=' ')
                print('%r' % p.value, end=' ')
            print()

def printGrpcError(e):
    print("gRPC Error:", e.details(), end=' ')
    status_code = e.code()
    print("(%s)" % status_code.name, end=' ')
    traceback = sys.exc_info()[2]
    print("[%s:%d]" % (traceback.tb_frame.f_code.co_filename, traceback.tb_lineno))
    
def initialize_switches(p4info_helper, bmv2_file_path):
    # Create a switch connection object for s1 and s2;
    # this is backed by a P4Runtime gRPC connection.
    # Also, dump all P4Runtime messages sent to switch to given txt files.
    global s1
    s1 = p4runtime_lib.bmv2.Bmv2SwitchConnection(
        name='s1',
        address='127.0.0.1:50051',
        device_id=0,
        proto_dump_file='logs/s1-p4runtime-requests.txt')

    # Send master arbitration update message to establish this controller as
    # master (required by P4Runtime before performing any other write operation)
    s1.MasterArbitrationUpdate()
    # s2.MasterArbitrationUpdate()

    # Install the P4 program on the switches
    s1.SetForwardingPipelineConfig(p4info=p4info_helper.p4info,
                                    bmv2_json_file_path=bmv2_file_path)
    print("Installed P4 Program using SetForwardingPipelineConfig on s1")

    writeTableEntriesS1(p4info_helper, sw=s1)
    readTableRules(p4info_helper, s1)

def main(p4info_file_path, bmv2_file_path):
    global p4info_helper
    p4info_helper = p4runtime_lib.helper.P4InfoHelper(p4info_file_path)

    try:
        initialize_switches(p4info_helper, bmv2_file_path)
    except KeyboardInterrupt:
        print("Shutting down.")
    except grpc.RpcError as e:
        printGrpcError(e)
        
    app.run(port=5000)

    ShutdownAllSwitchConnections()
    
    
@app.route('/insert_hop', methods=['POST'])
def insert_hop():
    global p4info_helper, s1
    data = request.get_json()

    ecmp_select = data.get('ecmp_select')
    dmac = data.get('dmac')
    ipv4 = data.get('ipv4')
    port = data.get('port')

    if not all([ecmp_select, dmac, ipv4, port]):
        return jsonify({'error': 'Missing parameters'}), 400

    try:
        updateEcmpNhopTable(p4info_helper, s1, ecmp_select, dmac, ipv4, port, update_type="INSERT")
        return jsonify({'status': 'success'}), 200
    except grpc.RpcError as e:
        printGrpcError(e)
        return jsonify({'error': str(e)}), 500

@app.route('/update_hop', methods=['POST'])
def update_hop():
    global p4info_helper, s1
    data = request.get_json()

    ecmp_select = data.get('ecmp_select')
    dmac = data.get('dmac')
    ipv4 = data.get('ipv4')
    port = data.get('port')

    if not all([ecmp_select, dmac, ipv4, port]):
        return jsonify({'error': 'Missing parameters'}), 400

    try:
        updateEcmpNhopTable(p4info_helper, s1, ecmp_select, dmac, ipv4, port, update_type="MODIFY")
        return jsonify({'status': 'success'}), 200
    except grpc.RpcError as e:
        printGrpcError(e)
        return jsonify({'error': str(e)}), 500
    
@app.route('/delete_hop', methods=['POST'])
def delete_hop():
    global p4info_helper, s1
    data = request.get_json()

    ecmp_select = data.get('ecmp_select')

    if ecmp_select is None:
        return jsonify({'error': 'Missing ecmp_select parameter'}), 400

    try:
        deleteEcmpNhopTable(p4info_helper, s1, ecmp_select)
        return jsonify({'status': 'success'}), 200
    except grpc.RpcError as e:
        printGrpcError(e)
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='P4Runtime Controller')
    parser.add_argument('--p4info', help='p4info proto in text format from p4c',
                        type=str, action="store", required=False,
                        default='./build/load_balance.p4.p4info.txt')
    parser.add_argument('--bmv2-json', help='BMv2 JSON file from p4c',
                        type=str, action="store", required=False,
                        default='./build/load_balance.json')
    args = parser.parse_args()

    if not os.path.exists(args.p4info):
        parser.print_help()
        print("\np4info file not found: %s\nHave you run 'make'?" % args.p4info)
        parser.exit(1)
    if not os.path.exists(args.bmv2_json):
        parser.print_help()
        print("\nBMv2 JSON file not found: %s\nHave you run 'make'?" % args.bmv2_json)
        parser.exit(1)
    main(args.p4info, args.bmv2_json)
