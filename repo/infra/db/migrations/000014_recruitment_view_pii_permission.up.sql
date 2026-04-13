-- Full candidate PII read (design.md §11.2) — explicit permission separate from recruitment.view.

SET NAMES utf8mb4;

INSERT INTO permissions (id, code, description, created_at) VALUES
  ('20000000-0000-4000-8000-000000000014', 'recruitment.view_pii', 'View full candidate PII (phone, government ID, email)', CURRENT_TIMESTAMP(3))
ON DUPLICATE KEY UPDATE description = VALUES(description);

INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT '10000000-0000-4000-8000-000000000020', id, CURRENT_TIMESTAMP(3)
FROM permissions
WHERE id = '20000000-0000-4000-8000-000000000014'
ON DUPLICATE KEY UPDATE role_id = role_id;

INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT '10000000-0000-4000-8000-000000000023', id, CURRENT_TIMESTAMP(3)
FROM permissions
WHERE id = '20000000-0000-4000-8000-000000000014'
ON DUPLICATE KEY UPDATE role_id = role_id;
