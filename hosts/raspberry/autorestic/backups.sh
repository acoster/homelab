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
  exec "${REPO_DIR}/hosts/raspberry/autorestic/backups.sh" "$@"
fi

docker exec linkding python manage.py full_backup /etc/linkding/data/backups/backup.zip
docker exec authentik_db pg_dump -U authentik authentik > ${OUTPUT_DIR}/authentik_backup.sql
docker exec miniflux-db pg_dump -U miniflux miniflux > ${OUTPUT_DIR}/miniflux.sql

sqlite3 ${DOCKER_DATA_DIR}/homebox/homebox.db ".backup '${OUTPUT_DIR}/homebox.bak'"
sqlite3 ${DOCKER_DATA_DIR}/grafana-data/grafana.db ".backup '${OUTPUT_DIR}/grafana.bak'"
sqlite3 /opt/crowdsec-db/crowdsec.db ".backup '${OUTPUT_DIR}/crowdsec.bak'"

dpkg --get-selections > ${OUTPUT_DIR}/installed_packages.txt

cd ${REPO_DIR}/hosts/raspberry/autorestic
sops exec-env secrets.enc.env 'docker compose run --rm -T --name autorestic-cron autorestic autorestic --ci cron'
