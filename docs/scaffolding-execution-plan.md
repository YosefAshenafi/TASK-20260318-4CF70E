# PharmaOps Scaffolding Execution Plan

This is the working execution plan we will follow step by step and remove after project completion.

## Scope Baseline

- UI-first delivery with professional engineering standards.
- Offline intranet deployment using Docker.
- No `sessions/` folder in repository structure.
- Session management is implemented in database/API (`sessions` table), not filesystem.

## Step-by-Step Delivery Order

### 1) Scaffolding

- Create project skeleton inside `repo/`:
  - `repo/apps/web`
  - `repo/apps/api`
  - `repo/infra/docker`
  - `repo/scripts`
  - `repo/API_tests`
  - `repo/e2e_tests`
  - `repo/unit_tests`
- Keep root only for `docs/`, `metadata.json`, and `repo/`.
- Add baseline files under `repo/`: `README.md`, `.env.example`, `Makefile`, contribution/security docs.
- Define conventions: branch strategy, commit format, review checklist.

Docker checkpoint:
- `cd repo && docker compose up -d --build`
- Validate container health and logs before moving forward.

### 2) Frontend

- Bootstrap Vue 3 + TypeScript + Vite + Element Plus.
- Build routing skeleton and domain-based module layout.
- Add role-aware menu shell and standardized feedback/confirmation patterns.

Docker checkpoint:
- Build and run frontend via Docker.
- Run first dockerized test cycle immediately after UI baseline is complete.

### 3) Database Design

- Define migration-first schema for identity, recruitment, compliance, cases, files, and audit.
- Add indexes and relationship constraints.
- Add security integration points (PII encryption-ready columns and policies).

Docker checkpoint:
- Run DB migrations in Dockerized environment.
- Verify schema creation and migration reproducibility.

### 4) Backend API (Connection + Integration)

- Bootstrap Go + Gin + GORM layered architecture.
- Integrate DB connectivity, auth/session flow, RBAC/data scope middleware.
- Integrate frontend contracts and consistent response envelope.

Docker checkpoint:
- Run API + DB (+ web where available) together.
- Validate health endpoints and integration smoke checks.

### 5) Testing (`run_tests.sh`)

- Add aggregated pipeline to run:
  - E2E tests
  - Backend unit tests
  - API tests
- Ensure non-zero exit on failure and clear pass/fail output.

Docker checkpoint:
- Run test pipeline against Dockerized services.

### 6) Checkup (Feature Criteria)

- Map required features to implementation and tests.
- Status values: `implemented`, `partial`, `missing`, `blocked`.
- Gate progress until critical items are covered.

Docker checkpoint:
- Re-run targeted regression checks in Docker before finalization.

### 7) Continuous Docker Verification (Per Step)

- Docker is not delayed to a final stage.
- At the end of every step, run docker build/start/check and fix environment issues immediately.

### 8) Final `testrun.sh`

- Execute full end-to-end release gate:
  - clean bootstrap
  - migrations
  - integrated tests
  - final report summary

## Required Test Artifacts

- `repo/API_tests/` (API test files and runner)
- `repo/e2e_tests/` (E2E test files and runner)
- `repo/unit_tests/` (unit test files/data and runner)
- `repo/run_tests.sh` (single entrypoint)

## Working Rule

We execute one step at a time, review output, and only then proceed to the next step.
