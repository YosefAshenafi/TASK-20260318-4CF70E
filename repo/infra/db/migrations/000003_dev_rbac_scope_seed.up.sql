-- Dev seed: institution hierarchy, data scope, system_admin role, full_access permission, admin bindings.

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

INSERT INTO roles (id, slug, name, description, created_at, updated_at)
VALUES (
  '10000000-0000-4000-8000-000000000020',
  'system_admin',
  'System Administrator',
  'Full application access (dev seed)',
  CURRENT_TIMESTAMP(3),
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO permissions (id, code, description, created_at)
VALUES (
  '10000000-0000-4000-8000-000000000030',
  'system.full_access',
  'Bypass permission checks for all route-level gates (dev)',
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE description = VALUES(description);

INSERT INTO role_permissions (role_id, permission_id, created_at)
VALUES (
  '10000000-0000-4000-8000-000000000020',
  '10000000-0000-4000-8000-000000000030',
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE role_id = role_id;

INSERT INTO user_roles (user_id, role_id, created_at)
VALUES (
  '00000000-0000-4000-8000-000000000001',
  '10000000-0000-4000-8000-000000000020',
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE user_id = user_id;

INSERT INTO user_data_scopes (user_id, data_scope_id, created_at)
VALUES (
  '00000000-0000-4000-8000-000000000001',
  '10000000-0000-4000-8000-000000000010',
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE user_id = user_id;
