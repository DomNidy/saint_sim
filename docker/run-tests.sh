#!/usr/bin/env sh

set -eu

printf 'Go: '
go version
printf 'SimC binary: %s\n' "$SIMC_BINARY_PATH"

set +e
simc_output="$("$SIMC_BINARY_PATH" 2>&1)"
simc_status=$?
set -e

if [ "$simc_status" -ne 0 ] && [ "$simc_status" -ne 50 ]; then
    printf '%s\n' "$simc_output"
    exit "$simc_status"
fi

printf '%s\n' "$simc_output"

cd /src/apps/api
go test ./...

cd /src/apps/simulation_worker
go test ./...
