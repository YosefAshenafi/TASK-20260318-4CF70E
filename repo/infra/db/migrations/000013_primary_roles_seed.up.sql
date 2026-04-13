-- Align `roles` with design.md §1 primary personas (Business Specialist, Compliance Administrator,
-- Recruitment Specialist). System Administrator (`system_admin`) already exists in 000003.

SET NAMES utf8mb4;

INSERT INTO roles (id, slug, name, description, created_at, updated_at) VALUES
  (
    '10000000-0000-4000-8000-000000000021',
    'business_specialist',
    'Business Specialist',
    'Cross-module operations: cases and read access to recruitment/compliance (design.md §1).',
    CURRENT_TIMESTAMP(3),
    CURRENT_TIMESTAMP(3)
  ),
  (
    '10000000-0000-4000-8000-000000000022',
    'compliance_administrator',
    'Compliance Administrator',
    'Qualifications, restrictions, medications compliance (design.md §1).',
    CURRENT_TIMESTAMP(3),
    CURRENT_TIMESTAMP(3)
  ),
  (
    '10000000-0000-4000-8000-000000000023',
    'recruitment_specialist',
    'Recruitment Specialist',
    'Candidates, positions, matching (design.md §1).',
    CURRENT_TIMESTAMP(3),
    CURRENT_TIMESTAMP(3)
  )
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description);

-- business_specialist: dashboard, cases, recruitment + compliance read, files view
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT '10000000-0000-4000-8000-000000000021', id, CURRENT_TIMESTAMP(3)
FROM permissions
WHERE id IN (
  '20000000-0000-4000-8000-000000000001',
  '20000000-0000-4000-8000-000000000002',
  '20000000-0000-4000-8000-000000000004',
  '20000000-0000-4000-8000-000000000005',
  '20000000-0000-4000-8000-000000000006',
  '20000000-0000-4000-8000-000000000010'
)
ON DUPLICATE KEY UPDATE role_id = role_id;

-- compliance_administrator
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT '10000000-0000-4000-8000-000000000022', id, CURRENT_TIMESTAMP(3)
FROM permissions
WHERE id IN (
  '20000000-0000-4000-8000-000000000001',
  '20000000-0000-4000-8000-000000000004',
  '20000000-0000-4000-8000-000000000006',
  '20000000-0000-4000-8000-000000000007',
  '20000000-0000-4000-8000-000000000009',
  '20000000-0000-4000-8000-000000000011'
)
ON DUPLICATE KEY UPDATE role_id = role_id;

-- recruitment_specialist
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT '10000000-0000-4000-8000-000000000023', id, CURRENT_TIMESTAMP(3)
FROM permissions
WHERE id IN (
  '20000000-0000-4000-8000-000000000001',
  '20000000-0000-4000-8000-000000000002',
  '20000000-0000-4000-8000-000000000003',
  '20000000-0000-4000-8000-000000000006',
  '20000000-0000-4000-8000-000000000011'
)
ON DUPLICATE KEY UPDATE role_id = role_id;
