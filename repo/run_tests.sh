#!/usr/bin/env bash
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO"

echo "=== PharmaOps Test Runner ==="
echo "Building and starting stack (Docker only)..."
docker compose up -d --build
docker compose ps

echo
echo "DB: apply migrations"
bash scripts/db_migrate.sh

bash scripts/run_integrated_tests.sh

echo
echo "=== Final Summary ==="
echo "All stages passed: unit, API, E2E, design conformance."
echo "For a clean-database release run with a saved log under reports/, use: bash testrun.sh"
