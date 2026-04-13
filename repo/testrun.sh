#!/usr/bin/env bash
# Full release verification: clean volumes, rebuild, migrate, integrated tests, timestamped log under reports/.
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO"

TS="$(date -u +"%Y%m%d-%H%M%S")"
START_ISO="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_DIR="$REPO/reports"
mkdir -p "$REPORT_DIR"
REPORT_FILE="$REPORT_DIR/release-${TS}.log"

run_gate() {
  echo "=== PharmaOps testrun.sh (release verification) ==="
  echo "Report file: $REPORT_FILE"
  echo "Started (UTC): $START_ISO"
  echo
  echo "Stopping stack and removing volumes (clean MySQL data)..."
  docker compose down -v

  echo "Building and starting stack..."
  docker compose up -d --build
  docker compose ps

  echo
  echo "DB: apply migrations"
  bash scripts/db_migrate.sh

  bash scripts/run_integrated_tests.sh

  echo
  echo "--- docker compose ps (post-run) ---"
  docker compose ps

  END_ISO="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  echo
  echo "=== Release verification completed successfully ==="
  echo "Finished (UTC): $END_ISO"
}

set -o pipefail
run_gate 2>&1 | tee "$REPORT_FILE"
exit "${PIPESTATUS[0]}"
