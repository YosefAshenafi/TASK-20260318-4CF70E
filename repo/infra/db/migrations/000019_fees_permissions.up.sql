SET NAMES utf8mb4;

INSERT INTO permissions (id, code, description, created_at) VALUES
  ('20000000-0000-4000-8000-000000000017', 'fees.view', 'View fees and billing records', CURRENT_TIMESTAMP(3)),
  ('20000000-0000-4000-8000-000000000018', 'fees.manage', 'Create and update fees and billing records', CURRENT_TIMESTAMP(3))
ON DUPLICATE KEY UPDATE description = VALUES(description);

-- system_admin
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT '10000000-0000-4000-8000-000000000020', id, CURRENT_TIMESTAMP(3)
FROM permissions
WHERE id IN (
  '20000000-0000-4000-8000-000000000017',
  '20000000-0000-4000-8000-000000000018'
)
ON DUPLICATE KEY UPDATE role_id = role_id;

-- business_specialist gets fees view/manage
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT '10000000-0000-4000-8000-000000000021', id, CURRENT_TIMESTAMP(3)
FROM permissions
WHERE id IN (
  '20000000-0000-4000-8000-000000000017',
  '20000000-0000-4000-8000-000000000018'
)
ON DUPLICATE KEY UPDATE role_id = role_id;
