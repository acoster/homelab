<!-- Generated from docs/services.md.tmpl by tools/docgen. Do not edit. -->

# Services

Every service in the compose files, by host. Hostnames are the names in the
service's Traefik `Host()` rules; a `${NAME}` is a value that lives in the
stack's encrypted secrets.

## cloudy

| Stack | Service | Image | Tag | Hostnames |
| ----- | ------- | ----- | --- | --------- |
| doco-cd | doco-cd | `ghcr.io/kimdre/doco-cd` | 0.133.2 | `${WEBHOOK_HOST}` |
| infra | socket-proxy | `ghcr.io/tecnativa/docker-socket-proxy` | v0.5.0 |  |
| infra | traefik | `traefik` | v3.7.18 |  |

## datengrab

| Stack | Service | Image | Tag | Hostnames |
| ----- | ------- | ----- | --- | --------- |
| apps | bambu-slicer-api | `ghcr.io/maziggy/bambu-studio-api` | bambuddy-1.2.5.7 |  |
| apps | bricktracker | `gitea.baerentsen.space/frederikbaerentsen/bricktracker` | 1.4.2 | `bricktracker.h.coster.ch` |
| apps | gyroidvault | `ghcr.io/teecodedev/gyroidvault` | 2.3.1 | `gyroidvault.h.coster.ch` |
| autorestic | autorestic | `cupcakearmy/autorestic` | latest |  |
| doco-cd | doco-cd | `ghcr.io/kimdre/doco-cd` | 0.133.2 | `${WEBHOOK_HOST}.h.coster.ch` |
| infra | traefik-kop | `ghcr.io/jittering/traefik-kop` | 0.21.0 |  |
| media | jellyfin | `jellyfin/jellyfin` | 12.4 | `jellyfin.h.coster.ch` |
| media | lidarr | `lscr.io/linuxserver/lidarr` | version-3.1.6.5079 | `lidarr.h.coster.ch` |
| media | prowlarr | `lscr.io/linuxserver/prowlarr` | 2.7.0 | `prowlarr.h.coster.ch` |
| media | radarr | `lscr.io/linuxserver/radarr` | 6.4.4 | `radarr.h.coster.ch` |
| media | sabnzbd | `lscr.io/linuxserver/sabnzbd` | 5.1.3 | `sabnzbd.h.coster.ch` |
| media | sonarr | `lscr.io/linuxserver/sonarr` | 4.0.20 | `sonarr.h.coster.ch` |
| media | transmission | `lscr.io/linuxserver/transmission` | 4.1.4 | `transmission.h.coster.ch` |
| monitoring | alloy | `grafana/alloy` | v1.22.0 | `alloy-datengrab.h.coster.ch` |
| monitoring | cadvisor | `ghcr.io/google/cadvisor` | v0.60.6 |  |
| paperless | broker | `redis` | 8.10.3 |  |
| paperless | db | `mariadb` | 13.0.2 |  |
| paperless | paperless-gpt | `icereed/paperless-gpt` | v0.31.0 | `paperless-gpt.h.coster.ch` |
| paperless | server | `ghcr.io/paperless-ngx/paperless-ngx` | 3.4.2 | `paperless.h.coster.ch` |

## framboesa

| Stack | Service | Image | Tag | Hostnames |
| ----- | ------- | ----- | --- | --------- |
| apps | bytestash | `ghcr.io/jordan-dalby/bytestash` | 1.6.2 | `bytestash.h.coster.ch` |
| apps | homebridge | `homebridge/homebridge` | 2026-11-10 | `homebridge.h.coster.ch` |
| apps | octoeverywhere-bambu-connect | `octoeverywhere/octoeverywhere` | 5.4.1 |  |
| apps | retroassembly | `arianrhodsandlot/retroassembly` | v7.261011.1345 | `retroassembly.h.coster.ch` |
| apps | spoolman | `ghcr.io/donkie/spoolman` | 0.27.1 | `spoolman.h.coster.ch` |
| autorestic | autorestic | `cupcakearmy/autorestic` | latest |  |
| doco-cd | doco-cd | `ghcr.io/kimdre/doco-cd` | 0.133.2 | `${WEBHOOK_HOST}.h.coster.ch` |
| infra | bambulab-relay | `fingerhutascode/udp-broadcast-relay` | latest |  |
| infra | technitium | `technitium/dns-server` | 15.6.2 |  |
| infra | traefik-kop | `ghcr.io/jittering/traefik-kop` | 0.21.0 |  |
| monitoring | alloy | `grafana/alloy` | v1.22.0 | `alloy-framboesa.h.coster.ch` |
| monitoring | cadvisor | `ghcr.io/google/cadvisor` | v0.60.6 |  |
| monitoring | gatus | `twinproduction/gatus` | v5.38.0 |  |

## raspberry

| Stack | Service | Image | Tag | Hostnames |
| ----- | ------- | ----- | --- | --------- |
| apps | bambuddy | `ghcr.io/maziggy/bambuddy` | 1.3.1 | `bambuddy.h.coster.ch` |
| apps | homeassistant | `ghcr.io/home-assistant/home-assistant` | 2026.11.1 | `hass.h.coster.ch` |
| apps | homebox | `ghcr.io/sysadminsmedia/homebox` | 0.27.1 | `homebox.h.coster.ch` |
| apps | linkding | `sissbruecker/linkding` | 1.47.0 | `linkding.h.coster.ch` |
| apps | matter-server | `ghcr.io/matter-js/matterjs-server` | 1.5.5 |  |
| apps | miniflux | `miniflux/miniflux` | 2.3.5 | `miniflux.h.coster.ch` |
| apps | miniflux-db | `postgres` | 18.7 |  |
| autorestic | autorestic | `cupcakearmy/autorestic` | latest |  |
| doco-cd | doco-cd | `ghcr.io/kimdre/doco-cd` | 0.133.2 | `${WEBHOOK_HOST}.h.coster.ch` |
| infra | authentik_db | `postgres` | 18.7 |  |
| infra | authentik_ldap | `ghcr.io/goauthentik/ldap` | 2026.8.5 |  |
| infra | authentik_server | `ghcr.io/goauthentik/server` | 2026.8.5 | `auth.h.coster.ch` |
| infra | authentik_worker | `ghcr.io/goauthentik/server` | 2026.8.5 |  |
| infra | cloudflare-ddns | `favonia/cloudflare-ddns` | 1.18.0 |  |
| infra | crowdsec | `crowdsecurity/crowdsec` | v1.8.2-debian |  |
| infra | redis | `redis` | 8.10.3 |  |
| infra | socket-proxy | `ghcr.io/tecnativa/docker-socket-proxy` | v0.5.0 |  |
| infra | technitium | `technitium/dns-server` | 15.6.2 | `technitium.h.coster.ch` |
| infra | traefik | `traefik` | v3.7.18 | `traefik.h.coster.ch` |
| monitoring | alloy | `grafana/alloy` | v1.22.0 | `alloy-raspberry.h.coster.ch` |
| monitoring | cadvisor | `ghcr.io/google/cadvisor` | v0.60.6 |  |
| monitoring | gatus | `twinproduction/gatus` | v5.38.0 | `gatus.h.coster.ch` |
| monitoring | grafana | `grafana/grafana` | 13.3.1 | `grafana.h.coster.ch` |
| monitoring | heartbeat | `curlimages/curl` | 8.23.0 |  |
| monitoring | jellyfin_exporter | `rebelcore/jellyfin-exporter` | v1.6.0 |  |
| monitoring | smartctl-exporter | `prometheuscommunity/smartctl-exporter` | v0.15.0 |  |
| monitoring | victorialogs | `victoriametrics/victoria-logs` | v1.53.0 | `vlogs.h.coster.ch` |
| monitoring | victoriametrics | `victoriametrics/victoria-metrics` | v1.156.0 | `vm.h.coster.ch` |

