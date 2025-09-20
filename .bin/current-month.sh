#!/bin/bash
set -euo pipefail

current_month=$(date +%m)
month=$((10#$current_month))

printf 'Running for month %d...\n' "$month"
go run ./cmd/cli/main.go --month "$month"
