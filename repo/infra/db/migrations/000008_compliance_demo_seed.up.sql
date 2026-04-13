-- Demo qualification and restriction rows for the dev institution.

SET NAMES utf8mb4;

INSERT INTO qualification_profiles (
  id, institution_id, client_id, display_name, status, expires_on, deactivated_at, metadata_json, created_at, updated_at
) VALUES
  (
    '40000000-0000-4000-8000-000000000001',
    '10000000-0000-4000-8000-000000000001',
    'client-demo-1',
    'Acme Wholesale — Active',
    'active',
    DATE_ADD(CURDATE(), INTERVAL 45 DAY),
    NULL,
    JSON_OBJECT('tier', 'gold'),
    CURRENT_TIMESTAMP(3),
    CURRENT_TIMESTAMP(3)
  ),
  (
    '40000000-0000-4000-8000-000000000002',
    '10000000-0000-4000-8000-000000000001',
    'client-demo-2',
    'Beta Pharmacy — Expiring soon',
    'active',
    DATE_ADD(CURDATE(), INTERVAL 20 DAY),
    NULL,
    NULL,
    CURRENT_TIMESTAMP(3),
    CURRENT_TIMESTAMP(3)
  )
ON DUPLICATE KEY UPDATE display_name = VALUES(display_name);

INSERT INTO purchase_restrictions (
  id, institution_id, client_id, medication_id, rule_json, is_active, created_at, updated_at
) VALUES
  (
    '40000000-0000-4000-8000-000000000010',
    '10000000-0000-4000-8000-000000000001',
    'client-demo-1',
    'med-controlled-1',
    JSON_OBJECT('requiresPrescription', true, 'frequencyDays', 7),
    1,
    CURRENT_TIMESTAMP(3),
    CURRENT_TIMESTAMP(3)
  )
ON DUPLICATE KEY UPDATE rule_json = VALUES(rule_json);

INSERT INTO restriction_violation_records (
  id, restriction_id, institution_id, client_id, medication_id, case_id, details_json, created_at
) VALUES
  (
    '40000000-0000-4000-8000-000000000020',
    '40000000-0000-4000-8000-000000000010',
    '10000000-0000-4000-8000-000000000001',
    'client-demo-1',
    'med-controlled-1',
    NULL,
    JSON_OBJECT('reason', 'purchase already made within last 7 days'),
    CURRENT_TIMESTAMP(3)
  )
ON DUPLICATE KEY UPDATE details_json = VALUES(details_json);
