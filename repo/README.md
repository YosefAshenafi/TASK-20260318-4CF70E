# PharmaOps Project

Run and verify everything through **Docker** from this directory (`repo/`).

## Stack

```bash
docker compose up -d --build
docker compose ps
```

- **Web UI:** http://127.0.0.1:8080/ (sign-in uses the API; dev seed `admin` / `password` after migrations)
- **API:** http://127.0.0.1:8080/api/v1/health (proxied through Nginx; direct container port 8080 on `pharmaops-api`)
- **MySQL:** `localhost:3306` (see `docker-compose.yml` for credentials)

Stop:

```bash
docker compose down
```

## Environment Variables

All configuration is environment-driven. See [`.env.example`](.env.example) for a documented template.

| Variable | Required | Default | Description |
|---|---|---|---|
| `HTTP_ADDR` | No | `:8080` | API listen address |
| `MYSQL_DSN` | No | `pharmaops:pharmaops@tcp(127.0.0.1:3306)/pharmaops?parseTime=true&loc=UTC&charset=utf8mb4` | MySQL connection string |
| `APP_ENV` | No | `development` | `development` or `production` (controls Gin mode) |
| `PII_AES_KEY_HEX` | **Yes** | *(none)* | 64 hex chars (32 bytes AES-256) for candidate PII encryption at rest. Generate: `openssl rand -hex 32` |
| `FILE_STORAGE_ROOT` | No | `$TMPDIR/pharmaops-uploads` | Absolute path for file uploads and chunks |
| `HEALTH_CHECK_TOKEN` | No | *(empty)* | When set, `/api/v1/health` requires `X-Internal-Health-Token` |

Docker Compose injects defaults for the API service. For local development outside Docker, copy `.env.example` to `.env` and source it.

## Database Migrations

Migrations live in `infra/db/migrations/` and are applied automatically on API startup via `scripts/db_migrate.sh`:

```bash
# Inside Docker (automatic on compose up):
docker compose exec api sh -c "cd /app && sh scripts/db_migrate.sh"

# Manual (requires go-migrate or goose):
bash scripts/db_migrate.sh
```

## Running Services Individually

```bash
# API (requires MySQL running):
cd apps/api && go run ./cmd/api

# Web dev server:
cd apps/web && npm install && npm run dev

# MySQL only:
docker compose up -d db
```

## Tests

```bash
bash run_tests.sh
```

This builds/starts the Compose stack, performs a **web smoke check** (HTTP 200 on port 8080), then runs unit, API, E2E, and design conformance checks against `docs/design.md`. Any failure exits non-zero.

### Individual test suites

```bash
# Go unit tests:
cd apps/api && go test ./...

# API contract tests (requires running stack):
bash API_tests/run_api_tests.sh

# E2E smoke tests (requires running stack):
bash e2e_tests/run_e2e_tests.sh
```

Full release verification (clean database volumes, same stages, log under `reports/`):

```bash
bash testrun.sh
```
