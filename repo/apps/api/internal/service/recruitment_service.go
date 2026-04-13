package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

// ErrForbiddenScope means the principal has no institution scope for this operation.
var ErrForbiddenScope = errors.New("forbidden scope")

type RecruitmentService struct {
	repo *repository.RecruitmentRepository
}

func NewRecruitmentService(repo *repository.RecruitmentRepository) *RecruitmentService {
	return &RecruitmentService{repo: repo}
}

func (s *RecruitmentService) scopeInstitutions(p *access.Principal) ([]string, error) {
	if p == nil {
		return nil, ErrForbiddenScope
	}
	ids := p.AllowedInstitutionIDs()
	if len(ids) == 0 {
		return nil, ErrForbiddenScope
	}
	return ids, nil
}

// CandidateDTO matches api-spec list/detail shape (masked PII).
type CandidateDTO struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	PhoneMasked      string         `json:"phoneMasked"`
	IDNumberMasked   string         `json:"idNumberMasked"`
	Email            string         `json:"email,omitempty"`
	Skills           []string       `json:"skills"`
	ExperienceYears  *int           `json:"experienceYears,omitempty"`
	EducationLevel   *string        `json:"educationLevel,omitempty"`
	Tags             []string       `json:"tags"`
	CustomFields     map[string]any `json:"customFields"`
	InstitutionID    string         `json:"institutionId"`
	DepartmentID     *string        `json:"departmentId,omitempty"`
	TeamID           *string        `json:"teamId,omitempty"`
	CreatedAt        string         `json:"createdAt"`
	UpdatedAt        string         `json:"updatedAt"`
}

// PositionDTO for API responses.
type PositionDTO struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	Status        string  `json:"status"`
	InstitutionID string  `json:"institutionId"`
	DepartmentID  *string `json:"departmentId,omitempty"`
	TeamID        *string `json:"teamId,omitempty"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

func maskPII(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return "••••••••"
}

func toCandidateDTO(c *model.Candidate, tags []string) CandidateDTO {
	skills := make([]string, 0, len(c.Skills))
	for _, sk := range c.Skills {
		skills = append(skills, sk.SkillName)
	}
	if tags == nil {
		tags = []string{}
	}
	return CandidateDTO{
		ID:              c.ID,
		Name:            c.Name,
		PhoneMasked:     maskPII(c.PhoneEnc),
		IDNumberMasked:  maskPII(c.IDNumberEnc),
		Email:           maskPII(c.EmailEnc),
		Skills:          skills,
		ExperienceYears: c.ExperienceYears,
		EducationLevel:  c.EducationLevel,
		Tags:            tags,
		CustomFields:    map[string]any{},
		InstitutionID:   c.InstitutionID,
		DepartmentID:    c.DepartmentID,
		TeamID:          c.TeamID,
		CreatedAt:       c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       c.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toPositionDTO(p *model.Position) PositionDTO {
	return PositionDTO{
		ID:            p.ID,
		Title:         p.Title,
		Description:   p.Description,
		Status:        p.Status,
		InstitutionID: p.InstitutionID,
		DepartmentID:  p.DepartmentID,
		TeamID:        p.TeamID,
		CreatedAt:     p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func candidateOrder(sortBy, sortOrder string) string {
	col := "name"
	switch sortBy {
	case "created_at", "updated_at", "experience_years", "name":
		col = sortBy
	}
	order := "ASC"
	if sortOrder == "desc" {
		order = "DESC"
	}
	return col + " " + order
}

func positionOrder(sortBy, sortOrder string) string {
	col := "title"
	switch sortBy {
	case "created_at", "updated_at", "title", "status":
		col = sortBy
	}
	order := "ASC"
	if sortOrder == "desc" {
		order = "DESC"
	}
	return col + " " + order
}

// ListCandidates returns paginated candidates in scope.
func (s *RecruitmentService) ListCandidates(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder string) ([]CandidateDTO, int64, int, int, error) {
	instIDs, err := s.scopeInstitutions(p)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListCandidates(ctx, instIDs, offset, pageSize, candidateOrder(sortBy, sortOrder))
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	tagMap, err := s.repo.TagsForCandidates(ctx, ids)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]CandidateDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toCandidateDTO(&r, tagMap[r.ID]))
	}
	return out, total, page, pageSize, nil
}

// GetCandidate returns one candidate if in scope.
func (s *RecruitmentService) GetCandidate(ctx context.Context, p *access.Principal, id string) (*CandidateDTO, error) {
	instIDs, err := s.scopeInstitutions(p)
	if err != nil {
		return nil, err
	}
	c, err := s.repo.GetCandidate(ctx, id, instIDs)
	if err != nil {
		return nil, err
	}
	tagMap, err := s.repo.TagsForCandidates(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	dto := toCandidateDTO(c, tagMap[id])
	return &dto, nil
}

// CreateCandidateInput holds POST body fields.
type CreateCandidateInput struct {
	Name            string
	InstitutionID   string
	DepartmentID    *string
	TeamID          *string
	ExperienceYears *int
	EducationLevel  *string
	Skills          []string
	Tags            []string
}

func (s *RecruitmentService) CreateCandidate(ctx context.Context, p *access.Principal, in CreateCandidateInput) (*CandidateDTO, error) {
	if !p.AllowsInstitution(in.InstitutionID) {
		return nil, ErrForbiddenScope
	}
	now := time.Now().UTC()
	c := &model.Candidate{
		ID:              uuid.NewString(),
		InstitutionID:   in.InstitutionID,
		DepartmentID:    in.DepartmentID,
		TeamID:          in.TeamID,
		Name:            in.Name,
		ExperienceYears: in.ExperienceYears,
		EducationLevel:  in.EducationLevel,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	skillSeen := make(map[string]struct{})
	skills := make([]model.CandidateSkill, 0, len(in.Skills))
	for _, name := range in.Skills {
		if name == "" {
			continue
		}
		if _, ok := skillSeen[name]; ok {
			continue
		}
		skillSeen[name] = struct{}{}
		skills = append(skills, model.CandidateSkill{
			ID:        uuid.NewString(),
			SkillName: name,
		})
	}
	tags := make([]string, 0, len(in.Tags))
	seen := map[string]struct{}{}
	for _, t := range in.Tags {
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		tags = append(tags, t)
	}
	if err := s.repo.CreateCandidate(ctx, c, skills, tags); err != nil {
		return nil, err
	}
	loaded, err := s.repo.GetCandidate(ctx, c.ID, []string{in.InstitutionID})
	if err != nil {
		return nil, err
	}
	tagMap, err := s.repo.TagsForCandidates(ctx, []string{c.ID})
	if err != nil {
		return nil, err
	}
	dto := toCandidateDTO(loaded, tagMap[c.ID])
	return &dto, nil
}

// UpdateCandidateInput for PATCH.
type UpdateCandidateInput struct {
	Name            *string
	DepartmentID    *string
	TeamID          *string
	ExperienceYears *int
	EducationLevel  *string
}

func (s *RecruitmentService) UpdateCandidate(ctx context.Context, p *access.Principal, id string, in UpdateCandidateInput) (*CandidateDTO, error) {
	instIDs, err := s.scopeInstitutions(p)
	if err != nil {
		return nil, err
	}
	c, err := s.repo.GetCandidate(ctx, id, instIDs)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		c.Name = *in.Name
	}
	if in.DepartmentID != nil {
		c.DepartmentID = in.DepartmentID
	}
	if in.TeamID != nil {
		c.TeamID = in.TeamID
	}
	if in.ExperienceYears != nil {
		c.ExperienceYears = in.ExperienceYears
	}
	if in.EducationLevel != nil {
		c.EducationLevel = in.EducationLevel
	}
	if err := s.repo.UpdateCandidate(ctx, c, instIDs); err != nil {
		return nil, err
	}
	return s.GetCandidate(ctx, p, id)
}

func (s *RecruitmentService) DeleteCandidate(ctx context.Context, p *access.Principal, id string) error {
	instIDs, err := s.scopeInstitutions(p)
	if err != nil {
		return err
	}
	return s.repo.SoftDeleteCandidate(ctx, id, instIDs)
}

// ListPositions returns paginated positions in scope.
func (s *RecruitmentService) ListPositions(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder string) ([]PositionDTO, int64, int, int, error) {
	instIDs, err := s.scopeInstitutions(p)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListPositions(ctx, instIDs, offset, pageSize, positionOrder(sortBy, sortOrder))
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]PositionDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toPositionDTO(&rows[i]))
	}
	return out, total, page, pageSize, nil
}

func (s *RecruitmentService) GetPosition(ctx context.Context, p *access.Principal, id string) (*PositionDTO, error) {
	instIDs, err := s.scopeInstitutions(p)
	if err != nil {
		return nil, err
	}
	pos, err := s.repo.GetPosition(ctx, id, instIDs)
	if err != nil {
		return nil, err
	}
	dto := toPositionDTO(pos)
	return &dto, nil
}

// CreatePositionInput for POST /positions.
type CreatePositionInput struct {
	InstitutionID string
	Title         string
	Description   *string
	Status        string
	DepartmentID  *string
	TeamID        *string
}

func (s *RecruitmentService) CreatePosition(ctx context.Context, p *access.Principal, in CreatePositionInput) (*PositionDTO, error) {
	if !p.AllowsInstitution(in.InstitutionID) {
		return nil, ErrForbiddenScope
	}
	status := in.Status
	if status == "" {
		status = "open"
	}
	now := time.Now().UTC()
	pos := &model.Position{
		ID:            uuid.NewString(),
		InstitutionID: in.InstitutionID,
		DepartmentID:  in.DepartmentID,
		TeamID:        in.TeamID,
		Title:         in.Title,
		Description:   in.Description,
		Status:        status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreatePosition(ctx, pos); err != nil {
		return nil, err
	}
	dto := toPositionDTO(pos)
	return &dto, nil
}

// UpdatePositionInput for PATCH.
type UpdatePositionInput struct {
	Title       *string
	Description *string
	Status      *string
	DepartmentID *string
	TeamID      *string
}

func (s *RecruitmentService) UpdatePosition(ctx context.Context, p *access.Principal, id string, in UpdatePositionInput) (*PositionDTO, error) {
	instIDs, err := s.scopeInstitutions(p)
	if err != nil {
		return nil, err
	}
	pos, err := s.repo.GetPosition(ctx, id, instIDs)
	if err != nil {
		return nil, err
	}
	if in.Title != nil {
		pos.Title = *in.Title
	}
	if in.Description != nil {
		pos.Description = in.Description
	}
	if in.Status != nil {
		pos.Status = *in.Status
	}
	if in.DepartmentID != nil {
		pos.DepartmentID = in.DepartmentID
	}
	if in.TeamID != nil {
		pos.TeamID = in.TeamID
	}
	if err := s.repo.UpdatePosition(ctx, pos, instIDs); err != nil {
		return nil, err
	}
	return s.GetPosition(ctx, p, id)
}
