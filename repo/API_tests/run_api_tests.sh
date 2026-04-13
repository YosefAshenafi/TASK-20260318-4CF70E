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

# ── Negative / authz matrix tests ────────────────────────────────────────

echo "[API] Invalid token returns 401"
CODE="$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer bad-token-value" "$BASE/api/v1/auth/me")"
if [[ "$CODE" != "401" ]]; then
  echo "[API] expected 401 for invalid token, got $CODE"
  exit 1
fi

echo "[API] POST /auth/login with bad credentials returns 401"
BAD_LOGIN_CODE="$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"wrongpassword"}')"
if [[ "$BAD_LOGIN_CODE" != "401" ]]; then
  echo "[API] expected 401 for bad credentials, got $BAD_LOGIN_CODE"
  exit 1
fi

echo "[API] POST /auth/login with short password returns 400"
SHORT_PW_CODE="$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"short"}')"
if [[ "$SHORT_PW_CODE" != "400" ]]; then
  echo "[API] expected 400 for short password, got $SHORT_PW_CODE"
  exit 1
fi

echo "[API] Login response shape: token + expiresAt + user object"
echo "$LOGIN_JSON" | python3 -c '
import json, sys
d = json.load(sys.stdin)["data"]
assert "token" in d, "missing token"
assert "expiresAt" in d, "missing expiresAt"
assert "user" in d, "missing user object"
u = d["user"]
assert "id" in u, "missing user.id"
assert "username" in u, "missing user.username"
assert "roles" in u, "missing user.roles"
' || {
  echo "[API] login response does not match api-spec contract"
  exit 1
}

echo "[API] /auth/me response includes scopes"
echo "$ME_JSON" | python3 -c '
import json, sys
d = json.load(sys.stdin)["data"]
assert isinstance(d.get("scopes"), list), "scopes must be a list"
if len(d["scopes"]) > 0:
    s = d["scopes"][0]
    assert "id" in s, "scope missing id"
    assert "institutionId" in s, "scope missing institutionId"
' || {
  echo "[API] /auth/me response does not include scopes per api-spec"
  exit 1
}

echo "[API] Envelope error shape for 401"
ERR_JSON="$(curl -s -H "Authorization: Bearer bad-token" "$BASE/api/v1/recruitment/candidates")"
echo "$ERR_JSON" | python3 -c '
import json, sys
d = json.load(sys.stdin)
assert d.get("code") not in (None, "", "OK"), "error code should not be OK: " + str(d)
assert "message" in d, "missing message"
assert "requestId" in d, "missing requestId"
' || {
  echo "[API] error envelope shape incorrect"
  exit 1
}

echo "[API] Recruitment extended routes are callable (imports, duplicates, match, recommendations)"
for ROUTE in \
  "recruitment/candidates/duplicates" \
  "recruitment/candidates/merge-history?page=1&pageSize=10"; do
  CODE="$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/$ROUTE")"
  if [[ "$CODE" -ge "500" ]]; then
    echo "[API] GET $ROUTE returned $CODE (server error)"
    exit 1
  fi
done

echo "[API] GET candidate detail (dev seed)"
CAND_IDS="$(echo "$CAND_JSON" | python3 -c 'import json,sys; items=json.load(sys.stdin)["data"]["items"]; print(items[0]["id"] if items else "")')"
if [[ -n "$CAND_IDS" ]]; then
  DETAIL_JSON="$(curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/recruitment/candidates/$CAND_IDS")"
  echo "$DETAIL_JSON" | python3 -c '
import json, sys
d = json.load(sys.stdin)["data"]
assert "phoneMasked" in d, "missing phoneMasked"
assert "idNumberMasked" in d, "missing idNumberMasked"
assert "institutionId" in d, "missing institutionId"
' || {
    echo "[API] candidate detail missing expected fields"
    exit 1
  }
fi

echo "[API] Audit logs contain entries (dev seed + mutations from earlier tests)"
AUDIT_COUNT="$(echo "$AUDIT_JSON" | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["total"])')"
if [[ "$AUDIT_COUNT" -lt 1 ]]; then
  echo "[API] expected at least 1 audit log entry"
  exit 1
fi

echo "[API] Contract checks passed."
