-- Demo case + sequence row so live-created cases continue serials after this date bucket.

SET NAMES utf8mb4;

INSERT INTO case_number_sequences (institution_id, sequence_date, last_serial)
VALUES ( '10000000-0000-4000-8000-000000000001', '2099-12-31', 1 )
ON DUPLICATE KEY UPDATE last_serial = GREATEST(last_serial, VALUES(last_serial));

INSERT INTO cases (
  id, case_number, institution_id, department_id, team_id,
  case_type, title, description, reported_at, status,
  assignee_user_id, duplicate_guard_hash, created_at, updated_at
) VALUES (
  '50000000-0000-4000-8000-000000000001',
  '20991231-dev-000001',
  '10000000-0000-4000-8000-000000000001',
  NULL,
  NULL,
  'quality',
  'Demo: packaging observation',
  'Demo seed case for ledger review (design §7.4).',
  CURRENT_TIMESTAMP(3),
  'submitted',
  NULL,
  NULL,
  CURRENT_TIMESTAMP(3),
  CURRENT_TIMESTAMP(3)
)
ON DUPLICATE KEY UPDATE title = VALUES(title);

INSERT INTO case_processing_records (id, case_id, step_code, actor_user_id, note, created_at)
SELECT
  '50000000-0000-4000-8000-000000000011',
  '50000000-0000-4000-8000-000000000001',
  'intake',
  u.id,
  'Recorded from demo seed.',
  CURRENT_TIMESTAMP(3)
FROM users u
WHERE u.id = '00000000-0000-4000-8000-000000000001'
ON DUPLICATE KEY UPDATE step_code = VALUES(step_code);
