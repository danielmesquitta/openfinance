#!/bin/bash
set -euo pipefail

start_date=$(date -d "7 days ago" --iso-8601)
end_date=$(date --iso-8601)

printf 'Running for last 7 days (%s to %s)...\n' "$start_date" "$end_date"
go run ./cmd/cli/main.go --start-date "$start_date" --end-date "$end_date"
