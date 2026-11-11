#!/bin/bash
set -euo pipefail

export REPO_DIR=/opt/homelab

# Sync the repo before doing anything else, then re-exec the fresh copy of
# this script (see the note in the backup script).
if [[ "${HOMELAB_REPO_SYNCED:-0}" != "1" ]]; then
  git -C "${REPO_DIR}" fetch origin main --prune
  git -C "${REPO_DIR}" reset --hard origin/main
  git -C "${REPO_DIR}" clean -fd -e "*.lock" -e "*autorestic.lock"
  export HOMELAB_REPO_SYNCED=1
  exec "${REPO_DIR}/hosts/framboesa/autorestic/check.sh" "$@"
fi

# Every run verifies the repository structure (snapshots, trees, pack index) and
# reads 1/N of the data packs, rotating by ISO week. Run weekly, this verifies
# all data every N weeks without downloading the whole repository each time.
SUBSETS="${CHECK_SUBSETS:-10}"
SUBSET="$((10#$(date +%V) % SUBSETS + 1))/${SUBSETS}"
export SUBSET

cd "${REPO_DIR}/hosts/framboesa/autorestic"
sops exec-env secrets.enc.env '
  ping() { curl -fsS -m 10 --retry 3 "${HC_CHECK_URL:?}${1:-}" > /dev/null; }
  ping /start
  if docker compose run --rm -T --name autorestic-check autorestic \
    autorestic --ci exec -va -- check --read-data-subset="${SUBSET}"; then
    ping
  else
    ping /fail
    exit 1
  fi
'
