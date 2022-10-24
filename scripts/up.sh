#!/bin/bash

set -e

readonly git_hash=$(git log --format="%h" -n 1)
readonly date=$(date -u +%Y-%m-%dT%H:%M:%S)
readonly ldflags="-X 'main.version=$git_hash' -X 'main.buildDate=$date'"
readonly project=$1

GOOSE_VERSION=3.6.0 LDFLAGS=$ldflags docker-compose \
  -f deployments/docker-compose.yml -p "$project" --env-file .env up -d --build