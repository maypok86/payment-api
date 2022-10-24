#!/bin/bash

set -e

find . -name '*.go' -not -wholename './vendor/*' -not -wholename '*.gen.go' |
while read -r file
do
  golines -m 120 -w "$file"
  gci -w "$file"
  gofumpt -w "$file"
done