#!/bin/bash

export DOCKER_DATA_DIR=/var/lib/docker-data
export OUTPUT_DIR=${DOCKER_DATA_DIR}/backups

curl -s -X POST http://127.0.0.1:8000/api/v1/backup 2>&1 > /dev/null
dpkg --get-selections > ${OUTPUT_DIR}/installed_packages.txt

sqlite3 ${DOCKER_DATA_DIR}/bytestash/snippets.db ".backup ${OUTPUT_DIR}/bytestash.bak"

SCRIPT=$(readlink -f "${0}")
cd $(dirname "${SCRIPT}")

sops exec-env secrets.enc.env 'docker compose run --rm autorestic autorestic  -v --ci cron'
