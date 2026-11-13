<!-- Generated from docs/ports.md.tmpl by tools/docgen. Do not edit. -->

# Published ports

Ports that containers publish on their host, taken from the compose files in
`hosts/`. `all` means the port is bound on every address of the host.

| Host | Address | Port | Container port | Container | Stack |
| ---- | ------- | ---- | -------------- | --------- | ----- |
| cloudy | all | 80/tcp | 80 | traefik | infra |
| cloudy | all | 80/udp | 80 | traefik | infra |
| cloudy | all | 443/tcp | 443 | traefik | infra |
| cloudy | all | 443/udp | 443 | traefik | infra |
| datengrab | `192.168.1.14` | 3003/tcp | 3000 | bambu-slicer-api | apps |
| datengrab | `192.168.1.14` | 3128/tcp | 8080 | sabnzbd | media |
| datengrab | `192.168.1.14` | 3333/tcp | 3333 | bricktracker | apps |
| datengrab | `192.168.1.14` | 4000/tcp | 3000 | gyroidvault | apps |
| datengrab | `192.168.1.14` | 7878/tcp | 7878 | radarr | media |
| datengrab | `192.168.1.14` | 8000/tcp | 8000 | paperless-server | paperless |
| datengrab | `192.168.1.14` | 8080/tcp | 8080 | paperless-gpt | paperless |
| datengrab | `192.168.1.14` | 8686/tcp | 8686 | lidarr | media |
| datengrab | `192.168.1.14` | 8787/tcp | 8080 | cadvisor | monitoring |
| datengrab | `192.168.1.14` | 8989/tcp | 8989 | sonarr | media |
| datengrab | `192.168.1.14` | 9000/tcp | 80 | doco-cd | doco-cd |
| datengrab | `192.168.1.14` | 9091/tcp | 9091 | transmission | media |
| datengrab | `192.168.1.14` | 9120/tcp | 9120 | doco-cd | doco-cd |
| datengrab | `192.168.1.14` | 9696/tcp | 9696 | prowlarr | media |
| datengrab | `192.168.1.14` | 12345/tcp | 12345 | alloy | monitoring |
| datengrab | `192.168.1.14` | 51413/tcp | 51413 | transmission | media |
| datengrab | `192.168.1.14` | 51413/udp | 51413 | transmission | media |
| framboesa | `192.168.1.39` | 3128/tcp | 8000 | retroassembly | apps |
| framboesa | `192.168.1.39` | 5000/tcp | 5000 | bytestash | apps |
| framboesa | `192.168.1.39` | 8000/tcp | 8000 | spoolman | apps |
| framboesa | `192.168.1.39` | 8080/tcp | 8080 | gatus | monitoring |
| framboesa | `192.168.1.39` | 8081/tcp | 8080 | cadvisor | monitoring |
| framboesa | `192.168.1.39` | 9000/tcp | 80 | doco-cd | doco-cd |
| framboesa | `192.168.1.39` | 9120/tcp | 9120 | doco-cd | doco-cd |
| framboesa | `192.168.1.39` | 12345/tcp | 12345 | alloy | monitoring |
| raspberry | `192.168.1.212` | 80/tcp | 80 | traefik | infra |
| raspberry | `192.168.1.212` | 389/tcp | 3389 | authentik_ldap | infra |
| raspberry | `192.168.1.212` | 443/tcp | 443 | traefik | infra |
| raspberry | `192.168.1.212` | 443/udp | 443 | traefik | infra |
| raspberry | `192.168.1.212` | 514/tcp | 514 | victorialogs | monitoring |
| raspberry | `192.168.1.212` | 514/udp | 514 | victorialogs | monitoring |
| raspberry | `192.168.1.212` | 636/tcp | 6636 | authentik_ldap | infra |
| raspberry | `192.168.1.212` | 6379/tcp | 6379 | redis | infra |
| raspberry | `127.0.0.1` | 8989/tcp | 8080 | crowdsec | infra |
| raspberry | `192.168.1.212` | 9428/tcp | 9428 | victorialogs | monitoring |

