#!/usr/bin/env bash
# Run smoke checks, then unit → API → E2E stages. Expect cwd repo root; stack up; migrations applied.
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE="${TEST_BASE_URL:-http://127.0.0.1:8080}"
ENVE="$REPO/scripts/assert_ok_envelope.py"

echo "Smoke: web UI ($BASE/)"
for _ in $(seq 1 60); do
  if curl -fsS "$BASE/" >/dev/null 2>&1; then
    echo "Web responded OK."
    break
  fi
  sleep 1
done
curl -fsS "$BASE/" >/dev/null

echo "Smoke: API health (via nginx → api)"
for _ in $(seq 1 60); do
  if curl -fsS "$BASE/api/v1/health" >/dev/null 2>&1; then
    echo "API /api/v1/health responded OK."
    break
  fi
  sleep 1
done
curl -fsS "$BASE/api/v1/health" | python3 "$ENVE"

echo
echo "1/4 Unit tests"
bash "$REPO/unit_tests/run_unit_tests.sh"

echo
echo "2/4 API tests"
bash "$REPO/API_tests/run_api_tests.sh"

echo
echo "3/4 E2E tests"
bash "$REPO/e2e_tests/run_e2e_tests.sh"

echo
echo "4/4 Design conformance (docs/design.md)"
python3 "$REPO/scripts/design_conformance_verify.py"
