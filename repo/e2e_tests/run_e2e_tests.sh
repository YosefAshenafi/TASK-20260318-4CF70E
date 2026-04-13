#!/usr/bin/env bash
set -euo pipefail

# End-to-end checks against the Dockerized stack (nginx → static web + proxied API).
# Requires: `docker compose up -d --build` and migrations applied (as in run_tests.sh).

echo "[E2E] Running stack integration checks..."

BASE="${E2E_BASE_URL:-http://127.0.0.1:8080}"

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
  -d '{"username":"admin","password":"password"}')"
echo "$LOGIN_JSON" | grep -q '"code":"OK"' || {
  echo "[E2E] Login envelope not OK"
  exit 1
}
TOKEN="$(echo "$LOGIN_JSON" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"
if [[ -z "$TOKEN" ]]; then
  echo "[E2E] Could not parse session token"
  exit 1
fi
ME_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/auth/me")"
echo "$ME_JSON" | grep -q '"code":"OK"' || {
  echo "[E2E] /auth/me not OK through nginx proxy"
  exit 1
}

echo "[E2E] All checks passed."
