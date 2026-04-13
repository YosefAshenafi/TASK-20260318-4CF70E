-- Roll back initial schema (dev / reset). Disables FK checks for drop order safety.

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS audit_exports;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS case_attachment_indexes;
DROP TABLE IF EXISTS case_status_transitions;
DROP TABLE IF EXISTS case_processing_records;
DROP TABLE IF EXISTS case_assignments;
DROP TABLE IF EXISTS cases;
DROP TABLE IF EXISTS case_number_sequences;
DROP TABLE IF EXISTS restriction_violation_records;
DROP TABLE IF EXISTS file_references;
DROP TABLE IF EXISTS file_dedup_index;
DROP TABLE IF EXISTS file_chunks;
DROP TABLE IF EXISTS upload_sessions;
DROP TABLE IF EXISTS file_objects;
DROP TABLE IF EXISTS prescription_attachments;
DROP TABLE IF EXISTS qualification_documents;
DROP TABLE IF EXISTS qualification_expiration_jobs;
DROP TABLE IF EXISTS purchase_restrictions;
DROP TABLE IF EXISTS qualification_profiles;
DROP TABLE IF EXISTS candidate_import_batches;
DROP TABLE IF EXISTS match_score_snapshots;
DROP TABLE IF EXISTS candidate_merge_history;
DROP TABLE IF EXISTS position_requirements;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS candidate_tags;
DROP TABLE IF EXISTS candidate_custom_fields;
DROP TABLE IF EXISTS candidate_education;
DROP TABLE IF EXISTS candidate_experience;
DROP TABLE IF EXISTS candidate_skills;
DROP TABLE IF EXISTS candidate_contacts;
DROP TABLE IF EXISTS candidates;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_data_scopes;
DROP TABLE IF EXISTS data_scopes;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS institutions;

DELETE FROM schema_migrations WHERE version = '000001_initial_schema';

SET FOREIGN_KEY_CHECKS = 1;
