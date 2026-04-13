package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
)

// ErrMergeValidationFailed is returned when merge preconditions fail.
var ErrMergeValidationFailed = errors.New("merge validation failed")

// ErrImportValidationFailed is returned when import rows fail validation.
var ErrImportValidationFailed = errors.New("import validation failed")

// MatchScoreDTO matches api-spec match score shape.
type MatchScoreDTO struct {
	Score     int `json:"score"`
	Breakdown struct {
		Skills     int `json:"skills"`
		Experience int `json:"experience"`
		Education  int `json:"education"`
	} `json:"breakdown"`
	Reasons []string `json:"reasons"`
}

// ImportBatchDTO is returned for import batch APIs.
type ImportBatchDTO struct {
	ID                   string          `json:"id"`
	InstitutionID        string          `json:"institutionId"`
	DepartmentID         *string         `json:"departmentId,omitempty"`
	TeamID               *string         `json:"teamId,omitempty"`
	Status               string          `json:"status"`
	ValidationReport     json.RawMessage `json:"validationReport,omitempty"`
	CreatedByUserID      string          `json:"createdByUserId"`
	CommittedAt          *string         `json:"committedAt,omitempty"`
	CreatedAt            string          `json:"createdAt"`
}

// MergeHistoryDTO is one merge record.
type MergeHistoryDTO struct {
	ID               string          `json:"id"`
	BaseCandidateID  string          `json:"baseCandidateId"`
	SourceCandidateIDs []string      `json:"sourceCandidateIds"`
	OperatorUserID   string          `json:"operatorUserId"`
	CreatedAt        string          `json:"createdAt"`
}

// DuplicateGroupDTO lists candidates sharing the same phone or ID number.
type DuplicateGroupDTO struct {
	MatchType     string   `json:"matchType"`
	InstitutionID string   `json:"institutionId"`
	CandidateIDs  []string `json:"candidateIds"`
}

type importStaging struct {
	Rows []ImportStagingRow `json:"rows"`
}

// ImportStagingRow is one row in a candidate import batch.
type ImportStagingRow struct {
	Name            string         `json:"name"`
	Phone           *string        `json:"phone,omitempty"`
	IDNumber        *string        `json:"idNumber,omitempty"`
	Email           *string        `json:"email,omitempty"`
	Skills          []string       `json:"skills"`
	Tags            []string       `json:"tags"`
	ExperienceYears *int           `json:"experienceYears"`
	EducationLevel  *string        `json:"educationLevel"`
	CustomFields    map[string]any `json:"customFields,omitempty"`
}

type importValidationReport struct {
	Rows   []ImportStagingRow `json:"rows"`
	Errors []importRowError   `json:"errors,omitempty"`
}

type importRowError struct {
	RowIndex int    `json:"rowIndex"`
	Message  string `json:"message"`
}

// CreateImportBatch stages rows for a later commit.
func (s *RecruitmentService) CreateImportBatch(ctx context.Context, p *access.Principal, userID string, institutionID string, rows []ImportStagingRow) (*ImportBatchDTO, error) {
	dept, team := access.DefaultOrgAssignment(p, institutionID)
	if !p.RowVisible(institutionID, dept, team) {
		return nil, ErrForbiddenScope
	}
	if len(rows) == 0 {
		return nil, ErrImportValidationFailed
	}
	report := importValidationReport{Rows: rows}
	for i, row := range rows {
		if strings.TrimSpace(row.Name) == "" {
			report.Errors = append(report.Errors, importRowError{RowIndex: i, Message: "name is required"})
		}
	}
	if len(report.Errors) > 0 {
		return nil, ErrImportValidationFailed
	}
	raw, err := json.Marshal(importStaging{Rows: rows})
	if err != nil {
		return nil, err
	}
	validJSON, err := json.Marshal(report)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	b := &model.CandidateImportBatch{
		ID:                   uuid.NewString(),
		InstitutionID:        institutionID,
		DepartmentID:         dept,
		TeamID:               team,
		Status:               "pending",
		MappingJSON:          raw,
		ValidationReportJSON: validJSON,
		CreatedByUserID:      userID,
		CreatedAt:            now,
	}
	if err := s.repo.CreateImportBatch(ctx, b); err != nil {
		return nil, err
	}
	return toImportDTO(b), nil
}

func toImportDTO(b *model.CandidateImportBatch) *ImportBatchDTO {
	d := &ImportBatchDTO{
		ID:              b.ID,
		InstitutionID:   b.InstitutionID,
		DepartmentID:    b.DepartmentID,
		TeamID:          b.TeamID,
		Status:          b.Status,
		CreatedByUserID: b.CreatedByUserID,
		CreatedAt:       b.CreatedAt.UTC().Format(time.RFC3339),
	}
	if len(b.ValidationReportJSON) > 0 {
		d.ValidationReport = b.ValidationReportJSON
	}
	if b.CommittedAt != nil {
		t := b.CommittedAt.UTC().Format(time.RFC3339)
		d.CommittedAt = &t
	}
	return d
}

// GetImportBatch returns one import batch in scope.
func (s *RecruitmentService) GetImportBatch(ctx context.Context, p *access.Principal, id string) (*ImportBatchDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	b, err := s.repo.GetImportBatch(ctx, id, p)
	if err != nil {
		return nil, err
	}
	return toImportDTO(b), nil
}

// CommitImportBatch creates candidates from a pending batch.
func (s *RecruitmentService) CommitImportBatch(ctx context.Context, p *access.Principal, id string) (*ImportBatchDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	b, err := s.repo.GetImportBatch(ctx, id, p)
	if err != nil {
		return nil, err
	}
	if b.Status != "pending" {
		return nil, ErrImportValidationFailed
	}
	var st importStaging
	if err := json.Unmarshal(b.MappingJSON, &st); err != nil {
		return nil, ErrImportValidationFailed
	}
	defDept, defTeam := access.DefaultOrgAssignment(p, b.InstitutionID)
	created := 0
	for _, row := range st.Rows {
		if strings.TrimSpace(row.Name) == "" {
			continue
		}
		_, err := s.CreateCandidate(ctx, p, CreateCandidateInput{
			Name:            strings.TrimSpace(row.Name),
			InstitutionID:   b.InstitutionID,
			DepartmentID:    defDept,
			TeamID:          defTeam,
			Phone:           row.Phone,
			IDNumber:        row.IDNumber,
			Email:           row.Email,
			ExperienceYears: row.ExperienceYears,
			EducationLevel:  row.EducationLevel,
			Skills:          row.Skills,
			Tags:            row.Tags,
			CustomFields:    row.CustomFields,
		}, GetCandidateOpts{})
		if err != nil {
			return nil, err
		}
		created++
	}
	if created == 0 {
		return nil, ErrImportValidationFailed
	}
	summary, _ := json.Marshal(map[string]any{"createdCount": created, "rows": len(st.Rows)})
	committedAt := time.Now().UTC()
	if err := s.repo.UpdateImportBatchCommitted(ctx, id, p, summary, committedAt); err != nil {
		return nil, err
	}
	out, err := s.repo.GetImportBatch(ctx, id, p)
	if err != nil {
		return nil, err
	}
	return toImportDTO(out), nil
}

// ListDuplicateGroups returns groups of candidates sharing the same phone or ID number.
func (s *RecruitmentService) ListDuplicateGroups(ctx context.Context, p *access.Principal) ([]DuplicateGroupDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	groups, err := s.repo.ListDuplicateGroups(ctx, p)
	if err != nil {
		return nil, err
	}
	out := make([]DuplicateGroupDTO, 0, len(groups))
	for _, g := range groups {
		out = append(out, DuplicateGroupDTO{
			MatchType:     g.MatchType,
			InstitutionID: g.InstitutionID,
			CandidateIDs:  g.CandidateIDs,
		})
	}
	return out, nil
}

// MergeCandidatesInput mirrors api-spec merge request.
type MergeCandidatesInput struct {
	BaseCandidateID    string
	SourceCandidateIDs []string
	Strategy           string
}

// MergeCandidates runs duplicate merge with history logging.
func (s *RecruitmentService) MergeCandidates(ctx context.Context, p *access.Principal, userID string, in MergeCandidatesInput, meta AuditRequestMeta) error {
	if err := requireScope(p); err != nil {
		return err
	}
	if in.Strategy != "" && in.Strategy != "latest_wins_fill_missing" {
		return ErrMergeValidationFailed
	}
	if len(in.SourceCandidateIDs) == 0 {
		return ErrMergeValidationFailed
	}
	strategy := in.Strategy
	if strategy == "" {
		strategy = "latest_wins_fill_missing"
	}
	if err := s.repo.MergeIntoBase(ctx, in.BaseCandidateID, in.SourceCandidateIDs, p, userID, strategy); err != nil {
		return err
	}
	op := meta
	if op.OperatorUserID == "" {
		op.OperatorUserID = userID
	}
	_ = s.audit.LogMutation(ctx, AuditMutationInput{
		Module:     "recruitment",
		Operation:  "candidate.merge",
		TargetType: "candidate",
		TargetID:   in.BaseCandidateID,
		Before: map[string]any{
			"sourceCandidateIds": in.SourceCandidateIDs,
		},
		After: map[string]any{
			"strategy": strategy,
			"merged":   true,
		},
		Meta: op,
	})
	return nil
}

// ListMergeHistory returns paginated merge records in scope.
func (s *RecruitmentService) ListMergeHistory(ctx context.Context, p *access.Principal, page, pageSize, offset int) ([]MergeHistoryDTO, int64, int, int, error) {
	if err := requireScope(p); err != nil {
		return nil, 0, page, pageSize, err
	}
	rows, total, err := s.repo.ListMergeHistory(ctx, p, offset, pageSize)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]MergeHistoryDTO, 0, len(rows))
	for _, r := range rows {
		var src []string
		_ = json.Unmarshal(r.SourceCandidateIDsJSON, &src)
		out = append(out, MergeHistoryDTO{
			ID:                 r.ID,
			BaseCandidateID:    r.BaseCandidateID,
			SourceCandidateIDs: src,
			OperatorUserID:     r.OperatorUserID,
			CreatedAt:          r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, total, page, pageSize, nil
}

// MatchCandidateToPosition scores a candidate against a position and stores a snapshot.
func (s *RecruitmentService) MatchCandidateToPosition(ctx context.Context, p *access.Principal, candidateID, positionID string) (*MatchScoreDTO, error) {
	return s.matchPair(ctx, p, candidateID, positionID)
}

// MatchPositionToCandidate is symmetric with MatchCandidateToPosition.
func (s *RecruitmentService) MatchPositionToCandidate(ctx context.Context, p *access.Principal, positionID, candidateID string) (*MatchScoreDTO, error) {
	return s.matchPair(ctx, p, candidateID, positionID)
}

func (s *RecruitmentService) matchPair(ctx context.Context, p *access.Principal, candidateID, positionID string) (*MatchScoreDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	cand, err := s.repo.GetCandidate(ctx, candidateID, p)
	if err != nil {
		return nil, err
	}
	pos, err := s.repo.GetPosition(ctx, positionID, p)
	if err != nil {
		return nil, err
	}
	reqs, err := s.repo.ListPositionRequirements(ctx, positionID)
	if err != nil {
		return nil, err
	}
	skills := make([]string, 0, len(cand.Skills))
	for _, sk := range cand.Skills {
		skills = append(skills, sk.SkillName)
	}
	dto, br, reasons := computeMatchScore(cand, skills, pos, reqs)
	bj, _ := json.Marshal(br)
	rj, _ := json.Marshal(reasons)
	snap := &model.MatchScoreSnapshot{
		ID:            uuid.NewString(),
		CandidateID:   candidateID,
		PositionID:    positionID,
		Score:         uint16(dto.Score),
		BreakdownJSON: bj,
		ReasonsJSON:   rj,
		ComputedAt:    time.Now().UTC(),
	}
	if err := s.repo.CreateMatchSnapshot(ctx, snap); err != nil {
		return nil, err
	}
	return dto, nil
}

type breakdownMap struct {
	Skills     int `json:"skills"`
	Experience int `json:"experience"`
	Education  int `json:"education"`
}

func computeMatchScore(cand *model.Candidate, skillNames []string, pos *model.Position, reqs []model.PositionRequirement) (*MatchScoreDTO, breakdownMap, []string) {
	reqSkillNames := make([]string, 0)
	for _, r := range reqs {
		if r.SkillName != "" {
			reqSkillNames = append(reqSkillNames, r.SkillName)
		}
	}
	if len(reqSkillNames) == 0 {
		reqSkillNames = tokenizeTitle(pos.Title)
	}
	cSet := map[string]struct{}{}
	for _, s := range skillNames {
		cSet[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
	}
	matched := 0
	for _, req := range reqSkillNames {
		if _, ok := cSet[strings.ToLower(strings.TrimSpace(req))]; ok {
			matched++
		}
	}
	n := len(reqSkillNames)
	if n == 0 {
		n = 1
	}
	skillRatio := float64(matched) / float64(n)
	skillsPart := int(math.Round(skillRatio * 50))
	if skillsPart > 50 {
		skillsPart = 50
	}

	expPart := 0
	if cand.ExperienceYears != nil {
		y := *cand.ExperienceYears
		if y < 0 {
			y = 0
		}
		if y > 10 {
			y = 10
		}
		expPart = int(math.Round(float64(y) / 10.0 * 30))
	}

	eduPart := educationMatchPoints(cand.EducationLevel)

	total := skillsPart + expPart + eduPart
	if total > 100 {
		total = 100
	}

	reasons := []string{
		"Skills alignment contributes " + strconv.Itoa(skillsPart) + "/50 (matched required skills vs profile).",
		"Experience contributes " + strconv.Itoa(expPart) + "/30.",
		"Education contributes " + strconv.Itoa(eduPart) + "/20.",
	}
	br := breakdownMap{Skills: skillsPart, Experience: expPart, Education: eduPart}
	dto := &MatchScoreDTO{Score: total, Reasons: reasons}
	dto.Breakdown.Skills = skillsPart
	dto.Breakdown.Experience = expPart
	dto.Breakdown.Education = eduPart
	return dto, br, reasons
}

func tokenizeTitle(title string) []string {
	parts := strings.FieldsFunc(strings.ToLower(title), func(r rune) bool {
		return r == ' ' || r == ',' || r == '/' || r == '-' || r == '(' || r == ')'
	})
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) < 3 {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func educationMatchPoints(level *string) int {
	if level == nil {
		return 0
	}
	s := strings.ToLower(strings.TrimSpace(*level))
	switch {
	case strings.Contains(s, "phd") || strings.Contains(s, "doctor"):
		return 20
	case strings.Contains(s, "master") || strings.Contains(s, "硕士"):
		return 16
	case strings.Contains(s, "bachelor") || strings.Contains(s, "本科"):
		return 12
	case s != "":
		return 8
	default:
		return 0
	}
}

// SimilarCandidates returns top similar profiles by skill Jaccard similarity.
func (s *RecruitmentService) SimilarCandidates(ctx context.Context, p *access.Principal, candidateID string, limit int) ([]CandidateDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	self, err := s.repo.GetCandidate(ctx, candidateID, p)
	if err != nil {
		return nil, err
	}
	selfSkills := skillSetOfCandidate(self)
	others, err := s.repo.ListCandidatesForSimilarity(ctx, p, candidateID, 50)
	if err != nil {
		return nil, err
	}
	type candScore struct {
		c CandidateDTO
		j float64
	}
	var buf []candScore
	for i := range others {
		other := &others[i]
		tagMap, err := s.repo.TagsForCandidates(ctx, []string{other.ID})
		if err != nil {
			return nil, err
		}
		dto := s.candidateDTO(other, tagMap[other.ID], false)
		j := jaccard(selfSkills, skillSetOfCandidate(other))
		buf = append(buf, candScore{c: dto, j: j})
	}
	sort.Slice(buf, func(i, j int) bool {
		if buf[i].j == buf[j].j {
			return buf[i].c.Name < buf[j].c.Name
		}
		return buf[i].j > buf[j].j
	})
	if limit <= 0 {
		limit = 10
	}
	if limit > len(buf) {
		limit = len(buf)
	}
	out := make([]CandidateDTO, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, buf[i].c)
	}
	return out, nil
}

func skillSetOfCandidate(c *model.Candidate) map[string]struct{} {
	m := make(map[string]struct{})
	for _, sk := range c.Skills {
		m[strings.ToLower(strings.TrimSpace(sk.SkillName))] = struct{}{}
	}
	return m
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1
	}
	inter := 0
	for k := range a {
		if _, ok := b[k]; ok {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

// SimilarPositions returns positions ranked by overlap with candidate skills and title tokens.
func (s *RecruitmentService) SimilarPositions(ctx context.Context, p *access.Principal, positionID string, limit int) ([]PositionDTO, error) {
	if err := requireScope(p); err != nil {
		return nil, err
	}
	pos, err := s.repo.GetPosition(ctx, positionID, p)
	if err != nil {
		return nil, err
	}
	reqs, err := s.repo.ListPositionRequirements(ctx, positionID)
	if err != nil {
		return nil, err
	}
	want := map[string]struct{}{}
	for _, r := range reqs {
		want[strings.ToLower(strings.TrimSpace(r.SkillName))] = struct{}{}
	}
	for _, t := range tokenizeTitle(pos.Title) {
		want[t] = struct{}{}
	}
	others, err := s.repo.ListPositionsForSimilarity(ctx, p, positionID, 50)
	if err != nil {
		return nil, err
	}
	type posScore struct {
		p PositionDTO
		s float64
	}
	var buf []posScore
	for i := range others {
		o := &others[i]
		r2, _ := s.repo.ListPositionRequirements(ctx, o.ID)
		got := map[string]struct{}{}
		for _, r := range r2 {
			got[strings.ToLower(strings.TrimSpace(r.SkillName))] = struct{}{}
		}
		for _, t := range tokenizeTitle(o.Title) {
			got[t] = struct{}{}
		}
		sim := jaccard(want, got)
		buf = append(buf, posScore{p: toPositionDTO(o), s: sim})
	}
	sort.Slice(buf, func(i, j int) bool {
		if buf[i].s == buf[j].s {
			return buf[i].p.Title < buf[j].p.Title
		}
		return buf[i].s > buf[j].s
	})
	if limit <= 0 {
		limit = 10
	}
	if limit > len(buf) {
		limit = len(buf)
	}
	out := make([]PositionDTO, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, buf[i].p)
	}
	return out, nil
}
