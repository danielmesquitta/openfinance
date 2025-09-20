#!/bin/bash
set -euo pipefail

current_month=$(date +%m)
end=$((10#$current_month))

for ((month=1; month<=end; month++)); do
  printf 'Running for month %d...\n' "$month"
  go run ./cmd/cli/main.go --month "$month"
done
