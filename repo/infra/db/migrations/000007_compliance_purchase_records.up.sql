-- Tracks approved purchase checks for frequency enforcement (per client/medication window).

SET NAMES utf8mb4;

CREATE TABLE compliance_purchase_records (
  id CHAR(36) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  client_id VARCHAR(64) NOT NULL,
  medication_id VARCHAR(64) NULL,
  recorded_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_cpr_lookup (institution_id, client_id, medication_id, recorded_at),
  CONSTRAINT fk_cpr_institution FOREIGN KEY (institution_id) REFERENCES institutions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
