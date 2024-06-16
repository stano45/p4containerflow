#!/usr/bin/env python3
import sys

from scapy.all import (
    get_if_list,
    sniff,
)


def get_if(iface):
    if iface in get_if_list():
        return iface
    else:
        print(f"Interface {iface} not found.")
        exit(1)


def handle_pkt(pkt):
    print("got a packet")
    pkt.show2()
    sys.stdout.flush()


def main():
    if len(sys.argv) != 2:
        print("Usage: %s <interface>" % sys.argv[0])
        sys.exit(1)

    iface = sys.argv[1]
    iface = get_if(iface)

    print(f"sniffing on {iface}")
    sys.stdout.flush()
    sniff(filter="tcp", iface=iface, prn=lambda x: handle_pkt(x))


if __name__ == "__main__":
    main()
