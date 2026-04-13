#!/usr/bin/env bash
set -euo pipefail

# Provision an initial API/E2E operator without seeding default credentials.
# The password hash must be provided as bcrypt (not plaintext).

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
USERNAME="${API_TEST_USERNAME:-}"
DISPLAY_NAME="${API_TEST_DISPLAY_NAME:-QA Admin}"
PASSWORD_BCRYPT="${API_TEST_PASSWORD_BCRYPT:-}"
DB_NAME="${MYSQL_DATABASE:-pharmaops}"
DB_USER="${MYSQL_USER:-pharmaops}"
DB_PASSWORD="${MYSQL_PASSWORD:-pharmaops}"
ROLE_ID="10000000-0000-4000-8000-000000000020"
SCOPE_ID="10000000-0000-4000-8000-000000000010"

if [[ -z "$USERNAME" ]]; then
  echo "Set API_TEST_USERNAME before running provisioning."
  exit 1
fi
if [[ -z "$PASSWORD_BCRYPT" ]]; then
  echo "Set API_TEST_PASSWORD_BCRYPT to a bcrypt hash before running provisioning."
  exit 1
fi
if [[ "$USERNAME" =~ [\'\"] || "$DISPLAY_NAME" =~ [\'\"] || "$PASSWORD_BCRYPT" =~ [\'\"] ]]; then
  echo "Username/display/hash must not include quotes."
  exit 1
fi

echo "[provision] Ensuring test user '$USERNAME' exists with system_admin + dev-root scope..."
docker compose -f "$REPO/docker-compose.yml" exec -T db \
  mysql -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" <<SQL
SET @existing_user_id := (SELECT id FROM users WHERE username = '$USERNAME' LIMIT 1);
SET @user_id := IFNULL(@existing_user_id, UUID());

INSERT INTO users (id, username, password_hash, display_name, is_active, created_at, updated_at)
VALUES (@user_id, '$USERNAME', '$PASSWORD_BCRYPT', '$DISPLAY_NAME', 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3))
ON DUPLICATE KEY UPDATE
  password_hash = VALUES(password_hash),
  display_name = VALUES(display_name),
  is_active = 1,
  updated_at = CURRENT_TIMESTAMP(3);

INSERT IGNORE INTO user_roles (user_id, role_id, created_at)
VALUES (@user_id, '$ROLE_ID', CURRENT_TIMESTAMP(3));

INSERT IGNORE INTO user_data_scopes (user_id, data_scope_id, created_at)
VALUES (@user_id, '$SCOPE_ID', CURRENT_TIMESTAMP(3));
SQL

echo "[provision] Done."
