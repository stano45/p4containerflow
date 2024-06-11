# Import P4Runtime lib from parent utils dir
# Probably there's a better way of doing this.
import json

import p4runtime_lib.bmv2
import p4runtime_lib.helper
from p4runtime_lib.switch import ShutdownAllSwitchConnections
from p4runtime_lib.simple_controller import program_from_file


class SwitchController(object):
    def __init__(
        self,
        p4info_file_path,
        bmv2_file_path,
        sw_name,
        sw_addr,
        sw_id,
        proto_dump_file,
        initial_table_rules_file=None,
    ):
        self.p4info_helper = p4runtime_lib.helper.P4InfoHelper(
            p4info_file_path
        )
        self.sw = p4runtime_lib.bmv2.Bmv2SwitchConnection(
            name=sw_name,
            address=sw_addr,
            device_id=sw_id,
            proto_dump_file=proto_dump_file,
        )
        self.initial_table_rules_file = initial_table_rules_file

        # Send master arbitration update message
        # to establish this controller as
        # master (required by P4Runtime before
        # performing any other write operation)
        self.sw.MasterArbitrationUpdate()

        # Install the P4 program on the switches
        self.sw.SetForwardingPipelineConfig(
            p4info=self.p4info_helper.p4info,
            bmv2_json_file_path=bmv2_file_path,
        )
        print(
            f"Installed P4 Program using"
            f"SetForwardingPipelineConfig on {sw_name}"
        )

        if initial_table_rules_file:
            with open(initial_table_rules_file, "r") as sw_conf_file:
                sw_conf = json.load(sw_conf_file)
                program_from_file(
                    sw=self.sw,
                    sw_conf=sw_conf,
                    p4info_helper=self.p4info_helper,
                    runtime_json=None,
                )

    def __del__(self):
        ShutdownAllSwitchConnections()

    def upsertEcmpNhopEntry(
        self, ecmp_select, dmac, ipv4, port, update_type="INSERT"
    ):
        table_entry = self.p4info_helper.buildTableEntry(
            table_name="MyIngress.ecmp_nhop",
            match_fields={"meta.ecmp_select": ecmp_select},
            action_name="MyIngress.set_nhop",
            action_params={"nhop_dmac": dmac, "nhop_ipv4": ipv4, "port": port},
        )
        self.sw.WriteTableEntry(table_entry, update_type=update_type)
        print(
            f"Updated the 'ecmp_nhop' table on "
            f"{self.sw.name=} with {ecmp_select=}, {ipv4=}, {dmac=}, {port=}"
        )

    def deleteEcmpNhopEntry(self, ecmp_select):
        table_entry = self.p4info_helper.buildTableEntry(
            table_name="MyIngress.ecmp_nhop",
            match_fields={"meta.ecmp_select": ecmp_select},
        )
        self.sw.WriteTableEntry(table_entry, update_type="DELETE")
        print(
            f"Deleted a 'ecmp_nhop' table entry on"
            f"{self.sw.name=} with {ecmp_select=}"
        )

    def upsertSendFrameEntry(self, egress_port, smac):
        table_entry = self.p4info_helper.buildTableEntry(
            table_name="MyEgress.send_frame",
            match_fields={"standard_metadata.egress_port": egress_port},
            action_name="MyEgress.rewrite_mac",
            action_params={"smac": smac},
        )
        self.sw.WriteTableEntry(table_entry)
        print(
            f"Updated the 'send_frame' table on "
            f"{self.sw.name=} with {egress_port=}, {smac=}"
        )

    def deleteSendFrameEntry(self, egress_port):
        table_entry = self.p4info_helper.buildTableEntry(
            table_name="MyEgress.send_frame",
            match_fields={"standard_metadata.egress_port": egress_port},
        )
        self.sw.WriteTableEntry(table_entry, update_type="DELETE")
        print(
            f"Deleted a 'send_frame' table entry on"
            f"{self.sw.name=} with {egress_port=}"
        )

    def readTableRules(self):
        """
        Reads the table entries from all tables on the switch.

        :param p4info_helper: the P4Info helper
        :param sw: the switch connection
        """
        print("\n----- Reading tables rules for %s -----" % self.sw.name)
        for response in self.sw.ReadTableEntries():
            for entity in response.entities:
                entry = entity.table_entry
                table_name = self.p4info_helper.get_tables_name(entry.table_id)
                print("%s: " % table_name, end=" ")
                for m in entry.match:
                    print(
                        self.p4info_helper.get_match_field_name(
                            table_name, m.field_id
                        ),
                        end=" ",
                    )
                    print(
                        "%r" % (self.p4info_helper.get_match_field_value(m),),
                        end=" ",
                    )
                action = entry.action.action
                action_name = self.p4info_helper.get_actions_name(
                    action.action_id
                )
                print("->", action_name, end=" ")
                for p in action.params:
                    print(
                        self.p4info_helper.get_action_param_name(
                            action_name, p.param_id
                        ),
                        end=" ",
                    )
                    print("%r" % p.value, end=" ")
                print()
