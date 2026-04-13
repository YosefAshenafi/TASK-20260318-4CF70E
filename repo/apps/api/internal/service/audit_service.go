package service

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"sort"
	"time"

	"github.com/google/uuid"

	"pharmaops/api/internal/model"
	"pharmaops/api/internal/oplog"
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

// AuditRequestMeta carries operator identity and HTTP request correlation (design §17.2).
type AuditRequestMeta struct {
	OperatorUserID string
	RequestID      string
	RequestSource  *string
}

// AuditMutationInput is one append-only audit row with optional before/after field maps.
type AuditMutationInput struct {
	Module      string
	Operation   string
	TargetType  string
	TargetID    string
	Before      map[string]any
	After       map[string]any
	Meta        AuditRequestMeta
}

// DTOToAuditMap converts a DTO to a map for audit JSON (masked / API-safe fields only).
func DTOToAuditMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

func auditJSONEqual(a, b any) bool {
	ba, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(ba) == string(bb)
}

// mutationChangedKeys lists top-level keys whose values differ between before and after snapshots.
func mutationChangedKeys(before, after map[string]any) []string {
	if before == nil || after == nil {
		return nil
	}
	keys := make(map[string]struct{})
	for k := range before {
		keys[k] = struct{}{}
	}
	for k := range after {
		keys[k] = struct{}{}
	}
	var out []string
	for k := range keys {
		if k == "_changedFields" {
			continue
		}
		bv, bok := before[k]
		av, aok := after[k]
		if !bok || !aok {
			out = append(out, k)
			continue
		}
		if !auditJSONEqual(bv, av) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
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

// LogCandidatePIIRead records that full plaintext PII was returned to the client (design.md §11.2).
// Only field names are stored, never secret values.
func (s *AuditService) LogCandidatePIIRead(ctx context.Context, operatorUserID, candidateID, requestID string, requestSource *string, fields []string) error {
	if s == nil || s.repo == nil {
		return nil
	}
	after, err := json.Marshal(map[string]any{
		"kind":   "full_pii_disclosure",
		"fields": fields,
	})
	if err != nil {
		return err
	}
	row := &model.AuditLog{
		ID:             uuid.NewString(),
		Module:         "recruitment",
		Operation:      "candidate.read_pii",
		OperatorUserID: operatorUserID,
		TargetType:     "candidate",
		TargetID:       candidateID,
		AfterJSON:      after,
		CreatedAt:      time.Now().UTC(),
	}
	if requestID != "" {
		row.RequestID = &requestID
	}
	row.RequestSource = requestSource
	return s.repo.CreateAuditLog(ctx, row)
}

// LogMutation appends a business-mutation audit event (non-repudiation; design §17).
func (s *AuditService) LogMutation(ctx context.Context, in AuditMutationInput) error {
	if s == nil || s.repo == nil {
		return nil
	}
	if in.Meta.OperatorUserID == "" {
		return nil
	}
	if in.Module == "" || in.Operation == "" || in.TargetType == "" || in.TargetID == "" {
		return nil
	}
	var beforeJSON, afterJSON []byte
	var err error
	if len(in.Before) > 0 {
		if beforeJSON, err = json.Marshal(in.Before); err != nil {
			return err
		}
	}
	afterForAudit := in.After
	if len(in.Before) > 0 && len(in.After) > 0 {
		if changed := mutationChangedKeys(in.Before, in.After); len(changed) > 0 {
			afterForAudit = maps.Clone(in.After)
			afterForAudit["_changedFields"] = changed
		}
	}
	if len(afterForAudit) > 0 {
		if afterJSON, err = json.Marshal(afterForAudit); err != nil {
			return err
		}
	}
	row := &model.AuditLog{
		ID:             uuid.NewString(),
		Module:         in.Module,
		Operation:      in.Operation,
		OperatorUserID: in.Meta.OperatorUserID,
		TargetType:     in.TargetType,
		TargetID:       in.TargetID,
		BeforeJSON:     beforeJSON,
		AfterJSON:      afterJSON,
		CreatedAt:      time.Now().UTC(),
	}
	if in.Meta.RequestID != "" {
		rid := in.Meta.RequestID
		row.RequestID = &rid
	}
	row.RequestSource = in.Meta.RequestSource
	oplog.AuditWrite(in.Meta.RequestID, in.Meta.OperatorUserID, in.Module, in.Operation, in.TargetType, in.TargetID)
	return s.repo.CreateAuditLog(ctx, row)
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

func (s *AuditService) RequestExport(ctx context.Context, userID string, filter AuditExportFilter, meta AuditRequestMeta) (*AuditExportDTO, error) {
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
	opMeta := meta
	if opMeta.OperatorUserID == "" {
		opMeta.OperatorUserID = userID
	}
	_ = s.LogMutation(ctx, AuditMutationInput{
		Module:     "audit",
		Operation:  "audit.export_requested",
		TargetType: "audit_export",
		TargetID:   e.ID,
		After: map[string]any{
			"filter": DTOToAuditMap(filter),
		},
		Meta: opMeta,
	})
	return &AuditExportDTO{
		ID:        e.ID,
		Status:    e.Status,
		CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}
