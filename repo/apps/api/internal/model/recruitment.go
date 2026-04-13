package model

import (
	"time"

	"gorm.io/gorm"
)

type Candidate struct {
	ID              string         `gorm:"column:id;type:char(36);primaryKey"`
	InstitutionID   string         `gorm:"column:institution_id;type:char(36);not null"`
	DepartmentID    *string      `gorm:"column:department_id;type:char(36)"`
	TeamID          *string        `gorm:"column:team_id;type:char(36)"`
	Name            string         `gorm:"column:name;not null"`
	PhoneEnc        []byte         `gorm:"column:phone_enc"`
	IDNumberEnc     []byte         `gorm:"column:id_number_enc"`
	EmailEnc        []byte         `gorm:"column:email_enc"`
	PIIKeyVersion   uint8          `gorm:"column:pii_key_version;not null;default:1"`
	ExperienceYears *int         `gorm:"column:experience_years"`
	EducationLevel  *string      `gorm:"column:education_level"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	Skills          []CandidateSkill `gorm:"foreignKey:CandidateID;references:ID"`
}

func (Candidate) TableName() string { return "candidates" }

type CandidateSkill struct {
	ID          string `gorm:"column:id;type:char(36);primaryKey"`
	CandidateID string `gorm:"column:candidate_id;type:char(36);not null;index"`
	SkillName   string `gorm:"column:skill_name;not null"`
}

func (CandidateSkill) TableName() string { return "candidate_skills" }

type CandidateTag struct {
	CandidateID string `gorm:"column:candidate_id;type:char(36);primaryKey"`
	Tag         string `gorm:"column:tag;primaryKey"`
}

func (CandidateTag) TableName() string { return "candidate_tags" }

type Position struct {
	ID             string    `gorm:"column:id;type:char(36);primaryKey"`
	InstitutionID  string    `gorm:"column:institution_id;type:char(36);not null"`
	DepartmentID   *string   `gorm:"column:department_id;type:char(36)"`
	TeamID         *string   `gorm:"column:team_id;type:char(36)"`
	Title          string    `gorm:"column:title;not null"`
	Description    *string   `gorm:"column:description"`
	Status         string    `gorm:"column:status;not null;default:open"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (Position) TableName() string { return "positions" }
