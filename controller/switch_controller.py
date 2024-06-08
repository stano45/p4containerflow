# Import P4Runtime lib from parent utils dir
# Probably there's a better way of doing this.
import os
import sys


sys.path.append(
    os.path.join(os.path.dirname(os.path.abspath(__file__)), "../utils/")
)
import p4runtime_lib.bmv2  # noqa
import p4runtime_lib.helper  # noqa
from p4runtime_lib.switch import ShutdownAllSwitchConnections  # noqa


class SwitchController(object):
    def __init__(self, p4info_file_path, bmv2_file_path):
        self.p4info_helper = p4runtime_lib.helper.P4InfoHelper(
            p4info_file_path
        )
        self.sw = p4runtime_lib.bmv2.Bmv2SwitchConnection(
            name="s1",
            address="127.0.0.1:50051",
            device_id=0,
            proto_dump_file="../load_balancer/logs/s1-p4runtime-requests.txt",
        )

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
        print("Installed P4 Program using SetForwardingPipelineConfig on s1")

        self._writeTableEntriesS1()

    def __del__(self):
        ShutdownAllSwitchConnections()

    def _writeTableEntriesS1(self):
        # Group IP (for load balancing)
        table_entry = self.p4info_helper.buildTableEntry(
            table_name="MyIngress.ecmp_group",
            default_action=True,
            action_name="MyIngress.drop",
            action_params={},
        )
        self.sw.WriteTableEntry(table_entry, update_type="MODIFY")
        print("Installed ecmp_group drop rule on %s" % self.sw.name)

        table_entry = self.p4info_helper.buildTableEntry(
            table_name="MyIngress.ecmp_group",
            match_fields={"hdr.ipv4.dstAddr": ("10.0.0.1", 32)},
            action_name="MyIngress.set_ecmp_select",
            action_params={"ecmp_base": 1, "ecmp_count": 2},
        )
        self.sw.WriteTableEntry(table_entry)
        print("Installed ecmp_group set_ecmp_select rule on %s" % self.sw.name)

        table_entry = self.p4info_helper.buildTableEntry(
            table_name="MyIngress.ecmp_group",
            match_fields={"hdr.ipv4.dstAddr": ("10.0.1.1", 32)},
            action_name="MyIngress.set_rewrite_src",
            action_params={"new_src": "10.0.0.1"},
        )
        self.sw.WriteTableEntry(table_entry)
        print("Installed ecmp_group set_rewrite_src rule on %s" % self.sw.name)

        # Hops
        self.upsertEcmpNhopEntry(
            ecmp_select=0,
            dmac="00:00:00:00:01:01",
            ipv4="10.0.1.1",
            port=1,
        )
        self.upsertEcmpNhopEntry(
            ecmp_select=1,
            dmac="00:00:00:00:01:02",
            ipv4="10.0.2.2",
            port=2,
        )
        self.upsertEcmpNhopEntry(
            ecmp_select=2,
            dmac="00:00:00:00:01:03",
            ipv4="10.0.3.3",
            port=3,
        )

        # Egress
        self.upsertSendFrameEntry(egress_port=1, smac="00:00:00:01:01:00")
        self.upsertSendFrameEntry(egress_port=2, smac="00:00:00:01:02:00")
        self.upsertSendFrameEntry(egress_port=3, smac="00:00:00:01:03:00")

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
