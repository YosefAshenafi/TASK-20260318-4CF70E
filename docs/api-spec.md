# PharmaOps Talent & Compliance Platform API Specification

## Runtime Reality (Important)

- PharmaOps Talent & Compliance Platform runs with a **real backend API** in offline intranet deployments.
- Frontend (`Vue 3 + TypeScript + Element Plus`) communicates with backend (`Go + Gin + GORM + MySQL`) over local network HTTP(S).
- Persistence is backend-owned: MySQL for relational data and local file system for attachments/import-export artifacts.
- This spec is the implementation target for backend endpoints and frontend integration.

## Source of Truth Used

- Product intent and constraints from `metadata.json`
- Assumptions and resolved open questions from `docs/questions.md`
- System design decisions from `docs/design.md`

## Contract Conventions

- Base URL prefix: `/api/v1`
- Authentication: `Authorization: Bearer <opaque-session-token>`
- IDs: string UUIDs (recommended), case number has business format
- Timestamps: ISO-8601 UTC strings (`YYYY-MM-DDTHH:mm:ssZ`)
- List APIs use `page`, `pageSize`, `sortBy`, `sortOrder`
- Protected endpoints enforce both:
  - RBAC permission check
  - institution/department/team data-scope check
- Standard response envelope:

```json
{
  "code": "OK",
  "message": "success",
  "requestId": "req_123",
  "data": {}
}
```

## Canonical DTO/Model Types

Primary contract entities:

- Identity and auth: `User`, `Role`, `Permission`, `DataScope`, `Session`
- Recruitment: `Candidate`, `Position`, `CandidateMergeRecord`, `MatchScore`
- Compliance: `QualificationProfile`, `PurchaseRestriction`, `RestrictionViolation`
- Case domain: `Case`, `CaseAssignment`, `CaseProcessingRecord`, `CaseStatusTransition`
- Files: `FileObject`, `UploadSession`, `FileChunk`, `FileReference`
- Auditing: `AuditLogEvent`, `AuditExportTask`

## Explicit REST Contract Groups

### Auth (`AuthController`)

Operations:

- `POST /auth/login`
- `GET /auth/me`
- `POST /auth/logout`

`POST /auth/login` request:

```json
{
  "username": "alice",
  "password": "StrongPassword123"
}
```

`POST /auth/login` response data:

```json
{
  "token": "opaque-session-token",
  "expiresAt": "2026-04-13T16:30:00Z",
  "user": {
    "id": "u_1",
    "username": "alice",
    "roles": ["recruitment_specialist"]
  }
}
```

`GET /auth/me` response data (authenticated session; includes resolved roles, permissions, and **data scopes** for UI defaults):

```json
{
  "id": "uuid-user",
  "username": "alice",
  "roles": ["recruitment_specialist"],
  "permissions": ["recruitment.view"],
  "scopes": [
    {
      "id": "uuid-scope-row",
      "scopeKey": "inst_acme",
      "institutionId": "uuid-inst",
      "departmentId": null,
      "teamId": null
    }
  ]
}
```

Typed errors (`AuthError.code`):

- `AUTH_INVALID_CREDENTIALS`
- `AUTH_PASSWORD_TOO_SHORT`
- `AUTH_ACCOUNT_DISABLED`
- `AUTH_SESSION_EXPIRED`
- `AUTH_SESSION_REVOKED`

### RBAC and Scope (`AccessController`)

Operations:

- `GET /users`
- `POST /users`
- `GET /users/{id}`
- `PATCH /users/{id}`
- `GET /roles`
- `POST /roles`
- `PATCH /roles/{id}`
- `POST /roles/{id}/permissions`
- `GET /scopes`
- `POST /scopes`
- `POST /users/{id}/scopes`

Typed errors (`AccessError.code`):

- `FORBIDDEN_PERMISSION`
- `FORBIDDEN_SCOPE`
- `USER_NOT_FOUND`
- `ROLE_NOT_FOUND`
- `SCOPE_NOT_FOUND`

### Recruitment (`RecruitmentController`)

Operations:

- `GET /recruitment/candidates`
- `POST /recruitment/candidates`
- `GET /recruitment/candidates/{id}`
- `PATCH /recruitment/candidates/{id}`
- `DELETE /recruitment/candidates/{id}` (soft delete recommended)
- `POST /recruitment/candidates/imports`
- `GET /recruitment/candidates/imports/{importId}`
- `POST /recruitment/candidates/imports/{importId}/commit`
- `GET /recruitment/candidates/duplicates`
- `POST /recruitment/candidates/merge`
- `GET /recruitment/candidates/merge-history`
- `GET /recruitment/positions`
- `POST /recruitment/positions`
- `GET /recruitment/positions/{id}`
- `PATCH /recruitment/positions/{id}`
- `POST /recruitment/match/candidate-to-position`
- `POST /recruitment/match/position-to-candidate`
- `GET /recruitment/recommendations/similar-candidates/{candidateId}`
- `GET /recruitment/recommendations/similar-positions/{positionId}`

Candidate DTO (response shape):

```ts
{
  id: string;
  name: string;
  phoneMasked: string;
  idNumberMasked: string;
  email?: string;
  skills: string[];
  experienceYears?: number;
  educationLevel?: string;
  tags: string[];
  customFields: Record<string, unknown>;
  institutionId: string;
  departmentId?: string;
  teamId?: string;
  createdAt: string;
  updatedAt: string;
}
```

Candidate list filters (`GET /recruitment/candidates`) include:

- `keyword`, `skills`, `educationLevel`, `minExperience`, `maxExperience`
- `createdFrom`, `createdTo`, `updatedFrom`, `updatedTo` (RFC3339)

Match score DTO:

```ts
{
  score: number; // 0-100
  breakdown: {
    skills: number;      // default 50%
    experience: number;  // default 30%
    education: number;   // default 20%
  };
  reasons: string[];
}
```

Merge request DTO:

```ts
{
  baseCandidateId: string;           // newest record by default
  sourceCandidateIds: string[];
  strategy: 'latest_wins_fill_missing';
  manualOverrides?: Record<string, unknown>;
}
```

Duplicate handling policy:

- Candidate create/import paths run deterministic duplicate auto-merge when normalized phone or ID matches.
- Strategy is `latest_wins_fill_missing` (newest row as base; fill missing fields; union skills/tags).
- Merge history is persisted in `candidate_merge_history` and an audit mutation is emitted.

Resume import staging request supports both structured rows and resume file ingestion:

```ts
{
  institutionId: string;
  rows?: Array<{
    name: string;
    phone?: string;
    idNumber?: string;
    email?: string;
    skills?: string[];
    tags?: string[];
    customFields?: Record<string, unknown>;
  }>;
  resumeFileIds?: string[]; // uploaded file IDs from /files/uploads/*
}
```

The import batch response includes `validationReport` with extracted rows, errors, and warnings. Commit imports only valid staged rows.

Candidate PATCH request supports structured updates for contact and profile fields:

```ts
{
  name?: string;
  departmentId?: string | null;
  teamId?: string | null;
  phone?: string;
  idNumber?: string;
  email?: string;
  experienceYears?: number;
  educationLevel?: string;
  skills?: string[];
  tags?: string[];
  customFields?: Record<string, unknown>;
}
```

Typed errors (`RecruitmentError.code`):

- `CANDIDATE_NOT_FOUND`
- `POSITION_NOT_FOUND`
- `DUPLICATE_CANDIDATE_CONFLICT`
- `MERGE_VALIDATION_FAILED`
- `IMPORT_VALIDATION_FAILED`
- `FORBIDDEN_SCOPE`

### Compliance (`ComplianceController`)

Operations:

- `GET /compliance/qualifications`
- `POST /compliance/qualifications`
- `GET /compliance/qualifications/{id}`
- `PATCH /compliance/qualifications/{id}`
- `POST /compliance/qualifications/{id}/activate`
- `POST /compliance/qualifications/{id}/deactivate`
- `GET /compliance/qualifications/expiring`
- `POST /compliance/jobs/qualifications/run` (manual admin trigger)
- `GET /compliance/restrictions`
- `POST /compliance/restrictions`
- `PATCH /compliance/restrictions/{id}`
- `POST /compliance/restrictions/check-purchase`
- `GET /compliance/restrictions/violations`

Restriction check request:

```json
{
  "clientId": "client_1",
  "medicationId": "med_1",
  "prescriptionAttachmentId": "file_1",
  "purchaseAt": "2026-04-13T09:00:00Z"
}
```

Prescription checks are enforced from server-side restriction rules and cannot be bypassed by client-provided control flags.

Restriction check response data:

```json
{
  "allowed": false,
  "reasons": [
    "purchase already made within last 7 days"
  ]
}
```

Typed errors (`ComplianceError.code`):

- `QUALIFICATION_NOT_FOUND`
- `QUALIFICATION_EXPIRED`
- `RESTRICTION_VIOLATION`
- `PRESCRIPTION_REQUIRED`
- `FORBIDDEN_SCOPE`

### Case Management (`CaseController`)

Operations:

- `GET /cases`
- `POST /cases`
- `GET /cases/{id}`
- `PATCH /cases/{id}`
- `POST /cases/{id}/assign`
- `POST /cases/{id}/processing-records`
- `GET /cases/{id}/processing-records`
- `POST /cases/{id}/status-transitions`
- `GET /cases/{id}/status-transitions`
- `GET /cases/{id}/attachments`
- `POST /cases/{id}/attachments`
- `DELETE /cases/{id}/attachments/{fileId}`
- `GET /case-ledger/search`

Case create DTO:

```ts
{
  institutionId: string;
  departmentId?: string;
  teamId?: string;
  caseType: string;
  title: string;
  description: string;
  reportedAt: string;
}
```

Case response DTO:

```ts
{
  id: string;
  caseNumber: string; // YYYYMMDD-{institution}-{6-digit-serial}
  status: 'submitted' | 'assigned' | 'in_progress' | 'pending_review' | 'closed';
  assigneeId?: string;
  createdAt: string;
  updatedAt: string;
}
```

Typed errors (`CaseError.code`):

- `CASE_NOT_FOUND`
- `CASE_MANDATORY_FIELDS_MISSING`
- `DUPLICATE_SUBMISSION_BLOCKED` // duplicate within 5 minutes
- `INVALID_STATUS_TRANSITION`
- `FORBIDDEN_SCOPE`

### Files and Attachments (`FileController`)

Operations:

- `POST /files/uploads/init`
- `PUT /files/uploads/{uploadId}/chunks/{chunkIndex}`
- `POST /files/uploads/{uploadId}/complete`
- `GET /files/uploads/{uploadId}`
- `GET /files/{fileId}`
- `GET /files/{fileId}/download`
- `POST /files/{fileId}/link`

Upload init DTO:

```ts
{
  fileName: string;
  size: number;
  mimeType: string;
  chunkSize: number;
}
```

Upload complete response data:

```ts
{
  fileId: string;
  sha256: string;
  deduplicated: boolean;
}
```

Typed errors (`FileError.code`):

- `FILE_TYPE_NOT_ALLOWED`
- `FILE_SIZE_EXCEEDED`
- `FILE_CHUNK_MISSING`
- `FILE_HASH_MISMATCH`
- `FILE_NOT_FOUND`
- `FORBIDDEN_SCOPE`

### Audit (`AuditController`)

Operations:

- `GET /audit/logs`
- `POST /audit/logs/export`

Audit log DTO:

```ts
{
  id: string;
  module: 'rbac' | 'recruitment' | 'compliance' | 'cases' | 'fees';
  operation: string;
  operatorId: string;
  requestSource: string;
  targetType: string;
  targetId: string;
  before?: Record<string, unknown>;
  after?: Record<string, unknown>;
  createdAt: string;
}
```

Typed errors (`AuditError.code`):

- `FORBIDDEN_PERMISSION`
- `EXPORT_VALIDATION_FAILED`

## System/Internal Operation Catalog (Background)

These are system-level operations, typically scheduler/job driven:

- `QualificationJob.runDailyExpirationSweep() -> deactivates expired qualifications`
- `UploadJob.cleanupStaleChunks() -> removes expired temporary chunks`
- `SessionJob.cleanupExpiredSessions() -> removes expired sessions`

These are not public user-facing endpoints except optional admin manual triggers.

## Security and PII Contract Notes

- Passwords are stored as bcrypt hashes; never returned by APIs.
- `phone` and `idNumber` are encrypted at rest with AES-256.
- List responses return masked PII; full PII requires explicit elevated permission.
- Audit list/export responses must not expose plaintext candidate PII to users without `recruitment.view_pii`.
- All permission changes and key business field changes must emit append-only audit logs.

## Persistence Contract Baseline

Primary relational tables (minimum):

- auth and access: `users`, `roles`, `permissions`, `user_roles`, `role_permissions`, `data_scopes`, `user_data_scopes`, `sessions`
- recruitment: `candidates`, `positions`, `candidate_merge_history`, `match_score_snapshots`
- compliance: `qualification_profiles`, `purchase_restrictions`, `restriction_violation_records`
- case ledger: `cases`, `case_assignments`, `case_processing_records`, `case_status_transitions`
- file system index: `file_objects`, `file_chunks`, `file_dedup_index`, `file_references`
- auditing: `audit_logs`, `audit_exports`

## Error Code Baseline (Cross-Module)

- `VALIDATION_ERROR`
- `FORBIDDEN_PERMISSION`
- `FORBIDDEN_SCOPE`
- `NOT_FOUND`
- `CONFLICT`
- `RATE_LIMITED`
- `INTERNAL_ERROR`

Module-specific typed codes are preferred over generic codes where available.

## Integration Note

For the avoidance of doubt: **these are intended live backend HTTP contracts for PharmaOps Talent & Compliance Platform**, not simulated frontend-only service contracts.

Frontend should integrate against these contracts through typed API clients, preserving:

- RBAC + data-scope consistency
- deterministic error handling per `*.Error.code`
- auditable high-impact operations

