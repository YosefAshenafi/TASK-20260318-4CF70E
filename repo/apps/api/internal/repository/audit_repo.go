package repository

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func auditLogQuery(db *gorm.DB, p *access.Principal, module, targetType string, from, to *time.Time) *gorm.DB {
	q := db.Model(&model.AuditLog{})
	if p != nil {
		q = applyScopeOrNullAudit(q, p)
	}
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

func applyScopeOrNullAudit(db *gorm.DB, p *access.Principal) *gorm.DB {
	expr, args, ok := buildDataScopeExpr(p, "institution_id", "department_id", "team_id")
	if !ok {
		return db.Where("institution_id IS NULL")
	}
	return db.Where("(institution_id IS NULL OR "+expr+")", args...)
}

func (r *AuditRepository) ListLogs(ctx context.Context, p *access.Principal, offset, limit int, orderClause, module, targetType string, from, to *time.Time) ([]model.AuditLog, int64, error) {
	base := auditLogQuery(r.db.WithContext(ctx), p, module, targetType, from, to)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.AuditLog
	err := auditLogQuery(r.db.WithContext(ctx), p, module, targetType, from, to).
		Order(orderClause).
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *AuditRepository) CreateExport(ctx context.Context, e *model.AuditExport) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *AuditRepository) GetExport(ctx context.Context, id string) (*model.AuditExport, error) {
	var e model.AuditExport
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error
	return &e, err
}

func (r *AuditRepository) UpdateExport(ctx context.Context, e *model.AuditExport) error {
	return r.db.WithContext(ctx).Model(e).Updates(map[string]interface{}{
		"status":           e.Status,
		"output_file_path": e.OutputFilePath,
		"completed_at":     e.CompletedAt,
	}).Error
}

// ExecuteExport materializes filtered audit logs to a CSV file and marks the export complete.
func (r *AuditRepository) ExecuteExport(ctx context.Context, p *access.Principal, e *model.AuditExport, outputDir string) error {
	var filter struct {
		Module     string `json:"module"`
		TargetType string `json:"targetType"`
		From       string `json:"from"`
		To         string `json:"to"`
	}
	if len(e.FilterJSON) > 0 {
		_ = json.Unmarshal(e.FilterJSON, &filter)
	}

	var fromPtr, toPtr *time.Time
	if filter.From != "" {
		if t, err := time.Parse(time.RFC3339, filter.From); err == nil {
			fromPtr = &t
		}
	}
	if filter.To != "" {
		if t, err := time.Parse(time.RFC3339, filter.To); err == nil {
			toPtr = &t
		}
	}

	var rows []model.AuditLog
	q := auditLogQuery(r.db.WithContext(ctx), p, filter.Module, filter.TargetType, fromPtr, toPtr)
	if err := q.Order("created_at ASC").Find(&rows).Error; err != nil {
		return err
	}

	if outputDir == "" {
		outputDir = os.TempDir()
	}
	_ = os.MkdirAll(outputDir, 0o700)
	filename := fmt.Sprintf("audit_export_%s.csv", e.ID)
	outPath := filepath.Join(outputDir, filename)
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"id", "module", "operation", "operator_user_id", "target_type", "target_id", "created_at"})
	for _, row := range rows {
		_ = w.Write([]string{row.ID, row.Module, row.Operation, row.OperatorUserID, row.TargetType, row.TargetID, row.CreatedAt.UTC().Format(time.RFC3339)})
	}
	w.Flush()

	now := time.Now().UTC()
	e.Status = "completed"
	e.OutputFilePath = &outPath
	e.CompletedAt = &now
	return r.UpdateExport(ctx, e)
}

// CreateAuditLog appends one immutable audit row (append-only store).
func (r *AuditRepository) CreateAuditLog(ctx context.Context, a *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(a).Error
}
