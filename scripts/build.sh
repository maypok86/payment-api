#!/bin/bash

set -e

readonly git_hash=$(git log --format="%h" -n 1)
readonly date=$(date -u +%Y-%m-%dT%H:%M:%S)
readonly ldflags="-X 'main.version=$git_hash' -X 'main.buildDate=$date'"

go build -v -o "$1" -ldflags "$ldflags" "$2"