# design.md

## 1. System Overview

PharmaOps Talent & Compliance Platform is an offline intranet integrated management platform for pharmaceutical compliance and talent operations.

Primary roles:

* Business Specialist
* Compliance Administrator
* Recruitment Specialist
* System Administrator

Core capabilities:

* RBAC-driven access with institution/department/team data scopes
* recruitment operations including bulk resume import, structured profile editing, duplicate merge, and intelligent matching
* compliance operations including qualification lifecycle management, expiration reminders, automatic deactivation, and medication purchase restrictions
* case management with unique numbering, mandatory field validation, duplicate submission protection, assignment, processing timeline, and searchable ledger
* secure attachment handling with resumable chunked uploads and SHA256 deduplication
* append-only, searchable, exportable audit logs for non-repudiation

The frontend is built with Vue 3 + TypeScript + Vite + Element Plus.  
The backend is built with Go + Gin + GORM + MySQL.  
Deployment targets a fully offline intranet environment with Docker Compose.

---

## 2. Design Goals

* Fully offline operation inside intranet environments
* Strong role and data-scope enforcement at both UI and API layers
* High integrity for compliance workflows and case ledger records
* Security-first handling of credentials, PII, files, and audit trails
* Explainable matching and search behavior for recruitment workflows
* Modular architecture that supports maintainability and future extension
* Deterministic and auditable behavior for all critical actions

---

## 3. Assumptions from Open Questions

The following assumptions from `docs/questions.md` are adopted as design constraints:

* frontend stack: Vue 3 + Vite
* component library: Element Plus
* frontend language: TypeScript
* data access layer: GORM
* session model: server-side stored random session token (not JWT-only auth)
* encryption algorithm for PII: AES-256
* candidate scoring: weighted by skills, experience, education with explainable output
* duplicate candidate merge: newest record as base, backfill missing fields from older record, log merge in audit trail
* qualification expiration deactivation: scheduled background job (daily)
* deployment: Docker Compose for app + database (+ optional Nginx)

---

## 4. High-Level Architecture

```text
Vue 3 Web UI (TypeScript, Element Plus)
          ↓ HTTPS/HTTP (Intranet)
Go Gin REST API
          ↓
Domain Services + Authorization Layer
          ↓
Repositories (GORM)
          ↓
MySQL + Local File System
```

Supporting runtime services:

* Auth & Session Service
* RBAC + Data Scope Policy Engine
* Candidate Matching/Search Service
* Qualification Expiration Scheduler
* File Upload/Chunk Merge Service
* Audit Logging Service (append-only)

### Architecture principle

All business rules are enforced in the backend service layer.  
The frontend handles presentation, validation hints, and user interaction flow, but cannot bypass backend authorization and scope checks.

### Offline intranet principle

All dependencies must run locally in intranet.  
No cloud-managed dependencies are required for normal operation.

---

## 5. Frontend Architecture

### 5.1 Framework and Tooling

* Vue 3
* TypeScript
* Vite
* Vue Router
* Pinia (state management)
* Element Plus (admin UI components)

### 5.2 Route Areas

* `/login`
* `/dashboard`
* `/recruitment/candidates`
* `/recruitment/positions`
* `/compliance/qualifications`
* `/compliance/restrictions`
* `/cases`
* `/files`
* `/audit-logs`
* `/system/rbac`

### 5.3 UI Design Rules

* role-aware menu rendering based on authenticated permissions
* data-scope aware filters and default query constraints
* all destructive or high-impact actions require secondary confirmation dialog
* operations return clear inline + toast feedback (success/failure/partial)
* sensitive list fields use masking by default

### 5.4 Major UI Components

* candidate import wizard (mapping, preview, validation result)
* candidate profile editor with custom fields and tags
* explainable match score panel
* qualification profile table with expiration warning highlights
* purchase restriction rule editor and violation history view
* case intake form with mandatory field validation
* case ledger timeline and status transition controls
* resumable upload widget with chunk progress and dedup status
* audit log query form and export dialog

---

## 6. Backend Architecture

### 6.1 Framework

* Go
* Gin for REST APIs
* GORM for ORM and transaction handling
* MySQL for relational persistence

### 6.2 Layering

* Router Layer: endpoint registration + middleware chain
* Handler Layer: request parsing, response shaping
* Service Layer: business rules, orchestration, policy checks
* Repository Layer: database access via GORM
* Infra Layer: file storage, crypto, scheduler, audit writer

### 6.3 API Style

* RESTful resource endpoints
* consistent response envelope (code, message, data, requestId)
* pagination/filtering/sorting conventions for list endpoints
* idempotency and anti-duplicate safeguards for critical writes

---

## 7. Core Domain Modules

### 7.1 Identity and Access

Responsibilities:

* user authentication (username/password with bcrypt verification)
* role assignment and role-permission mapping
* data scope assignment (institution/department/team)
* per-request role + data-scope secondary validation
* session issuance, validation, expiration, and logout invalidation

### 7.2 Recruitment

Responsibilities:

* candidate and position CRUD
* configurable tags and custom fields
* bulk resume import with structured extraction/mapping
* duplicate detection (phone/ID) and merge workflow
* intelligent search, filtering, sorting
* explainable 0-100 match scoring
* similar-candidate and similar-position recommendations

### 7.3 Compliance

Responsibilities:

* client/supplier qualification profile management
* expiration warning generation (default 30 days in advance)
* automatic deactivation upon expiration
* controlled/prescription medication purchase restrictions
* prescription attachment requirement checks
* frequency restriction checks (once per 7 days per client)
* restriction event retention

### 7.4 Case Management

Responsibilities:

* unique case numbering: `YYYYMMDD-{institution}-{6-digit serial}`
* mandatory key field validation
* duplicate submission blocking within 5 minutes
* assignment and ownership tracking
* processing record timeline
* attachment archive/index linking
* status transition workflow and searchable ledger

### 7.5 File and Attachment Management

Responsibilities:

* whitelist-based format validation
* resumable chunk upload and merge
* temporary offline chunk storage
* SHA256 fingerprint deduplication
* metadata indexing and secure retrieval

### 7.6 Audit and Compliance Logging

Responsibilities:

* append-only audit event writing
* permission change tracking
* field-level before/after diff logging for resumes/qualifications/cases/fees
* metadata capture: operator, timestamp, request source
* conditional search and export

---

## 8. Data Model Overview

### 8.1 Users and Authorization

Core tables:

* `users`
* `roles`
* `permissions`
* `user_roles`
* `role_permissions`
* `data_scopes`
* `user_data_scopes`
* `sessions`

### 8.2 Recruitment

Core tables:

* `candidates`
* `candidate_contacts`
* `candidate_skills`
* `candidate_experience`
* `candidate_education`
* `candidate_tags`
* `candidate_custom_fields`
* `positions`
* `position_requirements`
* `candidate_merge_history`
* `match_score_snapshots`

### 8.3 Compliance

Core tables:

* `qualification_profiles`
* `qualification_documents`
* `qualification_expiration_jobs`
* `purchase_restrictions`
* `prescription_attachments`
* `restriction_violation_records`

### 8.4 Case Ledger

Core tables:

* `cases`
* `case_assignments`
* `case_processing_records`
* `case_status_transitions`
* `case_attachment_indexes`

### 8.5 Files and Uploads

Core tables:

* `file_objects`
* `file_chunks`
* `file_dedup_index`
* `file_references`

### 8.6 Audit

Core tables:

* `audit_logs` (append-only)
* `audit_exports`

---

## 9. Authentication and Session Design

### 9.1 Login Model

* local username/password authentication
* minimum password length: 8
* password hash algorithm: bcrypt

### 9.2 Session Model

* random opaque session token issued on successful login
* token persisted server-side in `sessions` table
* default validity: 8 hours
* logout invalidates token immediately
* expired or revoked token rejected by auth middleware

### 9.3 Session Security Controls

* token rotation optional on sensitive operations
* device/source metadata stored with session
* idle timeout and absolute timeout supported by policy config

---

## 10. Authorization and Data Scope Enforcement

### 10.1 RBAC

Permission checks happen in two stages:

1. route-level permission gate
2. service-level secondary permission validation

### 10.2 Data Scope

Each query and mutation is constrained by user scope:

* institution
* department
* team

Scope is injected as a mandatory query predicate in repository methods and enforced for create/update/delete.

### 10.3 UI and API Consistency

Frontend hides unauthorized menus/actions, while backend remains the final authority.  
Unauthorized requests return standard forbidden responses and are logged.

---

## 11. Sensitive Data Protection

### 11.1 PII Encryption

Sensitive fields (e.g., ID number, phone) are encrypted at rest using AES-256 before DB storage.

### 11.2 Desensitization

List views return masked values by default (e.g., partial phone/ID display).  
Full value retrieval requires explicit permission and is audited.

### 11.3 Key Management

Encryption keys are loaded from local secure configuration (environment or secret file in intranet deployment) and rotated through controlled migration procedures.

---

## 12. Recruitment Matching and Recommendation Design

### 12.1 Search Inputs

Supported query dimensions:

* keywords
* skills
* education
* experience
* time ranges
* sortable fields

### 12.2 Match Score (0-100)

Default weighted formula:

* skills match: 50%
* experience alignment: 30%
* education alignment: 20%

Score output includes explainable reasons, such as:

* "3 required skills matched"
* "experience requirement met"
* "education level below preferred"

### 12.3 Similarity Recommendation

* similar candidates: vectorized feature overlap on skills/experience/domain tags
* similar positions: requirement profile similarity
* recommendation results are constrained by data scope and access policy

---

## 13. Duplicate Candidate Merge Design

### 13.1 Duplicate Trigger

Duplicates are detected when phone number or ID number matches existing records (after normalized comparison).

### 13.2 Merge Rule

* newest record becomes base record
* non-empty missing fields are backfilled from older record(s)
* conflicting fields follow "newest wins" unless manually overridden

### 13.3 Traceability

Each merge writes:

* source and target candidate IDs
* merged fields list
* before/after snapshots
* operator and timestamp

All merge events are appended to audit logs.

---

## 14. Qualification Lifecycle and Restriction Rules

### 14.1 Expiration Reminder

Default reminder threshold: 30 days before expiration.  
UI highlights near-expiry or expired records in red.

### 14.2 Auto-Deactivation

A scheduled daily job deactivates expired qualifications and records the action in audit logs.

### 14.3 Medication Purchase Restriction

For controlled/prescription medications:

* valid prescription attachment is mandatory
* purchase frequency limit: once every 7 days per client
* blocked attempts are recorded with reason and context

---

## 15. Case Ledger Workflow Design

### 15.1 Case Numbering

Format: `YYYYMMDD-{institution}-{6-digit serial}`  
Serial is generated atomically per institution per day.

### 15.2 Mandatory Fields

Case creation is blocked unless required fields are present and valid.

### 15.3 Duplicate Submission Guard

Submissions matching duplicate criteria within 5 minutes are blocked to prevent accidental replay.

### 15.4 Status Lifecycle

Example statuses:

* submitted
* assigned
* in_progress
* pending_review
* closed

Transitions are policy-controlled and logged with operator and timestamp.

### 15.5 Processing Records and Attachments

Each case keeps:

* assignment history
* processing notes/actions
* attachment index references

All records are searchable in the ledger interface.

---

## 16. File Upload and Storage Design

### 16.1 Storage Strategy

* local file system stores physical files
* MySQL stores metadata and references

### 16.2 Upload Validation

* extension/MIME whitelist enforcement
* size limit checks
* malware scanning hook point (optional on-prem integration)

### 16.3 Resumable Chunk Upload

* chunk metadata persisted in `file_chunks`
* interrupted uploads can resume from last confirmed chunk
* final merge operation validates chunk integrity and total hash

### 16.4 SHA256 Deduplication

* file fingerprint computed on final artifact
* existing fingerprint reuses file object reference
* duplicate upload is linked without storing another physical copy

---

## 17. Audit Logging and Non-Repudiation

### 17.1 Scope of Audited Events

* permission and role changes
* resume field modifications
* qualification field modifications
* case field modifications
* fee-related field modifications

### 17.2 Event Schema

Each event includes:

* event ID
* operator
* operation type
* request source
* timestamp
* target resource
* field diff (before/after)

### 17.3 Append-Only Policy

Audit records are immutable (no update/delete path in service APIs).  
Any export action is itself auditable.

---

## 18. Scheduler and Background Jobs

### 18.1 Job Types

* qualification expiration/deactivation (daily)
* reminder precompute and notification generation
* temporary chunk cleanup
* stale session cleanup

### 18.2 Execution Model

In single-instance mode, an internal scheduler runs in the Gin process.  
In multi-instance mode, a distributed lock strategy (DB-based) prevents duplicate execution.

---

## 19. Deployment and Environment Design

### 19.1 Target Deployment

Offline intranet deployment using Docker Compose.

### 19.2 Compose Services

* `web` (Vue static assets served by Nginx or internal static server)
* `api` (Go Gin backend)
* `db` (MySQL)
* optional `reverse-proxy` (Nginx for unified routing/TLS termination)

### 19.3 Persistence Mounts

* MySQL data volume
* local attachment storage volume
* temp chunk storage volume
* export/import directory volume

### 19.4 Configuration

Environment-driven configuration for:

* DB connection
* AES key source
* session TTL
* upload size/format whitelist
* scheduler toggles

---

## 20. Error Handling and UX Feedback Strategy

* all critical operations return explicit success/failure status
* actionable error messages are shown for validation failures
* secondary confirmation required for destructive/high-impact operations
* duplicate-submit and restriction violations return deterministic reason codes
* backend validation errors are standardized for frontend form mapping

---

## 21. Testing Strategy

### 21.1 Unit Tests

Focus areas:

* password/session validation
* RBAC and data scope policy checks
* match score calculation and explanation output
* duplicate merge rule behavior
* case number generation and duplicate-submit guard
* purchase restriction enforcement
* AES encrypt/decrypt helper correctness
* audit log immutability rules

### 21.2 Integration Tests

Focus areas:

* auth middleware + session lifecycle
* scoped query enforcement across modules
* file upload chunk/merge/dedup workflow
* scheduler-triggered deactivation behavior
* case ledger end-to-end transitions

### 21.3 UI / E2E Tests

Focus areas:

* login/logout and role-aware navigation
* recruitment import -> merge -> search -> scoring flow
* qualification warning and deactivation visibility
* prescription-required purchase flow
* case creation, assignment, processing, and ledger search
* audit log query and export operations

---

## 22. Future Evolution Considerations

* pluggable search engine upgrade for larger datasets
* configurable scoring profiles per institution or position type
* stronger key management integration with on-prem HSM/KMS if available
* reporting warehouse/export pipelines for BI tools in intranet
* workflow engine abstraction for more complex case/compliance processes

---
