#!/usr/bin/env bash
set -euo pipefail

# Full-stack checks: nginx static SPA + proxied API (same-origin session).

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVE="$REPO/scripts/assert_ok_envelope.py"

echo "[E2E] Running stack integration checks..."

BASE="${E2E_BASE_URL:-http://127.0.0.1:8080}"
E2E_TEST_USERNAME="${E2E_TEST_USERNAME:-${API_TEST_USERNAME:-}}"
E2E_TEST_PASSWORD="${E2E_TEST_PASSWORD:-${API_TEST_PASSWORD:-}}"

if [[ -z "$E2E_TEST_USERNAME" || -z "$E2E_TEST_PASSWORD" ]]; then
  echo "[E2E] Set E2E_TEST_USERNAME/E2E_TEST_PASSWORD (or API_TEST_USERNAME/API_TEST_PASSWORD) before running."
  exit 1
fi

echo "[E2E] GET $BASE/ (built app shell)"
HTML="$(curl -fsS "$BASE/")"
echo "$HTML" | grep -q 'id="app"' || {
  echo "[E2E] Expected Vue mount point #app in index response"
  exit 1
}
echo "$HTML" | grep -q 'PharmaOps' || {
  echo "[E2E] Expected PharmaOps title in index response"
  exit 1
}
if ! echo "$HTML" | grep -qE '/assets/index-[^/]+\.(js|mjs)' && ! echo "$HTML" | grep -q '/src/main.ts'; then
  echo "[E2E] Expected bundled asset or dev script reference in HTML"
  exit 1
fi

echo "[E2E] GET $BASE/dashboard (history mode → index.html)"
DASH="$(curl -fsS "$BASE/dashboard")"
echo "$DASH" | grep -q 'id="app"' || {
  echo "[E2E] Deep link should return SPA shell (try_files)"
  exit 1
}

echo "[E2E] GET $BASE/login"
LOGIN_PAGE="$(curl -fsS "$BASE/login")"
echo "$LOGIN_PAGE" | grep -q 'id="app"' || {
  echo "[E2E] Login route should return SPA shell"
  exit 1
}

echo "[E2E] Same-origin API via nginx: login + session"
LOGIN_JSON="$(curl -fsS -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$E2E_TEST_USERNAME\",\"password\":\"$E2E_TEST_PASSWORD\"}")"
echo "$LOGIN_JSON" | python3 "$ENVE"
TOKEN="$(echo "$LOGIN_JSON" | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["token"])')"
if [[ -z "$TOKEN" ]]; then
  echo "[E2E] Could not parse session token"
  exit 1
fi
ME_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/auth/me")"
echo "$ME_JSON" | python3 "$ENVE"

echo "[E2E] All checks passed."
