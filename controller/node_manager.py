from switch_controller import SwitchController


class NodeManager(object):
    def __init__(self, switch_controller: SwitchController):
        self.switch_controller = switch_controller
        # ipv4 -> ecmp_select_id
        self.node_map = {
            "10.0.2.2": 1,
            "10.0.3.3": 2,
        }

    def updateNode(self, old_ip, new_ip, dest_mac, egress_port):
        if new_ip in self.node_map:
            pass
            raise Exception(f"Node with IP {new_ip=} already exists")
        if old_ip not in self.node_map:
            raise Exception(f"Node with IP {old_ip=} does not exist")

        ecmp_select_id = self.node_map.pop(old_ip)

        self.switch_controller.upsertEcmpNhopEntry(
            ecmp_select=ecmp_select_id,
            dmac=dest_mac,
            ipv4=new_ip,
            port=egress_port,
            update_type="MODIFY",
        )

        self.node_map[new_ip] = ecmp_select_id
