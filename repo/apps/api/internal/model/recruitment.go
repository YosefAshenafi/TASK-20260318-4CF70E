package model

import (
	"time"

	"gorm.io/gorm"
)

type Candidate struct {
	ID               string         `gorm:"column:id;type:char(36);primaryKey"`
	InstitutionID    string         `gorm:"column:institution_id;type:char(36);not null"`
	DepartmentID     *string        `gorm:"column:department_id;type:char(36)"`
	TeamID           *string        `gorm:"column:team_id;type:char(36)"`
	Name             string         `gorm:"column:name;not null"`
	PhoneEnc         []byte         `gorm:"column:phone_enc"`
	IDNumberEnc      []byte         `gorm:"column:id_number_enc"`
	EmailEnc         []byte         `gorm:"column:email_enc"`
	PIIKeyVersion    uint8          `gorm:"column:pii_key_version;not null;default:1"`
	ExperienceYears  *int           `gorm:"column:experience_years"`
	EducationLevel   *string        `gorm:"column:education_level"`
	CustomFieldsJSON []byte         `gorm:"column:custom_fields_json;type:json"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
	Skills           []CandidateSkill `gorm:"foreignKey:CandidateID;references:ID"`
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

// PositionRequirement maps position_requirements (used for matching / recommendations).
type PositionRequirement struct {
	ID         string `gorm:"column:id;type:char(36);primaryKey"`
	PositionID string `gorm:"column:position_id;type:char(36);not null;index"`
	SkillName  string `gorm:"column:skill_name;not null"`
	WeightPct  uint8  `gorm:"column:weight_pct;not null;default:0"`
	IsRequired bool   `gorm:"column:is_required;not null;default:1"`
}

func (PositionRequirement) TableName() string { return "position_requirements" }

// CandidateImportBatch maps candidate_import_batches.
type CandidateImportBatch struct {
	ID                   string     `gorm:"column:id;type:char(36);primaryKey"`
	InstitutionID        string     `gorm:"column:institution_id;type:char(36);not null;index"`
	DepartmentID         *string    `gorm:"column:department_id;type:char(36)"`
	TeamID               *string    `gorm:"column:team_id;type:char(36)"`
	Status               string     `gorm:"column:status;not null;default:pending"`
	MappingJSON          []byte     `gorm:"column:mapping_json"`
	ValidationReportJSON []byte     `gorm:"column:validation_report_json"`
	CreatedByUserID      string     `gorm:"column:created_by_user_id;type:char(36);not null"`
	CommittedAt          *time.Time `gorm:"column:committed_at"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
}

func (CandidateImportBatch) TableName() string { return "candidate_import_batches" }

// CandidateMergeHistory maps candidate_merge_history.
type CandidateMergeHistory struct {
	ID                     string    `gorm:"column:id;type:char(36);primaryKey"`
	BaseCandidateID        string    `gorm:"column:base_candidate_id;type:char(36);not null;index"`
	SourceCandidateIDsJSON []byte    `gorm:"column:source_candidate_ids_json;type:json;not null"`
	MergedFieldsJSON       []byte    `gorm:"column:merged_fields_json"`
	BeforeSnapshotJSON     []byte    `gorm:"column:before_snapshot_json"`
	AfterSnapshotJSON      []byte    `gorm:"column:after_snapshot_json"`
	OperatorUserID         string    `gorm:"column:operator_user_id;type:char(36);not null"`
	CreatedAt              time.Time `gorm:"column:created_at"`
}

func (CandidateMergeHistory) TableName() string { return "candidate_merge_history" }

// MatchScoreSnapshot maps match_score_snapshots.
type MatchScoreSnapshot struct {
	ID            string    `gorm:"column:id;type:char(36);primaryKey"`
	CandidateID   string    `gorm:"column:candidate_id;type:char(36);not null;index"`
	PositionID    string    `gorm:"column:position_id;type:char(36);not null;index"`
	Score         uint16    `gorm:"column:score;not null"`
	BreakdownJSON []byte    `gorm:"column:breakdown_json"`
	ReasonsJSON   []byte    `gorm:"column:reasons_json"`
	ComputedAt    time.Time `gorm:"column:computed_at"`
}

func (MatchScoreSnapshot) TableName() string { return "match_score_snapshots" }
