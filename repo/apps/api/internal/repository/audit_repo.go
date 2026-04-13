package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/model"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func auditLogQuery(db *gorm.DB, module, targetType string, from, to *time.Time) *gorm.DB {
	q := db.Model(&model.AuditLog{})
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at < ?", *to)
	}
	return q
}

func (r *AuditRepository) ListLogs(ctx context.Context, offset, limit int, orderClause, module, targetType string, from, to *time.Time) ([]model.AuditLog, int64, error) {
	base := auditLogQuery(r.db.WithContext(ctx), module, targetType, from, to)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.AuditLog
	err := auditLogQuery(r.db.WithContext(ctx), module, targetType, from, to).
		Order(orderClause).
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *AuditRepository) CreateExport(ctx context.Context, e *model.AuditExport) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// CreateAuditLog appends one immutable audit row (append-only store).
func (r *AuditRepository) CreateAuditLog(ctx context.Context, a *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(a).Error
}
