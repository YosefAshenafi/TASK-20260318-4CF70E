-- Route-level permissions for recruitment + other UI modules (explicit codes alongside system.full_access).

SET NAMES utf8mb4;

INSERT INTO permissions (id, code, description, created_at) VALUES
  ('20000000-0000-4000-8000-000000000001', 'dashboard.view', 'View dashboard', CURRENT_TIMESTAMP(3)),
  ('20000000-0000-4000-8000-000000000002', 'recruitment.view', 'View recruitment data', CURRENT_TIMESTAMP(3)),
  ('20000000-0000-4000-8000-000000000003', 'recruitment.manage', 'Create/update recruitment records', CURRENT_TIMESTAMP(3)),
  ('20000000-0000-4000-8000-000000000004', 'compliance.view', 'View compliance data', CURRENT_TIMESTAMP(3)),
  ('20000000-0000-4000-8000-000000000005', 'cases.view', 'View cases', CURRENT_TIMESTAMP(3)),
  ('20000000-0000-4000-8000-000000000006', 'files.view', 'View files', CURRENT_TIMESTAMP(3)),
  ('20000000-0000-4000-8000-000000000007', 'audit.view', 'View audit logs', CURRENT_TIMESTAMP(3)),
  ('20000000-0000-4000-8000-000000000008', 'system.rbac', 'Manage RBAC', CURRENT_TIMESTAMP(3))
ON DUPLICATE KEY UPDATE description = VALUES(description);

INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT '10000000-0000-4000-8000-000000000020', id, CURRENT_TIMESTAMP(3)
FROM permissions
WHERE id IN (
  '20000000-0000-4000-8000-000000000001',
  '20000000-0000-4000-8000-000000000002',
  '20000000-0000-4000-8000-000000000003',
  '20000000-0000-4000-8000-000000000004',
  '20000000-0000-4000-8000-000000000005',
  '20000000-0000-4000-8000-000000000006',
  '20000000-0000-4000-8000-000000000007',
  '20000000-0000-4000-8000-000000000008'
)
ON DUPLICATE KEY UPDATE role_id = role_id;
