-- Deterministic duplicate-key hashes for phone and ID number.
-- Keeps ciphertext-at-rest while enabling reliable duplicate grouping.

SET NAMES utf8mb4;

ALTER TABLE candidates
  ADD COLUMN phone_norm_hash CHAR(64) NULL AFTER phone_enc,
  ADD COLUMN id_number_norm_hash CHAR(64) NULL AFTER id_number_enc,
  ADD KEY idx_candidates_phone_norm_hash (institution_id, phone_norm_hash, deleted_at),
  ADD KEY idx_candidates_id_number_norm_hash (institution_id, id_number_norm_hash, deleted_at);
