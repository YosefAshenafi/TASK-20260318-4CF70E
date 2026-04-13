SET NAMES utf8mb4;

CREATE TABLE fees (
  id CHAR(36) NOT NULL,
  institution_id CHAR(36) NOT NULL,
  department_id CHAR(36) NULL,
  team_id CHAR(36) NULL,
  case_id CHAR(36) NULL,
  candidate_id CHAR(36) NULL,
  fee_type VARCHAR(64) NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  currency VARCHAR(8) NOT NULL DEFAULT 'CNY',
  note TEXT NULL,
  created_by_user_id CHAR(36) NOT NULL,
  updated_by_user_id CHAR(36) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_fees_scope (institution_id, department_id, team_id, created_at),
  KEY idx_fees_case (case_id),
  KEY idx_fees_candidate (candidate_id),
  CONSTRAINT fk_fees_institution FOREIGN KEY (institution_id) REFERENCES institutions (id),
  CONSTRAINT fk_fees_department FOREIGN KEY (department_id) REFERENCES departments (id),
  CONSTRAINT fk_fees_team FOREIGN KEY (team_id) REFERENCES teams (id),
  CONSTRAINT fk_fees_case FOREIGN KEY (case_id) REFERENCES cases (id),
  CONSTRAINT fk_fees_candidate FOREIGN KEY (candidate_id) REFERENCES candidates (id),
  CONSTRAINT fk_fees_created_by FOREIGN KEY (created_by_user_id) REFERENCES users (id),
  CONSTRAINT fk_fees_updated_by FOREIGN KEY (updated_by_user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
