# PharmaOps Project

Run and verify everything through **Docker** from this directory (`repo/`).

## Stack

```bash
docker compose up -d --build
docker compose ps
```

- **Web UI:** http://127.0.0.1:8080/
- **MySQL:** `localhost:3306` (see `docker-compose.yml` for credentials)

Stop:

```bash
docker compose down
```

## Tests

```bash
bash run_tests.sh
```

This builds/starts the Compose stack, performs a **web smoke check** (HTTP 200 on port 8080), then runs unit, API, and E2E stages. Any failure exits non-zero.
