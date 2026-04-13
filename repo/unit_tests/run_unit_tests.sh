#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
echo "[UNIT] Running Go unit tests (apps/api)..."

(cd "$ROOT/apps/api" && go test ./...)
