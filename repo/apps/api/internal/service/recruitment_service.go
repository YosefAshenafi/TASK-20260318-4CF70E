package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"pharmaops/api/internal/access"
	cryptopii "pharmaops/api/internal/crypto/pii"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/oplog"
	"pharmaops/api/internal/repository"
)

// ErrForbiddenScope means the principal has no usable data scope for this operation.
var ErrForbiddenScope = errors.New("forbidden scope")

func requireScope(p *access.Principal) error {
	if p == nil || len(p.Scopes) == 0 {
		return ErrForbiddenScope
	}
	return nil
}

// ErrPIINotConfigured means PII_AES_KEY_HEX is missing but plaintext PII was submitted.
var ErrPIINotConfigured = errors.New("PII encryption key not configured")

const permissionRecruitmentViewPII = "recruitment.view_pii"

type RecruitmentService struct {
	repo      *repository.RecruitmentRepository
	piiCipher *cryptopii.Cipher
	audit     *AuditService
}

func NewRecruitmentService(repo *repository.RecruitmentRepository, piiCipher *cryptopii.Cipher, audit *AuditService) *RecruitmentService {
	return &RecruitmentService{repo: repo, piiCipher: piiCipher, audit: audit}
}

// GetCandidateOpts carries request metadata for audit logging when full PII is returned.
type GetCandidateOpts struct {
	OperatorUserID string
	RequestID      string
	RequestSource  *string
}

// AuditMeta returns operator/request correlation for mutation audit rows (design §17).
func (o GetCandidateOpts) AuditMeta() AuditRequestMeta {
	return AuditRequestMeta{
		OperatorUserID: o.OperatorUserID,
		RequestID:      o.RequestID,
		RequestSource:  o.RequestSource,
	}
}

// CandidateDTO matches api-spec list/detail shape (masked PII; full fields when recruitment.view_pii).
type CandidateDTO struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	PhoneMasked     string         `json:"phoneMasked"`
	IDNumberMasked  string         `json:"idNumberMasked"`
	Email           string         `json:"email,omitempty"`
	Phone           *string        `json:"phone,omitempty"`
	IDNumber        *string        `json:"idNumber,omitempty"`
	Skills          []string       `json:"skills"`
	ExperienceYears *int           `json:"experienceYears,omitempty"`
	EducationLevel  *string        `json:"educationLevel,omitempty"`
	Tags            []string       `json:"tags"`
	CustomFields    map[string]any `json:"customFields"`
	InstitutionID   string         `json:"institutionId"`
	DepartmentID    *string        `json:"departmentId,omitempty"`
	TeamID          *string        `json:"teamId,omitempty"`
	CreatedAt       string         `json:"createdAt"`
	UpdatedAt       string         `json:"updatedAt"`
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

func (s *RecruitmentService) revealFullPII(p *access.Principal) bool {
	return p != nil && p.Has(permissionRecruitmentViewPII)
}

func disclosedPIIFieldNames(d *CandidateDTO) []string {
	if d == nil {
		return nil
	}
	var fields []string
	if d.Phone != nil && *d.Phone != "" {
		fields = append(fields, "phone")
	}
	if d.IDNumber != nil && *d.IDNumber != "" {
		fields = append(fields, "idNumber")
	}
	if d.Email != "" {
		fields = append(fields, "email")
	}
	return fields
}

// maybeAuditPIIRead appends an audit row when full PII was included in the response. Logging failure is ignored (best-effort).
func (s *RecruitmentService) maybeAuditPIIRead(ctx context.Context, p *access.Principal, candidateID string, dto *CandidateDTO, opts GetCandidateOpts) {
	if s.audit == nil || dto == nil || p == nil || !s.revealFullPII(p) {
		return
	}
	if opts.OperatorUserID == "" {
		return
	}
	fields := disclosedPIIFieldNames(dto)
	if len(fields) == 0 {
		return
	}
	reqID := opts.RequestID
	oplog.PIIAccess(reqID, opts.OperatorUserID, candidateID)
	_ = s.audit.LogCandidatePIIRead(ctx, opts.OperatorUserID, candidateID, reqID, opts.RequestSource, fields)
}

func (s *RecruitmentService) decryptPIIField(blob []byte) (plain string, ok bool) {
	if len(blob) == 0 {
		return "", true
	}
	if s.piiCipher == nil || !s.piiCipher.Valid() {
		return "", false
	}
	out, err := s.piiCipher.DecryptString(blob)
	if err != nil {
		return "", false
	}
	return out, true
}

func (s *RecruitmentService) sealPII(plain string) ([]byte, error) {
	t := strings.TrimSpace(plain)
	if t == "" {
		return nil, nil
	}
	if s.piiCipher == nil || !s.piiCipher.Valid() {
		oplog.EncryptionError("", "PII encryption key not configured")
		return nil, ErrPIINotConfigured
	}
	ct, err := s.piiCipher.EncryptString(t)
	if err != nil {
		oplog.EncryptionError("", "PII encrypt failed: "+err.Error())
		return nil, err
	}
	return ct, nil
}

func (s *RecruitmentService) encryptOptional(in *string) ([]byte, error) {
	if in == nil {
		return nil, nil
	}
	return s.sealPII(*in)
}

func normalizePhoneForDuplicate(raw string) string {
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeIDForDuplicate(raw string) string {
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range strings.ToUpper(strings.TrimSpace(raw)) {
		if unicode.IsDigit(r) || (r >= 'A' && r <= 'Z') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (s *RecruitmentService) duplicateHash(plain *string, normalizer func(string) string) (*string, error) {
	if plain == nil {
		return nil, nil
	}
	norm := normalizer(*plain)
	if norm == "" {
		return nil, nil
	}
	if s.piiCipher == nil || !s.piiCipher.Valid() {
		return nil, ErrPIINotConfigured
	}
	sum, err := s.piiCipher.DigestHex(norm)
	if err != nil {
		return nil, fmt.Errorf("duplicate hash failed: %w", err)
	}
	if sum == "" {
		return nil, nil
	}
	return &sum, nil
}

func (s *RecruitmentService) candidateDTO(c *model.Candidate, tags []string, reveal bool) CandidateDTO {
	skills := make([]string, 0, len(c.Skills))
	for _, sk := range c.Skills {
		skills = append(skills, sk.SkillName)
	}
	if tags == nil {
		tags = []string{}
	}

	phonePlain, phoneOK := s.decryptPIIField(c.PhoneEnc)
	idPlain, idOK := s.decryptPIIField(c.IDNumberEnc)
	emailPlain, emailOK := s.decryptPIIField(c.EmailEnc)

	phoneMasked := cryptopii.PartialMaskPhone(phonePlain)
	if !phoneOK && len(c.PhoneEnc) > 0 {
		phoneMasked = "••••••••"
	}
	idMasked := cryptopii.PartialMaskID(idPlain)
	if !idOK && len(c.IDNumberEnc) > 0 {
		idMasked = "••••••••"
	}

	var emailOut string
	if reveal {
		if emailOK {
			emailOut = emailPlain
		} else if len(c.EmailEnc) > 0 {
			emailOut = "••••••••"
		}
	} else {
		if emailOK {
			emailOut = cryptopii.PartialMaskEmail(emailPlain)
		} else if len(c.EmailEnc) > 0 {
			emailOut = "••••••••"
		}
	}

	customFields := map[string]any{}
	if len(c.CustomFieldsJSON) > 0 {
		_ = json.Unmarshal(c.CustomFieldsJSON, &customFields)
	}

	dto := CandidateDTO{
		ID:              c.ID,
		Name:            c.Name,
		PhoneMasked:     phoneMasked,
		IDNumberMasked:  idMasked,
		Email:           emailOut,
		Skills:          skills,
		ExperienceYears: c.ExperienceYears,
		EducationLevel:  c.EducationLevel,
		Tags:            tags,
		CustomFields:    customFields,
		InstitutionID:   c.InstitutionID,
		DepartmentID:    c.DepartmentID,
		TeamID:          c.TeamID,
		CreatedAt:       c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       c.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if reveal {
		if phoneOK && phonePlain != "" {
			p := phonePlain
			dto.Phone = &p
		}
		if idOK && idPlain != "" {
			x := idPlain
			dto.IDNumber = &x
		}
	}
	return dto
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

// CandidateSearchParams mirrors the query parameters for candidate search/filter.
type CandidateSearchParams struct {
	Keyword        string
	Skills         []string
	EducationLevel string
	MinExperience  *int
	MaxExperience  *int
}

// ListCandidates returns paginated candidates in scope with optional filters.
func (s *RecruitmentService) ListCandidates(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder string, search CandidateSearchParams) ([]CandidateDTO, int64, int, int, error) {
	if err := requireScope(p); err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListCandidates(ctx, p, offset, pageSize, candidateOrder(sortBy, sortOrder), repository.CandidateFilter{
		Keyword:        search.Keyword,
		Skills:         search.Skills,
		EducationLevel: search.EducationLevel,
		MinExperience:  search.MinExperience,
		MaxExperience:  search.MaxExperience,
	})
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
		out = append(out, s.candidateDTO(&r, tagMap[r.ID], false))
	}
	return out, total, page, pageSize, nil
}

// GetCandidate returns one candidate if in scope. Full plaintext PII is included only when the principal has recruitment.view_pii.
func (s *RecruitmentService) GetCandidate(ctx context.Context, p *access.Principal, id string, opts GetCandidateOpts) (*CandidateDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	c, err := s.repo.GetCandidate(ctx, id, p)
	if err != nil {
		return nil, err
	}
	tagMap, err := s.repo.TagsForCandidates(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	dto := s.candidateDTO(c, tagMap[id], s.revealFullPII(p))
	s.maybeAuditPIIRead(ctx, p, id, &dto, opts)
	return &dto, nil
}

// CreateCandidateInput holds POST body fields.
type CreateCandidateInput struct {
	Name            string
	InstitutionID   string
	DepartmentID    *string
	TeamID          *string
	Phone           *string
	IDNumber        *string
	Email           *string
	ExperienceYears *int
	EducationLevel  *string
	Skills          []string
	Tags            []string
	CustomFields    map[string]any
}

func (s *RecruitmentService) CreateCandidate(ctx context.Context, p *access.Principal, in CreateCandidateInput, opts GetCandidateOpts) (*CandidateDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if !p.RowVisible(in.InstitutionID, in.DepartmentID, in.TeamID) {
		return nil, ErrForbiddenScope
	}
	phoneEnc, err := s.encryptOptional(in.Phone)
	if err != nil {
		return nil, err
	}
	idEnc, err := s.encryptOptional(in.IDNumber)
	if err != nil {
		return nil, err
	}
	emailEnc, err := s.encryptOptional(in.Email)
	if err != nil {
		return nil, err
	}
	phoneNormHash, err := s.duplicateHash(in.Phone, normalizePhoneForDuplicate)
	if err != nil {
		return nil, err
	}
	idNormHash, err := s.duplicateHash(in.IDNumber, normalizeIDForDuplicate)
	if err != nil {
		return nil, err
	}
	var cfJSON []byte
	if len(in.CustomFields) > 0 {
		cfJSON, _ = json.Marshal(in.CustomFields)
	}
	now := time.Now().UTC()
	c := &model.Candidate{
		ID:               uuid.NewString(),
		InstitutionID:    in.InstitutionID,
		DepartmentID:     in.DepartmentID,
		TeamID:           in.TeamID,
		Name:             in.Name,
		PhoneEnc:         phoneEnc,
		PhoneNormHash:    phoneNormHash,
		IDNumberEnc:      idEnc,
		IDNumberNormHash: idNormHash,
		EmailEnc:         emailEnc,
		PIIKeyVersion:    1,
		ExperienceYears:  in.ExperienceYears,
		EducationLevel:   in.EducationLevel,
		CustomFieldsJSON: cfJSON,
		CreatedAt:        now,
		UpdatedAt:        now,
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
	loaded, err := s.repo.GetCandidate(ctx, c.ID, p)
	if err != nil {
		return nil, err
	}
	tagMap, err := s.repo.TagsForCandidates(ctx, []string{c.ID})
	if err != nil {
		return nil, err
	}
	dto := s.candidateDTO(loaded, tagMap[c.ID], s.revealFullPII(p))
	s.maybeAuditPIIRead(ctx, p, c.ID, &dto, opts)
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "recruitment",
		Operation:  "candidate.create",
		TargetType: "candidate",
		TargetID:   c.ID,
		After:      DTOToAuditMap(&dto),
		Meta:       opts.AuditMeta(),
	})
	return &dto, nil
}

// UpdateCandidateInput for PATCH.
type UpdateCandidateInput struct {
	Name            *string
	DepartmentID    *string
	TeamID          *string
	Phone           *string
	IDNumber        *string
	Email           *string
	ExperienceYears *int
	EducationLevel  *string
	CustomFields    map[string]any
}

func (s *RecruitmentService) UpdateCandidate(ctx context.Context, p *access.Principal, id string, in UpdateCandidateInput, opts GetCandidateOpts) (*CandidateDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	c, err := s.repo.GetCandidate(ctx, id, p)
	if err != nil {
		return nil, err
	}
	beforeTags, _ := s.repo.TagsForCandidates(ctx, []string{id})
	beforeMap := DTOToAuditMap(s.candidateDTO(c, beforeTags[id], false))
	if in.Name != nil {
		c.Name = *in.Name
	}
	if in.DepartmentID != nil {
		c.DepartmentID = in.DepartmentID
	}
	if in.TeamID != nil {
		c.TeamID = in.TeamID
	}
	if in.Phone != nil {
		b, err := s.sealPII(*in.Phone)
		if err != nil {
			return nil, err
		}
		c.PhoneEnc = b
		h, err := s.duplicateHash(in.Phone, normalizePhoneForDuplicate)
		if err != nil {
			return nil, err
		}
		c.PhoneNormHash = h
	}
	if in.IDNumber != nil {
		b, err := s.sealPII(*in.IDNumber)
		if err != nil {
			return nil, err
		}
		c.IDNumberEnc = b
		h, err := s.duplicateHash(in.IDNumber, normalizeIDForDuplicate)
		if err != nil {
			return nil, err
		}
		c.IDNumberNormHash = h
	}
	if in.Email != nil {
		b, err := s.sealPII(*in.Email)
		if err != nil {
			return nil, err
		}
		c.EmailEnc = b
	}
	if in.ExperienceYears != nil {
		c.ExperienceYears = in.ExperienceYears
	}
	if in.EducationLevel != nil {
		c.EducationLevel = in.EducationLevel
	}
	if in.CustomFields != nil {
		cfJSON, _ := json.Marshal(in.CustomFields)
		c.CustomFieldsJSON = cfJSON
	}
	if !p.RowVisible(c.InstitutionID, c.DepartmentID, c.TeamID) {
		return nil, ErrForbiddenScope
	}
	if err := s.repo.UpdateCandidate(ctx, c, p); err != nil {
		return nil, err
	}
	afterC, err := s.repo.GetCandidate(ctx, id, p)
	if err != nil {
		return nil, err
	}
	afterTags, _ := s.repo.TagsForCandidates(ctx, []string{id})
	afterMap := DTOToAuditMap(s.candidateDTO(afterC, afterTags[id], false))
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "recruitment",
		Operation:  "candidate.update",
		TargetType: "candidate",
		TargetID:   id,
		Before:     beforeMap,
		After:      afterMap,
		Meta:       opts.AuditMeta(),
	})
	return s.GetCandidate(ctx, p, id, opts)
}

func (s *RecruitmentService) DeleteCandidate(ctx context.Context, p *access.Principal, id string, meta AuditRequestMeta) error {
	if err := requireScope(p); err != nil {
		return err
	}
	beforeC, err := s.repo.GetCandidate(ctx, id, p)
	if err != nil {
		return err
	}
	tags, _ := s.repo.TagsForCandidates(ctx, []string{id})
	beforeMap := DTOToAuditMap(s.candidateDTO(beforeC, tags[id], false))
	if err := s.repo.SoftDeleteCandidate(ctx, id, p); err != nil {
		return err
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "recruitment",
		Operation:  "candidate.delete",
		TargetType: "candidate",
		TargetID:   id,
		Before:     beforeMap,
		After:      map[string]any{"deleted": true},
		Meta:       meta,
	})
	return nil
}

// ListPositions returns paginated positions in scope.
func (s *RecruitmentService) ListPositions(ctx context.Context, p *access.Principal, page, pageSize, offset int, sortBy, sortOrder string) ([]PositionDTO, int64, int, int, error) {
	if err := requireScope(p); err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListPositions(ctx, p, offset, pageSize, positionOrder(sortBy, sortOrder))
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
	if err := requireScope(p); err != nil {
		return nil, err
	}
	pos, err := s.repo.GetPosition(ctx, id, p)
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

func (s *RecruitmentService) CreatePosition(ctx context.Context, p *access.Principal, in CreatePositionInput, meta AuditRequestMeta) (*PositionDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	if !p.RowVisible(in.InstitutionID, in.DepartmentID, in.TeamID) {
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
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "recruitment",
		Operation:  "position.create",
		TargetType: "position",
		TargetID:   pos.ID,
		After:      DTOToAuditMap(&dto),
		Meta:       meta,
	})
	return &dto, nil
}

// UpdatePositionInput for PATCH.
type UpdatePositionInput struct {
	Title        *string
	Description  *string
	Status       *string
	DepartmentID *string
	TeamID       *string
}

func (s *RecruitmentService) UpdatePosition(ctx context.Context, p *access.Principal, id string, in UpdatePositionInput, meta AuditRequestMeta) (*PositionDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	pos, err := s.repo.GetPosition(ctx, id, p)
	if err != nil {
		return nil, err
	}
	beforeMap := DTOToAuditMap(toPositionDTO(pos))
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
	if !p.RowVisible(pos.InstitutionID, pos.DepartmentID, pos.TeamID) {
		return nil, ErrForbiddenScope
	}
	if err := s.repo.UpdatePosition(ctx, pos, p); err != nil {
		return nil, err
	}
	loaded, err := s.repo.GetPosition(ctx, id, p)
	if err != nil {
		return nil, err
	}
	afterMap := DTOToAuditMap(toPositionDTO(loaded))
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "recruitment",
		Operation:  "position.update",
		TargetType: "position",
		TargetID:   id,
		Before:     beforeMap,
		After:      afterMap,
		Meta:       meta,
	})
	return s.GetPosition(ctx, p, id)
}
