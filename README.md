# homelab

This is a collection of configs for my homelab. A few notes for my future self:

- Secrets are encrypted with sops + age. To run docker compose commands, use
  `sops exec-env secrets.enc.env 'docker compose ...'`
- Backups use restic. The keys are somewhere safe

## Backup integrity checks

Each host has `hosts/<host>/autorestic/check.sh`. It runs `restic check` against
every backend of that host, reading 1/10 of the data packs each run (rotating by
ISO week, so a weekly run verifies all data every 10 weeks). Set `CHECK_SUBSETS`
to change the number of slices. It pings a healthchecks.io check through
`HC_CHECK_URL` in the host's `autorestic/secrets.enc.env`.

The host crontab is not managed from this repo. Add one line per host, on a
weekday slot that doesn't overlap the nightly backup, for example:

    0 11 * * 0  /opt/homelab/hosts/raspberry/autorestic/check.sh >> /var/log/restic-check.log 2>&1

Healthchecks.io settings: period 1 week, grace 1 day (a full check of the large
datengrab repository can take hours).
