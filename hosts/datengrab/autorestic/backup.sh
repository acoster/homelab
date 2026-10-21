#!/bin/bash
OUTPUT_ROOT=/volume1/docker/backups
PATH=$PATH:/var/services/homes/coster/bin

echo "Exporting synology config"
sudo /usr/syno/bin/synoconfbkp export --filepath=${OUTPUT_ROOT}/datengrab.dss

echo "Exporting bricktracker DB"
sqlite3 /volume1/docker/bricktracker/data/app.db ".backup ${OUTPUT_ROOT}/bricktracker.bak"

echo "Exporting paperless-gpt DB"
sqlite3 /volume1/docker/paperless-gpt/db/modification_history.db ".backup ${OUTPUT_ROOT}/paperless-gpt.bak"

for i in sonarr radarr prowlarr lidarr; do
  echo "Exporting ${i} DB"
  sqlite3 /volume1/docker/$i/$i.db ".backup /volume1/docker/backups/$i.bak"
done

echo "Exporting paperless DB"
cd /volume1/homes/coster/docker/paperless
docker compose exec db mariadb-dump -u paperless -ppaperless  paperless > ${OUTPUT_ROOT}/paperless.sql

echo "Running autorestic"
cd /volume1/homes/coster/homelab/hosts/datengrab/autorestic
sops exec-env secrets.enc.env 'docker compose run --rm autorestic autorestic  --ci cron'
