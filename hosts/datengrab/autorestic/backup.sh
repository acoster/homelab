#!/bin/bash
set -euo pipefail

export REPO_DIR=/opt/homelab
export DOCKER_DATA_DIR=/volume1/docker
export OUTPUT_ROOT=/volume1/docker/backups
export PATH=$PATH:/var/services/homes/coster/bin

echo "Exporting synology config"
sudo rm ${OUTPUT_ROOT}/datengrab.dss
sudo /usr/syno/bin/synoconfbkp export --filepath=${OUTPUT_ROOT}/datengrab.dss

echo "Exporting bricktracker DB"
sqlite3 ${DOCKER_DATA_DIR}/bricktracker/data/app.db ".backup ${OUTPUT_ROOT}/bricktracker.bak"

echo "Exporting Gyroidvault DB"
sqlite3 ${DOCKER_DATA_DIR}/gyroidvault/gyroidvault.db ".backup ${OUTPUT_ROOT}/gyroidvault.bak"

echo "Exporting paperless-gpt DB"
sqlite3 ${DOCKER_DATA_DIR}/paperless-gpt/db/modification_history.db ".backup ${OUTPUT_ROOT}/paperless-gpt.bak"

echo "Exporting Jelyfin DB"
sqlite ${DOCKER_DATA_DIR}/jellyfin/config/data/jellyfin.db ".backup ${OUTPUT_ROOT}/jellyfin.bak"

for i in sonarr radarr prowlarr lidarr; do
  echo "Exporting ${i} DB"
  sqlite3 ${DOCKER_DATA_DIR}/$i/$i.db ".backup /volume1/docker/backups/$i.bak"
done

echo "Exporting paperless DB"
cd ${REPO_DIR}/hosts/datengrab/paperless
sops exec-env secrets.enc.env 'docker compose exec db mariadb-dump -u paperless -ppaperless  paperless' > ${OUTPUT_ROOT}/paperless.sql

cd ${REPO_DIR}
git fetch origin main --prune
git reset --hard origin/main
git clean -fd -e "*.lock" -e "*autorestic.lock"

echo "Running autorestic"
cd ${REPO_DIR}/hosts/datengrab/autorestic
sops exec-env secrets.enc.env 'docker compose run --rm autorestic autorestic  --ci cron'
