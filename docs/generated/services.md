<!-- Generated from docs/services.md.tmpl by tools/docgen. Do not edit. -->

# Services

Every service in the compose files, by host. Hostnames are the names in the
service's Traefik `Host()` rules; a `${NAME}` is a value that lives in the
stack's encrypted secrets.

## cloudy

| Stack | Service | Image | Hostnames |
| ----- | ------- | ----- | --------- |
| doco-cd | doco-cd | `ghcr.io/kimdre/doco-cd` | `${WEBHOOK_HOST}` |
| infra | socket-proxy | `ghcr.io/tecnativa/docker-socket-proxy` |  |
| infra | traefik | `traefik` |  |

## datengrab

| Stack | Service | Image | Hostnames |
| ----- | ------- | ----- | --------- |
| apps | bambu-slicer-api | `ghcr.io/maziggy/bambu-studio-api` |  |
| apps | bricktracker | `gitea.baerentsen.space/frederikbaerentsen/bricktracker` | `bricktracker.h.coster.ch` |
| apps | gyroidvault | `ghcr.io/teecodedev/gyroidvault` | `gyroidvault.h.coster.ch` |
| autorestic | autorestic | `cupcakearmy/autorestic` |  |
| doco-cd | doco-cd | `ghcr.io/kimdre/doco-cd` | `${WEBHOOK_HOST}.h.coster.ch` |
| infra | traefik-kop | `ghcr.io/jittering/traefik-kop` |  |
| media | jellyfin | `jellyfin/jellyfin` | `jellyfin.h.coster.ch` |
| media | lidarr | `lscr.io/linuxserver/lidarr` | `lidarr.h.coster.ch` |
| media | prowlarr | `lscr.io/linuxserver/prowlarr` | `prowlarr.h.coster.ch` |
| media | radarr | `lscr.io/linuxserver/radarr` | `radarr.h.coster.ch` |
| media | sabnzbd | `lscr.io/linuxserver/sabnzbd` | `sabnzbd.h.coster.ch` |
| media | sonarr | `lscr.io/linuxserver/sonarr` | `sonarr.h.coster.ch` |
| media | transmission | `lscr.io/linuxserver/transmission` | `transmission.h.coster.ch` |
| monitoring | alloy | `grafana/alloy` | `alloy-datengrab.h.coster.ch` |
| monitoring | cadvisor | `ghcr.io/google/cadvisor` |  |
| paperless | broker | `redis` |  |
| paperless | db | `mariadb` |  |
| paperless | paperless-gpt | `icereed/paperless-gpt` | `paperless-gpt.h.coster.ch` |
| paperless | server | `ghcr.io/paperless-ngx/paperless-ngx` | `paperless.h.coster.ch` |

## framboesa

| Stack | Service | Image | Hostnames |
| ----- | ------- | ----- | --------- |
| apps | bytestash | `ghcr.io/jordan-dalby/bytestash` | `bytestash.h.coster.ch` |
| apps | homebridge | `homebridge/homebridge` | `homebridge.h.coster.ch` |
| apps | octoeverywhere-bambu-connect | `octoeverywhere/octoeverywhere` |  |
| apps | retroassembly | `arianrhodsandlot/retroassembly` | `retroassembly.h.coster.ch` |
| apps | spoolman | `ghcr.io/donkie/spoolman` | `spoolman.h.coster.ch` |
| autorestic | autorestic | `cupcakearmy/autorestic` |  |
| doco-cd | doco-cd | `ghcr.io/kimdre/doco-cd` | `${WEBHOOK_HOST}.h.coster.ch` |
| infra | bambulab-relay | `fingerhutascode/udp-broadcast-relay` |  |
| infra | technitium | `technitium/dns-server` |  |
| infra | traefik-kop | `ghcr.io/jittering/traefik-kop` |  |
| monitoring | alloy | `grafana/alloy` | `alloy-framboesa.h.coster.ch` |
| monitoring | cadvisor | `ghcr.io/google/cadvisor` |  |
| monitoring | gatus | `twinproduction/gatus` |  |

## raspberry

| Stack | Service | Image | Hostnames |
| ----- | ------- | ----- | --------- |
| apps | bambuddy | `ghcr.io/maziggy/bambuddy` | `bambuddy.h.coster.ch` |
| apps | homeassistant | `ghcr.io/home-assistant/home-assistant` | `hass.h.coster.ch` |
| apps | homebox | `ghcr.io/sysadminsmedia/homebox` | `homebox.h.coster.ch` |
| apps | linkding | `sissbruecker/linkding` | `linkding.h.coster.ch` |
| apps | matter-server | `ghcr.io/matter-js/matterjs-server` |  |
| apps | miniflux | `miniflux/miniflux` | `miniflux.h.coster.ch` |
| apps | miniflux-db | `postgres` |  |
| autorestic | autorestic | `cupcakearmy/autorestic` |  |
| doco-cd | doco-cd | `ghcr.io/kimdre/doco-cd` | `${WEBHOOK_HOST}.h.coster.ch` |
| infra | authentik_db | `postgres` |  |
| infra | authentik_ldap | `ghcr.io/goauthentik/ldap` |  |
| infra | authentik_server | `ghcr.io/goauthentik/server` | `auth.h.coster.ch` |
| infra | authentik_worker | `ghcr.io/goauthentik/server` |  |
| infra | cloudflare-ddns | `favonia/cloudflare-ddns` |  |
| infra | crowdsec | `crowdsecurity/crowdsec` |  |
| infra | redis | `redis` |  |
| infra | socket-proxy | `ghcr.io/tecnativa/docker-socket-proxy` |  |
| infra | technitium | `technitium/dns-server` | `technitium.h.coster.ch` |
| infra | traefik | `traefik` | `traefik.h.coster.ch` |
| monitoring | alloy | `grafana/alloy` | `alloy-raspberry.h.coster.ch` |
| monitoring | cadvisor | `ghcr.io/google/cadvisor` |  |
| monitoring | gatus | `twinproduction/gatus` | `gatus.h.coster.ch` |
| monitoring | grafana | `grafana/grafana` | `grafana.h.coster.ch` |
| monitoring | heartbeat | `curlimages/curl` |  |
| monitoring | jellyfin_exporter | `rebelcore/jellyfin-exporter` |  |
| monitoring | smartctl-exporter | `prometheuscommunity/smartctl-exporter` |  |
| monitoring | victorialogs | `victoriametrics/victoria-logs` | `vlogs.h.coster.ch` |
| monitoring | victoriametrics | `victoriametrics/victoria-metrics` | `vm.h.coster.ch` |

