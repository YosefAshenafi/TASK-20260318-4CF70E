-- Sample audit row for UI verification (append-only table; seed is idempotent).

SET NAMES utf8mb4;

INSERT INTO audit_logs (
  id, module, operation, operator_user_id, request_source, request_id,
  target_type, target_id, before_json, after_json, created_at
) VALUES (
  '60000000-0000-4000-8000-000000000001',
  'rbac',
  'seed.demo',
  '00000000-0000-4000-8000-000000000001',
  'demo-seed',
  NULL,
  'system',
  '10000000-0000-4000-8000-000000000020',
  NULL,
  JSON_OBJECT('note', 'Demo audit entry (design §7.6)'),
  CURRENT_TIMESTAMP(3)
)
ON DUPLICATE KEY UPDATE operation = VALUES(operation);
