ALTER TABLE candidates DROP COLUMN custom_fields_json;

DROP INDEX idx_audit_logs_institution ON audit_logs;
ALTER TABLE audit_logs DROP COLUMN team_id;
ALTER TABLE audit_logs DROP COLUMN department_id;
ALTER TABLE audit_logs DROP COLUMN institution_id;
