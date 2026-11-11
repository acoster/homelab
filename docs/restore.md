# Restore runbook (DRAFT)

Status: first draft written from the repo. Anything marked **TODO(alex)** is
something the repo can't tell me; fill it in or correct it. Nothing here has been
tested in an actual restore yet.

## 1. What runs where

| Host      | Address       | Role                                                                                                     |
| --------- | ------------- | -------------------------------------------------------------------------------------------------------- |
| raspberry | 192.168.1.212 | Edge and main dev host: Traefik, Authentik, CrowdSec, Technitium (DNS), Redis, apps, monitoring, doco-cd |
| framboesa | 192.168.5.202 | Second DNS, bytestash, homebridge, retroassembly, spoolman, bambu-connect, Gatus                         |
| datengrab | 192.168.1.14  | Synology NAS (DSM): media stack, paperless, bricktracker, gyroidvault, kop                               |
| cloudy    | VPS           | Traefik and doco-cd only. **TODO(alex): what does cloudy serve?**                                        |

Deployment is GitOps: each host runs doco-cd, which polls `main` every 30
minutes and applies the stacks listed in `.doco-cd.<host>.yaml`. Secrets are
`*.enc.env` files (sops + age). `$DOCKER_DATA_DIR` is `/var/lib/docker-data` on
raspberry and framboesa, `/volume1/docker` on datengrab.

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
   the backup it unlocks. An SSH key is also kept in 1Password, so recovery doesn't
   depend on that folder. **TODO(alex): confirm it is the key for both `rsyncnet`
   and `datengrab`, and that the SSH config (host aliases) is stored with it.**
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

1.  Install the OS and packages. **TODO(alex): how are the Pis provisioned?** The
    `installed_packages.txt` file in the backups folder lists the `dpkg` selections.
2.  Install docker, git, sops, age and sqlite3.
3.  Get the admin age key from 1Password. You only need it to decrypt the host's
    `doco-cd/secrets.enc.env` below. Afterwards, recreate the host's key file at
    its path from section 2 (the value of `SOPS_AGE_KEY` in those secrets), and on
    the Pis point `SOPS_AGE_KEY_FILE` at it or copy it to `~/.config/sops/age/keys.txt`.
4.  Clone the repo to `/opt/homelab`.
5.  Restore `restic-ssh`, then restore the data you need into `$DOCKER_DATA_DIR`
    (section 4). Skip dumps for now.
6.  **Raspberry only:** the `servers` network is created by the `infra` stack, but
    doco-cd and the apps mark it external. Start `infra` by hand once, using the
    admin key:

         cd /opt/homelab/hosts/raspberry/infra
         SOPS_AGE_KEY_FILE=<admin key file> sops exec-env secrets.enc.env 'docker compose up -d'

    The other hosts don't need this: framboesa and datengrab's doco-cd has no
    external network, and cloudy's creates its own.

7.  Load the database dumps (section 6) before starting the apps that use them.
8.  Start doco-cd from `hosts/<host>/doco-cd` the same way. Because it carries
    `SOPS_AGE_KEY` and `GIT_AUTH_TOKEN`, it takes over from here and deploys every
    stack in `.doco-cd.<host>.yaml` on its next poll (every 30 minutes).
    **TODO(alex): how do you trigger it sooner (webhook or restart)?**
9.  Verify: Gatus is green, healthchecks.io is receiving pings, you can log in
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

For every `.bak`, copy it over the live `.db` with the container stopped, make
the owner match the folder (`ls -ln` on the parent), then start the container.
**TODO(alex): confirm the owner for each, since only Grafana's is in the repo.**

Services with their own backups, not covered by the dumps above:

- **linkding:** `linkding/backups/backup.zip`, made by `full_backup`. **TODO(alex): confirm the restore steps.**
- **spoolman:** its own backups folder under `spoolman/`. Restore by replacing `spoolman.db` per its docs.
- **Home Assistant, bambuddy:** their own backups. **TODO(alex): where do they land and how do you restore them?**
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

Do a restore drill at least once a year: restore one small service (for example
bytestash) from `latest` into a scratch directory and confirm it opens.
**TODO(alex): add the date of the last drill here.**
