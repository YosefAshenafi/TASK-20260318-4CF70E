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
- Use **Bun** only inside the `web` Docker build (install + `bun run build`); do not rely on host dev servers for verification.
- Build routing skeleton and domain-based module layout.
- Add role-aware menu shell and standardized feedback/confirmation patterns.

Docker checkpoint:
- Update `repo/docker-compose.yml` / `repo/apps/web/Dockerfile` when the UI changes how it is built or served.
- `cd repo && docker compose up -d --build` then open **http://127.0.0.1:8080/** (nginx static bundle).
- Run `bash repo/run_tests.sh` (includes web HTTP smoke + test stages).

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

### 4b) Domain feature implementation (replace placeholders)

Step **4** delivers **integration**: auth, RBAC/data scope, envelope, DB connectivity.  
Step **4b** delivers **product**: real screens and APIs from `docs/design.md` and `docs/api-spec.md`, replacing `ModulePlaceholderView` and stub routes with working flows.

Implementation must **trace to `docs/design.md`** (routes §5.2, UI rules §5.3, domain §7.x, data §8 as applicable). Do not invent off-spec product behavior; if the design is silent on a detail, reconcile with `docs/api-spec.md` or clarify before coding.

Work in **domain slices** (order flexible; complete one slice before claiming the next):

- **Recruitment** — candidates, positions, imports, duplicates/merge, matching (see design §7.2, §5.4).
- **Compliance** — qualifications lifecycle, restrictions, violations (§7.3).
- **Cases** — intake, assignment, timeline, status transitions, ledger (§7.4).
- **Files** — upload sessions, chunks, dedup, references (§7.5).
- **Audit** — query, export, append-only events (§7.6).
- **System / RBAC UI** — user/role/scope admin as specified (routes under `/system/...`).

Each slice should include:

- Vue views + routing (no long-lived placeholders for that slice’s primary screens).
- Go handlers/services/repos aligned with `docs/api-spec.md`.
- RBAC permission + data-scope enforcement on new endpoints.
- Tests (unit/API/E2E) for that slice; extend `repo/API_tests/` and `repo/e2e_tests/` as behavior appears.

Docker checkpoint (per slice or per milestone):

- `cd repo && docker compose up -d --build`, migrations applied, manual smoke of new flows through **http://127.0.0.1:8080/** (same-origin `/api`).

**Relationship to Step 5:** Step 5 formalizes the **aggregated test pipeline**. You may start Step 5 in parallel once the pipeline skeleton exists; **test coverage should grow as Step 4b domains land** (do not treat Step 5 as “only after all of 4b is done” unless you choose that gate).

### 5) Testing (`run_tests.sh`)

- Add aggregated pipeline to run:
  - E2E tests
  - Backend unit tests
  - API tests
- Ensure non-zero exit on failure and clear pass/fail output.
- Expand stages and fixtures as Step **4b** adds real behavior (smoke tests alone are not sufficient long-term).

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

Step **4b** is intentionally **iterative**: ship domain slices (recruitment, compliance, etc.) in sequence; each slice should be reviewable and testable before moving to the next.
