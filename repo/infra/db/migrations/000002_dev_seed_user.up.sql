-- Dev-only seed user (password: password). bcrypt hash matches golang.org/x/crypto/bcrypt default test vectors pattern.
-- Replace or remove in production deployments.

SET NAMES utf8mb4;

INSERT INTO users (id, username, password_hash, display_name, is_active, created_at, updated_at)
VALUES (
  '00000000-0000-4000-8000-000000000001',
  'admin',
  '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
  'Admin',
  1,
  CURRENT_TIMESTAMP(3),
  CURRENT_TIMESTAMP(3)
) ON DUPLICATE KEY UPDATE username = username;
