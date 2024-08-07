from switch_controller import SwitchController


class NodeManager(object):
    def __init__(self, switch_controller: SwitchController, lb_nodes):
        self.switch_controller = switch_controller
        # ipv4 -> ecmp_select_id
        self.node_map = {}
        for i, node in enumerate(lb_nodes):
            # 0 is reserved for 10.0.1.1 (client), the rest for nodes
            ip = node["ip"]
            mac = node["mac"]
            port = node["port"]
            self.node_map[ip] = i + 1
            self.switch_controller.insertEcmpNhopEntry(
                ecmp_select=i + 1,
                dmac=mac,
                ipv4=ip,
                port=port,
                update_type="INSERT",
            )
        self.switch_controller.insertEcmpGroupEntry(
            matchDstAddr=["10.0.1.10", 32],
            ecmp_base=1,
            ecmp_count=len(lb_nodes),
        )
        self.switch_controller.readTableRules()

    def updateNode(self, old_ip, new_ip, dest_mac, egress_port):
        if new_ip in self.node_map:
            pass
            raise Exception(f"Node with IP {new_ip=} already exists")
        if old_ip not in self.node_map:
            raise Exception(f"Node with IP {old_ip=} does not exist")

        ecmp_select_id = self.node_map.pop(old_ip)

        self.switch_controller.insertEcmpNhopEntry(
            ecmp_select=ecmp_select_id,
            dmac=dest_mac,
            ipv4=new_ip,
            port=egress_port,
            update_type="MODIFY",
        )

        self.node_map[new_ip] = ecmp_select_id
