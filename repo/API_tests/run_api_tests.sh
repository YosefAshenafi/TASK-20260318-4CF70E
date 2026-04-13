#!/usr/bin/env bash
set -euo pipefail

# Contract checks against the running Compose stack (nginx → api). Requires migrations + dev seed.

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVE="$REPO/scripts/assert_ok_envelope.py"

echo "[API] Running API contract checks (Compose stack must be up)..."

BASE="${API_BASE_URL:-http://127.0.0.1:8080}"

echo "[API] GET $BASE/api/v1/health"
curl -fsS "$BASE/api/v1/health" | python3 "$ENVE"

echo "[API] GET $BASE/api/v1/recruitment/candidates without Authorization (expect 401)"
CODE="$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/recruitment/candidates?page=1&pageSize=10")"
if [[ "$CODE" != "401" ]]; then
  echo "[API] expected HTTP 401 without bearer token, got $CODE"
  exit 1
fi

echo "[API] POST $BASE/api/v1/auth/login (dev admin)"
LOGIN_JSON="$(curl -fsS -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"password"}')"
echo "$LOGIN_JSON" | python3 "$ENVE"
TOKEN="$(echo "$LOGIN_JSON" | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["token"])')"
if [[ -z "$TOKEN" ]]; then
  echo "[API] Empty token after validated envelope"
  exit 1
fi

echo "[API] GET $BASE/api/v1/auth/me (session + RBAC)"
ME_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/auth/me")"
echo "$ME_JSON" | python3 "$ENVE"
echo "$ME_JSON" | python3 -c 'import json,sys; d=json.load(sys.stdin); data=d.get("data") or {}; assert "system_admin" in (data.get("roles") or []), data' || {
  echo "[API] /auth/me expected system_admin role from dev seed"
  exit 1
}
echo "$ME_JSON" | python3 -c 'import json,sys; d=json.load(sys.stdin); data=d.get("data") or {}; perms=data.get("permissions") or []; assert "system.full_access" in perms, data' || {
  echo "[API] /auth/me expected system.full_access permission"
  exit 1
}

echo "[API] GET $BASE/api/v1/recruitment/candidates"
CAND_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/recruitment/candidates?page=1&pageSize=10")"
echo "$CAND_JSON" | python3 "$ENVE"
echo "$CAND_JSON" | python3 -c 'import json,sys; assert "items" in (json.load(sys.stdin).get("data") or {})' || {
  echo "[API] recruitment list expected data.items"
  exit 1
}

echo "[API] GET $BASE/api/v1/recruitment/positions"
POS_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/recruitment/positions?page=1&pageSize=10")"
echo "$POS_JSON" | python3 "$ENVE"

echo "[API] GET $BASE/api/v1/compliance/qualifications"
QUAL_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/compliance/qualifications?page=1&pageSize=10")"
echo "$QUAL_JSON" | python3 "$ENVE"
echo "$QUAL_JSON" | python3 -c 'import json,sys; assert "items" in (json.load(sys.stdin).get("data") or {})' || {
  echo "[API] compliance qualifications expected data.items"
  exit 1
}

echo "[API] GET $BASE/api/v1/compliance/restrictions"
REST_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/compliance/restrictions?page=1&pageSize=10")"
echo "$REST_JSON" | python3 "$ENVE"

echo "[API] GET $BASE/api/v1/cases"
CASES_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/cases?page=1&pageSize=10")"
echo "$CASES_JSON" | python3 "$ENVE"
echo "$CASES_JSON" | python3 -c 'import json,sys; assert "items" in (json.load(sys.stdin).get("data") or {})' || {
  echo "[API] cases list expected data.items"
  exit 1
}

echo "[API] GET $BASE/api/v1/audit/logs"
AUDIT_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/audit/logs?page=1&pageSize=10")"
echo "$AUDIT_JSON" | python3 "$ENVE"

echo "[API] GET $BASE/api/v1/users"
RBAC_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/users")"
echo "$RBAC_JSON" | python3 "$ENVE"

echo "[API] GET $BASE/api/v1/roles"
ROLES_LIST_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/roles")"
echo "$ROLES_LIST_JSON" | python3 "$ENVE"
echo "$ROLES_LIST_JSON" | grep -q 'business_specialist' || {
  echo "[API] roles list expected seeded business_specialist slug"
  exit 1
}

echo "[API] GET $BASE/api/v1/users/{id}"
USER_DETAIL_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/users/00000000-0000-4000-8000-000000000001")"
echo "$USER_DETAIL_JSON" | python3 "$ENVE"
echo "$USER_DETAIL_JSON" | python3 -c 'import json,sys; assert "roleIds" in (json.load(sys.stdin).get("data") or {})' || {
  echo "[API] user detail expected data.roleIds"
  exit 1
}

echo "[API] GET $BASE/api/v1/files"
FILES_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/files?page=1&pageSize=10")"
echo "$FILES_JSON" | python3 "$ENVE"
echo "$FILES_JSON" | python3 -c 'import json,sys; assert "items" in (json.load(sys.stdin).get("data") or {})' || {
  echo "[API] files list expected data.items"
  exit 1
}

echo "[API] Contract checks passed."
