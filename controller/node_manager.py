from switch_controller import SwitchController

# TODO: pass this in a config file
LB_IP = "10.244.1.244"

class NodeManager(object):
    def __init__(self, switch_controller: SwitchController, lb_nodes):
        self.switch_controller = switch_controller
        # ipv4 -> ecmp_select_id
        self.node_map = {}
        self.client = None

        if lb_nodes is not None:
            for i, node in enumerate(lb_nodes):
                # Port 0 is reserved for client, the rest for server nodes
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
            self.switch_controller.insertEcmpGroupSelectEntry(
                matchDstAddr=[LB_IP, 32],
                ecmp_base=1,
                ecmp_count=len(lb_nodes),
            )
        self.switch_controller.readTableRules()

    def updateNode(self, old_ip, new_ip, dest_mac, egress_port):
        if new_ip in self.node_map:
            raise Exception(f"Node with IP {new_ip} already exists")
        if old_ip not in self.node_map:
            raise Exception(f"Node with IP {old_ip} does not exist")

        ecmp_select_id = self.node_map.pop(old_ip)

        self.switch_controller.insertEcmpNhopEntry(
            ecmp_select=ecmp_select_id,
            dmac=dest_mac,
            ipv4=new_ip,
            port=egress_port,
            update_type="MODIFY",
        )

        self.node_map[new_ip] = ecmp_select_id

    def _addServerNode(self, ip, smac, dmac, port):
        ecmp_select_id = len(self.node_map) + 1
        self.switch_controller.insertEcmpNhopEntry(
            ecmp_select=ecmp_select_id,
            dmac=dmac,
            ipv4=ip,
            port=port,
            update_type="INSERT",
        )

        self.node_map[ip] = ecmp_select_id

        self.switch_controller.insertEcmpGroupSelectEntry(
            matchDstAddr=[LB_IP, 32],
            ecmp_base=1,
            ecmp_count=len(self.node_map),
            update_type="MODIFY",
        )

        self.switch_controller.insertSendFrameEntry(egress_port=port, smac=smac)
    
    def _addClientNode(self, ip, smac, dmac, port):
        # TODO: 
        # if self.client is not None:
        #     raise Exception("Client node already exists")

        self.switch_controller.insertEcmpNhopEntry(
            ecmp_select=0,
            dmac=dmac,
            ipv4=ip,
            port=port,
            update_type="INSERT",
        )

        self.switch_controller.insertEcmpGroupRewriteSrcEntry(
            matchDstAddr=[ip, 32],
            new_src=LB_IP,
            update_type="INSERT",
        )

        self.switch_controller.insertSendFrameEntry(egress_port=port, smac=smac, update_type="INSERT")

        self.client = (ip, smac, dmac, port)

    def addNode(self, ip, smac, dmac, port, isClient):
        if ip in self.node_map:
            raise Exception(f"Node with IP {ip} already exists")
        
        addNodeHandlerFn = self._addClientNode if isClient else self._addServerNode
        addNodeHandlerFn(ip, smac, dmac, port)

        self.switch_controller.readTableRules()
