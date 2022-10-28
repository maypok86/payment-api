#!/bin/bash

set -e

readonly project=$1

GOOSE_VERSION=3.6.0 docker-compose \
  -f deployments/docker-compose.test.yml -p "$project" --env-file .env.test up --build -d

test_status_code=0

GOOSE_VERSION=3.6.0 docker-compose \
  -f deployments/docker-compose.test.yml -p "$project" --env-file .env.test run integration go test -v ./tests/integration || test_status_code=$?
GOOSE_VERSION=3.6.0 docker-compose \
  -f deployments/docker-compose.test.yml -p "$project" --env-file .env.test logs > integration.log
GOOSE_VERSION=3.6.0 docker-compose \
  -f deployments/docker-compose.test.yml -p "$project" --env-file .env.test down

exit $test_status_code

GOOSE_VERSION=3.6.0 docker-compose \
  -f deployments/docker-compose.test.yml -p "$project" --env-file .env.test up --build --exit-code-from integration