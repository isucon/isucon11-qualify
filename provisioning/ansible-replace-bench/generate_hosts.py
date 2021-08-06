#!/usr/bin/env python3

import json
import ipaddress

print("[target]")

base_addr = {
        "zone_a": ipaddress.ip_network('192.168.1.0/24'),
        "zone_c": ipaddress.ip_network('192.168.2.0/24'),
        "zone_d": ipaddress.ip_network('192.168.3.0/24'),
        }

with open("teams.json", 'r') as teams_raw:
    teams = json.load(teams_raw)

    for zone in base_addr.keys():
        idx = 0
        for h in base_addr[zone].hosts():
            idx += 1
            if idx == 1:
                # 192.168.n.1 は gateway addr のため skip
                continue
            elif idx > len(teams[zone])+1:
                break
            print(h)
