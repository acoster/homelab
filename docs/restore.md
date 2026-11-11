# Restore runbook

Written from the repo and the current setup. It has not been tested in a real
restore yet: see issue #50 for the first drill.

## 1. What runs where

| Host      | Address       | Role                                                                                                     |
| --------- | ------------- | -------------------------------------------------------------------------------------------------------- |
| raspberry | 192.168.1.212 | Edge and main dev host: Traefik, Authentik, CrowdSec, Technitium (DNS), Redis, apps, monitoring, doco-cd |
| framboesa | 192.168.5.202 | Second DNS, bytestash, homebridge, retroassembly, spoolman, bambu-connect, Gatus                         |
| datengrab | 192.168.1.14  | Synology NAS (DSM): media stack, paperless, bricktracker, gyroidvault, kop                               |
| cloudy    | VPS           | VPN exit node (Traefik and doco-cd only)                                                                 |

Deployment is GitOps: each host runs doco-cd, which polls `main` every 30
minutes and applies the stacks listed in `.doco-cd.<host>.yaml`. Secrets are
`*.enc.env` files (sops + age). `$DOCKER_DATA_DIR` is `/var/lib/docker-data` on
raspberry and framboesa, `/volume1/docker` on datengrab.

### Macvlan addresses

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

## 2. Keys you need first

Everything else depends on these. Check that you have an offline copy of each.

1. **The admin age key.** It decrypts every `*.enc.env` in the repo. It is the
   root of recovery. An offline copy is in 1Password.
2. **Host age keys.** Each host's `hosts/<host>/doco-cd/secrets.enc.env` contains
   `SOPS_AGE_KEY`, that host's own key. doco-cd uses it to decrypt every other
   secret and deploy the stacks, so a rebuild only needs the admin key once, to
   start doco-cd. The same key is also kept as a file for host scripts that call
   `sops exec-env` (backups, checks, manual compose commands):
   `/var/lib/docker-data/age_key` on raspberry and framboesa, and
   `/var/services/homes/coster/.config/sops/age/keys.txt` on datengrab. The Pi key
   files are inside the data backup, which you can only open with a key first. The
   datengrab file is not backed up. If a host key is lost, make a new key, put its
   public key in `.sops.yaml`, and run `sops updatekeys` on that host's files using
   the admin key.
3. **Restic repository passwords** are in each host's `autorestic/secrets.enc.env`
   (`AUTORESTIC_<BACKEND>_RESTIC_PASSWORD`), so you need key 1 or a surviving host
   key before you can open any backup.
4. **The rsync.net / datengrab SSH key and SSH config** (aliases `rsyncnet` and
   `datengrab`) live in `$DOCKER_DATA_DIR/restic-ssh` on each host, which is inside
   the backup it unlocks. An emergency SSH key is also kept in 1Password and is known
   to work with both rsync.net and datengrab. The SSH config is stored separately in
   a safe place too, and is short enough to rewrite by hand if needed.
5. **`GIT_AUTH_TOKEN` and `WEBHOOK_TOKEN`** for doco-cd are in
   `hosts/<host>/doco-cd/secrets.enc.env`.

## 3. Where the backups are

| Host      | Repositories                                                       | Nightly run (cron in `autorestic.yml`) |
| --------- | ------------------------------------------------------------------ | -------------------------------------- |
| framboesa | `rsyncnet:framboesa-backup`, `datengrab:pi_backups/framboesa`      | 01:00                                  |
| raspberry | `rsyncnet:raspberry-backup`, `datengrab:pi_backups/raspberry`      | 03:00                                  |
| datengrab | `rsyncnet:datengrab-backup` only (one offsite copy, no second one) | 05:00                                  |

Retention: 7 daily, 4 weekly, 6 monthly. Each run first exports database dumps
(section 6), then runs autorestic. Weekly `check.sh` runs verify repository
integrity; see the README.

Paths inside a snapshot: `/backup_source/bind_mounts/<dir>` is
`$DOCKER_DATA_DIR/<dir>` on raspberry and framboesa. On datengrab the folders are
mounted directly, for example `/backup_source/backups` and `/backup_source/paperless`.
Host files are under `/backup_source/host/...`.

## 4. Reading from a backup

On a machine with restic and the SSH config:

    export RESTIC_REPOSITORY=sftp:rsyncnet:raspberry-backup
    export RESTIC_PASSWORD=...        # from the sops-decrypted secrets
    restic snapshots
    restic ls latest /backup_source/bind_mounts/backups
    restic restore latest --target /restore \
      --include /backup_source/bind_mounts/grafana-data

On a working host, use the autorestic container instead:

    cd /opt/homelab/hosts/raspberry/autorestic
    sops exec-env secrets.enc.env \
      'docker compose run --rm -T autorestic autorestic exec -b rsyncnet -- ls latest'

## 5. Rebuilding a blank host

1.  Install the OS and packages. The Pis are set up by hand, so there is no
    automation to rely on. The `installed_packages.txt` file in the backups folder
    lists the `dpkg` selections.
2.  Install docker, git, sops, age and sqlite3.
3.  Get the admin age key from 1Password. You only need it to decrypt the host's
    `doco-cd/secrets.enc.env` below. Afterwards, recreate the host's key file at
    its path from section 2 (the value of `SOPS_AGE_KEY` in those secrets), and on
    the Pis point `SOPS_AGE_KEY_FILE` at it or copy it to `~/.config/sops/age/keys.txt`.
4.  Clone the repo to `/opt/homelab`.
5.  Restore `restic-ssh`, then restore the data you need into `$DOCKER_DATA_DIR`
    (section 4). Skip dumps for now.
6.  Start doco-cd from `hosts/<host>/doco-cd`, decrypting with the admin key:

         cd /opt/homelab/hosts/<host>/doco-cd
         SOPS_AGE_KEY_FILE=<admin key file> sops exec-env secrets.enc.env \
           'docker compose --all-resources up -d'

    On raspberry this stack also creates the shared `servers` network, so nothing
    else has to be started first. `--all-resources` is a global Compose option, so
    it goes before `up` (after it, Compose says `unknown flag`). It makes Compose
    create every network the stack declares, even ones no service uses. The
    `servers` network is used by doco-cd, so it would be created anyway; the flag
    is a safeguard if more networks are added to this stack.

    Because doco-cd carries `SOPS_AGE_KEY` and `GIT_AUTH_TOKEN`, it takes over from
    here. On startup it pulls the latest `main` and updates the containers of every
    stack in `.doco-cd.<host>.yaml`, so you don't wait for the 30-minute poll.
    Restarting doco-cd triggers the same run at any time.

7.  Once the stacks are up, restore the databases (section 6). The dump restores
    need the database containers running, so for each service stop the app, load
    the dump or copy the `.bak` into place, then start the app again.
8.  Verify: Gatus is green, healthchecks.io is receiving pings, you can log in
    through Authentik, and DNS resolves on both DNS hosts.

## 6. Databases

Dumps are written to `$DOCKER_DATA_DIR/backups` (datengrab: `/volume1/docker/backups`)
before every backup. Stop the service, restore, then start it.

| Service                          | File in snapshot                   | Restore                                                                                          |
| -------------------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------ |
| authentik                        | `authentik_backup.sql` (raspberry) | start only `authentik_db`, then `docker exec -i authentik_db psql -U authentik authentik < file` |
| miniflux                         | `miniflux.sql` (raspberry)         | same with `miniflux-db`, user and database `miniflux`                                            |
| paperless                        | `paperless.sql` (datengrab)        | `docker compose exec -T db sh -c 'mariadb -u paperless -p"$MARIADB_PASSWORD" paperless' < file`  |
| homebox                          | `homebox.bak` (raspberry)          | copy to `homebox/homebox.db`                                                                     |
| grafana                          | `grafana.bak` (raspberry)          | copy to `grafana-data/grafana.db`, owner `1000:1000`                                             |
| bytestash                        | `bytestash.bak` (framboesa)        | copy to `bytestash/snippets.db`                                                                  |
| retroassembly                    | `retroassembly.bak` (framboesa)    | copy to `retroassembly/retroassembly.sqlite`                                                     |
| bricktracker                     | `bricktracker.bak` (datengrab)     | copy to `bricktracker/data/app.db`                                                               |
| gyroidvault                      | `gyroidvault.bak` (datengrab)      | copy to `gyroidvault/gyroidvault.db`                                                             |
| paperless-gpt                    | `paperless-gpt.bak` (datengrab)    | copy to `paperless-gpt/db/modification_history.db`                                               |
| jellyfin                         | `jellyfin.bak` (datengrab)         | copy to `jellyfin/config/data/jellyfin.db`                                                       |
| sonarr, radarr, prowlarr, lidarr | `<app>.bak` (datengrab)            | copy to `<app>/<app>.db`                                                                         |

For every `.bak`, copy it over the live database file with the container stopped,
set the owner from the table below, then start the container. These are the owners
(`uid:gid`) of the live files, read with `ls -ln`:

| Service                                 | Owner           |
| --------------------------------------- | --------------- |
| homebox                                 | `0:0`           |
| bytestash                               | `0:0`           |
| retroassembly                           | `0:0`           |
| bricktracker                            | `0:0`           |
| gyroidvault                             | `0:0`           |
| grafana                                 | `1000:1000`     |
| paperless-gpt                           | `10001:10001`   |
| jellyfin                                | `297536:297536` |
| sonarr, radarr, prowlarr, lidarr        | `1028:100`      |
| linkding (the whole `linkding/` folder) | `33:33`         |

Files copied as root already end up as `0:0`. For the others, run
`chown <owner> <file>` after copying (for linkding, `chown -R 33:33 linkding/`
after extracting the zip). The datengrab IDs (`1028:100`, `297536`) are Synology's,
so check them again if you ever rebuild DSM.

Services with their own backups, not covered by the dumps above:

- **linkding:** `linkding/backups/backup.zip`, made by `full_backup` (database,
  assets, favicons and previews). To restore, stop linkding, move the existing
  `linkding/` folder aside, extract the zip into a new `linkding/` folder, and start
  it again. The folder is mounted at `/etc/linkding/data`, which is what linkding's
  [backup docs](https://linkding.link/backups/) describe.
- **Home Assistant:** backups are in the `backups` folder of the config directory,
  so `hass/config/backups`, which is inside the file backup. In the UI, go to
  Settings > System > Backups and restore the file, or choose "Upload backup" on a
  fresh install's welcome screen. You need the backup's encryption key from its
  emergency kit. See the [HA docs](https://www.home-assistant.io/common-tasks/general/backups/#restoring-a-backup).
  Those docs are written for Home Assistant OS and don't cover a plain Container
  install, so if the UI restore doesn't work, fall back to restoring the `hass/config`
  folder from restic. Only the recorder history (`home-assistant_v2.db`) is lost.
  The encryption key is kept in 1Password.
- **Bambuddy:** a scheduled backup writes a ZIP (database, print archives and
  library files) into one of Bambuddy's bind mounts (`bambuddy/backups` is mounted
  for this). To restore, open Settings > Backup & Restore > Restore Backup, pick
  the ZIP, wait for it to finish and restart Bambuddy when asked. Restore into the
  same Bambuddy version or a newer one, never an older one. See the
  [Bambuddy docs](https://wiki.bambuddy.cool/features/backup/#full-restore).
- **spoolman:** its own backups folder under `spoolman/`. Restore by replacing `spoolman.db` per its docs.
- **Synology settings:** `datengrab.dss`, created by `synoconfbkp`. Restore through DSM's configuration restore.

## 7. Not backed up (and what to do instead)

- **CrowdSec database:** regenerates. Re-add the Traefik bouncer with
  `cscli bouncers add <name> -k "$CROWDSEC_LAPI_KEY"`. The config in `/opt/crowdsec`
  is backed up; the `hub/` folder is re-downloaded at startup.
- **VictoriaMetrics, VictoriaLogs, Redis cache, Traefik logs:** history only.
- **Gatus history** on both Pis: history only. The config is in git.
- **Raw Postgres and MariaDB data folders:** the dumps replace them.
- **Certificates:** `traefik/acme.json` is gitignored but is inside the data
  backup. If it is lost, Traefik re-issues through the Cloudflare DNS challenge.

## 8. Practising

No restore drill has been done yet (issue #50 tracks the first one). After that,
repeat it at least once a year: restore one small service (for example bytestash)
from `latest` into a scratch directory and confirm it opens. Record the date of
each drill here.
