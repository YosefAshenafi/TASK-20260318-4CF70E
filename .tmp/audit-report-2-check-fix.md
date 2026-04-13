# PharmaOps Audit Report 2 - Check Fix

## Verdict

- **Overall conclusion: Pass**
- Scope of this re-check is limited to the five issues listed in `audit-report-2.md`.

## Fix Validation (Issue-by-Issue)

### 1) Default privileged admin seed (High) - Fixed

- **Status:** Fixed
- **What changed:**
  - Removed default credentialed user seed from migration.
  - Removed seeded `system.full_access` role/permission binding to default user.
  - Updated README to state that no default credentialed user is seeded.
- **Evidence:**
  - `repo/infra/db/migrations/000002_dev_seed_user.up.sql:1`
  - `repo/infra/db/migrations/000003_dev_rbac_scope_seed.up.sql:1`
  - `repo/README.md:5`

### 2) Resume extraction format-naive for PDF/DOCX (High) - Fixed

- **Status:** Fixed
- **What changed:**
  - Replaced raw binary regex parsing path with format-aware extraction behavior.
  - Added DOCX text extraction path via ZIP/XML content extraction.
  - Added deterministic validation failure for unsupported structured extraction formats (e.g. PDF/DOC).
- **Evidence:**
  - `repo/apps/api/internal/service/recruitment_extended.go:22`
  - `repo/apps/api/internal/service/recruitment_extended.go:143`
  - `repo/apps/api/internal/service/recruitment_extended.go:247`

### 3) Health endpoint hardening optional/default-open (Medium) - Fixed

- **Status:** Fixed
- **What changed:**
  - `/api/v1/health` now requires configured token and rejects calls when token is missing.
  - Removed unauthenticated `/healthz` endpoint registration.
  - Updated env and README documentation to mark health token as required.
- **Evidence:**
  - `repo/apps/api/internal/handler/health.go:21`
  - `repo/apps/api/internal/httpserver/server.go:79`
  - `repo/.env.example:31`
  - `repo/README.md:42`

### 4) Missing substantive auth/session lifecycle tests (Medium) - Fixed

- **Status:** Fixed
- **What changed:**
  - Added integration tests covering:
    - login success
    - session token validation
    - logout invalidation
    - short password rejection
    - invalid credential rejection
    - expired session rejection
- **Evidence:**
  - `repo/apps/api/internal/service/auth_service_integration_test.go:49`
  - `repo/apps/api/internal/service/auth_service_integration_test.go:78`

### 5) No frontend automated tests (Medium) - Fixed

- **Status:** Fixed
- **What changed:**
  - Added frontend test runner (`vitest`) and `npm test` script.
  - Added baseline frontend unit tests for auth permission logic and data-scope default selection.
  - Added `vitest` config.
- **Evidence:**
  - `repo/apps/web/package.json:6`
  - `repo/apps/web/vitest.config.ts:1`
  - `repo/apps/web/src/stores/auth.test.ts:1`
  - `repo/apps/web/src/utils/dataScope.test.ts:1`

## Final Re-check Conclusion

- All five issues from `audit-report-2.md` have been addressed with direct code/documentation changes.
- Re-check outcome for the listed issue set is **Pass**.
