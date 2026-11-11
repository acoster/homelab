# Network addresses

Taken from the compose files and `shared.env` files in this repo. Update this
page when an address changes.

## Hosts

| Host      | Address       |
| --------- | ------------- |
| raspberry | 192.168.1.212 |
| framboesa | 192.168.1.39  |
| datengrab | 192.168.1.14  |
| cloudy    | VPS           |

This is each host's `HOST_IP`. Published ports are bound to this address only,
so a service on a host does not answer on the host's other addresses (for
example a VLAN 5 address handed out by DHCP).

## Macvlan addresses

Containers with their own address on the LAN or on VLAN 5, taken from the
compose files. Nothing else needs a fixed address; the rest is reached through
Traefik or the host's `HOST_IP`.

| Host      | Container  | IP                                                 |
| --------- | ---------- | -------------------------------------------------- |
| raspberry | hass       | 192.168.1.246, 2a02:169:4bce:1:4e81:beff:fe81:babe |
| raspberry | hass       | 192.168.5.200, 2a02:169:4bce:2::50                 |
| raspberry | technitium | 192.168.5.201, 2a02:169:4bce:2::201                |
| framboesa | technitium | 192.168.5.202, 2a02:169:4bce:2::202                |
| framboesa | homebridge | 192.168.1.241                                      |
| datengrab | jellyfin   | 192.168.1.231                                      |

The networks behind them:

- raspberry: `main_macvlan` on `eth0` (192.168.1.0/24) and `servers_macvlan` on
  `eth0.5` (VLAN 5, 192.168.5.0/24). `hass` is on both.
- framboesa: `macvlan_1` on `eth0` and `macvlan_5` on `eth0.5`.
- datengrab: `media-macvlan` on `bond0` (192.168.1.0/24).

A Docker host cannot reach its own macvlan containers, so test these addresses
from another machine. That is why each Pi's `technitium` is checked by the other
Pi's Gatus, and why VictoriaMetrics on raspberry scrapes `jellyfin` directly.
