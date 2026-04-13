package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

type restrictionRule struct {
	RequiresPrescription bool `json:"requiresPrescription"`
	FrequencyDays        int  `json:"frequencyDays"`
}

// QualificationDTO for API responses.
type QualificationDTO struct {
	ID             string         `json:"id"`
	InstitutionID  string         `json:"institutionId"`
	ClientID       string         `json:"clientId"`
	DisplayName    string         `json:"displayName"`
	Status         string         `json:"status"`
	ExpiresOn      *string        `json:"expiresOn,omitempty"`
	DeactivatedAt  *string        `json:"deactivatedAt,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      string         `json:"createdAt"`
	UpdatedAt      string         `json:"updatedAt"`
}

// RestrictionDTO for API responses.
type RestrictionDTO struct {
	ID            string         `json:"id"`
	InstitutionID string         `json:"institutionId"`
	ClientID      string         `json:"clientId"`
	MedicationID  *string        `json:"medicationId,omitempty"`
	Rule          map[string]any `json:"rule"`
	IsActive      bool           `json:"isActive"`
	CreatedAt     string         `json:"createdAt"`
	UpdatedAt     string         `json:"updatedAt"`
}

// ViolationDTO for API responses.
type ViolationDTO struct {
	ID            string         `json:"id"`
	RestrictionID *string        `json:"restrictionId,omitempty"`
	InstitutionID string         `json:"institutionId"`
	ClientID      string         `json:"clientId"`
	MedicationID  *string        `json:"medicationId,omitempty"`
	Details       map[string]any `json:"details,omitempty"`
	CreatedAt     string         `json:"createdAt"`
}

type ComplianceService struct {
	repo  *repository.ComplianceRepository
	audit *AuditService
}

func NewComplianceService(repo *repository.ComplianceRepository, audit *AuditService) *ComplianceService {
	return &ComplianceService{repo: repo, audit: audit}
}

func qualificationOrder(sortBy, sortOrder string) string {
	col := "created_at"
	switch sortBy {
	case "created_at", "updated_at", "expires_on", "display_name", "client_id", "status":
		col = sortBy
	}
	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}
	return col + " " + order
}

func restrictionOrder(sortBy, sortOrder string) string {
	col := "created_at"
	switch sortBy {
	case "created_at", "updated_at", "client_id", "is_active":
		col = sortBy
	}
	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}
	return col + " " + order
}

func violationOrder(sortBy, sortOrder string) string {
	col := "created_at"
	switch sortBy {
	case "created_at", "client_id":
		col = sortBy
	}
	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}
	return col + " " + order
}

func parseMetadata(b []byte) map[string]any {
	if len(b) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}

func parseRuleMap(b []byte) map[string]any {
	if len(b) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]any{}
	}
	return m
}

func toQualificationDTO(q *model.QualificationProfile) QualificationDTO {
	dto := QualificationDTO{
		ID:            q.ID,
		InstitutionID: q.InstitutionID,
		ClientID:      q.ClientID,
		DisplayName:   q.DisplayName,
		Status:        q.Status,
		Metadata:      parseMetadata(q.MetadataJSON),
		CreatedAt:     q.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     q.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if q.ExpiresOn != nil {
		s := q.ExpiresOn.UTC().Format("2006-01-02")
		dto.ExpiresOn = &s
	}
	if q.DeactivatedAt != nil {
		s := q.DeactivatedAt.UTC().Format(time.RFC3339)
		dto.DeactivatedAt = &s
	}
	return dto
}

func toRestrictionDTO(r *model.PurchaseRestriction) RestrictionDTO {
	return RestrictionDTO{
		ID:            r.ID,
		InstitutionID: r.InstitutionID,
		ClientID:      r.ClientID,
		MedicationID:  r.MedicationID,
		Rule:          parseRuleMap(r.RuleJSON),
		IsActive:      r.IsActive,
		CreatedAt:     r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     r.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toViolationDTO(v *model.RestrictionViolationRecord) ViolationDTO {
	return ViolationDTO{
		ID:            v.ID,
		RestrictionID: v.RestrictionID,
		InstitutionID: v.InstitutionID,
		ClientID:      v.ClientID,
		MedicationID:  v.MedicationID,
		Details:       parseMetadata(v.DetailsJSON),
		CreatedAt:     v.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (s *ComplianceService) ListQualifications(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder string) ([]QualificationDTO, int64, int, int, error) {
	if err := requireScope(p); err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListQualifications(ctx, p, offset, pageSize, qualificationOrder(sortBy, sortOrder))
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]QualificationDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toQualificationDTO(&rows[i]))
	}
	return out, total, page, pageSize, nil
}

func (s *ComplianceService) GetQualification(ctx context.Context, p *access.Principal, id string) (*QualificationDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	q, err := s.repo.GetQualification(ctx, id, p)
	if err != nil {
		return nil, err
	}
	dto := toQualificationDTO(q)
	return &dto, nil
}

// CreateQualificationInput for POST.
type CreateQualificationInput struct {
	InstitutionID string
	ClientID      string
	DisplayName   string
	ExpiresOn     *string
	Metadata      map[string]any
}

func (s *ComplianceService) CreateQualification(ctx context.Context, p *access.Principal, in CreateQualificationInput, meta AuditRequestMeta) (*QualificationDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if !p.AllowsInstitution(in.InstitutionID) {
		return nil, ErrForbiddenScope
	}
	now := time.Now().UTC()
	var metaBytes []byte
	if in.Metadata != nil && len(in.Metadata) > 0 {
		var err error
		metaBytes, err = json.Marshal(in.Metadata)
		if err != nil {
			return nil, err
		}
	}
	var exp *time.Time
	if in.ExpiresOn != nil && *in.ExpiresOn != "" {
		t, err := time.Parse("2006-01-02", *in.ExpiresOn)
		if err != nil {
			return nil, err
		}
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		exp = &t
	}
	q := &model.QualificationProfile{
		ID:            uuid.NewString(),
		InstitutionID: in.InstitutionID,
		ClientID:      in.ClientID,
		DisplayName:   in.DisplayName,
		Status:        "active",
		ExpiresOn:     exp,
		MetadataJSON:  metaBytes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateQualification(ctx, q); err != nil {
		return nil, err
	}
	loaded, err := s.repo.GetQualification(ctx, q.ID, p)
	if err != nil {
		return nil, err
	}
	dto := toQualificationDTO(loaded)
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "compliance",
		Operation:  "qualification.create",
		TargetType: "qualification",
		TargetID:   dto.ID,
		After:      DTOToAuditMap(&dto),
		Meta:       meta,
	})
	return &dto, nil
}

// UpdateQualificationInput for PATCH.
type UpdateQualificationInput struct {
	DisplayName *string
	ExpiresOn   *string // empty string clears
	Metadata    map[string]any
	Status      *string
}

func (s *ComplianceService) UpdateQualification(ctx context.Context, p *access.Principal, id string, in UpdateQualificationInput, meta AuditRequestMeta) (*QualificationDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	q, err := s.repo.GetQualification(ctx, id, p)
	if err != nil {
		return nil, err
	}
	beforeDTO := toQualificationDTO(q)
	beforeMap := DTOToAuditMap(&beforeDTO)
	if in.DisplayName != nil {
		q.DisplayName = *in.DisplayName
	}
	if in.Status != nil {
		q.Status = *in.Status
	}
	if in.ExpiresOn != nil {
		if *in.ExpiresOn == "" {
			q.ExpiresOn = nil
		} else {
			t, err := time.Parse("2006-01-02", *in.ExpiresOn)
			if err != nil {
				return nil, err
			}
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			q.ExpiresOn = &t
		}
	}
	if in.Metadata != nil {
		b, err := json.Marshal(in.Metadata)
		if err != nil {
			return nil, err
		}
		q.MetadataJSON = b
	}
	if err := s.repo.UpdateQualification(ctx, q, p); err != nil {
		return nil, err
	}
	out, err := s.GetQualification(ctx, p, id)
	if err != nil {
		return nil, err
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "compliance",
		Operation:  "qualification.update",
		TargetType: "qualification",
		TargetID:   id,
		Before:     beforeMap,
		After:      DTOToAuditMap(out),
		Meta:       meta,
	})
	return out, nil
}

func (s *ComplianceService) ActivateQualification(ctx context.Context, p *access.Principal, id string, meta AuditRequestMeta) (*QualificationDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	q, err := s.repo.GetQualification(ctx, id, p)
	if err != nil {
		return nil, err
	}
	beforeDTO := toQualificationDTO(q)
	beforeMap := DTOToAuditMap(&beforeDTO)
	q.Status = "active"
	q.DeactivatedAt = nil
	if err := s.repo.UpdateQualification(ctx, q, p); err != nil {
		return nil, err
	}
	out, err := s.GetQualification(ctx, p, id)
	if err != nil {
		return nil, err
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "compliance",
		Operation:  "qualification.activate",
		TargetType: "qualification",
		TargetID:   id,
		Before:     beforeMap,
		After:      DTOToAuditMap(out),
		Meta:       meta,
	})
	return out, nil
}

func (s *ComplianceService) DeactivateQualification(ctx context.Context, p *access.Principal, id string, meta AuditRequestMeta) (*QualificationDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	q, err := s.repo.GetQualification(ctx, id, p)
	if err != nil {
		return nil, err
	}
	beforeDTO := toQualificationDTO(q)
	beforeMap := DTOToAuditMap(&beforeDTO)
	now := time.Now().UTC()
	q.Status = "inactive"
	q.DeactivatedAt = &now
	if err := s.repo.UpdateQualification(ctx, q, p); err != nil {
		return nil, err
	}
	out, err := s.GetQualification(ctx, p, id)
	if err != nil {
		return nil, err
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "compliance",
		Operation:  "qualification.deactivate",
		TargetType: "qualification",
		TargetID:   id,
		Before:     beforeMap,
		After:      DTOToAuditMap(out),
		Meta:       meta,
	})
	return out, nil
}

// ListExpiringQualifications returns active qualifications expiring within the next `days` (inclusive window from today).
func (s *ComplianceService) ListExpiringQualifications(ctx context.Context, p *access.Principal, days int) ([]QualificationDTO, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	if err := requireScope(p); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := startOfToday.AddDate(0, 0, days)
	rows, err := s.repo.ListQualificationsExpiringBetween(ctx, p, startOfToday, end)
	if err != nil {
		return nil, err
	}
	out := make([]QualificationDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toQualificationDTO(&rows[i]))
	}
	return out, nil
}

// RunQualificationExpirationJob deactivates active qualifications whose expires_on is before today.
func (s *ComplianceService) RunQualificationExpirationJob(ctx context.Context, p *access.Principal) (deactivated int64, err error) {
	if err := requireScope(p); err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return s.repo.DeactivateExpiredQualifications(ctx, p, startOfToday)
}

func (s *ComplianceService) ListRestrictions(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder string) ([]RestrictionDTO, int64, int, int, error) {
	if err := requireScope(p); err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListRestrictions(ctx, p, offset, pageSize, restrictionOrder(sortBy, sortOrder))
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]RestrictionDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toRestrictionDTO(&rows[i]))
	}
	return out, total, page, pageSize, nil
}

func (s *ComplianceService) GetRestriction(ctx context.Context, p *access.Principal, id string) (*RestrictionDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	r, err := s.repo.GetRestriction(ctx, id, p)
	if err != nil {
		return nil, err
	}
	dto := toRestrictionDTO(r)
	return &dto, nil
}

// CreateRestrictionInput for POST.
type CreateRestrictionInput struct {
	InstitutionID string
	ClientID      string
	MedicationID  *string
	Rule          map[string]any
	IsActive      bool
}

func (s *ComplianceService) CreateRestriction(ctx context.Context, p *access.Principal, in CreateRestrictionInput, meta AuditRequestMeta) (*RestrictionDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if !p.AllowsInstitution(in.InstitutionID) {
		return nil, ErrForbiddenScope
	}
	ruleBytes, err := json.Marshal(in.Rule)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	active := in.IsActive
	row := &model.PurchaseRestriction{
		ID:            uuid.NewString(),
		InstitutionID: in.InstitutionID,
		ClientID:      in.ClientID,
		MedicationID:  in.MedicationID,
		RuleJSON:      ruleBytes,
		IsActive:      active,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateRestriction(ctx, row); err != nil {
		return nil, err
	}
	loaded, err := s.repo.GetRestriction(ctx, row.ID, p)
	if err != nil {
		return nil, err
	}
	dto := toRestrictionDTO(loaded)
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "fees",
		Operation:  "restriction.create",
		TargetType: "restriction",
		TargetID:   dto.ID,
		After:      DTOToAuditMap(&dto),
		Meta:       meta,
	})
	return &dto, nil
}

// UpdateRestrictionInput for PATCH.
type UpdateRestrictionInput struct {
	ClientID     *string
	MedicationID *string
	Rule         map[string]any
	IsActive     *bool
}

func (s *ComplianceService) UpdateRestriction(ctx context.Context, p *access.Principal, id string, in UpdateRestrictionInput, meta AuditRequestMeta) (*RestrictionDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	row, err := s.repo.GetRestriction(ctx, id, p)
	if err != nil {
		return nil, err
	}
	beforeDTO := toRestrictionDTO(row)
	beforeMap := DTOToAuditMap(&beforeDTO)
	if in.ClientID != nil {
		row.ClientID = *in.ClientID
	}
	if in.MedicationID != nil {
		row.MedicationID = in.MedicationID
	}
	if in.Rule != nil {
		b, err := json.Marshal(in.Rule)
		if err != nil {
			return nil, err
		}
		row.RuleJSON = b
	}
	if in.IsActive != nil {
		row.IsActive = *in.IsActive
	}
	if err := s.repo.UpdateRestriction(ctx, row, p); err != nil {
		return nil, err
	}
	out, err := s.GetRestriction(ctx, p, id)
	if err != nil {
		return nil, err
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "fees",
		Operation:  "restriction.update",
		TargetType: "restriction",
		TargetID:   id,
		Before:     beforeMap,
		After:      DTOToAuditMap(out),
		Meta:       meta,
	})
	return out, nil
}

func (s *ComplianceService) ListViolations(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder string) ([]ViolationDTO, int64, int, int, error) {
	if err := requireScope(p); err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListViolations(ctx, p, offset, pageSize, violationOrder(sortBy, sortOrder))
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]ViolationDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toViolationDTO(&rows[i]))
	}
	return out, total, page, pageSize, nil
}

// CheckPurchaseInput for POST check-purchase.
type CheckPurchaseInput struct {
	InstitutionID            string
	ClientID                 string
	MedicationID             string
	IsControlled             bool
	PrescriptionAttachmentID *string
	PurchaseAt               time.Time
}

// CheckPurchaseResult is returned in API data.
type CheckPurchaseResult struct {
	Allowed bool     `json:"allowed"`
	Reasons []string `json:"reasons"`
}

func parseRestrictionRule(b []byte) (restrictionRule, error) {
	var r restrictionRule
	if len(b) == 0 {
		return r, nil
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return r, err
	}
	return r, nil
}

func (s *ComplianceService) CheckPurchase(ctx context.Context, p *access.Principal, in CheckPurchaseInput) (*CheckPurchaseResult, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if !p.AllowsInstitution(in.InstitutionID) {
		return nil, ErrForbiddenScope
	}
	rules, err := s.repo.ListActiveRestrictionsForPurchase(ctx, in.InstitutionID, in.ClientID, in.MedicationID)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &CheckPurchaseResult{Allowed: true, Reasons: []string{}}, nil
	}

	var medPtr *string
	if in.MedicationID != "" {
		medPtr = &in.MedicationID
	}

	for _, row := range rules {
		rule, err := parseRestrictionRule(row.RuleJSON)
		if err != nil {
			continue
		}
		if rule.RequiresPrescription && in.IsControlled {
			hasRx := in.PrescriptionAttachmentID != nil && *in.PrescriptionAttachmentID != ""
			if !hasRx {
				details, _ := json.Marshal(map[string]any{"reason": "PRESCRIPTION_REQUIRED", "restrictionId": row.ID})
				_ = s.repo.InsertViolation(ctx, &model.RestrictionViolationRecord{
					ID:             uuid.NewString(),
					RestrictionID:  &row.ID,
					InstitutionID:  in.InstitutionID,
					ClientID:       in.ClientID,
					MedicationID:   medPtr,
					DetailsJSON:    details,
					CreatedAt:      time.Now().UTC(),
				})
				return &CheckPurchaseResult{Allowed: false, Reasons: []string{"prescription attachment required for controlled medication"}}, nil
			}
		}
	}

	for _, row := range rules {
		rule, err := parseRestrictionRule(row.RuleJSON)
		if err != nil {
			continue
		}
		if rule.FrequencyDays > 0 {
			since := in.PurchaseAt.Add(-time.Duration(rule.FrequencyDays) * 24 * time.Hour)
			n, err := s.repo.CountPurchaseRecordsSince(ctx, in.InstitutionID, in.ClientID, medPtr, since)
			if err != nil {
				return nil, err
			}
			if n > 0 {
				msg := "purchase already made within last " + strconv.Itoa(rule.FrequencyDays) + " days"
				details, _ := json.Marshal(map[string]any{"reason": "FREQUENCY", "restrictionId": row.ID})
				_ = s.repo.InsertViolation(ctx, &model.RestrictionViolationRecord{
					ID:             uuid.NewString(),
					RestrictionID:  &row.ID,
					InstitutionID:  in.InstitutionID,
					ClientID:       in.ClientID,
					MedicationID:   medPtr,
					DetailsJSON:    details,
					CreatedAt:      time.Now().UTC(),
				})
				return &CheckPurchaseResult{Allowed: false, Reasons: []string{msg}}, nil
			}
		}
	}

	rec := &model.CompliancePurchaseRecord{
		ID:             uuid.NewString(),
		InstitutionID:  in.InstitutionID,
		ClientID:       in.ClientID,
		MedicationID:   medPtr,
		RecordedAt:     in.PurchaseAt.UTC(),
	}
	if err := s.repo.InsertPurchaseRecord(ctx, rec); err != nil {
		return nil, err
	}
	return &CheckPurchaseResult{Allowed: true, Reasons: []string{}}, nil
}
