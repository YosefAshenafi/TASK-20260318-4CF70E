package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

// AuditLogDTO matches api-spec audit log shape.
type AuditLogDTO struct {
	ID            string         `json:"id"`
	Module        string         `json:"module"`
	Operation     string         `json:"operation"`
	OperatorID    string         `json:"operatorId"`
	RequestSource string         `json:"requestSource,omitempty"`
	RequestID     string         `json:"requestId,omitempty"`
	TargetType    string         `json:"targetType"`
	TargetID      string         `json:"targetId"`
	Before        map[string]any `json:"before,omitempty"`
	After         map[string]any `json:"after,omitempty"`
	CreatedAt     string         `json:"createdAt"`
}

type AuditExportDTO struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func parseJSONObj(b []byte) map[string]any {
	if len(b) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil || m == nil {
		return nil
	}
	return m
}

func toAuditLogDTO(a *model.AuditLog) AuditLogDTO {
	dto := AuditLogDTO{
		ID:         a.ID,
		Module:     a.Module,
		Operation:  a.Operation,
		OperatorID: a.OperatorUserID,
		TargetType: a.TargetType,
		TargetID:   a.TargetID,
		CreatedAt:  a.CreatedAt.UTC().Format(time.RFC3339),
	}
	if a.RequestSource != nil {
		dto.RequestSource = *a.RequestSource
	}
	if a.RequestID != nil {
		dto.RequestID = *a.RequestID
	}
	dto.Before = parseJSONObj(a.BeforeJSON)
	dto.After = parseJSONObj(a.AfterJSON)
	return dto
}

func auditLogOrder(sortBy, sortOrder string) string {
	col := "created_at"
	switch sortBy {
	case "created_at", "module", "operation", "target_type":
		col = sortBy
	}
	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}
	return col + " " + order
}

// ListAuditLogsInput holds query filters.
type ListAuditLogsInput struct {
	Module     string
	TargetType string
	From       *time.Time
	To         *time.Time
}

func (s *AuditService) ListAuditLogs(ctx context.Context, page, pageSize, offset int, sortBy, sortOrder string, in ListAuditLogsInput) ([]AuditLogDTO, int64, int, int, error) {
	rows, total, err := s.repo.ListLogs(ctx, offset, pageSize, auditLogOrder(sortBy, sortOrder), in.Module, in.TargetType, in.From, in.To)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]AuditLogDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toAuditLogDTO(&rows[i]))
	}
	return out, total, page, pageSize, nil
}

// AuditExportFilter is stored as filter_json on export requests.
type AuditExportFilter struct {
	Module     string `json:"module,omitempty"`
	TargetType string `json:"targetType,omitempty"`
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
}

// ErrAuditExportValidation when export filter cannot be encoded or validated.
var ErrAuditExportValidation = errors.New("audit export validation failed")

func (s *AuditService) RequestExport(ctx context.Context, userID string, filter AuditExportFilter) (*AuditExportDTO, error) {
	b, err := json.Marshal(filter)
	if err != nil {
		return nil, ErrAuditExportValidation
	}
	now := time.Now().UTC()
	e := &model.AuditExport{
		ID:                  uuid.NewString(),
		RequestedByUserID:   userID,
		FilterJSON:          b,
		Status:              "pending",
		CreatedAt:           now,
	}
	if err := s.repo.CreateExport(ctx, e); err != nil {
		return nil, err
	}
	return &AuditExportDTO{
		ID:        e.ID,
		Status:    e.Status,
		CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}
