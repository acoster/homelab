#!/bin/bash

export DOCKER_DATA_DIR=/var/lib/docker-data
export OUTPUT_DIR=${DOCKER_DATA_DIR}/backups

docker exec linkding python manage.py full_backup /etc/linkding/data/backups/backup.zip   
docker exec authentik_db pg_dump -U authentik authentik > ${OUTPUT_DIR}/authentik_backup.sql
docker exec miniflux-db pg_dump -U miniflux miniflux > ${OUTPUT_DIR}/miniflux.sql

sqlite3 ${DOCKER_DATA_DIR}/homebox/homebox.db ".backup '${OUTPUT_DIR}/homebox.bak'"
sqlite3 ${DOCKER_DATA_DIR}/manyfold/manyfold.sqlite3 ".backup '${OUTPUT_DIR}/manyfold.bak'"
sqlite3 ${DOCKER_DATA_DIR}/kuma/kuma.sqlite3 ".backup '${OUTPUT_DIR}/kuma.bak'"

dpkg --get-selections > ${OUTPUT_DIR}/installed_packages.txt

SCRIPT=$(readlink -f "${0}")
cd $(dirname "${SCRIPT}")
sops exec-env secrets.enc.env 'docker compose run --rm autorestic autorestic --ci cron'
