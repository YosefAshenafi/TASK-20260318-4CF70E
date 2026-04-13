-- Dev seed: institution hierarchy and a base data scope only.
-- Security hardening: no privileged role/permission/user bindings are seeded.

SET NAMES utf8mb4;

INSERT INTO institutions (id, code, name, created_at, updated_at)
VALUES (
  '10000000-0000-4000-8000-000000000001',
  'dev',
  'Development Institution',
  CURRENT_TIMESTAMP(3),
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO departments (id, institution_id, name, created_at, updated_at)
VALUES (
  '10000000-0000-4000-8000-000000000002',
  '10000000-0000-4000-8000-000000000001',
  'General',
  CURRENT_TIMESTAMP(3),
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO teams (id, department_id, name, created_at, updated_at)
VALUES (
  '10000000-0000-4000-8000-000000000003',
  '10000000-0000-4000-8000-000000000002',
  'Default',
  CURRENT_TIMESTAMP(3),
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO data_scopes (id, scope_key, institution_id, department_id, team_id, created_at)
VALUES (
  '10000000-0000-4000-8000-000000000010',
  'inst:dev-root',
  '10000000-0000-4000-8000-000000000001',
  NULL,
  NULL,
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE scope_key = VALUES(scope_key);
