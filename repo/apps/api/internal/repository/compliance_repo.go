package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
)

type ComplianceRepository struct {
	db *gorm.DB
}

func NewComplianceRepository(db *gorm.DB) *ComplianceRepository {
	return &ComplianceRepository{db: db}
}

func (r *ComplianceRepository) ListQualifications(ctx context.Context, p *access.Principal, offset, limit int, orderClause string) ([]model.QualificationProfile, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.QualificationProfile{})
	base = applyDataScope(base, p, "institution_id", "department_id", "team_id")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.QualificationProfile
	q := r.db.WithContext(ctx).Model(&model.QualificationProfile{})
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	err := q.Order(orderClause).Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *ComplianceRepository) GetQualification(ctx context.Context, id string, p *access.Principal) (*model.QualificationProfile, error) {
	var q model.QualificationProfile
	db := r.db.WithContext(ctx).Where("id = ?", id)
	db = applyDataScope(db, p, "institution_id", "department_id", "team_id")
	err := db.First(&q).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *ComplianceRepository) CreateQualification(ctx context.Context, q *model.QualificationProfile) error {
	return r.db.WithContext(ctx).Create(q).Error
}

func (r *ComplianceRepository) UpdateQualification(ctx context.Context, q *model.QualificationProfile, p *access.Principal) error {
	db := r.db.WithContext(ctx).Model(&model.QualificationProfile{}).Where("id = ?", q.ID)
	db = applyDataScope(db, p, "institution_id", "department_id", "team_id")
	res := db.Updates(map[string]interface{}{
			"display_name":    q.DisplayName,
			"expires_on":      q.ExpiresOn,
			"metadata_json":   q.MetadataJSON,
			"status":          q.Status,
			"deactivated_at":  q.DeactivatedAt,
			"updated_at":      time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ComplianceRepository) ListQualificationsExpiringBetween(ctx context.Context, p *access.Principal, from, to time.Time) ([]model.QualificationProfile, error) {
	var rows []model.QualificationProfile
	q := r.db.WithContext(ctx).Model(&model.QualificationProfile{}).
		Where("status = ? AND expires_on IS NOT NULL AND expires_on >= ? AND expires_on <= ?", "active", from, to)
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	err := q.Order("expires_on ASC").Find(&rows).Error
	return rows, err
}

func (r *ComplianceRepository) DeactivateExpiredQualifications(ctx context.Context, p *access.Principal, before time.Time) (int64, error) {
	q := r.db.WithContext(ctx).Model(&model.QualificationProfile{}).
		Where("status = ? AND expires_on IS NOT NULL AND expires_on < ?", "active", before)
	if p != nil && len(p.Scopes) > 0 && p.Scopes[0].InstitutionID != "*" {
		q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	}
	res := q.Updates(map[string]interface{}{
		"status":         "inactive",
		"deactivated_at": time.Now().UTC(),
		"updated_at":     time.Now().UTC(),
	})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// ListActiveRestrictionsForPurchase returns active rules for a client/medication pair (medication-specific or institution-wide when medication_id is NULL), scoped to the principal.
func (r *ComplianceRepository) ListActiveRestrictionsForPurchase(ctx context.Context, p *access.Principal, institutionID, clientID, medicationID string) ([]model.PurchaseRestriction, error) {
	var rows []model.PurchaseRestriction
	q := r.db.WithContext(ctx).
		Where("institution_id = ? AND client_id = ? AND is_active = ?", institutionID, clientID, true).
		Where("(medication_id IS NULL OR medication_id = ?)", medicationID)
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	err := q.Find(&rows).Error
	return rows, err
}

func (r *ComplianceRepository) ListRestrictions(ctx context.Context, p *access.Principal, offset, limit int, orderClause string) ([]model.PurchaseRestriction, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.PurchaseRestriction{})
	base = applyDataScope(base, p, "institution_id", "department_id", "team_id")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.PurchaseRestriction
	q := r.db.WithContext(ctx).Model(&model.PurchaseRestriction{})
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	err := q.Order(orderClause).Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *ComplianceRepository) GetRestriction(ctx context.Context, id string, p *access.Principal) (*model.PurchaseRestriction, error) {
	var row model.PurchaseRestriction
	db := r.db.WithContext(ctx).Where("id = ?", id)
	db = applyDataScope(db, p, "institution_id", "department_id", "team_id")
	err := db.First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ComplianceRepository) CreateRestriction(ctx context.Context, row *model.PurchaseRestriction) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *ComplianceRepository) UpdateRestriction(ctx context.Context, row *model.PurchaseRestriction, p *access.Principal) error {
	db := r.db.WithContext(ctx).Model(&model.PurchaseRestriction{}).Where("id = ?", row.ID)
	db = applyDataScope(db, p, "institution_id", "department_id", "team_id")
	res := db.Updates(map[string]interface{}{
			"client_id":      row.ClientID,
			"medication_id":  row.MedicationID,
			"rule_json":      row.RuleJSON,
			"is_active":      row.IsActive,
			"updated_at":     time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ComplianceRepository) ListViolations(ctx context.Context, p *access.Principal, offset, limit int, orderClause string) ([]model.RestrictionViolationRecord, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.RestrictionViolationRecord{})
	base = applyDataScope(base, p, "institution_id", "department_id", "team_id")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.RestrictionViolationRecord
	q := r.db.WithContext(ctx).Model(&model.RestrictionViolationRecord{})
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	err := q.Order(orderClause).Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *ComplianceRepository) CountPurchaseRecordsSince(ctx context.Context, institutionID, clientID string, medicationID *string, since time.Time, deptID, teamID *string, scopedByOrg bool) (int64, error) {
	q := r.db.WithContext(ctx).Model(&model.CompliancePurchaseRecord{}).
		Where("institution_id = ? AND client_id = ? AND recorded_at >= ?", institutionID, clientID, since)
	if medicationID != nil && *medicationID != "" {
		q = q.Where("medication_id = ?", *medicationID)
	} else {
		q = q.Where("medication_id IS NULL OR medication_id = ''")
	}
	if scopedByOrg {
		if deptID == nil {
			q = q.Where("department_id IS NULL")
		} else {
			q = q.Where("department_id = ?", *deptID)
		}
		if teamID == nil {
			q = q.Where("team_id IS NULL")
		} else {
			q = q.Where("team_id = ?", *teamID)
		}
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (r *ComplianceRepository) InsertPurchaseRecord(ctx context.Context, rec *model.CompliancePurchaseRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *ComplianceRepository) InsertViolation(ctx context.Context, v *model.RestrictionViolationRecord) error {
	return r.db.WithContext(ctx).Create(v).Error
}
