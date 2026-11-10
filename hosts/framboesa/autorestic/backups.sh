#!/bin/bash
set -euo pipefail

export REPO_DIR=/opt/homelab
export DOCKER_DATA_DIR=/var/lib/docker-data
export OUTPUT_DIR=${DOCKER_DATA_DIR}/backups

# Sync the repo before doing anything else, then re-exec the fresh copy of
# this script. Bash reads scripts incrementally, so resetting this file while
# it is running can make bash execute garbage. The whole `if` block is parsed
# before it runs, so the reset inside it is safe.
if [[ "${HOMELAB_REPO_SYNCED:-0}" != "1" ]]; then
  git -C "${REPO_DIR}" fetch origin main --prune
  git -C "${REPO_DIR}" reset --hard origin/main
  git -C "${REPO_DIR}" clean -fd -e "*.lock" -e "*autorestic.lock"
  export HOMELAB_REPO_SYNCED=1
  exec "${REPO_DIR}/hosts/framboesa/autorestic/backups.sh" "$@"
fi

curl -fsS -m 60 -X POST http://127.0.0.1:8000/api/v1/backup > /dev/null \
  || echo "WARNING: spoolman backup API call failed" >&2
dpkg --get-selections > ${OUTPUT_DIR}/installed_packages.txt

sqlite3 ${DOCKER_DATA_DIR}/bytestash/snippets.db ".backup ${OUTPUT_DIR}/bytestash.bak"


cd ${REPO_DIR}/hosts/framboesa/autorestic
sops exec-env secrets.enc.env 'docker compose run --rm -T --name autorestic-cron autorestic autorestic  -v --ci cron'
