# homelab

Four machines, about sixty containers, one person, and a Git repository that
claims to be in charge of all of it. This is the configuration for my homelab.

It started as "I'll just run one container on the NAS". It is now a
production environment with a single user who is also the on-call engineer, the
change advisory board and the person who gets paged when the DNS stops working
during a video call. The SLA is whatever my family tolerates, which is less
than you would think.

I should know better. That is not a defence.

## What's in here

| Host        | What it is           | What it does                                                                |
| ----------- | -------------------- | --------------------------------------------------------------------------- |
| `raspberry` | A Raspberry Pi       | Edge proxy, auth, DNS, monitoring, and most of my questionable life choices |
| `framboesa` | Another Raspberry Pi | Second DNS, so that one Pi dying is merely an annoyance                     |
| `datengrab` | A Synology NAS       | Media, paperless, and everything that needs more disk than sense            |
| `cloudy`    | A VPS                | A VPN exit node, so that some traffic can leave from somewhere else         |

Every host runs [doco-cd](https://github.com/kimdre/doco-cd), which polls `main`
every 30 minutes and deploys whatever it finds. That is GitOps: I push a typo
to `main`, and a robot installs it across the house on a schedule. The
`.doco-cd.<host>.yaml` files say which stack goes where.

A full list of what runs, and on which ports, is generated from the compose
files, so it is at least wrong in a consistent way:

- [docs/generated/services.md](docs/generated/services.md): every service, by host
- [docs/generated/ports.md](docs/generated/ports.md): who is listening on what
- [docs/network.md](docs/network.md): addresses, written by hand and therefore already stale

## Rules of the house

- **Secrets** are encrypted with sops + age. To run docker compose commands,
  use `sops exec-env secrets.enc.env 'docker compose ...'`. If you commit a
  plaintext secret, gitleaks will find it, and then I will find out how it feels
  to rotate everything.
- **Images** are pinned by digest and kept current by Renovate. Branches are
  created daily, CI runs on every push, and passing digest updates are merged
  straight to `main` overnight, because nothing says "trust" like a bot with
  push access. Infra goes in a narrower window. Nobody enjoys a DNS restart
  in the middle of an evening.
- **Docs** under `docs/generated/` are rendered from the compose files by
  `tools/docgen`. Enable the pre-commit hook once per clone with
  `git config core.hooksPath .githooks`, or CI will remind you. CI is not shy.
- **Alerts** go to Pushover, from Gatus and from Grafana rules. If my phone
  buzzes at an inconvenient moment, that is the system working as designed,
  not as intended.
- **The keys** are somewhere safe. I have checked. More than once. Eventually.

## Backups

Backups use restic, driven by autorestic, and go to two places so that one
mistake is not enough. They report to healthchecks.io, which is how I find out
they have stopped.

A backup you have never restored is a hope. The restore drill has now been run,
it worked, and I am as surprised as you are. The runbook is in
[docs/restore.md](docs/restore.md). It covers reading from a backup, rebuilding
a blank host, and where the secrets live, and it is correct as of the last time
somebody tried it, which was recently enough to count.

### Integrity checks

Each host has `hosts/<host>/autorestic/check.sh`. It runs `restic check` against
every backend of that host, reading 1/10 of the data packs each run (rotating by
ISO week, so a weekly run verifies all data every 10 weeks). Set `CHECK_SUBSETS`
to change the number of slices. It pings a healthchecks.io check through
`HC_CHECK_URL` in the host's `autorestic/secrets.enc.env`.

The host crontab is not managed from this repo. Yes, in a repo about
declaratively managing everything. Add one line per host, on a weekday slot that
doesn't overlap the nightly backup, for example:

    0 11 * * 0  /opt/homelab/hosts/raspberry/autorestic/check.sh >> /var/log/restic-check.log 2>&1

Healthchecks.io settings: period 1 week, grace 1 day (a full check of the large
datengrab repository can take hours).

## Contributing

No. It is a single-user repo with a ruleset that consists of my own good
intentions. Take what is useful, and if you find something wrong, you are
probably right. I will be seeing it again at 3am anyway.
