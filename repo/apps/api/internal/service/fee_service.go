package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

var ErrFeeValidation = errors.New("fee validation failed")

type FeeDTO struct {
	ID            string  `json:"id"`
	InstitutionID string  `json:"institutionId"`
	DepartmentID  *string `json:"departmentId,omitempty"`
	TeamID        *string `json:"teamId,omitempty"`
	CaseID        *string `json:"caseId,omitempty"`
	CandidateID   *string `json:"candidateId,omitempty"`
	FeeType       string  `json:"feeType"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Note          *string `json:"note,omitempty"`
	CreatedByUser string  `json:"createdByUserId"`
	UpdatedByUser *string `json:"updatedByUserId,omitempty"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

type FeeService struct {
	repo  *repository.FeeRepository
	audit *AuditService
}

func NewFeeService(repo *repository.FeeRepository, audit *AuditService) *FeeService {
	return &FeeService{repo: repo, audit: audit}
}

func feeOrder(sortBy, sortOrder string) string {
	col := "created_at"
	switch sortBy {
	case "created_at", "updated_at", "amount", "fee_type":
		col = sortBy
	}
	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}
	return col + " " + order
}

func toFeeDTO(f *model.FeeRecord) FeeDTO {
	return FeeDTO{
		ID:            f.ID,
		InstitutionID: f.InstitutionID,
		DepartmentID:  f.DepartmentID,
		TeamID:        f.TeamID,
		CaseID:        f.CaseID,
		CandidateID:   f.CandidateID,
		FeeType:       f.FeeType,
		Amount:        f.Amount,
		Currency:      f.Currency,
		Note:          f.Note,
		CreatedByUser: f.CreatedByUserID,
		UpdatedByUser: f.UpdatedByUserID,
		CreatedAt:     f.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     f.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (s *FeeService) ListFees(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder string) ([]FeeDTO, int64, int, int, error) {
	if err := requireScope(p); err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListFees(ctx, p, offset, pageSize, feeOrder(sortBy, sortOrder))
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]FeeDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toFeeDTO(&rows[i]))
	}
	return out, total, page, pageSize, nil
}

type CreateFeeInput struct {
	InstitutionID string
	DepartmentID  *string
	TeamID        *string
	CaseID        *string
	CandidateID   *string
	FeeType       string
	Amount        float64
	Currency      string
	Note          *string
}

func (s *FeeService) CreateFee(ctx context.Context, p *access.Principal, in CreateFeeInput, meta AuditRequestMeta) (*FeeDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if !p.RowVisible(in.InstitutionID, in.DepartmentID, in.TeamID) {
		return nil, ErrForbiddenScope
	}
	ft := strings.TrimSpace(in.FeeType)
	ccy := strings.ToUpper(strings.TrimSpace(in.Currency))
	if ft == "" || in.Amount <= 0 {
		return nil, ErrFeeValidation
	}
	if ccy == "" {
		ccy = "CNY"
	}
	now := time.Now().UTC()
	row := &model.FeeRecord{
		ID:              uuid.NewString(),
		InstitutionID:   in.InstitutionID,
		DepartmentID:    in.DepartmentID,
		TeamID:          in.TeamID,
		CaseID:          in.CaseID,
		CandidateID:     in.CandidateID,
		FeeType:         ft,
		Amount:          in.Amount,
		Currency:        ccy,
		Note:            in.Note,
		CreatedByUserID: meta.OperatorUserID,
		UpdatedByUserID: &meta.OperatorUserID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if row.CreatedByUserID == "" {
		return nil, ErrFeeValidation
	}
	if err := s.repo.CreateFee(ctx, row); err != nil {
		return nil, err
	}
	dto := toFeeDTO(row)
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "fees",
		Operation:  "fee.create",
		TargetType: "fee",
		TargetID:   row.ID,
		After:      DTOToAuditMap(&dto),
		Meta:       meta,
	})
	return &dto, nil
}

type UpdateFeeInput struct {
	FeeType  *string
	Amount   *float64
	Currency *string
	Note     *string
}

func (s *FeeService) UpdateFee(ctx context.Context, p *access.Principal, id string, in UpdateFeeInput, meta AuditRequestMeta) (*FeeDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	row, err := s.repo.GetFee(ctx, id, p)
	if err != nil {
		return nil, err
	}
	before := toFeeDTO(row)
	if in.FeeType != nil {
		v := strings.TrimSpace(*in.FeeType)
		if v == "" {
			return nil, ErrFeeValidation
		}
		row.FeeType = v
	}
	if in.Amount != nil {
		if *in.Amount <= 0 {
			return nil, ErrFeeValidation
		}
		row.Amount = *in.Amount
	}
	if in.Currency != nil {
		c := strings.ToUpper(strings.TrimSpace(*in.Currency))
		if c == "" {
			return nil, ErrFeeValidation
		}
		row.Currency = c
	}
	if in.Note != nil {
		row.Note = in.Note
	}
	if meta.OperatorUserID != "" {
		row.UpdatedByUserID = &meta.OperatorUserID
	}
	if err := s.repo.UpdateFee(ctx, row, p); err != nil {
		return nil, err
	}
	out := toFeeDTO(row)
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "fees",
		Operation:  "fee.update",
		TargetType: "fee",
		TargetID:   id,
		Before:     DTOToAuditMap(&before),
		After:      DTOToAuditMap(&out),
		Meta:       meta,
	})
	return &out, nil
}
