#!/usr/bin/env bash
set -euo pipefail

echo "=== PharmaOps Test Runner ==="
echo "Starting docker dependencies..."
docker compose up -d --build db
docker compose ps

echo
echo "1/3 Unit tests"
bash unit_tests/run_unit_tests.sh

echo
echo "2/3 API tests"
bash API_tests/run_api_tests.sh

echo
echo "3/3 E2E tests"
bash e2e_tests/run_e2e_tests.sh

echo
echo "=== Final Summary ==="
echo "All test stages completed successfully."
