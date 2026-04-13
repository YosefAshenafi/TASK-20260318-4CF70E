-- Demo recruitment rows for dev institution (optional visual data).

SET NAMES utf8mb4;

INSERT INTO candidates (id, institution_id, department_id, team_id, name, experience_years, education_level, created_at, updated_at)
VALUES
  (
    '30000000-0000-4000-8000-000000000001',
    '10000000-0000-4000-8000-000000000001',
    '10000000-0000-4000-8000-000000000002',
    '10000000-0000-4000-8000-000000000003',
    'Alex Rivera',
    6,
    'Bachelor',
    CURRENT_TIMESTAMP(3),
    CURRENT_TIMESTAMP(3)
  ),
  (
    '30000000-0000-4000-8000-000000000002',
    '10000000-0000-4000-8000-000000000001',
    NULL,
    NULL,
    'Jordan Chen',
    10,
    'Master',
    CURRENT_TIMESTAMP(3),
    CURRENT_TIMESTAMP(3)
  )
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO candidate_skills (id, candidate_id, skill_name) VALUES
  ('30000000-0000-4000-8000-000000000011', '30000000-0000-4000-8000-000000000001', 'GMP'),
  ('30000000-0000-4000-8000-000000000012', '30000000-0000-4000-8000-000000000001', 'QA'),
  ('30000000-0000-4000-8000-000000000013', '30000000-0000-4000-8000-000000000002', 'Regulatory')
ON DUPLICATE KEY UPDATE skill_name = VALUES(skill_name);

INSERT INTO candidate_tags (candidate_id, tag) VALUES
  ('30000000-0000-4000-8000-000000000001', 'priority'),
  ('30000000-0000-4000-8000-000000000002', 'referral')
ON DUPLICATE KEY UPDATE tag = tag;

INSERT INTO positions (id, institution_id, title, description, status, created_at, updated_at)
VALUES (
  '30000000-0000-4000-8000-000000000020',
  '10000000-0000-4000-8000-000000000001',
  'Clinical Quality Specialist',
  'Support GMP documentation and audit readiness.',
  'open',
  CURRENT_TIMESTAMP(3),
  CURRENT_TIMESTAMP(3)
)
ON DUPLICATE KEY UPDATE title = VALUES(title);
