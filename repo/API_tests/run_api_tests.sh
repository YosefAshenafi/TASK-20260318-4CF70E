#!/usr/bin/env bash
set -euo pipefail

echo "[API] Running API smoke checks (Compose stack must be up)..."

BASE="${API_BASE_URL:-http://127.0.0.1:8080}"

echo "[API] GET $BASE/api/v1/health"
curl -fsS "$BASE/api/v1/health" | grep -q '"code":"OK"' || {
  echo "[API] Expected envelope code OK"
  exit 1
}

echo "[API] POST $BASE/api/v1/auth/login (dev admin)"
LOGIN_JSON="$(curl -fsS -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"password"}')"
echo "$LOGIN_JSON" | grep -q '"code":"OK"' || {
  echo "[API] Login expected OK"
  exit 1
}
TOKEN="$(echo "$LOGIN_JSON" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"
if [[ -z "$TOKEN" ]]; then
  echo "[API] Could not parse token from login response"
  exit 1
fi

echo "[API] GET $BASE/api/v1/auth/me (session + RBAC)"
ME_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/auth/me")"
echo "$ME_JSON" | grep -q '"code":"OK"' || {
  echo "[API] /auth/me expected OK"
  exit 1
}
echo "$ME_JSON" | grep -q 'system_admin' || {
  echo "[API] /auth/me expected system_admin role from dev seed"
  exit 1
}
echo "$ME_JSON" | grep -q 'system.full_access' || {
  echo "[API] /auth/me expected system.full_access permission"
  exit 1
}

echo "[API] Smoke checks passed."
