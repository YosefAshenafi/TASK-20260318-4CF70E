#!/usr/bin/env bash
# Apply SQL migrations in infra/db/migrations/*.up.sql (lexicographic order).
# Records applied versions in schema_migrations. Idempotent.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

MYSQL_USER="${MYSQL_USER:-pharmaops}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-pharmaops}"
MYSQL_DATABASE="${MYSQL_DATABASE:-pharmaops}"

mysql_exec() {
  docker compose exec -T db mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" "$@" "$MYSQL_DATABASE"
}

echo "[db_migrate] Waiting for MySQL..."
for _ in $(seq 1 60); do
  if docker compose exec -T db mysqladmin ping -h localhost -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" --silent 2>/dev/null; then
    break
  fi
  sleep 1
done

mysql_exec -e "
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(64) NOT NULL,
  applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
"

shopt -s nullglob
for f in "$REPO_ROOT/infra/db/migrations"/*.up.sql; do
  base="$(basename "$f" .up.sql)"
  count="$(mysql_exec -N -e "SELECT COUNT(*) FROM schema_migrations WHERE version='${base}'" | tr -d '\r')"
  if [[ "$count" == "1" ]]; then
    echo "[db_migrate] Skip (already applied): $base"
    continue
  fi
  echo "[db_migrate] Applying: $base"
  mysql_exec <"$f"
  mysql_exec -e "INSERT INTO schema_migrations (version) VALUES ('${base}')"
done

echo "[db_migrate] Done."
