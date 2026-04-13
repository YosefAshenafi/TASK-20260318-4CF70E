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

echo "[API] GET $BASE/api/v1/recruitment/candidates (Step 4b recruitment)"
CAND_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/recruitment/candidates?page=1&pageSize=10")"
echo "$CAND_JSON" | grep -q '"code":"OK"' || {
  echo "[API] recruitment candidates expected OK"
  exit 1
}
echo "$CAND_JSON" | grep -q '"items"' || {
  echo "[API] recruitment list expected items array"
  exit 1
}

echo "[API] GET $BASE/api/v1/recruitment/positions"
POS_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/recruitment/positions?page=1&pageSize=10")"
echo "$POS_JSON" | grep -q '"code":"OK"' || {
  echo "[API] recruitment positions expected OK"
  exit 1
}

echo "[API] GET $BASE/api/v1/compliance/qualifications (Step 4b compliance)"
QUAL_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/compliance/qualifications?page=1&pageSize=10")"
echo "$QUAL_JSON" | grep -q '"code":"OK"' || {
  echo "[API] compliance qualifications expected OK"
  exit 1
}
echo "$QUAL_JSON" | grep -q '"items"' || {
  echo "[API] compliance qualifications expected items array"
  exit 1
}

echo "[API] GET $BASE/api/v1/compliance/restrictions"
REST_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/compliance/restrictions?page=1&pageSize=10")"
echo "$REST_JSON" | grep -q '"code":"OK"' || {
  echo "[API] compliance restrictions expected OK"
  exit 1
}

echo "[API] GET $BASE/api/v1/cases (Step 4b cases)"
CASES_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/cases?page=1&pageSize=10")"
echo "$CASES_JSON" | grep -q '"code":"OK"' || {
  echo "[API] cases list expected OK"
  exit 1
}
echo "$CASES_JSON" | grep -q '"items"' || {
  echo "[API] cases list expected items array"
  exit 1
}

echo "[API] GET $BASE/api/v1/audit/logs (Step 4b audit)"
AUDIT_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/audit/logs?page=1&pageSize=10")"
echo "$AUDIT_JSON" | grep -q '"code":"OK"' || {
  echo "[API] audit logs expected OK"
  exit 1
}

echo "[API] GET $BASE/api/v1/users (Step 4b RBAC)"
RBAC_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/users")"
echo "$RBAC_JSON" | grep -q '"code":"OK"' || {
  echo "[API] users list expected OK"
  exit 1
}

echo "[API] GET $BASE/api/v1/roles (primary personas from design migration 000013)"
ROLES_LIST_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/roles")"
echo "$ROLES_LIST_JSON" | grep -q '"code":"OK"' || {
  echo "[API] roles list expected OK"
  exit 1
}
echo "$ROLES_LIST_JSON" | grep -q 'business_specialist' || {
  echo "[API] roles list expected seeded business_specialist"
  exit 1
}

echo "[API] GET $BASE/api/v1/users/{id} (RBAC user detail)"
USER_DETAIL_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/users/00000000-0000-4000-8000-000000000001")"
echo "$USER_DETAIL_JSON" | grep -q '"code":"OK"' || {
  echo "[API] user detail expected OK"
  exit 1
}
echo "$USER_DETAIL_JSON" | grep -q '"roleIds"' || {
  echo "[API] user detail expected roleIds"
  exit 1
}

echo "[API] GET $BASE/api/v1/files (file objects list)"
FILES_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/files?page=1&pageSize=10")"
echo "$FILES_JSON" | grep -q '"code":"OK"' || {
  echo "[API] files list expected OK"
  exit 1
}
echo "$FILES_JSON" | grep -q '"items"' || {
  echo "[API] files list expected items array"
  exit 1
}

echo "[API] Smoke checks passed."
