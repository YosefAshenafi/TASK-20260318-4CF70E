# PharmaOps Project

Minimal usage guide for Docker-first workflow.

## Run with Docker

1. Start dependencies:
   - `docker compose up -d --build`
2. Check service status:
   - `docker compose ps`
3. Stop services:
   - `docker compose down`

## Run Tests

Use the single test entrypoint:

- `bash run_tests.sh`

Expected output:
- Unit test stage result
- API test stage result
- E2E test stage result
- Final summary with pass/fail status

If any stage fails, the script exits with a non-zero code.
