#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

go test -tags=integration -count=1 ./quickstart/agnes-chat
