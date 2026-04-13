#!/usr/bin/env bash
set -euo pipefail

echo "=== PharmaOps Test Runner ==="
echo "Building and starting stack (Docker only)..."
docker compose up -d --build
docker compose ps

echo
echo "DB: apply migrations"
bash scripts/db_migrate.sh

echo
echo "Smoke: web UI (http://127.0.0.1:8080/)"
for _ in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:8080/" >/dev/null 2>&1; then
    echo "Web responded OK."
    break
  fi
  sleep 1
done
curl -fsS "http://127.0.0.1:8080/" >/dev/null

echo
echo "Smoke: API health (via nginx → api)"
for _ in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:8080/api/v1/health" >/dev/null 2>&1; then
    echo "API /api/v1/health responded OK."
    break
  fi
  sleep 1
done
curl -fsS "http://127.0.0.1:8080/api/v1/health" >/dev/null

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
