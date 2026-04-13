-- PharmaOps initial schema (identity, recruitment, compliance, cases, files, audit)
-- MySQL 8.x, utf8mb4. PII columns use VARBINARY for AES-256 ciphertext at rest.

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- Organizations (data-scope anchors)
-- ---------------------------------------------------------------------------
CREATE TABLE institutions (
  id CHAR(36) NOT NULL,
  code VARCHAR(32) NOT NULL,
  name VARCHAR(255) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_institutions_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE departments (
  id CHAR(36) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  name VARCHAR(255) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_departments_institution (institution_id),
  CONSTRAINT fk_departments_institution FOREIGN KEY (institution_id) REFERENCES institutions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE teams (
  id CHAR(36) NOT NULL,
  department_id CHAR(36) NOT NULL,
  name VARCHAR(255) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_teams_department (department_id),
  CONSTRAINT fk_teams_department FOREIGN KEY (department_id) REFERENCES departments (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- Identity & RBAC
-- ---------------------------------------------------------------------------
CREATE TABLE users (
  id CHAR(36) NOT NULL,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(255) NOT NULL,
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE roles (
  id CHAR(36) NOT NULL,
  slug VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(512) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_roles_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE permissions (
  id CHAR(36) NOT NULL,
  code VARCHAR(128) NOT NULL,
  description VARCHAR(512) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_permissions_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE user_roles (
  user_id CHAR(36) NOT NULL,
  role_id CHAR(36) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (user_id, role_id),
  KEY idx_user_roles_role (role_id),
  CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users (id),
  CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE role_permissions (
  role_id CHAR(36) NOT NULL,
  permission_id CHAR(36) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (role_id, permission_id),
  KEY idx_role_permissions_perm (permission_id),
  CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles (id),
  CONSTRAINT fk_role_permissions_perm FOREIGN KEY (permission_id) REFERENCES permissions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE data_scopes (
  id CHAR(36) NOT NULL,
  scope_key VARCHAR(160) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  department_id CHAR(36) NULL,
  team_id CHAR(36) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_data_scopes_key (scope_key),
  KEY idx_data_scopes_institution (institution_id),
  KEY idx_data_scopes_department (department_id),
  KEY idx_data_scopes_team (team_id),
  CONSTRAINT fk_data_scopes_institution FOREIGN KEY (institution_id) REFERENCES institutions (id),
  CONSTRAINT fk_data_scopes_department FOREIGN KEY (department_id) REFERENCES departments (id),
  CONSTRAINT fk_data_scopes_team FOREIGN KEY (team_id) REFERENCES teams (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE user_data_scopes (
  user_id CHAR(36) NOT NULL,
  data_scope_id CHAR(36) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (user_id, data_scope_id),
  KEY idx_user_data_scopes_scope (data_scope_id),
  CONSTRAINT fk_user_data_scopes_user FOREIGN KEY (user_id) REFERENCES users (id),
  CONSTRAINT fk_user_data_scopes_scope FOREIGN KEY (data_scope_id) REFERENCES data_scopes (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE sessions (
  id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  token_hash CHAR(64) NOT NULL,
  expires_at DATETIME(3) NOT NULL,
  revoked_at DATETIME(3) NULL,
  client_ip VARCHAR(45) NULL,
  user_agent VARCHAR(512) NULL,
  metadata JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_sessions_token_hash (token_hash),
  KEY idx_sessions_user (user_id),
  KEY idx_sessions_expires (expires_at),
  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- Recruitment
-- ---------------------------------------------------------------------------
CREATE TABLE candidates (
  id CHAR(36) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  department_id CHAR(36) NULL,
  team_id CHAR(36) NULL,
  name VARCHAR(255) NOT NULL,
  phone_enc VARBINARY(512) NULL,
  id_number_enc VARBINARY(512) NULL,
  email_enc VARBINARY(512) NULL,
  pii_key_version TINYINT UNSIGNED NOT NULL DEFAULT 1,
  experience_years INT NULL,
  education_level VARCHAR(128) NULL,
  deleted_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_candidates_institution (institution_id),
  KEY idx_candidates_deleted (deleted_at),
  CONSTRAINT fk_candidates_institution FOREIGN KEY (institution_id) REFERENCES institutions (id),
  CONSTRAINT fk_candidates_department FOREIGN KEY (department_id) REFERENCES departments (id),
  CONSTRAINT fk_candidates_team FOREIGN KEY (team_id) REFERENCES teams (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE candidate_contacts (
  id CHAR(36) NOT NULL,
  candidate_id CHAR(36) NOT NULL,
  contact_type VARCHAR(32) NOT NULL,
  value_enc VARBINARY(512) NULL,
  is_primary TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_candidate_contacts_candidate (candidate_id),
  CONSTRAINT fk_candidate_contacts_candidate FOREIGN KEY (candidate_id) REFERENCES candidates (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE candidate_skills (
  id CHAR(36) NOT NULL,
  candidate_id CHAR(36) NOT NULL,
  skill_name VARCHAR(128) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_candidate_skill (candidate_id, skill_name),
  CONSTRAINT fk_candidate_skills_candidate FOREIGN KEY (candidate_id) REFERENCES candidates (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE candidate_experience (
  id CHAR(36) NOT NULL,
  candidate_id CHAR(36) NOT NULL,
  employer VARCHAR(255) NULL,
  title VARCHAR(255) NULL,
  start_date DATE NULL,
  end_date DATE NULL,
  description TEXT NULL,
  PRIMARY KEY (id),
  KEY idx_candidate_experience_candidate (candidate_id),
  CONSTRAINT fk_candidate_experience_candidate FOREIGN KEY (candidate_id) REFERENCES candidates (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE candidate_education (
  id CHAR(36) NOT NULL,
  candidate_id CHAR(36) NOT NULL,
  institution_name VARCHAR(255) NULL,
  degree VARCHAR(128) NULL,
  field_of_study VARCHAR(255) NULL,
  graduation_year SMALLINT NULL,
  PRIMARY KEY (id),
  KEY idx_candidate_education_candidate (candidate_id),
  CONSTRAINT fk_candidate_education_candidate FOREIGN KEY (candidate_id) REFERENCES candidates (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE candidate_tags (
  candidate_id CHAR(36) NOT NULL,
  tag VARCHAR(64) NOT NULL,
  PRIMARY KEY (candidate_id, tag),
  CONSTRAINT fk_candidate_tags_candidate FOREIGN KEY (candidate_id) REFERENCES candidates (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE candidate_custom_fields (
  candidate_id CHAR(36) NOT NULL,
  field_key VARCHAR(128) NOT NULL,
  value_json JSON NULL,
  PRIMARY KEY (candidate_id, field_key),
  CONSTRAINT fk_candidate_custom_fields_candidate FOREIGN KEY (candidate_id) REFERENCES candidates (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE positions (
  id CHAR(36) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  department_id CHAR(36) NULL,
  team_id CHAR(36) NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'open',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_positions_institution (institution_id),
  CONSTRAINT fk_positions_institution FOREIGN KEY (institution_id) REFERENCES institutions (id),
  CONSTRAINT fk_positions_department FOREIGN KEY (department_id) REFERENCES departments (id),
  CONSTRAINT fk_positions_team FOREIGN KEY (team_id) REFERENCES teams (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE position_requirements (
  id CHAR(36) NOT NULL,
  position_id CHAR(36) NOT NULL,
  skill_name VARCHAR(128) NOT NULL,
  weight_pct TINYINT UNSIGNED NOT NULL DEFAULT 0,
  is_required TINYINT(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (id),
  KEY idx_position_requirements_position (position_id),
  CONSTRAINT fk_position_requirements_position FOREIGN KEY (position_id) REFERENCES positions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE candidate_merge_history (
  id CHAR(36) NOT NULL,
  base_candidate_id CHAR(36) NOT NULL,
  source_candidate_ids_json JSON NOT NULL,
  merged_fields_json JSON NULL,
  before_snapshot_json JSON NULL,
  after_snapshot_json JSON NULL,
  operator_user_id CHAR(36) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_merge_history_base (base_candidate_id),
  CONSTRAINT fk_merge_history_base FOREIGN KEY (base_candidate_id) REFERENCES candidates (id),
  CONSTRAINT fk_merge_history_operator FOREIGN KEY (operator_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE match_score_snapshots (
  id CHAR(36) NOT NULL,
  candidate_id CHAR(36) NOT NULL,
  position_id CHAR(36) NOT NULL,
  score SMALLINT UNSIGNED NOT NULL,
  breakdown_json JSON NULL,
  reasons_json JSON NULL,
  computed_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_match_candidate (candidate_id),
  KEY idx_match_position (position_id),
  CONSTRAINT fk_match_candidate FOREIGN KEY (candidate_id) REFERENCES candidates (id),
  CONSTRAINT fk_match_position FOREIGN KEY (position_id) REFERENCES positions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE candidate_import_batches (
  id CHAR(36) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  mapping_json JSON NULL,
  validation_report_json JSON NULL,
  created_by_user_id CHAR(36) NOT NULL,
  committed_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_import_batches_institution (institution_id),
  CONSTRAINT fk_import_batches_institution FOREIGN KEY (institution_id) REFERENCES institutions (id),
  CONSTRAINT fk_import_batches_user FOREIGN KEY (created_by_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- Compliance
-- ---------------------------------------------------------------------------
CREATE TABLE qualification_profiles (
  id CHAR(36) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  client_id VARCHAR(64) NOT NULL,
  display_name VARCHAR(255) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  expires_on DATE NULL,
  deactivated_at DATETIME(3) NULL,
  metadata_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_qualification_institution (institution_id),
  KEY idx_qualification_expires (expires_on),
  CONSTRAINT fk_qualification_institution FOREIGN KEY (institution_id) REFERENCES institutions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE qualification_documents (
  id CHAR(36) NOT NULL,
  qualification_id CHAR(36) NOT NULL,
  file_object_id CHAR(36) NOT NULL,
  doc_type VARCHAR(64) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_qual_docs_qual (qualification_id),
  CONSTRAINT fk_qual_docs_qual FOREIGN KEY (qualification_id) REFERENCES qualification_profiles (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE qualification_expiration_jobs (
  id CHAR(36) NOT NULL,
  run_at DATETIME(3) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  finished_at DATETIME(3) NULL,
  summary_json JSON NULL,
  PRIMARY KEY (id),
  KEY idx_qual_jobs_run (run_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE purchase_restrictions (
  id CHAR(36) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  client_id VARCHAR(64) NOT NULL,
  medication_id VARCHAR(64) NULL,
  rule_json JSON NOT NULL,
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_restrictions_institution (institution_id),
  CONSTRAINT fk_restrictions_institution FOREIGN KEY (institution_id) REFERENCES institutions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE prescription_attachments (
  id CHAR(36) NOT NULL,
  qualification_id CHAR(36) NOT NULL,
  file_object_id CHAR(36) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_rx_attach_qual (qualification_id),
  CONSTRAINT fk_rx_attach_qual FOREIGN KEY (qualification_id) REFERENCES qualification_profiles (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE restriction_violation_records (
  id CHAR(36) NOT NULL,
  restriction_id CHAR(36) NULL,
  institution_id CHAR(36) NOT NULL,
  client_id VARCHAR(64) NOT NULL,
  medication_id VARCHAR(64) NULL,
  case_id CHAR(36) NULL,
  details_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_violations_institution (institution_id),
  CONSTRAINT fk_violations_restriction FOREIGN KEY (restriction_id) REFERENCES purchase_restrictions (id),
  CONSTRAINT fk_violations_institution FOREIGN KEY (institution_id) REFERENCES institutions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- Files (created before FKs from compliance/cases that reference file_objects)
-- ---------------------------------------------------------------------------
CREATE TABLE file_objects (
  id CHAR(36) NOT NULL,
  sha256 CHAR(64) NOT NULL,
  size_bytes BIGINT UNSIGNED NOT NULL,
  mime_type VARCHAR(128) NULL,
  storage_path VARCHAR(1024) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_file_objects_sha256 (sha256)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE upload_sessions (
  id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  file_name VARCHAR(512) NOT NULL,
  total_size BIGINT UNSIGNED NOT NULL,
  chunk_size INT UNSIGNED NOT NULL,
  mime_type VARCHAR(128) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'initialized',
  merged_file_id CHAR(36) NULL,
  expires_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_upload_sessions_user (user_id),
  CONSTRAINT fk_upload_sessions_user FOREIGN KEY (user_id) REFERENCES users (id),
  CONSTRAINT fk_upload_sessions_file FOREIGN KEY (merged_file_id) REFERENCES file_objects (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE file_chunks (
  id CHAR(36) NOT NULL,
  upload_session_id CHAR(36) NOT NULL,
  chunk_index INT UNSIGNED NOT NULL,
  chunk_sha256 CHAR(64) NOT NULL,
  storage_path VARCHAR(1024) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_chunk_index (upload_session_id, chunk_index),
  CONSTRAINT fk_file_chunks_session FOREIGN KEY (upload_session_id) REFERENCES upload_sessions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE file_dedup_index (
  sha256 CHAR(64) NOT NULL,
  file_object_id CHAR(36) NOT NULL,
  PRIMARY KEY (sha256),
  KEY idx_dedup_file (file_object_id),
  CONSTRAINT fk_dedup_file FOREIGN KEY (file_object_id) REFERENCES file_objects (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE file_references (
  id CHAR(36) NOT NULL,
  file_object_id CHAR(36) NOT NULL,
  ref_type VARCHAR(64) NOT NULL,
  ref_id CHAR(36) NOT NULL,
  created_by_user_id CHAR(36) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_file_ref_target (ref_type, ref_id),
  CONSTRAINT fk_file_ref_object FOREIGN KEY (file_object_id) REFERENCES file_objects (id),
  CONSTRAINT fk_file_ref_user FOREIGN KEY (created_by_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add deferred FKs to file_objects
ALTER TABLE qualification_documents
  ADD CONSTRAINT fk_qual_docs_file FOREIGN KEY (file_object_id) REFERENCES file_objects (id);

ALTER TABLE prescription_attachments
  ADD CONSTRAINT fk_rx_attach_file FOREIGN KEY (file_object_id) REFERENCES file_objects (id);

-- ---------------------------------------------------------------------------
-- Cases
-- ---------------------------------------------------------------------------
CREATE TABLE case_number_sequences (
  institution_id CHAR(36) NOT NULL,
  sequence_date DATE NOT NULL,
  last_serial INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (institution_id, sequence_date),
  CONSTRAINT fk_case_seq_institution FOREIGN KEY (institution_id) REFERENCES institutions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE cases (
  id CHAR(36) NOT NULL,
  case_number VARCHAR(64) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  department_id CHAR(36) NULL,
  team_id CHAR(36) NULL,
  case_type VARCHAR(64) NOT NULL,
  title VARCHAR(512) NOT NULL,
  description TEXT NOT NULL,
  reported_at DATETIME(3) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'submitted',
  assignee_user_id CHAR(36) NULL,
  duplicate_guard_hash CHAR(64) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_cases_number (case_number),
  KEY idx_cases_institution (institution_id),
  KEY idx_cases_status (status),
  KEY idx_cases_duplicate_guard (institution_id, duplicate_guard_hash),
  CONSTRAINT fk_cases_institution FOREIGN KEY (institution_id) REFERENCES institutions (id),
  CONSTRAINT fk_cases_department FOREIGN KEY (department_id) REFERENCES departments (id),
  CONSTRAINT fk_cases_team FOREIGN KEY (team_id) REFERENCES teams (id),
  CONSTRAINT fk_cases_assignee FOREIGN KEY (assignee_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE restriction_violation_records
  ADD CONSTRAINT fk_violations_case FOREIGN KEY (case_id) REFERENCES cases (id);

CREATE TABLE case_assignments (
  id CHAR(36) NOT NULL,
  case_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  assigned_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_case_assignments_case (case_id),
  CONSTRAINT fk_case_assignments_case FOREIGN KEY (case_id) REFERENCES cases (id),
  CONSTRAINT fk_case_assignments_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE case_processing_records (
  id CHAR(36) NOT NULL,
  case_id CHAR(36) NOT NULL,
  step_code VARCHAR(64) NOT NULL,
  actor_user_id CHAR(36) NOT NULL,
  note TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_case_proc_case (case_id),
  CONSTRAINT fk_case_proc_case FOREIGN KEY (case_id) REFERENCES cases (id),
  CONSTRAINT fk_case_proc_actor FOREIGN KEY (actor_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE case_status_transitions (
  id CHAR(36) NOT NULL,
  case_id CHAR(36) NOT NULL,
  from_status VARCHAR(32) NOT NULL,
  to_status VARCHAR(32) NOT NULL,
  actor_user_id CHAR(36) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_case_status_case (case_id),
  CONSTRAINT fk_case_status_case FOREIGN KEY (case_id) REFERENCES cases (id),
  CONSTRAINT fk_case_status_actor FOREIGN KEY (actor_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE case_attachment_indexes (
  id CHAR(36) NOT NULL,
  case_id CHAR(36) NOT NULL,
  file_object_id CHAR(36) NOT NULL,
  purpose VARCHAR(64) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_case_attach_case (case_id),
  CONSTRAINT fk_case_attach_case FOREIGN KEY (case_id) REFERENCES cases (id),
  CONSTRAINT fk_case_attach_file FOREIGN KEY (file_object_id) REFERENCES file_objects (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- Audit (append-only writer enforced in application layer)
-- ---------------------------------------------------------------------------
CREATE TABLE audit_logs (
  id CHAR(36) NOT NULL,
  module VARCHAR(32) NOT NULL,
  operation VARCHAR(128) NOT NULL,
  operator_user_id CHAR(36) NOT NULL,
  request_source VARCHAR(255) NULL,
  request_id VARCHAR(64) NULL,
  target_type VARCHAR(64) NOT NULL,
  target_id CHAR(36) NOT NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_audit_created (created_at),
  KEY idx_audit_operator (operator_user_id),
  KEY idx_audit_target (target_type, target_id),
  CONSTRAINT fk_audit_operator FOREIGN KEY (operator_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE audit_exports (
  id CHAR(36) NOT NULL,
  requested_by_user_id CHAR(36) NOT NULL,
  filter_json JSON NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  output_file_path VARCHAR(1024) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  completed_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_audit_exports_user (requested_by_user_id),
  CONSTRAINT fk_audit_exports_user FOREIGN KEY (requested_by_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;
