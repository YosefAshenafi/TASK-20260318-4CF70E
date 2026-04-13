SET NAMES utf8mb4;

ALTER TABLE candidate_import_batches
  DROP FOREIGN KEY fk_import_batches_team,
  DROP FOREIGN KEY fk_import_batches_department,
  DROP COLUMN team_id,
  DROP COLUMN department_id;

ALTER TABLE compliance_purchase_records
  DROP FOREIGN KEY fk_purchase_team,
  DROP FOREIGN KEY fk_purchase_department,
  DROP COLUMN team_id,
  DROP COLUMN department_id;

ALTER TABLE restriction_violation_records
  DROP FOREIGN KEY fk_violations_team,
  DROP FOREIGN KEY fk_violations_department,
  DROP COLUMN team_id,
  DROP COLUMN department_id;

ALTER TABLE purchase_restrictions
  DROP FOREIGN KEY fk_restrictions_team,
  DROP FOREIGN KEY fk_restrictions_department,
  DROP COLUMN team_id,
  DROP COLUMN department_id;

ALTER TABLE qualification_profiles
  DROP FOREIGN KEY fk_qualification_team,
  DROP FOREIGN KEY fk_qualification_department,
  DROP COLUMN team_id,
  DROP COLUMN department_id;
