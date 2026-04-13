# PharmaOps Static Delivery Acceptance & Architecture Audit

## 1. Verdict

- **Overall conclusion: Partial Pass**
- Core full-stack capabilities are implemented and traceable, but there are material gaps against the prompt/security bar:
  - **High:** purchase restriction check accepts any existing file ID (no scope/ownership authorization on prescription attachment).
  - **High:** compliance frequency rule is configurable/optional, while prompt requires a strict once-per-7-days control.
  - **Medium:** audit export omits request source and before/after field diffs, weakening non-repudiation/export intent.

## 2. Scope and Static Verification Boundary

- **Reviewed**
  - docs/contracts: `docs/design.md`, `docs/api-spec.md`, `repo/README.md`, `repo/.env.example`
  - backend: bootstrapping, routing, middleware, handlers, services, repositories, models, migrations
  - frontend: route/layout and major module views for recruitment/compliance/cases/files/audit/RBAC
  - tests: Go test files under `repo/apps/api/internal`, Vitest files under `repo/apps/web/src`
- **Not reviewed**
  - runtime behavior under real deployment, browser interaction, network timing, long-running scheduler behavior in production
- **Intentionally not executed**
  - project startup, Docker, tests, external services (per static-only boundary)
- **Manual verification required**
  - end-user visual polish/accessibility/responsiveness
  - runtime scheduler cadence and ops behavior over long uptime

## 3. Repository / Requirement Mapping Summary

- **Prompt core goals mapped:** offline intranet pharma compliance + recruitment platform; RBAC + institution/department/team scope; secure auth/session; PII encryption/masking; resumable file upload + dedup; case ledger; append-only audit.
- **Primary implementation areas mapped:** API route/middleware matrix in `repo/apps/api/internal/httpserver/server.go`; domain logic in service/repository layers; schema/contracts in `repo/infra/db/migrations`; UI modules in `repo/apps/web/src/views`.

## 4. Section-by-section Review

### 1. Hard Gates

#### 1.1 Documentation and static verifiability
- **Conclusion: Pass**
- **Rationale:** repo has clear setup/env/test docs and statically consistent entrypoints/routes.
- **Evidence:** `repo/README.md:39`, `repo/README.md:54`, `repo/apps/api/cmd/api/main.go:12`, `repo/apps/api/internal/httpserver/server.go:79`, `repo/apps/web/src/router/index.ts:6`

#### 1.2 Material deviation from prompt
- **Conclusion: Partial Pass**
- **Rationale:** platform is centered on the prompt domains, but compliance control semantics diverge on 7-day purchase-limit strictness.
- **Evidence:** prompt-aligned requirement in `docs/design.md:463`; implementation uses arbitrary `frequencyDays` and enforces only when `> 0` in `repo/apps/api/internal/service/compliance_service.go:16`, `repo/apps/api/internal/service/compliance_service.go:726`

### 2. Delivery Completeness

#### 2.1 Coverage of explicit core requirements
- **Conclusion: Partial Pass**
- **Rationale:** most core requirements are implemented (RBAC/scope, auth/session, recruitment merge/match/recommendations, compliance qualification/restriction checks, case numbering/duplicate guard, file chunk/dedup, audit append-only), but key compliance restriction semantics and attachment authorization are materially weak.
- **Evidence:** RBAC/scope routes `repo/apps/api/internal/httpserver/server.go:85`; auth/session `repo/apps/api/internal/service/auth_service.go:49`; case controls `repo/apps/api/internal/service/case_service.go:123`, `repo/apps/api/internal/service/case_service.go:163`; upload/dedup `repo/apps/api/internal/service/file_service.go:130`, `repo/apps/api/internal/service/file_service.go:291`; audit DB append-only trigger `repo/infra/db/migrations/000020_audit_immutability_guards.up.sql:4`

#### 2.2 0→1 deliverable completeness
- **Conclusion: Pass**
- **Rationale:** complete backend/frontend/migrations/scripts/docs structure; not a fragment/demo-only snippet.
- **Evidence:** `repo/apps/api/internal/httpserver/server.go:90`, `repo/apps/web/src/views/recruitment/CandidatesView.vue:401`, `repo/infra/db/migrations/000001_initial_schema.up.sql:10`, `repo/README.md:22`

### 3. Engineering and Architecture Quality

#### 3.1 Structure and module decomposition
- **Conclusion: Pass**
- **Rationale:** clean layered composition (router/handler/service/repository), modular domains, clear route ownership.
- **Evidence:** `repo/apps/api/internal/httpserver/server.go:40`, `repo/apps/api/internal/service/recruitment_service.go:36`, `repo/apps/api/internal/repository/recruitment_repo.go:46`

#### 3.2 Maintainability/extensibility
- **Conclusion: Partial Pass**
- **Rationale:** generally extensible, but compliance rule model is overly permissive relative to required fixed policy.
- **Evidence:** flexible rule JSON and optional checks `repo/apps/api/internal/service/compliance_service.go:16`, `repo/apps/api/internal/service/compliance_service.go:533`, `repo/apps/api/internal/service/compliance_service.go:726`

### 4. Engineering Details and Professionalism

#### 4.1 Error handling, logging, validation, API design
- **Conclusion: Partial Pass**
- **Rationale:** consistent error envelopes and structured logs are present, but security-critical validation/authorization gaps remain in compliance check flow.
- **Evidence:** envelope/error handling `repo/apps/api/internal/response/envelope.go:11`; structured ops logs `repo/apps/api/internal/oplog/oplog.go:12`; insecure Rx attachment existence-only check `repo/apps/api/internal/service/compliance_service.go:693`, `repo/apps/api/internal/repository/file_repo.go:229`

#### 4.2 Product/service realism
- **Conclusion: Pass**
- **Rationale:** implementation resembles a real product service (multi-domain CRUD/workflows, role/scope gates, persistence, exports, UI modules).
- **Evidence:** `repo/apps/api/internal/httpserver/server.go:112`, `repo/apps/api/internal/httpserver/server.go:128`, `repo/apps/web/src/views/cases/CasesView.vue:448`

### 5. Prompt Understanding and Requirement Fit

#### 5.1 Business/constraint fit
- **Conclusion: Partial Pass**
- **Rationale:** broad fit is strong, but two requirement-semantic mismatches remain:
  - 7-day purchase frequency is not enforced as a fixed policy.
  - prescription attachment authorization is not scope-aware.
- **Evidence:** fixed 7-day requirement text `docs/design.md:464`; dynamic optional enforcement `repo/apps/api/internal/service/compliance_service.go:726`; attachment check without scope/owner authorization `repo/apps/api/internal/service/compliance_service.go:694`, `repo/apps/api/internal/repository/file_repo.go:229`

### 6. Aesthetics (frontend/full-stack)

#### 6.1 Visual/interaction quality
- **Conclusion: Cannot Confirm Statistically**
- **Rationale:** static code shows coherent layout/theme and interaction feedback patterns, but final visual quality requires live rendering.
- **Evidence:** layout/theme `repo/apps/web/src/layouts/AppLayout.vue:83`; confirmation+toast patterns in high-impact flows `repo/apps/web/src/views/recruitment/CandidatesView.vue:255`, `repo/apps/web/src/views/cases/CasesView.vue:351`, `repo/apps/web/src/views/compliance/QualificationsView.vue:223`
- **Manual verification note:** verify cross-page visual consistency/interaction states in browser.

## 5. Issues / Suggestions (Severity-Rated)

### High

1) **Severity:** High  
**Title:** Prescription attachment authorization bypass in purchase checks  
**Conclusion:** Fail  
**Evidence:** `repo/apps/api/internal/service/compliance_service.go:693`, `repo/apps/api/internal/repository/file_repo.go:229`  
**Impact:** purchase approval can be satisfied using any existing file ID, without verifying the file is accessible to the caller’s data scope/ownership; weakens controlled-medication guardrails.  
**Minimum actionable fix:** replace existence-only check with scope-aware authorization (e.g., `IsFileObjectAccessible`-style check using principal + actor), and reject attachments outside caller scope.

2) **Severity:** High  
**Title:** Prompt-required 7-day purchase limit is not strictly enforced  
**Conclusion:** Fail  
**Evidence:** `docs/design.md:464`, `repo/apps/api/internal/service/compliance_service.go:16`, `repo/apps/api/internal/service/compliance_service.go:726`, `repo/apps/web/src/views/compliance/RestrictionsView.vue:454`  
**Impact:** system can persist and enforce non-7-day (or zero-day) frequency values, violating explicit business rule semantics for controlled/prescription medications.  
**Minimum actionable fix:** enforce fixed `frequencyDays = 7` for the relevant medication-restriction class in backend validation; disallow disabling/overriding this guard where prompt requires mandatory policy.

### Medium

3) **Severity:** Medium  
**Title:** Audit export omits field diffs and request source  
**Conclusion:** Partial Fail  
**Evidence:** export columns only include id/module/op/operator/target/time in `repo/apps/api/internal/repository/audit_repo.go:132`; diff/source fields exist in stored schema `repo/apps/api/internal/model/audit.go:13`, `repo/apps/api/internal/model/audit.go:17`  
**Impact:** exported audit artifacts do not carry key non-repudiation details required for compliance review (before/after changes, request source).  
**Minimum actionable fix:** include `request_source`, `request_id`, and serialized `before_json`/`after_json` in export output (CSV or richer format).

4) **Severity:** Medium  
**Title:** Static test coverage misses critical purchase-attachment authorization abuse path  
**Conclusion:** Partial Fail  
**Evidence:** existing compliance tests validate existence/frequency behavior but not scope/ownership authorization on prescription attachment IDs `repo/apps/api/internal/service/compliance_purchase_enforcement_integration_test.go:16`, `repo/apps/api/internal/service/compliance_check_purchase_test.go:42`  
**Impact:** severe data-scope bypass risk could remain undetected while tests pass.  
**Minimum actionable fix:** add negative tests proving purchase is denied when attachment exists but is outside caller-accessible scope/ownership.

## 6. Security Review Summary

- **Authentication entry points — Pass**  
  Evidence: login/logout/me endpoints and middleware chain `repo/apps/api/internal/httpserver/server.go:82`, `repo/apps/api/internal/middleware/auth.go:27`; bcrypt + opaque token hashing in `repo/apps/api/internal/service/auth_service.go:64`, `repo/apps/api/internal/service/auth_service.go:76`.

- **Route-level authorization — Pass**  
  Evidence: permission guards per protected route `repo/apps/api/internal/httpserver/server.go:90`; permission middleware denies missing permission `repo/apps/api/internal/middleware/require_permission.go:23`.

- **Object-level authorization — Partial Pass**  
  Evidence: repository scope predicates for core entities (cases/recruitment/compliance) `repo/apps/api/internal/repository/case_repo.go:112`, `repo/apps/api/internal/repository/recruitment_repo.go:121`, `repo/apps/api/internal/repository/compliance_repo.go:38`; **gap** in Rx attachment validation path (`exists` only) `repo/apps/api/internal/service/compliance_service.go:694`.

- **Function-level authorization — Partial Pass**  
  Evidence: service-level `requireScope` + `RowVisible` guards `repo/apps/api/internal/service/recruitment_service.go:355`, `repo/apps/api/internal/service/compliance_service.go:665`, `repo/apps/api/internal/service/case_service.go:114`; **gap** for purchase-attachment access check as above.

- **Tenant/user data isolation — Partial Pass**  
  Evidence: centralized scope expression builder `repo/apps/api/internal/repository/scope_where.go:45`; file access utilities exist `repo/apps/api/internal/repository/file_repo.go:163`; **gap** because compliance check does not use these utilities for attachment IDs.

- **Admin/internal/debug protection — Partial Pass**  
  Evidence: RBAC admin endpoints behind `system.rbac` permission `repo/apps/api/internal/httpserver/server.go:160`; health endpoint relies on internal token header `repo/apps/api/internal/handler/health.go:26`.  
  Boundary: manual runtime network exposure review required.

## 7. Tests and Logging Review

- **Unit tests — Pass (with risk gaps)**
  - Broad Go unit/integration-style coverage across services/repos/middleware exists `repo/apps/api/internal/**/*_test.go`
  - Auth/session lifecycle has meaningful service integration tests `repo/apps/api/internal/service/auth_service_integration_test.go:51`

- **API/integration tests — Partial Pass**
  - Contract and route-guard tests exist, plus compliance/case/file/audit integration tests `repo/apps/api/internal/httpserver/auth_matrix_contract_test.go:8`, `repo/apps/api/internal/service/compliance_purchase_enforcement_integration_test.go:16`, `repo/apps/api/internal/service/file_upload_dedup_integration_test.go:17`
  - Missing targeted integration test for Rx attachment scope authorization gap.

- **Logging categories/observability — Pass**
  - Structured event categories for auth, authz, session, audit, PII access, crypto errors exist in `repo/apps/api/internal/oplog/oplog.go:36`.

- **Sensitive-data leakage risk in logs/responses — Partial Pass**
  - PII masking/sanitization in candidate/audit paths exists `repo/apps/api/internal/service/recruitment_service.go:241`, `repo/apps/api/internal/service/audit_service.go:159`
  - Risk remains via authorization bypass path in compliance attachment usage (not leakage per se, but sensitive control bypass).

## 8. Test Coverage Assessment (Static Audit)

### 8.1 Test Overview

- **Unit tests exist:** Yes (`repo/apps/api/internal/**/*_test.go`)
- **API/integration tests exist:** Yes (multiple `*_integration_test.go` + route contract tests)
- **Frontend tests exist:** Limited (2 Vitest files) `repo/apps/web/src/stores/auth.test.ts:1`, `repo/apps/web/src/utils/dataScope.test.ts:1`
- **Frameworks:** Go `testing` with SQLite in-memory fixtures; Vitest for frontend
- **Test entry points documented:** `repo/README.md:79`, `repo/apps/web/package.json:10`, `repo/apps/web/vitest.config.ts:10`
- **Documentation test commands:** present, but runtime-only scripts (not executed in this audit) `repo/run_tests.sh:9`

### 8.2 Coverage Mapping Table

| Requirement / Risk Point | Mapped Test Case(s) | Key Assertion / Fixture / Mock | Coverage Assessment | Gap | Minimum Test Addition |
|---|---|---|---|---|---|
| Auth login/session/logout lifecycle | `repo/apps/api/internal/service/auth_service_integration_test.go:51` | login success, token lookup, logout invalidation, expiry behavior | sufficient | handler-level auth matrix beyond basics | add HTTP-level auth contract tests for typed errors |
| Route authz middleware presence | `repo/apps/api/internal/httpserver/auth_matrix_contract_test.go:8` | required middleware/permission substrings in routes | basically covered | static substring checks only | add black-box endpoint authorization tests |
| Scope isolation | `repo/apps/api/internal/service/scope_isolation_test.go:12`, `repo/apps/api/internal/repository/case_scope_integration_test.go:15` | mismatched scope returns forbidden/filtered rows | basically covered | not all object references covered | add cross-resource OLA tests (file link + purchase attachment) |
| Case numbering + duplicate 5-minute guard | `repo/apps/api/internal/service/case_create_integration_test.go:18` | serial increment + duplicate block | sufficient | no concurrency stress/race checks | add concurrent serial allocation tests |
| Compliance Rx/frequency logic | `repo/apps/api/internal/service/compliance_purchase_enforcement_integration_test.go:16` | Rx required, frequency scopes | basically covered | does not test attachment scope/ownership authorization | add negative case with existing but unauthorized file ID |
| File chunk upload + dedup + spoof validation | `repo/apps/api/internal/service/file_upload_dedup_integration_test.go:17` | dedup true on repeat; MIME spoof rejected | sufficient | limited cross-scope download/link abuse coverage | add cross-scope file access tests |
| Audit mutation persistence/sanitization | `repo/apps/api/internal/service/audit_mutation_integration_test.go:16`, `repo/apps/api/internal/service/audit_log_persistence_test.go:14` | changed-fields, PII redaction, persistence | basically covered | export payload completeness not tested | add tests for export content columns and ownership checks |
| Frontend critical business flows | only auth/data-scope utility tests: `repo/apps/web/src/stores/auth.test.ts:6`, `repo/apps/web/src/utils/dataScope.test.ts:5` | permission helper/default scope selection | insufficient | no UI flow tests for recruitment/compliance/cases | add component/E2E smoke tests for core prompts |

### 8.3 Security Coverage Audit

- **Authentication:** basically covered (service integration present), but could use endpoint-level negative-path expansion.
- **Route authorization:** basically covered (middleware/route matrix checks).
- **Object-level authorization:** insufficient for prescription-attachment path (key high-risk gap untested).
- **Tenant/data isolation:** basically covered for primary entities; insufficient for cross-resource attachment usage in compliance check.
- **Admin/internal protection:** partial; no dedicated tests for health token behavior and export/ownership edge cases.

### 8.4 Final Coverage Judgment

- **Partial Pass**
- Major risks are covered for auth lifecycle, scope filtering, case duplicate guard, compliance core logic, file dedup, and audit mutation behavior.
- However, uncovered authorization edges (notably prescription attachment scope/ownership) mean severe defects could still pass current tests.

## 9. Final Notes

- This report is static-only and avoids runtime success claims.
- Findings are root-cause oriented and evidence-linked; repeated symptoms are merged.
- Manual verification should prioritize the high-severity compliance security/control gaps before acceptance.
