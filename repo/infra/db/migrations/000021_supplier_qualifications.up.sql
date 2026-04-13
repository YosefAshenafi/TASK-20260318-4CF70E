-- Add supplier qualification support via party_type discriminator
-- party_type: 'client' (default) or 'supplier'
-- supplier_id: populated when party_type is 'supplier'

ALTER TABLE qualification_profiles
  ADD COLUMN party_type VARCHAR(16) NOT NULL DEFAULT 'client' AFTER client_id,
  ADD COLUMN supplier_id VARCHAR(64) NULL AFTER party_type,
  ADD INDEX idx_qualification_party (party_type),
  ADD INDEX idx_qualification_supplier (supplier_id);

ALTER TABLE qualification_profiles
  ADD CONSTRAINT chk_qualification_party_type CHECK (party_type IN ('client', 'supplier'));
