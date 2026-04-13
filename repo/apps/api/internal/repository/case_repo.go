package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/model"
)

type CaseRepository struct {
	db *gorm.DB
}

func NewCaseRepository(db *gorm.DB) *CaseRepository {
	return &CaseRepository{db: db}
}

// GetDB exposes the underlying DB for transactions orchestrated in services.
func (r *CaseRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *CaseRepository) GetInstitutionCode(ctx context.Context, institutionID string) (string, error) {
	var code string
	err := r.db.WithContext(ctx).Raw(`SELECT code FROM institutions WHERE id = ? LIMIT 1`, institutionID).Scan(&code).Error
	if err != nil {
		return "", err
	}
	if code == "" {
		return "", gorm.ErrRecordNotFound
	}
	return code, nil
}

// AllocateCaseSerial increments the per-institution, per-day serial inside a transaction (caller supplies tx).
func (r *CaseRepository) AllocateCaseSerial(ctx context.Context, tx *gorm.DB, institutionID string, dayUTC time.Time) (uint32, error) {
	d := time.Date(dayUTC.Year(), dayUTC.Month(), dayUTC.Day(), 0, 0, 0, 0, time.UTC)
	if err := tx.WithContext(ctx).Exec(`
		INSERT INTO case_number_sequences (institution_id, sequence_date, last_serial)
		VALUES (?, ?, 1)
		ON DUPLICATE KEY UPDATE last_serial = last_serial + 1
	`, institutionID, d).Error; err != nil {
		return 0, err
	}
	var seq uint32
	if err := tx.WithContext(ctx).Raw(`
		SELECT last_serial FROM case_number_sequences
		WHERE institution_id = ? AND sequence_date = ?
	`, institutionID, d).Scan(&seq).Error; err != nil {
		return 0, err
	}
	return seq, nil
}

func (r *CaseRepository) RecentDuplicateCount(ctx context.Context, institutionIDs []string, hash string, since time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.CaseRecord{}).
		Where("institution_id IN ? AND duplicate_guard_hash = ? AND created_at >= ?", institutionIDs, hash, since).
		Count(&n).Error
	return n, err
}

func (r *CaseRepository) CreateCase(ctx context.Context, tx *gorm.DB, c *model.CaseRecord) error {
	return tx.WithContext(ctx).Create(c).Error
}

func caseListQuery(db *gorm.DB, institutionIDs []string, search, status string) *gorm.DB {
	q := db.Model(&model.CaseRecord{}).Where("institution_id IN ?", institutionIDs)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("(title LIKE ? OR case_number LIKE ? OR case_type LIKE ?)", like, like, like)
	}
	return q
}

func (r *CaseRepository) ListCases(ctx context.Context, institutionIDs []string, offset, limit int, orderClause, search, status string) ([]model.CaseRecord, int64, error) {
	base := caseListQuery(r.db.WithContext(ctx), institutionIDs, search, status)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.CaseRecord
	err := caseListQuery(r.db.WithContext(ctx), institutionIDs, search, status).
		Order(orderClause).
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *CaseRepository) GetCase(ctx context.Context, id string, institutionIDs []string) (*model.CaseRecord, error) {
	var row model.CaseRecord
	err := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", id, institutionIDs).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CaseRepository) UpdateCase(ctx context.Context, row *model.CaseRecord, institutionIDs []string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", row.ID, institutionIDs).
		Updates(map[string]interface{}{
			"title":        row.Title,
			"description":  row.Description,
			"department_id": row.DepartmentID,
			"team_id":      row.TeamID,
			"updated_at":   time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CaseRepository) UpdateCaseStatus(ctx context.Context, id string, institutionIDs []string, status string) error {
	res := r.db.WithContext(ctx).Model(&model.CaseRecord{}).
		Where("id = ? AND institution_id IN ?", id, institutionIDs).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CaseRepository) SetAssignee(ctx context.Context, id string, institutionIDs []string, assigneeID *string, newStatus string) error {
	updates := map[string]interface{}{
		"assignee_user_id": assigneeID,
		"updated_at":       time.Now().UTC(),
	}
	if newStatus != "" {
		updates["status"] = newStatus
	}
	res := r.db.WithContext(ctx).Model(&model.CaseRecord{}).
		Where("id = ? AND institution_id IN ?", id, institutionIDs).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CaseRepository) InsertAssignment(ctx context.Context, a *model.CaseAssignment) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *CaseRepository) ListProcessingRecords(ctx context.Context, caseID string, orderClause string) ([]model.CaseProcessingRecord, error) {
	var rows []model.CaseProcessingRecord
	err := r.db.WithContext(ctx).
		Where("case_id = ?", caseID).
		Order(orderClause).
		Find(&rows).Error
	return rows, err
}

func (r *CaseRepository) CreateProcessingRecord(ctx context.Context, rec *model.CaseProcessingRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *CaseRepository) ListStatusTransitions(ctx context.Context, caseID string, orderClause string) ([]model.CaseStatusTransition, error) {
	var rows []model.CaseStatusTransition
	err := r.db.WithContext(ctx).
		Where("case_id = ?", caseID).
		Order(orderClause).
		Find(&rows).Error
	return rows, err
}

func (r *CaseRepository) CreateStatusTransition(ctx context.Context, t *model.CaseStatusTransition) error {
	return r.db.WithContext(ctx).Create(t).Error
}
