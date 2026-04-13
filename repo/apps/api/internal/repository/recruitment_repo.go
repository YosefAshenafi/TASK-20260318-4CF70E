package repository

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
)

type mergeSnapshot struct {
	Name            string   `json:"name"`
	ExperienceYears *int     `json:"experienceYears,omitempty"`
	EducationLevel  *string  `json:"educationLevel,omitempty"`
	SkillNames      []string `json:"skills"`
	Tags            []string `json:"tags"`
}

func skillNamesFrom(c *model.Candidate) []string {
	out := make([]string, 0, len(c.Skills))
	for _, s := range c.Skills {
		out = append(out, s.SkillName)
	}
	sort.Strings(out)
	return out
}

func sortedSkillNamesFromMap(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		if k != "" {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

type RecruitmentRepository struct {
	db *gorm.DB
}

func NewRecruitmentRepository(db *gorm.DB) *RecruitmentRepository {
	return &RecruitmentRepository{db: db}
}

// CandidateFilter holds optional search/filter parameters.
type CandidateFilter struct {
	Keyword        string
	Skills         []string
	EducationLevel string
	MinExperience  *int
	MaxExperience  *int
}

func applyCandidateFilters(q *gorm.DB, f CandidateFilter) *gorm.DB {
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("name LIKE ?", kw)
	}
	if f.EducationLevel != "" {
		q = q.Where("education_level = ?", f.EducationLevel)
	}
	if f.MinExperience != nil {
		q = q.Where("experience_years >= ?", *f.MinExperience)
	}
	if f.MaxExperience != nil {
		q = q.Where("experience_years <= ?", *f.MaxExperience)
	}
	if len(f.Skills) > 0 {
		sub := q.Session(&gorm.Session{}).Model(&model.CandidateSkill{}).
			Select("DISTINCT candidate_id").
			Where("skill_name IN ?", f.Skills)
		q = q.Where("id IN (?)", sub)
	}
	return q
}

func (r *RecruitmentRepository) ListCandidates(ctx context.Context, p *access.Principal, offset, limit int, orderClause string, f CandidateFilter) ([]model.Candidate, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.Candidate{})
	base = applyDataScope(base, p, "institution_id", "department_id", "team_id")
	base = applyCandidateFilters(base, f)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Candidate
	q := r.db.WithContext(ctx).Model(&model.Candidate{}).Preload("Skills")
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	q = applyCandidateFilters(q, f)
	err := q.Order(orderClause).Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *RecruitmentRepository) GetCandidate(ctx context.Context, id string, p *access.Principal) (*model.Candidate, error) {
	var c model.Candidate
	q := r.db.WithContext(ctx).Model(&model.Candidate{}).Preload("Skills").Where("id = ?", id)
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	err := q.First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *RecruitmentRepository) TagsForCandidates(ctx context.Context, candidateIDs []string) (map[string][]string, error) {
	if len(candidateIDs) == 0 {
		return map[string][]string{}, nil
	}
	var rows []model.CandidateTag
	if err := r.db.WithContext(ctx).
		Where("candidate_id IN ?", candidateIDs).
		Order("tag").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string][]string)
	for _, row := range rows {
		out[row.CandidateID] = append(out[row.CandidateID], row.Tag)
	}
	return out, nil
}

func (r *RecruitmentRepository) CreateCandidate(ctx context.Context, c *model.Candidate, skills []model.CandidateSkill, tags []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(c).Error; err != nil {
			return err
		}
		for i := range skills {
			skills[i].CandidateID = c.ID
			if err := tx.Create(&skills[i]).Error; err != nil {
				return err
			}
		}
		for _, t := range tags {
			if err := tx.Create(&model.CandidateTag{CandidateID: c.ID, Tag: t}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RecruitmentRepository) UpdateCandidate(ctx context.Context, c *model.Candidate, p *access.Principal) error {
	q := r.db.WithContext(ctx).Model(&model.Candidate{}).Where("id = ?", c.ID)
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	res := q.Updates(map[string]interface{}{
		"name":                c.Name,
		"department_id":       c.DepartmentID,
		"team_id":             c.TeamID,
		"experience_years":    c.ExperienceYears,
		"education_level":     c.EducationLevel,
		"phone_enc":           c.PhoneEnc,
		"phone_norm_hash":     c.PhoneNormHash,
		"id_number_enc":       c.IDNumberEnc,
		"id_number_norm_hash": c.IDNumberNormHash,
		"email_enc":           c.EmailEnc,
		"pii_key_version":     c.PIIKeyVersion,
		"custom_fields_json":  c.CustomFieldsJSON,
		"updated_at":          time.Now().UTC(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *RecruitmentRepository) SoftDeleteCandidate(ctx context.Context, id string, p *access.Principal) error {
	q := r.db.WithContext(ctx).Model(&model.Candidate{}).Where("id = ?", id)
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	res := q.Delete(&model.Candidate{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *RecruitmentRepository) ListPositions(ctx context.Context, p *access.Principal, offset, limit int, orderClause string) ([]model.Position, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.Position{})
	base = applyDataScope(base, p, "institution_id", "department_id", "team_id")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Position
	q := r.db.WithContext(ctx).Model(&model.Position{})
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	err := q.Order(orderClause).Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *RecruitmentRepository) GetPosition(ctx context.Context, id string, p *access.Principal) (*model.Position, error) {
	var pos model.Position
	q := r.db.WithContext(ctx).Where("id = ?", id)
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	err := q.First(&pos).Error
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

func (r *RecruitmentRepository) CreatePosition(ctx context.Context, p *model.Position) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *RecruitmentRepository) UpdatePosition(ctx context.Context, pos *model.Position, p *access.Principal) error {
	q := r.db.WithContext(ctx).Model(&model.Position{}).Where("id = ?", pos.ID)
	q = applyDataScope(q, p, "institution_id", "department_id", "team_id")
	res := q.Updates(map[string]interface{}{
		"title":         pos.Title,
		"description":   pos.Description,
		"status":        pos.Status,
		"department_id": pos.DepartmentID,
		"team_id":       pos.TeamID,
		"updated_at":    time.Now().UTC(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// --- Import batches ---

func (r *RecruitmentRepository) CreateImportBatch(ctx context.Context, b *model.CandidateImportBatch) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *RecruitmentRepository) GetImportBatch(ctx context.Context, id string, pr *access.Principal) (*model.CandidateImportBatch, error) {
	var b model.CandidateImportBatch
	q := r.db.WithContext(ctx).Where("id = ?", id)
	q = applyDataScope(q, pr, "institution_id", "department_id", "team_id")
	err := q.First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *RecruitmentRepository) UpdateImportBatchCommitted(ctx context.Context, id string, pr *access.Principal, validationJSON []byte, committedAt time.Time) error {
	q := r.db.WithContext(ctx).Model(&model.CandidateImportBatch{}).
		Where("id = ? AND status = ?", id, "pending")
	q = applyDataScope(q, pr, "institution_id", "department_id", "team_id")
	res := q.Updates(map[string]interface{}{
		"status":                 "committed",
		"validation_report_json": validationJSON,
		"committed_at":           committedAt,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// --- Merge history ---

func (r *RecruitmentRepository) CreateMergeHistory(ctx context.Context, h *model.CandidateMergeHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *RecruitmentRepository) ListMergeHistory(ctx context.Context, pr *access.Principal, offset, limit int) ([]model.CandidateMergeHistory, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.CandidateMergeHistory{}).
		Joins("INNER JOIN candidates ON candidates.id = candidate_merge_history.base_candidate_id")
	base = applyDataScope(base, pr, "candidates.institution_id", "candidates.department_id", "candidates.team_id")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.CandidateMergeHistory
	q := r.db.WithContext(ctx).Model(&model.CandidateMergeHistory{}).
		Joins("INNER JOIN candidates ON candidates.id = candidate_merge_history.base_candidate_id")
	q = applyDataScope(q, pr, "candidates.institution_id", "candidates.department_id", "candidates.team_id")
	err := q.Order("candidate_merge_history.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

// --- Match snapshots ---

func (r *RecruitmentRepository) CreateMatchSnapshot(ctx context.Context, s *model.MatchScoreSnapshot) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// --- Position requirements ---

func (r *RecruitmentRepository) ListPositionRequirements(ctx context.Context, positionID string) ([]model.PositionRequirement, error) {
	var rows []model.PositionRequirement
	err := r.db.WithContext(ctx).
		Where("position_id = ?", positionID).
		Order("skill_name").
		Find(&rows).Error
	return rows, err
}

// --- Duplicates (same normalized phone or ID hash within institution) ---

type DuplicateGroup struct {
	MatchKey      string
	MatchType     string
	InstitutionID string
	CandidateIDs  []string
}

func (r *RecruitmentRepository) ListDuplicateGroups(ctx context.Context, pr *access.Principal) ([]DuplicateGroup, error) {
	scopeSQL, scopeArgs, ok := buildDataScopeExpr(pr, "institution_id", "department_id", "team_id")
	if !ok {
		return nil, nil
	}

	type row struct {
		MatchKey      string
		MatchType     string
		InstitutionID string
		IDsCSV        string `gorm:"column:ids_csv"`
	}

	phoneQuery := `
SELECT phone_norm_hash AS match_key, 'phone' AS match_type, institution_id,
       GROUP_CONCAT(id ORDER BY created_at) AS ids_csv
FROM candidates
WHERE deleted_at IS NULL AND phone_norm_hash IS NOT NULL AND LENGTH(phone_norm_hash) = 64 AND ` + scopeSQL + `
GROUP BY phone_norm_hash, institution_id
HAVING COUNT(*) > 1
`
	var phoneRaw []row
	if err := r.db.WithContext(ctx).Raw(phoneQuery, scopeArgs...).Scan(&phoneRaw).Error; err != nil {
		return nil, err
	}

	idQuery := `
SELECT id_number_norm_hash AS match_key, 'id_number' AS match_type, institution_id,
       GROUP_CONCAT(id ORDER BY created_at) AS ids_csv
FROM candidates
WHERE deleted_at IS NULL AND id_number_norm_hash IS NOT NULL AND LENGTH(id_number_norm_hash) = 64 AND ` + scopeSQL + `
GROUP BY id_number_norm_hash, institution_id
HAVING COUNT(*) > 1
`
	var idRaw []row
	if err := r.db.WithContext(ctx).Raw(idQuery, scopeArgs...).Scan(&idRaw).Error; err != nil {
		return nil, err
	}

	all := append(phoneRaw, idRaw...)
	seen := make(map[string]struct{})
	out := make([]DuplicateGroup, 0, len(all))
	for _, r0 := range all {
		dedup := r0.MatchType + ":" + r0.MatchKey + ":" + r0.InstitutionID
		if _, ok := seen[dedup]; ok {
			continue
		}
		seen[dedup] = struct{}{}
		parts := strings.Split(r0.IDsCSV, ",")
		ids := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				ids = append(ids, p)
			}
		}
		if len(ids) < 2 {
			continue
		}
		out = append(out, DuplicateGroup{
			MatchKey:      r0.MatchKey,
			MatchType:     r0.MatchType,
			InstitutionID: r0.InstitutionID,
			CandidateIDs:  ids,
		})
	}
	return out, nil
}

// ListCandidatesForSimilarity returns non-deleted candidates in scope (for recommendations).
func (r *RecruitmentRepository) ListCandidatesForSimilarity(ctx context.Context, pr *access.Principal, excludeID string, limit int) ([]model.Candidate, error) {
	q := r.db.WithContext(ctx).Model(&model.Candidate{}).
		Preload("Skills").
		Where("deleted_at IS NULL")
	q = applyDataScope(q, pr, "institution_id", "department_id", "team_id")
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	var rows []model.Candidate
	err := q.Order("updated_at DESC").Limit(limit * 3).Find(&rows).Error
	return rows, err
}

// ListPositionsForSimilarity returns positions in scope (for recommendations).
func (r *RecruitmentRepository) ListPositionsForSimilarity(ctx context.Context, pr *access.Principal, excludeID string, limit int) ([]model.Position, error) {
	q := r.db.WithContext(ctx).Model(&model.Position{})
	q = applyDataScope(q, pr, "institution_id", "department_id", "team_id")
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	var rows []model.Position
	err := q.Order("updated_at DESC").Limit(limit * 3).Find(&rows).Error
	return rows, err
}

// MergeIntoBase merges source candidates into base (soft-deletes sources) and writes merge history with snapshots.
func (r *RecruitmentRepository) MergeIntoBase(ctx context.Context, baseID string, sourceIDs []string, pr *access.Principal, operatorUserID string, strategy string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var base model.Candidate
		q := tx.Preload("Skills").Where("id = ?", baseID)
		q = applyDataScope(q, pr, "institution_id", "department_id", "team_id")
		if err := q.First(&base).Error; err != nil {
			return err
		}
		tagMapBefore, err := r.TagsForCandidates(ctx, append(sourceIDs, baseID))
		if err != nil {
			return err
		}
		beforeSnap, err := json.Marshal(mergeSnapshot{
			Name:            base.Name,
			ExperienceYears: base.ExperienceYears,
			EducationLevel:  base.EducationLevel,
			SkillNames:      skillNamesFrom(&base),
			Tags:            tagMapBefore[baseID],
		})
		if err != nil {
			return err
		}
		var sources []model.Candidate
		sq := tx.Preload("Skills").Where("id IN ?", sourceIDs)
		sq = applyDataScope(sq, pr, "institution_id", "department_id", "team_id")
		if err := sq.Find(&sources).Error; err != nil {
			return err
		}
		if len(sources) != len(sourceIDs) {
			return gorm.ErrRecordNotFound
		}
		for _, s := range sources {
			if s.ID == baseID {
				return errors.New("source equals base")
			}
			if s.InstitutionID != base.InstitutionID {
				return errors.New("institution mismatch")
			}
		}

		mergedSkills := make(map[string]struct{})
		for _, sk := range base.Skills {
			mergedSkills[sk.SkillName] = struct{}{}
		}
		for _, src := range sources {
			for _, sk := range src.Skills {
				mergedSkills[sk.SkillName] = struct{}{}
			}
		}
		mergedTags := map[string]struct{}{}
		for _, t := range tagMapBefore[baseID] {
			mergedTags[t] = struct{}{}
		}
		for _, sid := range sourceIDs {
			for _, t := range tagMapBefore[sid] {
				mergedTags[t] = struct{}{}
			}
		}
		var tags []string
		for t := range mergedTags {
			tags = append(tags, t)
		}
		sort.Strings(tags)

		exp := base.ExperienceYears
		edu := base.EducationLevel
		for _, src := range sources {
			if exp == nil && src.ExperienceYears != nil {
				exp = src.ExperienceYears
			}
			if (edu == nil || *edu == "") && src.EducationLevel != nil && *src.EducationLevel != "" {
				edu = src.EducationLevel
			}
		}

		base.ExperienceYears = exp
		base.EducationLevel = edu
		now := time.Now().UTC()
		base.UpdatedAt = now

		uq := tx.Model(&model.Candidate{}).Where("id = ?", base.ID)
		uq = applyDataScope(uq, pr, "institution_id", "department_id", "team_id")
		if err := uq.Updates(map[string]interface{}{
			"experience_years": base.ExperienceYears,
			"education_level":  base.EducationLevel,
			"updated_at":       now,
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("candidate_id = ?", base.ID).Delete(&model.CandidateSkill{}).Error; err != nil {
			return err
		}
		for name := range mergedSkills {
			if name == "" {
				continue
			}
			row := model.CandidateSkill{
				ID:          uuid.NewString(),
				CandidateID: base.ID,
				SkillName:   name,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("candidate_id = ?", base.ID).Delete(&model.CandidateTag{}).Error; err != nil {
			return err
		}
		for _, t := range tags {
			if t == "" {
				continue
			}
			if err := tx.Create(&model.CandidateTag{CandidateID: base.ID, Tag: t}).Error; err != nil {
				return err
			}
		}

		for _, sid := range sourceIDs {
			dq := tx.Model(&model.Candidate{}).Where("id = ?", sid)
			dq = applyDataScope(dq, pr, "institution_id", "department_id", "team_id")
			res := dq.Delete(&model.Candidate{})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}

		afterSnap, err := json.Marshal(mergeSnapshot{
			Name:            base.Name,
			ExperienceYears: base.ExperienceYears,
			EducationLevel:  base.EducationLevel,
			SkillNames:      sortedSkillNamesFromMap(mergedSkills),
			Tags:            tags,
		})
		if err != nil {
			return err
		}
		srcJSON, err := json.Marshal(sourceIDs)
		if err != nil {
			return err
		}
		mergedFields, err := json.Marshal(map[string]any{"strategy": strategy})
		if err != nil {
			return err
		}
		h := &model.CandidateMergeHistory{
			ID:                     uuid.NewString(),
			BaseCandidateID:        baseID,
			SourceCandidateIDsJSON: srcJSON,
			MergedFieldsJSON:       mergedFields,
			BeforeSnapshotJSON:     beforeSnap,
			AfterSnapshotJSON:      afterSnap,
			OperatorUserID:         operatorUserID,
			CreatedAt:              now,
		}
		return tx.Create(h).Error
	})
}
