-- Add data-scope columns to audit_logs for scope-filtered queries
ALTER TABLE audit_logs ADD COLUMN institution_id CHAR(36) NULL AFTER operator_user_id;
ALTER TABLE audit_logs ADD COLUMN department_id CHAR(36) NULL AFTER institution_id;
ALTER TABLE audit_logs ADD COLUMN team_id CHAR(36) NULL AFTER department_id;
CREATE INDEX idx_audit_logs_institution ON audit_logs (institution_id);

-- Add custom_fields_json to candidates for structured extensible data
ALTER TABLE candidates ADD COLUMN custom_fields_json JSON NULL AFTER education_level;
