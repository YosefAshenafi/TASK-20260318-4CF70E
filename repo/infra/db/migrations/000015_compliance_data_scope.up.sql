-- Nullable department/team anchors for data-scope predicates (design §10.2).

SET NAMES utf8mb4;

ALTER TABLE qualification_profiles
  ADD COLUMN department_id CHAR(36) NULL AFTER institution_id,
  ADD COLUMN team_id CHAR(36) NULL AFTER department_id,
  ADD KEY idx_qualification_department (department_id),
  ADD KEY idx_qualification_team (team_id),
  ADD CONSTRAINT fk_qualification_department FOREIGN KEY (department_id) REFERENCES departments (id),
  ADD CONSTRAINT fk_qualification_team FOREIGN KEY (team_id) REFERENCES teams (id);

ALTER TABLE purchase_restrictions
  ADD COLUMN department_id CHAR(36) NULL AFTER institution_id,
  ADD COLUMN team_id CHAR(36) NULL AFTER department_id,
  ADD KEY idx_restrictions_department (department_id),
  ADD KEY idx_restrictions_team (team_id),
  ADD CONSTRAINT fk_restrictions_department FOREIGN KEY (department_id) REFERENCES departments (id),
  ADD CONSTRAINT fk_restrictions_team FOREIGN KEY (team_id) REFERENCES teams (id);

ALTER TABLE restriction_violation_records
  ADD COLUMN department_id CHAR(36) NULL AFTER institution_id,
  ADD COLUMN team_id CHAR(36) NULL AFTER department_id,
  ADD KEY idx_violations_department (department_id),
  ADD KEY idx_violations_team (team_id),
  ADD CONSTRAINT fk_violations_department FOREIGN KEY (department_id) REFERENCES departments (id),
  ADD CONSTRAINT fk_violations_team FOREIGN KEY (team_id) REFERENCES teams (id);

ALTER TABLE compliance_purchase_records
  ADD COLUMN department_id CHAR(36) NULL AFTER institution_id,
  ADD COLUMN team_id CHAR(36) NULL AFTER department_id,
  ADD KEY idx_purchase_department (department_id),
  ADD KEY idx_purchase_team (team_id),
  ADD CONSTRAINT fk_purchase_department FOREIGN KEY (department_id) REFERENCES departments (id),
  ADD CONSTRAINT fk_purchase_team FOREIGN KEY (team_id) REFERENCES teams (id);

ALTER TABLE candidate_import_batches
  ADD COLUMN department_id CHAR(36) NULL AFTER institution_id,
  ADD COLUMN team_id CHAR(36) NULL AFTER department_id,
  ADD KEY idx_import_batches_department (department_id),
  ADD KEY idx_import_batches_team (team_id),
  ADD CONSTRAINT fk_import_batches_department FOREIGN KEY (department_id) REFERENCES departments (id),
  ADD CONSTRAINT fk_import_batches_team FOREIGN KEY (team_id) REFERENCES teams (id);
