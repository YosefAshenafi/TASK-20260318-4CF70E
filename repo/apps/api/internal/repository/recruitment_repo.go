package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/model"
)

type RecruitmentRepository struct {
	db *gorm.DB
}

func NewRecruitmentRepository(db *gorm.DB) *RecruitmentRepository {
	return &RecruitmentRepository{db: db}
}

func (r *RecruitmentRepository) ListCandidates(ctx context.Context, institutionIDs []string, offset, limit int, orderClause string) ([]model.Candidate, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Candidate{}).
		Where("institution_id IN ?", institutionIDs)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Candidate
	err := r.db.WithContext(ctx).
		Preload("Skills").
		Where("institution_id IN ?", institutionIDs).
		Order(orderClause).
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *RecruitmentRepository) GetCandidate(ctx context.Context, id string, institutionIDs []string) (*model.Candidate, error) {
	var c model.Candidate
	err := r.db.WithContext(ctx).
		Preload("Skills").
		Where("id = ? AND institution_id IN ?", id, institutionIDs).
		First(&c).Error
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

func (r *RecruitmentRepository) UpdateCandidate(ctx context.Context, c *model.Candidate, institutionIDs []string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", c.ID, institutionIDs).
		Updates(map[string]interface{}{
			"name":             c.Name,
			"department_id":    c.DepartmentID,
			"team_id":          c.TeamID,
			"experience_years": c.ExperienceYears,
			"education_level":  c.EducationLevel,
			"updated_at":       time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *RecruitmentRepository) SoftDeleteCandidate(ctx context.Context, id string, institutionIDs []string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", id, institutionIDs).
		Delete(&model.Candidate{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *RecruitmentRepository) ListPositions(ctx context.Context, institutionIDs []string, offset, limit int, orderClause string) ([]model.Position, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Position{}).
		Where("institution_id IN ?", institutionIDs)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Position
	err := r.db.WithContext(ctx).
		Where("institution_id IN ?", institutionIDs).
		Order(orderClause).
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *RecruitmentRepository) GetPosition(ctx context.Context, id string, institutionIDs []string) (*model.Position, error) {
	var p model.Position
	err := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", id, institutionIDs).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *RecruitmentRepository) CreatePosition(ctx context.Context, p *model.Position) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *RecruitmentRepository) UpdatePosition(ctx context.Context, p *model.Position, institutionIDs []string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND institution_id IN ?", p.ID, institutionIDs).
		Updates(map[string]interface{}{
			"title":         p.Title,
			"description":   p.Description,
			"status":        p.Status,
			"department_id": p.DepartmentID,
			"team_id":       p.TeamID,
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
