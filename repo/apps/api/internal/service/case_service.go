package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

// ErrDuplicateCaseSubmission blocks identical case intake within the duplicate window.
var ErrDuplicateCaseSubmission = errors.New("duplicate case submission")

// CaseDTO matches api-spec case response shape (extended for list/detail).
type CaseDTO struct {
	ID             string  `json:"id"`
	CaseNumber     string  `json:"caseNumber"`
	InstitutionID  string  `json:"institutionId"`
	DepartmentID   *string `json:"departmentId,omitempty"`
	TeamID         *string `json:"teamId,omitempty"`
	CaseType       string  `json:"caseType"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Status         string  `json:"status"`
	AssigneeID     *string `json:"assigneeId,omitempty"`
	ReportedAt     string  `json:"reportedAt"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type ProcessingRecordDTO struct {
	ID          string  `json:"id"`
	StepCode    string  `json:"stepCode"`
	ActorUserID string  `json:"actorUserId"`
	Note        *string `json:"note,omitempty"`
	CreatedAt   string  `json:"createdAt"`
}

type StatusTransitionDTO struct {
	ID          string `json:"id"`
	FromStatus  string `json:"fromStatus"`
	ToStatus    string `json:"toStatus"`
	ActorUserID string `json:"actorUserId"`
	CreatedAt   string `json:"createdAt"`
}

type CaseService struct {
	repo  *repository.CaseRepository
	audit *AuditService
}

func NewCaseService(repo *repository.CaseRepository, audit *AuditService) *CaseService {
	return &CaseService{repo: repo, audit: audit}
}

func duplicateContentHash(institutionID, caseType, title, description, reportedAt string) string {
	s := strings.Join([]string{institutionID, caseType, title, description, reportedAt}, "\x1e")
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func toCaseDTO(c *model.CaseRecord) CaseDTO {
	return CaseDTO{
		ID:            c.ID,
		CaseNumber:    c.CaseNumber,
		InstitutionID: c.InstitutionID,
		DepartmentID:  c.DepartmentID,
		TeamID:        c.TeamID,
		CaseType:      c.CaseType,
		Title:         c.Title,
		Description:   c.Description,
		Status:        c.Status,
		AssigneeID:    c.AssigneeUserID,
		ReportedAt:    c.ReportedAt.UTC().Format(time.RFC3339),
		CreatedAt:     c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     c.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func caseOrder(sortBy, sortOrder string) string {
	col := "created_at"
	switch sortBy {
	case "created_at", "updated_at", "reported_at", "status", "case_number", "title":
		col = sortBy
	}
	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}
	return col + " " + order
}

// CreateCaseInput for POST /cases.
type CreateCaseInput struct {
	InstitutionID string
	DepartmentID  *string
	TeamID        *string
	CaseType      string
	Title         string
	Description   string
	ReportedAt    time.Time
}

func (s *CaseService) CreateCase(ctx context.Context, p *access.Principal, in CreateCaseInput, meta AuditRequestMeta) (*CaseDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if !p.RowVisible(in.InstitutionID, in.DepartmentID, in.TeamID) {
		return nil, ErrForbiddenScope
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.CaseType = strings.TrimSpace(in.CaseType)
	if in.Title == "" || in.Description == "" || in.CaseType == "" {
		return nil, ErrCaseMandatoryFields
	}
	reportedRFC := in.ReportedAt.UTC().Format(time.RFC3339)
	h := duplicateContentHash(in.InstitutionID, in.CaseType, in.Title, in.Description, reportedRFC)
	since := time.Now().UTC().Add(-5 * time.Minute)

	code, err := s.repo.GetInstitutionCode(ctx, in.InstitutionID)
	if err != nil {
		return nil, err
	}

	day := time.Date(in.ReportedAt.Year(), in.ReportedAt.Month(), in.ReportedAt.Day(), 0, 0, 0, 0, time.UTC)

	var created *model.CaseRecord
	err = s.repo.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var n int64
		dupQ := tx.Model(&model.CaseRecord{}).
			Where("institution_id = ? AND duplicate_guard_hash = ? AND created_at >= ?", in.InstitutionID, h, since)
		if in.DepartmentID != nil {
			dupQ = dupQ.Where("department_id = ?", *in.DepartmentID)
		} else {
			dupQ = dupQ.Where("department_id IS NULL")
		}
		if in.TeamID != nil {
			dupQ = dupQ.Where("team_id = ?", *in.TeamID)
		} else {
			dupQ = dupQ.Where("team_id IS NULL")
		}
		if err := dupQ.Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return ErrDuplicateCaseSubmission
		}
		serial, err := s.repo.AllocateCaseSerial(ctx, tx, in.InstitutionID, day)
		if err != nil {
			return err
		}
		y := day.Format("20060102")
		caseNumber := fmt.Sprintf("%s-%s-%06d", y, code, serial)
		now := time.Now().UTC()
		rec := &model.CaseRecord{
			ID:                 uuid.NewString(),
			CaseNumber:         caseNumber,
			InstitutionID:      in.InstitutionID,
			DepartmentID:       in.DepartmentID,
			TeamID:             in.TeamID,
			CaseType:           in.CaseType,
			Title:              in.Title,
			Description:        in.Description,
			ReportedAt:         in.ReportedAt.UTC(),
			Status:             "submitted",
			DuplicateGuardHash: strPtr(h),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := tx.Create(rec).Error; err != nil {
			return err
		}
		created = rec
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateCaseSubmission) {
			return nil, err
		}
		return nil, err
	}
	dto := toCaseDTO(created)
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "cases",
		Operation:  "case.create",
		TargetType: "case",
		TargetID:   created.ID,
		After:      DTOToAuditMap(&dto),
		Meta:       meta,
	})
	return &dto, nil
}

func strPtr(s string) *string { return &s }

// ErrCaseMandatoryFields when required fields are empty.
var ErrCaseMandatoryFields = errors.New("case mandatory fields missing")

func (s *CaseService) ListCases(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder, search, status string) ([]CaseDTO, int64, int, int, error) {
	if err := requireScope(p); err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListCases(ctx, p, offset, pageSize, caseOrder(sortBy, sortOrder), search, status)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]CaseDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toCaseDTO(&rows[i]))
	}
	return out, total, page, pageSize, nil
}

func (s *CaseService) GetCase(ctx context.Context, p *access.Principal, id string) (*CaseDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	c, err := s.repo.GetCase(ctx, id, p)
	if err != nil {
		return nil, err
	}
	dto := toCaseDTO(c)
	return &dto, nil
}

// UpdateCaseInput for PATCH (non-status fields).
type UpdateCaseInput struct {
	Title        *string
	Description  *string
	DepartmentID *string
	TeamID       *string
}

func (s *CaseService) UpdateCase(ctx context.Context, p *access.Principal, id string, in UpdateCaseInput, meta AuditRequestMeta) (*CaseDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	c, err := s.repo.GetCase(ctx, id, p)
	if err != nil {
		return nil, err
	}
	beforeDTO := toCaseDTO(c)
	beforeMap := DTOToAuditMap(&beforeDTO)
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return nil, ErrCaseMandatoryFields
		}
		c.Title = t
	}
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		if d == "" {
			return nil, ErrCaseMandatoryFields
		}
		c.Description = d
	}
	if in.DepartmentID != nil {
		c.DepartmentID = in.DepartmentID
	}
	if in.TeamID != nil {
		c.TeamID = in.TeamID
	}
	if !p.RowVisible(c.InstitutionID, c.DepartmentID, c.TeamID) {
		return nil, ErrForbiddenScope
	}
	if err := s.repo.UpdateCase(ctx, c, p); err != nil {
		return nil, err
	}
	out, err := s.GetCase(ctx, p, id)
	if err != nil {
		return nil, err
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "cases",
		Operation:  "case.update",
		TargetType: "case",
		TargetID:   id,
		Before:     beforeMap,
		After:      DTOToAuditMap(out),
		Meta:       meta,
	})
	return out, nil
}

// AssignCase sets assignee; moves status from submitted → assigned when applicable.
func (s *CaseService) AssignCase(ctx context.Context, p *access.Principal, caseID, assigneeUserID string, meta AuditRequestMeta) (*CaseDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	c, err := s.repo.GetCase(ctx, caseID, p)
	if err != nil {
		return nil, err
	}
	beforeDTO := toCaseDTO(c)
	beforeMap := DTOToAuditMap(&beforeDTO)
	aid := assigneeUserID
	newStatus := ""
	if c.Status == "submitted" {
		newStatus = "assigned"
	}
	if err := s.repo.SetAssignee(ctx, caseID, p, &aid, newStatus); err != nil {
		return nil, err
	}
	_ = s.repo.InsertAssignment(ctx, &model.CaseAssignment{
		ID:         uuid.NewString(),
		CaseID:     caseID,
		UserID:     assigneeUserID,
		AssignedAt: time.Now().UTC(),
	})
	out, err := s.GetCase(ctx, p, caseID)
	if err != nil {
		return nil, err
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "cases",
		Operation:  "case.assign",
		TargetType: "case",
		TargetID:   caseID,
		Before:     beforeMap,
		After:      DTOToAuditMap(out),
		Meta:       meta,
	})
	return out, nil
}

func (s *CaseService) AddProcessingRecord(ctx context.Context, p *access.Principal, caseID, actorUserID, stepCode string, note *string, meta AuditRequestMeta) (*ProcessingRecordDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetCase(ctx, caseID, p); err != nil {
		return nil, err
	}
	stepCode = strings.TrimSpace(stepCode)
	if stepCode == "" {
		return nil, ErrCaseMandatoryFields
	}
	rec := &model.CaseProcessingRecord{
		ID:          uuid.NewString(),
		CaseID:      caseID,
		StepCode:    stepCode,
		ActorUserID: actorUserID,
		Note:        note,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.CreateProcessingRecord(ctx, rec); err != nil {
		return nil, err
	}
	dto := &ProcessingRecordDTO{
		ID:          rec.ID,
		StepCode:    rec.StepCode,
		ActorUserID: rec.ActorUserID,
		Note:        rec.Note,
		CreatedAt:   rec.CreatedAt.UTC().Format(time.RFC3339),
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "cases",
		Operation:  "case.processing_record.create",
		TargetType: "case",
		TargetID:   caseID,
		After: map[string]any{
			"recordId":    dto.ID,
			"stepCode":    dto.StepCode,
			"actorUserId": dto.ActorUserID,
		},
		Meta: meta,
	})
	return dto, nil
}

func (s *CaseService) ListProcessingRecords(ctx context.Context, p *access.Principal, caseID string) ([]ProcessingRecordDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetCase(ctx, caseID, p); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListProcessingRecords(ctx, caseID, "created_at ASC")
	if err != nil {
		return nil, err
	}
	out := make([]ProcessingRecordDTO, 0, len(rows))
	for i := range rows {
		r := rows[i]
		out = append(out, ProcessingRecordDTO{
			ID:          r.ID,
			StepCode:    r.StepCode,
			ActorUserID: r.ActorUserID,
			Note:        r.Note,
			CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

// ErrInvalidStatusTransition when transition is not allowed by workflow rules.
var ErrInvalidStatusTransition = errors.New("invalid status transition")

func allowedTransition(from, to string) bool {
	if from == to {
		return false
	}
	m := map[string]map[string]bool{
		"submitted":      {"assigned": true, "in_progress": true, "closed": true},
		"assigned":       {"in_progress": true, "pending_review": true, "closed": true},
		"in_progress":    {"pending_review": true, "closed": true, "assigned": true},
		"pending_review": {"closed": true, "in_progress": true},
		"closed":         {},
	}
	if m[from] == nil {
		return false
	}
	return m[from][to]
}

func (s *CaseService) AddStatusTransition(ctx context.Context, p *access.Principal, caseID, actorUserID, toStatus string, meta AuditRequestMeta) (*CaseDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	c, err := s.repo.GetCase(ctx, caseID, p)
	if err != nil {
		return nil, err
	}
	beforeDTO := toCaseDTO(c)
	beforeMap := DTOToAuditMap(&beforeDTO)
	toStatus = strings.TrimSpace(toStatus)
	if toStatus == "" || !allowedTransition(c.Status, toStatus) {
		return nil, ErrInvalidStatusTransition
	}
	tr := &model.CaseStatusTransition{
		ID:          uuid.NewString(),
		CaseID:      caseID,
		FromStatus:  c.Status,
		ToStatus:    toStatus,
		ActorUserID: actorUserID,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.CreateStatusTransition(ctx, tr); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateCaseStatus(ctx, caseID, p, toStatus); err != nil {
		return nil, err
	}
	out, err := s.GetCase(ctx, p, caseID)
	if err != nil {
		return nil, err
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "cases",
		Operation:  "case.status_transition",
		TargetType: "case",
		TargetID:   caseID,
		Before:     beforeMap,
		After:      DTOToAuditMap(out),
		Meta:       meta,
	})
	return out, nil
}

func (s *CaseService) ListStatusTransitions(ctx context.Context, p *access.Principal, caseID string) ([]StatusTransitionDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetCase(ctx, caseID, p); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListStatusTransitions(ctx, caseID, "created_at ASC")
	if err != nil {
		return nil, err
	}
	out := make([]StatusTransitionDTO, 0, len(rows))
	for i := range rows {
		r := rows[i]
		out = append(out, StatusTransitionDTO{
			ID:          r.ID,
			FromStatus:  r.FromStatus,
			ToStatus:    r.ToStatus,
			ActorUserID: r.ActorUserID,
			CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

// SearchCaseLedger is GET /case-ledger/search — same filters as list with required pagination.
func (s *CaseService) SearchCaseLedger(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder, q, status string) ([]CaseDTO, int64, int, int, error) {
	return s.ListCases(ctx, p, page, pageSize, offset, sortBy, sortOrder, q, status)
}
