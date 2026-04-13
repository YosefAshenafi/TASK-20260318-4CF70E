# PharmaOps Static Delivery Acceptance & Architecture Audit

## 1. Verdict
- Overall conclusion: **Fail**

## 2. Scope and Static Verification Boundary
- Reviewed: `docs/design.md`, `docs/api-spec.md`, `repo/README.md`, Docker/manifests/scripts, backend handlers/services/repos/models/migrations, frontend routes/views/stores, and all in-repo test files.
- Not reviewed: runtime behavior in a live stack, browser interaction outcomes, actual DB state after migrations, upload filesystem behavior under real load.
- Intentionally not executed: project startup, Docker, tests, E2E/browser flows (per static-only boundary).
- Manual verification required for: true runtime correctness, scheduler execution behavior over time, upload resume reliability under interruption, UI rendering details across browsers.

## 3. Repository / Requirement Mapping Summary
- Prompt core goals: offline full-stack pharma compliance + talent ops platform with RBAC + institution/department/team scope controls, recruitment import/merge/match/recommendation, compliance expiry/deactivation/restrictions, case numbering + duplicate guard + ledger, file chunk upload + SHA256 dedup, append-only searchable/exportable audit logs, and strong security (bcrypt auth, token lifecycle, encrypted PII at rest, masked PII in list).
- Implemented areas found: authenticated Gin API with permission middleware, institution-level scope checks, CRUD-heavy modules for recruitment/compliance/cases/files/RBAC, Vue route shell and module views, MySQL schema including many target tables, Dockerized deployment wiring.
- Key gap pattern: schema/design include many required capabilities, but service/router/test implementation covers a materially smaller subset.

## 4. Section-by-section Review

### 1. Hard Gates

#### 1.1 Documentation and static verifiability
- Conclusion: **Partial Pass**
- Rationale: startup/test entrypoints are documented and statically coherent for Docker flow, but configuration documentation is minimal and relies on implicit defaults.
- Evidence: `repo/README.md:3`, `repo/README.md:24`, `repo/docker-compose.yml:1`, `repo/scripts/db_migrate.sh:1`, `.env example missing` (`repo` has no `.env*` file).
- Manual verification note: actual end-to-end runability cannot be confirmed statically.

#### 1.2 Material deviation from Prompt
- Conclusion: **Fail**
- Rationale: delivered implementation omits multiple prompt-critical behaviors (recruitment import/merge/matching/recommendation, audited mutation trails, encrypted PII write path), so delivery materially deviates from the business prompt.
- Evidence: required endpoints listed at `docs/api-spec.md:122`, `docs/api-spec.md:126`, `docs/api-spec.md:132`, `docs/api-spec.md:134`; absent from registered routes in `repo/apps/api/internal/httpserver/server.go:78`-`141`. No encryption utilities/use in API internals (`repo/apps/api/internal/model/recruitment.go:15` and no encrypt/decrypt implementation usage).

### 2. Delivery Completeness

#### 2.1 Core requirements coverage
- Conclusion: **Fail**
- Rationale: only partial domain slices are implemented; many explicit requirements are missing or partially represented.
- Evidence:
  - Recruitment: only candidate/position CRUD (`repo/apps/api/internal/handler/recruitment.go:23`-`308`), no import/merge/match/recommend APIs from spec (`docs/api-spec.md:122`-`135`).
  - Compliance: restriction checks and qualification CRUD exist (`repo/apps/api/internal/service/compliance_service.go:185`-`617`), but no prescription attachment management endpoints (spec expects attachment-linked flows: `docs/api-spec.md:205`-`209` + schema table `prescription_attachments` at `repo/infra/db/migrations/000001_initial_schema.up.sql:356` not used in code).
  - Cases: numbering + duplicate guard + transitions exist (`repo/apps/api/internal/service/case_service.go:123`-`191`, `355`-`403`), but attachment archive/index APIs are not implemented though schema includes `case_attachment_indexes` (`repo/infra/db/migrations/000001_initial_schema.up.sql:529`).
  - Audit: query/export exists (`repo/apps/api/internal/handler/audit.go:23`-`109`), but no cross-module audit event writes for business mutations.

#### 2.2 Basic end-to-end deliverable (not demo fragment)
- Conclusion: **Partial Pass**
- Rationale: project has complete multi-app structure and routable UI/API, but notable demo coupling/hardcoding and missing core flows keep it below full 0-to-1 deliverable.
- Evidence: complete structure (`repo/apps/web`, `repo/apps/api`, `repo/infra/db/migrations`), but dev-seed hardcoded IDs in UI create flows (`repo/apps/web/src/config/devSeed.ts:1`-`12`, `repo/apps/web/src/views/recruitment/CandidatesView.vue:86`, `repo/apps/web/src/views/compliance/RestrictionsView.vue:277`, `repo/apps/web/src/views/cases/CasesView.vue:142`).

### 3. Engineering and Architecture Quality

#### 3.1 Engineering structure and module decomposition
- Conclusion: **Pass**
- Rationale: backend layering (handler/service/repository/model) and frontend module routing are clear and consistent.
- Evidence: backend wiring in `repo/apps/api/internal/httpserver/server.go:34`-`58`; module handlers under `repo/apps/api/internal/handler/*`; frontend route/module decomposition at `repo/apps/web/src/router/index.ts:17`-`73`.

#### 3.2 Maintainability and extensibility
- Conclusion: **Partial Pass**
- Rationale: architecture is extensible, but several business-critical paths are hardcoded (seed institution IDs in UI) or only scaffolded in schema without service/API integration.
- Evidence: hardcoded dev IDs (`repo/apps/web/src/config/devSeed.ts:1`-`12`), unused designed tables (`repo/infra/db/migrations/000001_initial_schema.up.sql:256`, `271`, `286`, `529`) with no corresponding route/service usage in server/router (`repo/apps/api/internal/httpserver/server.go:78`-`141`).

### 4. Engineering Details and Professionalism

#### 4.1 Error handling, logging, validation, API design
- Conclusion: **Partial Pass**
- Rationale: envelope/error conventions and basic validation are present, but logging/audit implementation is weak relative to compliance-grade requirements.
- Evidence:
  - Positive: envelope + request id middleware (`repo/apps/api/internal/response/envelope.go:11`-`40`, `repo/apps/api/internal/middleware/requestid.go:10`-`19`).
  - Gap: no structured application logs beyond defaults (`repo/apps/api/internal/db/db.go:11`) and no module mutation audit writes.
  - Validation present but limited to basic shape in handlers (`repo/apps/api/internal/handler/*`).

#### 4.2 Real product/service vs demo-only
- Conclusion: **Partial Pass**
- Rationale: resembles a real service skeleton, but critical behaviors remain missing while demo-seed assumptions appear in user flows.
- Evidence: demo/dev seeding flagged in migrations (`repo/infra/db/migrations/000002_dev_seed_user.up.sql:1`, `000005_recruitment_demo_seed.up.sql:1`), seed-bound UI defaults (`repo/apps/web/src/config/devSeed.ts:1`-`12`).

### 5. Prompt Understanding and Requirement Fit

#### 5.1 Business goal and constraints fit
- Conclusion: **Fail**
- Rationale: partial implementation misses central recruitment intelligence and compliance-grade traceability/security obligations required by prompt.
- Evidence: missing recruitment import/merge/scoring/recommendation endpoints (`docs/api-spec.md:122`-`135` vs `repo/apps/api/internal/httpserver/server.go:78`-`88`); no PII encryption implementation path despite requirement (`docs/design.md:375` and absence of crypto use in API internals except hashing tokens/files).

### 6. Aesthetics (frontend)

#### 6.1 Visual and interaction quality
- Conclusion: **Partial Pass**
- Rationale: static code shows consistent Element Plus theming and generally clear layout/feedback patterns, but runtime rendering/interaction quality is not statically provable.
- Evidence: themed layout/menu styles (`repo/apps/web/src/layouts/AppLayout.vue:81`-`203`), action feedback via `ElMessage`/dialogs across views (e.g., `repo/apps/web/src/views/recruitment/CandidatesView.vue:101`-`113`).
- Manual verification note: hover/click/timing behaviors require manual UI run.

## 5. Issues / Suggestions (Severity-Rated)

### Blocker / High

1) **Severity: Blocker**  
**Title:** Recruitment core capabilities are missing (import, merge, match, recommendations)  
**Conclusion:** Fail  
**Evidence:** `docs/api-spec.md:122`-`135`; `repo/apps/api/internal/httpserver/server.go:78`-`88`; `repo/apps/api/internal/handler/recruitment.go:23`-`308`  
**Impact:** Prompt’s recruitment objective is not met; delivery cannot satisfy core business workflow.  
**Minimum actionable fix:** Implement missing recruitment endpoints/services/repos/UI flows for import batches, duplicate merge, score breakdown/reasons, and similar candidate/position recommendations.

2) **Severity: High**  
**Title:** PII encryption-at-rest behavior is not implemented in write/read logic  
**Conclusion:** Fail  
**Evidence:** requirement in `docs/design.md:375`; candidate schema has encrypted columns `repo/infra/db/migrations/000001_initial_schema.up.sql:150`-`153`; model fields exist `repo/apps/api/internal/model/recruitment.go:15`-`18`; no AES/encrypt/decrypt usage in services creating/updating candidates (`repo/apps/api/internal/service/recruitment_service.go:198`-`253`).  
**Impact:** Sensitive-data protection requirement is unmet; compliance/security risk is material.  
**Minimum actionable fix:** Add encryption/decryption component (AES-256 key-managed), encrypt writes for phone/ID/email, provide scoped masked/full read policy and auditing.

3) **Severity: High**  
**Title:** Data-scope enforcement is institution-only; department/team scope not enforced  
**Conclusion:** Fail  
**Evidence:** principal helper reduces scope to institution IDs only (`repo/apps/api/internal/access/principal.go:49`-`66`); services use only `AllowedInstitutionIDs` checks (e.g., `repo/apps/api/internal/service/recruitment_service.go:26`-`35`, `compliance_service.go:66`-`75`, `case_service.go:101`-`110`).  
**Impact:** Users may access broader data than allowed by department/team scope constraints from prompt.  
**Minimum actionable fix:** propagate full scope predicates (institution + optional department/team) into repository queries/mutations and object-level checks.

4) **Severity: High**  
**Title:** Audit log non-repudiation scope is not implemented for business mutations  
**Conclusion:** Fail  
**Evidence:** prompt requires mutation diff logs (`docs/design.md:228`-`233`, `535`-`560`); implementation only lists/exports logs (`repo/apps/api/internal/handler/audit.go:23`-`109`) and creates export tasks (`repo/apps/api/internal/service/audit_service.go:120`-`141`); no mutation handlers/services write `audit_logs`.  
**Impact:** Permission and domain field changes are not reliably traceable; non-repudiation objective is not met.  
**Minimum actionable fix:** add centralized audit writer invoked on RBAC/recruitment/compliance/case/file high-impact mutations with before/after diffs and request source metadata.

5) **Severity: High**  
**Title:** Static test coverage is insufficient for security and core business risks  
**Conclusion:** Fail  
**Evidence:** tests are mostly constructors/helpers (`repo/apps/api/internal/service/service_constructors_test.go:9`, `repo/apps/api/internal/repository/repo_constructors_test.go:5`, `repo/apps/api/internal/model/*_test.go`); API/E2E scripts are smoke checks (`repo/API_tests/run_api_tests.sh:13`-`110`, `repo/e2e_tests/run_e2e_tests.sh:13`-`55`).  
**Impact:** Severe defects in authz/scope/business rules can remain undetected while tests still pass.  
**Minimum actionable fix:** add targeted unit/integration suites for 401/403/object-scope, duplicate guards, restriction policy branches, file chunk edge cases, and audit-write assertions.

6) **Severity: High**  
**Title:** Frontend create flows are hardcoded to dev seed identifiers  
**Conclusion:** Partial Fail  
**Evidence:** `repo/apps/web/src/config/devSeed.ts:1`-`12`; create APIs send `DEV_INSTITUTION_ID` (`repo/apps/web/src/views/recruitment/CandidatesView.vue:86`, `PositionsView.vue:77`, `QualificationsView.vue:98`, `RestrictionsView.vue:277`, `CasesView.vue:142`).  
**Impact:** Real multi-institution deployments are not supported by UI without code changes; weak fit for production acceptance.  
**Minimum actionable fix:** derive institution/department/team context from authenticated scope selection instead of hardcoded constants.

### Medium / Low

7) **Severity: Medium**  
**Title:** API contract drift in auth login response shape  
**Conclusion:** Partial Fail  
**Evidence:** spec expects token + expiresAt + user object (`docs/api-spec.md:67`-`78`); implementation returns only token/expiresAt (`repo/apps/api/internal/handler/auth.go:68`-`71`).  
**Impact:** frontend/API clients built to contract may break or require undocumented behavior.  
**Minimum actionable fix:** return contract-compliant user payload in login response or formally update contract.

8) **Severity: Medium**  
**Title:** File listing/download lacks data-scope isolation  
**Conclusion:** Suspected Risk  
**Evidence:** file list/get in service/repo are global (`repo/apps/api/internal/service/file_service.go:423`-`433`, `402`-`411`; `repo/apps/api/internal/repository/file_repo.go:122`-`133`, `82`-`89`) and do not filter by institution/scope.  
**Impact:** potential cross-scope file metadata/content exposure if references span institutions.  
**Minimum actionable fix:** enforce scope-aware file ownership/reference model and filter list/get/download by caller scope.

9) **Severity: Medium**  
**Title:** Secondary confirmation is inconsistent for high-impact actions  
**Conclusion:** Partial Fail  
**Evidence:** confirmations exist for some destructive actions (`repo/apps/web/src/views/recruitment/CandidatesView.vue:101`-`106`, `cases/CasesView.vue:242`-`245`), but not for toggling restrictions active state (`restrictions` switch invokes `toggleActive` directly at `repo/apps/web/src/views/compliance/RestrictionsView.vue:349`-`355`).  
**Impact:** accidental policy changes are easier than required by prompt UX rules.  
**Minimum actionable fix:** add confirmation dialogs for compliance rule activation/deactivation and other high-impact mutations.

## 6. Security Review Summary

- **Authentication entry points:** **Pass**  
  Evidence: login/logout/me routes and bcrypt/session-token validation (`repo/apps/api/internal/httpserver/server.go:70`-`77`, `repo/apps/api/internal/service/auth_service.go:55`, `63`, `79`-`102`).
- **Route-level authorization:** **Pass**  
  Evidence: protected groups with `SessionAuth`, `AccessContext`, `RequirePermission` (`repo/apps/api/internal/httpserver/server.go:72`-`141`; `repo/apps/api/internal/middleware/require_permission.go:13`-`27`).
- **Object-level authorization:** **Fail**  
  Evidence: object filtering keyed only by institution IDs, not department/team (`repo/apps/api/internal/access/principal.go:49`-`66`; representative repo filters `institution_id IN ?` in `recruitment_repo.go:23`, `compliance_repo.go:22`, `case_repo.go:70`).
- **Function-level authorization:** **Partial Pass**  
  Evidence: permission middleware covers endpoints; however mutation-level compliance controls like audited diff logging are absent.
- **Tenant / user data isolation:** **Fail**  
  Evidence: scope granularity reduced to institution; file module lacks scope-bound list/get (`file_service.go:423`-`433`).
- **Admin / internal / debug protection:** **Partial Pass**  
  Evidence: RBAC endpoints protected by `system.rbac` (`server.go:128`-`140`); health endpoints are unauthenticated (`server.go:63`-`70`) and should be manually reviewed against deployment policy.

## 7. Tests and Logging Review

- **Unit tests:** **Fail**  
  Mostly constructor/table-name/helper checks, minimal business-rule/security assertions (`repo/apps/api/internal/*_test.go`).
- **API / integration tests:** **Fail**  
  Shell scripts are smoke-level envelope checks and happy-path fetches; no negative authz/object scope cases (`repo/API_tests/run_api_tests.sh:13`-`110`).
- **Logging categories / observability:** **Partial Pass**  
  Request IDs and standardized envelopes exist (`response/envelope.go`, `middleware/requestid.go`), but operational/business logging depth is limited.
- **Sensitive-data leakage risk in logs/responses:** **Cannot Confirm Statistically**  
  Current code masks candidate PII display placeholder (`maskPII`, `recruitment_service.go:69`-`90`), but true leakage risk depends on runtime logging/config and unimplemented encryption/full-value retrieval paths.

## 8. Test Coverage Assessment (Static Audit)

### 8.1 Test Overview
- Unit tests exist under `repo/apps/api/internal/**/*_test.go`; runner: `repo/unit_tests/run_unit_tests.sh:7`.
- API/integration tests are shell scripts with curl assertions: `repo/API_tests/run_api_tests.sh:1`.
- E2E tests are shell-based HTTP/smoke checks, not browser workflow tests: `repo/e2e_tests/run_e2e_tests.sh:1`.
- Aggregated test entrypoint documented: `repo/run_tests.sh:24`, `repo/scripts/run_integrated_tests.sh:30`-`43`.
- Test commands are documented in README (`repo/README.md:21`-`33`), but execution depends on Docker.

### 8.2 Coverage Mapping Table

| Requirement / Risk Point | Mapped Test Case(s) | Key Assertion / Fixture / Mock | Coverage Assessment | Gap | Minimum Test Addition |
|---|---|---|---|---|---|
| Auth password bcrypt + session basics | `repo/apps/api/internal/middleware/auth_test.go:5` | Bearer token parsing only | insufficient | No login/session lifecycle assertions | Add service+handler tests for login success/failure, expiry, revoke |
| 401 unauthenticated handling | none meaningful | N/A | missing | No endpoint-level 401 tests | Add handler tests for missing/invalid token on protected routes |
| 403 permission authorization | none meaningful | N/A | missing | No `RequirePermission` behavior test | Add route tests with principal lacking permission |
| Data-scope isolation (institution/department/team) | none | N/A | missing | No scoped query/mutation tests | Add repo+handler integration tests for institution/dept/team constraints |
| Recruitment import/merge/matching | none | N/A | missing | Core prompt flow untested/mostly unimplemented | Add tests once endpoints/services are implemented |
| Compliance restriction checks (Rx + frequency) | none | N/A | missing | No branch tests for allowed/blocked outcomes | Add unit tests for `CheckPurchase` branch matrix |
| Case duplicate submit within 5 min | none (format-only test exists) | `case_number_design_test` checks formatting only | insufficient | Duplicate guard not covered | Add service test with repeated request window |
| File chunk upload/complete/dedup | none | N/A | missing | No resumable/chunk-missing/hash mismatch tests | Add integration tests with temp filesystem fixture |
| Audit append-only + field diffs | none | N/A | missing | No mutation-to-audit assertions | Add tests asserting audit entries emitted and immutable |
| Error envelope consistency | `repo/apps/api/internal/response/envelope_test.go:11` | status 200 only | basically covered | No error-path envelope assertions | Add tests for representative 4xx/5xx envelope shapes |

### 8.3 Security Coverage Audit
- **Authentication:** **Insufficient** — only token header parser + helper checks; login/session revoke/expiry behavior not tested.
- **Route authorization:** **Missing** — no tests proving forbidden access behavior (`403`) for permission gates.
- **Object-level authorization:** **Missing** — no tests for cross-scope object access denial.
- **Tenant/data isolation:** **Missing** — no tests for institution/department/team boundaries.
- **Admin/internal protection:** **Missing** — no tests for RBAC endpoint access control matrix.

Severe defects could remain undetected while current tests pass.

### 8.4 Final Coverage Judgment
- **Fail**
- Major risks covered: only minimal utility/smoke checks.
- Major risks uncovered: authz boundaries, scope isolation, core business rules, audit non-repudiation, and file integrity edge paths; therefore tests could pass while severe production defects remain.

## 9. Final Notes
- This report is static-only; runtime-dependent claims are intentionally not asserted as working.
- Findings were consolidated by root cause to avoid repetitive symptom listing.
- Highest-priority remediation should focus on missing core recruitment capabilities, scope-enforcement correctness, PII crypto implementation, and mutation audit logging with test coverage.
