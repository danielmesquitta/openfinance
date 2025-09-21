#!/bin/bash
set -euo pipefail

start_date=$(date -d "15 days ago" --iso-8601)
end_date=$(date --iso-8601)

printf 'Running for last 15 days (%s to %s)...\n' "$start_date" "$end_date"
go run ./cmd/cli/main.go --start-date "$start_date" --end-date "$end_date"
