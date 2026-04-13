SET NAMES utf8mb4;

ALTER TABLE candidates
  DROP INDEX idx_candidates_phone_norm_hash,
  DROP INDEX idx_candidates_id_number_norm_hash,
  DROP COLUMN phone_norm_hash,
  DROP COLUMN id_number_norm_hash;
