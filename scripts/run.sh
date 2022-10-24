#!/bin/bash

set -e

readonly app="$1"

env $(cat .env | grep -Ev '^#' | xargs) "$app"