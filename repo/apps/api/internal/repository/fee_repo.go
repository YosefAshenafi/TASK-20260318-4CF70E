package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
)

type FeeRepository struct {
	db *gorm.DB
}

func NewFeeRepository(db *gorm.DB) *FeeRepository {
	return &FeeRepository{db: db}
}

func (r *FeeRepository) ListFees(ctx context.Context, p *access.Principal, offset, limit int, orderClause string) ([]model.FeeRecord, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.FeeRecord{})
	base = applyDataScope(base, p, "institution_id", "department_id", "team_id")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.FeeRecord
	q := r.db.WithContext(ctx).Model(&model.FeeRecord{})
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	err := q.Order(orderClause).Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *FeeRepository) GetFee(ctx context.Context, id string, p *access.Principal) (*model.FeeRecord, error) {
	var row model.FeeRecord
	q := r.db.WithContext(ctx).Where("id = ?", id)
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	if err := q.First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FeeRepository) CreateFee(ctx context.Context, row *model.FeeRecord) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *FeeRepository) UpdateFee(ctx context.Context, row *model.FeeRecord, p *access.Principal) error {
	q := r.db.WithContext(ctx).Model(&model.FeeRecord{}).Where("id = ?", row.ID)
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	res := q.Updates(map[string]any{
		"fee_type":           row.FeeType,
		"amount":             row.Amount,
		"currency":           row.Currency,
		"note":               row.Note,
		"updated_by_user_id": row.UpdatedByUserID,
		"updated_at":         time.Now().UTC(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
