# PharmaOps Static Delivery Acceptance & Architecture Audit

## 1. Verdict

- **Overall conclusion: Partial Pass**
- The repository is a substantial full-stack delivery with broad prompt alignment, but there are **material security and requirement-fit risks** (notably default privileged seed credentials and weak resume parsing approach for binary formats) that prevent a full pass.

## 2. Scope and Static Verification Boundary

- **Reviewed**
  - Product/contract docs: `docs/design.md`, `docs/api-spec.md`, `repo/README.md`
  - Backend bootstrap, routing, authN/authZ, service/repository layers, migrations, and static tests
  - Frontend routing/layout and core views for recruitment/compliance/cases/files/audit
- **Not reviewed**
  - Runtime behavior, deployment runtime stability, real browser interactions, real DB state evolution under load
  - Non-primary generated artifacts (`dist`, `node_modules`) as implementation evidence
- **Intentionally not executed**
  - Project startup, Docker, tests, external services (per audit constraints)
- **Manual verification required**
  - Real-world resume parsing quality for PDF/DOCX inputs
  - Scheduler behavior over time (expiration/deactivation cadence in deployed environment)
  - UX polish and interaction consistency under actual browser usage

## 3. Repository / Requirement Mapping Summary

- **Prompt core goal mapped:** offline intranet pharma compliance + recruitment platform with RBAC/data scope, secure auth/session, PII controls, file uploads, case ledger, and append-only audit.
- **Mapped implementation areas:**
  - API route and middleware enforcement in `apps/api/internal/httpserver/server.go`
  - Domain logic in recruitment/compliance/case/file/audit services and repositories
  - DB contracts in `infra/db/migrations`
  - UI route/modules and operator flows in `apps/web/src/views/**`
  - Static tests in `apps/api/internal/**/*test.go`

## 4. Section-by-section Review

### 1. Hard Gates

#### 1.1 Documentation and static verifiability
- **Conclusion: Pass**
- **Rationale:** README provides setup/env/test entry points, and those are statically consistent with project layout and route registration.
- **Evidence:** `repo/README.md:31`, `repo/README.md:58`, `repo/apps/api/cmd/api/main.go:12`, `repo/apps/api/internal/httpserver/server.go:83`, `repo/apps/web/src/router/index.ts:6`

#### 1.2 Material deviation from Prompt
- **Conclusion: Partial Pass**
- **Rationale:** Core domains are implemented, but resume ingestion for binary files is likely below prompt intent for structured extraction quality.
- **Evidence:** Prompt-required resume import and structured extraction target in `docs/design.md:184`; current parser reads raw file bytes and regex-matches text without format-specific parsers in `repo/apps/api/internal/service/recruitment_extended.go:143`, `repo/apps/api/internal/service/recruitment_extended.go:208`, `repo/apps/api/internal/service/recruitment_extended.go:105`
- **Manual verification note:** Validate extraction precision/recall on realistic PDF/DOCX resumes.

### 2. Delivery Completeness

#### 2.1 Core prompt requirement coverage
- **Conclusion: Partial Pass**
- **Rationale:** Most explicit requirements are implemented (RBAC/scope, case numbering/duplicate guard, compliance checks, file chunk upload/dedup, audit append-only), but resume extraction depth for binary formats appears weak.
- **Evidence:** RBAC+scope middleware and permissions `repo/apps/api/internal/httpserver/server.go:89`; case numbering/duplicate guard `repo/apps/api/internal/service/case_service.go:127`; compliance purchase checks `repo/apps/api/internal/service/compliance_service.go:656`; upload chunk+dedup `repo/apps/api/internal/service/file_service.go:130`; audit append-only DB guard `repo/infra/db/migrations/000020_audit_immutability_guards.up.sql:4`; resume parser concern `repo/apps/api/internal/service/recruitment_extended.go:143`

#### 2.2 0-to-1 deliverable vs demo fragment
- **Conclusion: Pass**
- **Rationale:** Complete multi-module backend/frontend structure with migrations, handlers/services/repos, and documented commands exists; not a toy single-file sample.
- **Evidence:** `repo/apps/api/internal/httpserver/server.go:83`, `repo/apps/web/src/router/index.ts:14`, `repo/infra/db/migrations/000001_initial_schema.up.sql:10`, `repo/README.md:14`

### 3. Engineering and Architecture Quality

#### 3.1 Structure and decomposition
- **Conclusion: Pass**
- **Rationale:** Router/handler/service/repository layering is clear and consistent; domain modules are separated.
- **Evidence:** Layered composition in `repo/apps/api/internal/httpserver/server.go:40`, service modules under `repo/apps/api/internal/service/*.go`, repositories under `repo/apps/api/internal/repository/*.go`

#### 3.2 Maintainability/extensibility
- **Conclusion: Partial Pass**
- **Rationale:** Architecture is generally maintainable, but secure deployment hygiene is weakened by unconditional dev seed migrations.
- **Evidence:** unconditional seeded admin credentials `repo/infra/db/migrations/000002_dev_seed_user.up.sql:6`; unconditional full-access role binding `repo/infra/db/migrations/000003_dev_rbac_scope_seed.up.sql:67`

### 4. Engineering Details and Professionalism

#### 4.1 Error handling, logging, validation, API design
- **Conclusion: Partial Pass**
- **Rationale:** Response envelopes, typed errors, and validations are broadly present; however security posture is weakened by default privileged seed data and optional health protection.
- **Evidence:** envelope shape `repo/apps/api/internal/response/envelope.go:11`; validation examples `repo/apps/api/internal/handler/case.go:101`; logging categories `repo/apps/api/internal/oplog/oplog.go:36`; optional health token only when configured `repo/apps/api/internal/handler/health.go:22`, `repo/.env.example:33`

#### 4.2 Product-level realism vs demo-only
- **Conclusion: Pass**
- **Rationale:** Multi-role modules, persistence layer, and cross-domain workflows indicate product-style implementation rather than a teaching stub.
- **Evidence:** full route surface `repo/apps/api/internal/httpserver/server.go:94`; recruitment/compliance/cases/files/audit UI modules in `repo/apps/web/src/views/**`

### 5. Prompt Understanding and Requirement Fit

#### 5.1 Business goal/scenario/constraints fit
- **Conclusion: Partial Pass**
- **Rationale:** Strong coverage of prompt semantics overall, but two key fit risks remain: (1) binary resume extraction quality, and (2) insecure default privileged account posture conflicting with production-grade compliance expectations.
- **Evidence:** recruitment parsing approach `repo/apps/api/internal/service/recruitment_extended.go:143`; seeded `admin/password` `repo/README.md:9`, `repo/infra/db/migrations/000002_dev_seed_user.up.sql:9`; seeded `system.full_access` role binding `repo/infra/db/migrations/000003_dev_rbac_scope_seed.up.sql:55`

### 6. Aesthetics (frontend)

#### 6.1 Visual and interaction quality
- **Conclusion: Cannot Confirm Statistically**
- **Rationale:** Static code shows coherent Element Plus usage, spacing/theme, and interaction feedback patterns, but final rendering quality requires manual UI execution.
- **Evidence:** layout/theme styling `repo/apps/web/src/layouts/AppLayout.vue:83`; confirmations/toasts in high-impact flows e.g. `repo/apps/web/src/views/cases/CasesView.vue:285`, `repo/apps/web/src/views/recruitment/CandidatesView.vue:257`
- **Manual verification note:** Run UI and inspect cross-page visual consistency, hover/click states, and responsive behavior.

## 5. Issues / Suggestions (Severity-Rated)

### Blocker / High

1) **Severity: High**  
**Title:** Default privileged admin account seeded by migrations  
**Conclusion:** Fail  
**Evidence:** `repo/infra/db/migrations/000002_dev_seed_user.up.sql:6`, `repo/infra/db/migrations/000003_dev_rbac_scope_seed.up.sql:67`, `repo/README.md:9`  
**Impact:** Any deployment running migrations as-is receives a known-credential admin mapped to full-access role/scope, creating a critical takeover risk.  
**Minimum actionable fix:** Gate dev seed migrations by environment, or move them to explicit optional seed scripts not run in production paths; enforce first-login password rotation if seeds are retained for local-only use.

2) **Severity: High**  
**Title:** Resume extraction for PDF/DOCX is format-naive and likely unreliable  
**Conclusion:** Partial Fail  
**Evidence:** raw byte-to-string regex parsing in `repo/apps/api/internal/service/recruitment_extended.go:143`, file read in `repo/apps/api/internal/service/recruitment_extended.go:208`, MIME whitelist includes binary docs in `repo/apps/api/internal/service/file_service.go:32`  
**Impact:** Bulk resume import can produce poor/incorrect structured data for common resume formats, undermining core recruitment workflow quality and downstream matching/merge behavior.  
**Minimum actionable fix:** Add format-aware extractors for PDF/DOCX (or clearly constrain accepted import formats), and return deterministic validation errors for unsupported parse fidelity.

### Medium

3) **Severity: Medium**  
**Title:** Health endpoint hardening is optional and defaults open  
**Conclusion:** Partial Fail  
**Evidence:** token check only when env var set `repo/apps/api/internal/handler/health.go:22`; default empty token `repo/.env.example:33`  
**Impact:** Infrastructure/DB health metadata can be queried without auth in default config, increasing information disclosure surface.  
**Minimum actionable fix:** Require token (or auth) for health by default; provide explicit opt-out for local dev only.

4) **Severity: Medium**  
**Title:** Critical auth/session lifecycle lacks substantive automated coverage  
**Conclusion:** Partial Fail  
**Evidence:** no functional auth service/session tests found; only constructor reference `repo/apps/api/internal/service/service_constructors_test.go:14`; middleware bearer parsing only `repo/apps/api/internal/middleware/auth_test.go:5`  
**Impact:** Severe defects in login hash validation, TTL expiry, or logout invalidation could pass current test suite undetected.  
**Minimum actionable fix:** Add integration tests for login success/failure, session expiry behavior, logout revocation, and token reuse rejection.

5) **Severity: Medium**  
**Title:** Frontend has no automated test suite  
**Conclusion:** Partial Fail  
**Evidence:** no test files under web app `repo/apps/web` (static scan), while `package.json` has no `test` script `repo/apps/web/package.json:6`  
**Impact:** UI regressions in critical operator flows (confirmations, role-aware visibility, import UX) are not automatically detected.  
**Minimum actionable fix:** Add at least smoke-level component/E2E coverage for login, RBAC menu visibility, candidate import, compliance checks, and case transitions.

## 6. Security Review Summary

- **Authentication entry points: Partial Pass**  
  Evidence: login/logout/me routes in `repo/apps/api/internal/httpserver/server.go:86`; bcrypt compare and token hashing in `repo/apps/api/internal/service/auth_service.go:64`, `repo/apps/api/internal/service/auth_service.go:76`.  
  Note: Seeded default admin credentials materially weaken deployment security (`000002`, `000003` migrations).

- **Route-level authorization: Pass**  
  Evidence: permission middleware applied per protected route in `repo/apps/api/internal/httpserver/server.go:94`; permission checks in `repo/apps/api/internal/middleware/require_permission.go:23`.

- **Object-level authorization: Pass**  
  Evidence: repository `Get/Update/Delete` paths apply data-scope filters, e.g. candidates `repo/apps/api/internal/repository/recruitment_repo.go:120`, cases `repo/apps/api/internal/repository/case_repo.go:112`, compliance `repo/apps/api/internal/repository/compliance_repo.go:37`.

- **Function-level authorization: Partial Pass**  
  Evidence: service-level scope guards (`requireScope`, `RowVisible`) in multiple services, e.g. `repo/apps/api/internal/service/recruitment_service.go:355`, `repo/apps/api/internal/service/compliance_service.go:657`.  
  Gap: `GET /api/v1/health` protection is optional by env, not mandatory (`repo/apps/api/internal/handler/health.go:22`).

- **Tenant / user data isolation: Pass**  
  Evidence: centralized scope expression builder and scoped query application `repo/apps/api/internal/repository/scope_where.go:45`; file accessibility bound to scope or uploader `repo/apps/api/internal/repository/file_repo.go:125`.

- **Admin / internal / debug endpoint protection: Partial Pass**  
  Evidence: admin RBAC routes protected by `system.rbac` in `repo/apps/api/internal/httpserver/server.go:164`; open `/healthz` and optionally open `/api/v1/health` in `repo/apps/api/internal/httpserver/server.go:79`, `repo/apps/api/internal/handler/health.go:22`.

## 7. Tests and Logging Review

- **Unit tests: Partial Pass**  
  Exists across service/repository/middleware packages (e.g. `match_score_test.go`, `case_duplicate_window_test.go`, `endpoint_permission_matrix_test.go`), but limited auth/session lifecycle depth.

- **API / integration tests: Partial Pass**  
  Integration-style tests exist for compliance purchase enforcement, case creation duplicate guard, file dedup, audit mutation, and scope checks (`service/*integration_test.go`, `repository/case_scope_integration_test.go`), but auth/session lifecycle integration coverage is weak.

- **Logging categories / observability: Pass**  
  Structured operational logging categories exist for auth/authz/session/audit/PII/crypto events in `repo/apps/api/internal/oplog/oplog.go:36`.

- **Sensitive-data leakage risk in logs/responses: Partial Pass**  
  Candidate list/detail masking and audit payload sanitization exist (`recruitment_service.go`, `audit_service.go`), but seeded admin credential pattern and optional open health endpoint remain security posture concerns.

## 8. Test Coverage Assessment (Static Audit)

### 8.1 Test Overview

- **Unit tests exist:** yes (`apps/api/internal/**/*test.go`)
- **API/integration-style tests exist:** yes (multiple `*_integration_test.go` files)
- **Frontend tests exist:** no (no test files in web app; no test script)
- **Framework(s):** Go `testing` package, with in-memory SQLite for many service/repo tests
- **Test entry points documented:** yes (`repo/README.md:71`, `repo/README.md:82`)
- **Boundary note:** Documented scripts depend on Docker/runtime and were not executed in this static audit (`repo/run_tests.sh:9`, `repo/API_tests/run_api_tests.sh:9`, `repo/e2e_tests/run_e2e_tests.sh:9`)

### 8.2 Coverage Mapping Table

| Requirement / Risk Point | Mapped Test Case(s) | Key Assertion / Fixture / Mock | Coverage Assessment | Gap | Minimum Test Addition |
|---|---|---|---|---|---|
| Route authn/authz wiring | `apps/api/internal/httpserver/auth_matrix_contract_test.go:8`; `apps/api/internal/middleware/endpoint_permission_matrix_test.go:13` | Static route guard substring checks and middleware 403 matrix | basically covered | Does not validate real DB-backed principal/session path | Add black-box HTTP tests with real auth middleware + seeded DB |
| Data-scope isolation | `apps/api/internal/service/scope_isolation_test.go:12`; `apps/api/internal/repository/case_scope_integration_test.go:15` | Scope mismatch returns forbidden / filtered rows | sufficient | Mostly institution/dept/team checks; limited cross-module object scenarios | Add cross-resource object-level scope abuse tests (case/file/audit export) |
| Case numbering + 5-minute duplicate guard | `apps/api/internal/service/case_create_integration_test.go:18`; `apps/api/internal/service/case_duplicate_window_test.go:7` | Serial suffix increments; duplicate rejected | sufficient | No concurrency race test for serial allocation | Add concurrent create test for same institution/day |
| Compliance purchase restrictions (Rx + frequency) | `apps/api/internal/service/compliance_purchase_enforcement_integration_test.go:16`; `apps/api/internal/service/compliance_check_purchase_test.go:12` | Rx requirement + frequency checks + scoped behavior | sufficient | No end-to-end handler-level tests for typed errors | Add handler/API tests for code/status matrix |
| File chunk upload + dedup + type validation | `apps/api/internal/service/file_upload_dedup_integration_test.go:17` | Dedup true on second upload; MIME spoof rejected | sufficient | No permission/scope misuse tests on download/link endpoints | Add authz tests for file link/download across scopes |
| Audit mutation behavior + PII sanitization | `apps/api/internal/service/audit_mutation_integration_test.go:16`; `apps/api/internal/service/audit_log_persistence_test.go:14` | `_changedFields` presence, PII redaction, persistence | sufficient | Limited coverage for export authorization paths | Add export ownership and full_access override tests |
| Auth/session lifecycle (login, TTL, logout invalidation) | only constructor-level mention in `apps/api/internal/service/service_constructors_test.go:14`; bearer parsing in `apps/api/internal/middleware/auth_test.go:5` | No end-to-end assertions for session semantics | missing | Critical requirement not meaningfully tested | Add integration tests for valid login, expired token, logout revocation, invalid credentials |
| Recruitment matching scoring logic | `apps/api/internal/service/match_score_test.go:9` | Score bounds, breakdown weights, title token fallback | basically covered | No tests for recommendation endpoints and merge-triggered scoring impact | Add service tests for similar-candidate/position ranking behavior |
| Frontend critical flows | none | n/a | missing | UI regressions undetected | Add frontend unit/E2E smoke tests |

### 8.3 Security Coverage Audit

- **Authentication:** **insufficient**  
  Minimal direct tests; no robust session lifecycle integration tests.
- **Route authorization:** **basically covered**  
  Permission matrix and route guard presence are tested statically.
- **Object-level authorization:** **basically covered**  
  Scope enforcement tests exist in services/repos, but not exhaustive across all resources.
- **Tenant/data isolation:** **basically covered**  
  Institution/dept/team predicates are exercised in multiple tests.
- **Admin/internal protection:** **insufficient**  
  Little/no direct test coverage for health endpoint hardening and export-owner access edge cases.

### 8.4 Final Coverage Judgment

- **Final Coverage Judgment: Partial Pass**
- Major business/security areas (scope enforcement, case duplicate guard, compliance restriction logic, file dedup, audit mutation sanitation) have meaningful static tests.  
- However, **critical auth/session lifecycle coverage is missing**, and frontend has no automated tests, so severe defects could still remain undetected while many tests pass.

## 9. Final Notes

- This is a static-only assessment; no runtime claims are made beyond code/test evidence.
- The strongest blockers are security posture and requirement-fit risk, not general code organization.
- Priority remediation should start with privileged seed handling and robust resume parsing strategy, then strengthen auth/session and frontend testing.
