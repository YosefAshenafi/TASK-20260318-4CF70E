# PharmaOps Project

Run and verify everything through **Docker** from this directory (`repo/`).

## Stack

```bash
docker compose up -d --build
docker compose ps
```

- **Web UI:** http://127.0.0.1:8080/ (sign-in uses the API; dev seed `admin` / `password` after migrations)
- **MySQL:** `localhost:3306` (see `docker-compose.yml` for credentials)

Stop:

```bash
docker compose down
```

## Tests

```bash
bash run_tests.sh
```

This builds/starts the Compose stack, performs a **web smoke check** (HTTP 200 on port 8080), then runs unit, API, E2E, and design conformance checks against `docs/design.md`. Any failure exits non-zero.

Full release verification (clean database volumes, same stages, log under `reports/`):

```bash
bash testrun.sh
```
