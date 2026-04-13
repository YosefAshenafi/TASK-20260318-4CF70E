package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/model"
)

type ComplianceRepository struct {
	db *gorm.DB
}

func NewComplianceRepository(db *gorm.DB) *ComplianceRepository {
	return &ComplianceRepository{db: db}
}

func (r *ComplianceRepository) ListQualifications(ctx context.Context, institutionIDs []string, offset, limit int, orderClause string) ([]model.QualificationProfile, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.QualificationProfile{}).
		Where("institution_id IN ?", institutionIDs)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.QualificationProfile
	err := r.db.WithContext(ctx).
		Where("institution_id IN ?", institutionIDs).
		Order(orderClause).
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *ComplianceRepository) GetQualification(ctx context.Context, id string, institutionIDs []string) (*model.QualificationProfile, error) {
	var q model.QualificationProfile
	err := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", id, institutionIDs).
		First(&q).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *ComplianceRepository) CreateQualification(ctx context.Context, q *model.QualificationProfile) error {
	return r.db.WithContext(ctx).Create(q).Error
}

func (r *ComplianceRepository) UpdateQualification(ctx context.Context, q *model.QualificationProfile, institutionIDs []string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", q.ID, institutionIDs).
		Updates(map[string]interface{}{
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

func (r *ComplianceRepository) ListQualificationsExpiringBetween(ctx context.Context, institutionIDs []string, from, to time.Time) ([]model.QualificationProfile, error) {
	var rows []model.QualificationProfile
	err := r.db.WithContext(ctx).
		Where("institution_id IN ? AND status = ? AND expires_on IS NOT NULL AND expires_on >= ? AND expires_on <= ?",
			institutionIDs, "active", from, to).
		Order("expires_on ASC").
		Find(&rows).Error
	return rows, err
}

func (r *ComplianceRepository) DeactivateExpiredQualifications(ctx context.Context, institutionIDs []string, before time.Time) (int64, error) {
	q := r.db.WithContext(ctx).Model(&model.QualificationProfile{}).
		Where("institution_id IN ? AND status = ? AND expires_on IS NOT NULL AND expires_on < ?", institutionIDs, "active", before)
	res := q.Updates(map[string]interface{}{
		"status":          "inactive",
		"deactivated_at":  time.Now().UTC(),
		"updated_at":        time.Now().UTC(),
	})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// ListActiveRestrictionsForPurchase returns active rules for a client/medication pair (medication-specific or institution-wide when medication_id is NULL).
func (r *ComplianceRepository) ListActiveRestrictionsForPurchase(ctx context.Context, institutionID, clientID, medicationID string) ([]model.PurchaseRestriction, error) {
	var rows []model.PurchaseRestriction
	err := r.db.WithContext(ctx).
		Where("institution_id = ? AND client_id = ? AND is_active = ?", institutionID, clientID, true).
		Where("(medication_id IS NULL OR medication_id = ?)", medicationID).
		Find(&rows).Error
	return rows, err
}

func (r *ComplianceRepository) ListRestrictions(ctx context.Context, institutionIDs []string, offset, limit int, orderClause string) ([]model.PurchaseRestriction, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.PurchaseRestriction{}).
		Where("institution_id IN ?", institutionIDs)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.PurchaseRestriction
	err := r.db.WithContext(ctx).
		Where("institution_id IN ?", institutionIDs).
		Order(orderClause).
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *ComplianceRepository) GetRestriction(ctx context.Context, id string, institutionIDs []string) (*model.PurchaseRestriction, error) {
	var row model.PurchaseRestriction
	err := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", id, institutionIDs).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ComplianceRepository) CreateRestriction(ctx context.Context, row *model.PurchaseRestriction) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *ComplianceRepository) UpdateRestriction(ctx context.Context, row *model.PurchaseRestriction, institutionIDs []string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", row.ID, institutionIDs).
		Updates(map[string]interface{}{
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

func (r *ComplianceRepository) ListViolations(ctx context.Context, institutionIDs []string, offset, limit int, orderClause string) ([]model.RestrictionViolationRecord, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.RestrictionViolationRecord{}).
		Where("institution_id IN ?", institutionIDs)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.RestrictionViolationRecord
	err := r.db.WithContext(ctx).
		Where("institution_id IN ?", institutionIDs).
		Order(orderClause).
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *ComplianceRepository) CountPurchaseRecordsSince(ctx context.Context, institutionID, clientID string, medicationID *string, since time.Time) (int64, error) {
	q := r.db.WithContext(ctx).Model(&model.CompliancePurchaseRecord{}).
		Where("institution_id = ? AND client_id = ? AND recorded_at >= ?", institutionID, clientID, since)
	if medicationID != nil && *medicationID != "" {
		q = q.Where("medication_id = ?", *medicationID)
	} else {
		q = q.Where("medication_id IS NULL OR medication_id = ''")
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
