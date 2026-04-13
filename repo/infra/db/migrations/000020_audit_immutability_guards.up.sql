SET NAMES utf8mb4;

-- Enforce append-only semantics for audit logs at DB level.
CREATE TRIGGER trg_audit_logs_block_update
BEFORE UPDATE ON audit_logs
FOR EACH ROW
SIGNAL SQLSTATE '45000'
  SET MESSAGE_TEXT = 'audit_logs is append-only: UPDATE is not allowed';

CREATE TRIGGER trg_audit_logs_block_delete
BEFORE DELETE ON audit_logs
FOR EACH ROW
SIGNAL SQLSTATE '45000'
  SET MESSAGE_TEXT = 'audit_logs is append-only: DELETE is not allowed';
