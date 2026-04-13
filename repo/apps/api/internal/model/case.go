package model

import "time"

// CaseRecord maps to `cases` (Go keyword `case` avoided in type name).
type CaseRecord struct {
	ID                 string    `gorm:"column:id;type:char(36);primaryKey"`
	CaseNumber         string    `gorm:"column:case_number;not null;uniqueIndex"`
	InstitutionID      string    `gorm:"column:institution_id;type:char(36);not null;index"`
	DepartmentID       *string   `gorm:"column:department_id;type:char(36)"`
	TeamID             *string   `gorm:"column:team_id;type:char(36)"`
	CaseType           string    `gorm:"column:case_type;not null"`
	Title              string    `gorm:"column:title;not null"`
	Description        string    `gorm:"column:description;type:text;not null"`
	ReportedAt         time.Time `gorm:"column:reported_at"`
	Status             string    `gorm:"column:status;not null;default:submitted;index"`
	AssigneeUserID     *string   `gorm:"column:assignee_user_id;type:char(36)"`
	DuplicateGuardHash *string  `gorm:"column:duplicate_guard_hash;type:char(64);index"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
}

func (CaseRecord) TableName() string { return "cases" }

type CaseNumberSequence struct {
	InstitutionID string    `gorm:"column:institution_id;type:char(36);primaryKey"`
	SequenceDate  time.Time `gorm:"column:sequence_date;type:date;primaryKey"`
	LastSerial    uint32    `gorm:"column:last_serial;not null"`
}

func (CaseNumberSequence) TableName() string { return "case_number_sequences" }

type CaseAssignment struct {
	ID         string    `gorm:"column:id;type:char(36);primaryKey"`
	CaseID     string    `gorm:"column:case_id;type:char(36);not null;index"`
	UserID     string    `gorm:"column:user_id;type:char(36);not null"`
	AssignedAt time.Time `gorm:"column:assigned_at"`
}

func (CaseAssignment) TableName() string { return "case_assignments" }

type CaseProcessingRecord struct {
	ID          string    `gorm:"column:id;type:char(36);primaryKey"`
	CaseID      string    `gorm:"column:case_id;type:char(36);not null;index"`
	StepCode    string    `gorm:"column:step_code;not null"`
	ActorUserID string    `gorm:"column:actor_user_id;type:char(36);not null"`
	Note        *string   `gorm:"column:note;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (CaseProcessingRecord) TableName() string { return "case_processing_records" }

type CaseStatusTransition struct {
	ID          string    `gorm:"column:id;type:char(36);primaryKey"`
	CaseID      string    `gorm:"column:case_id;type:char(36);not null;index"`
	FromStatus  string    `gorm:"column:from_status;not null"`
	ToStatus    string    `gorm:"column:to_status;not null"`
	ActorUserID string    `gorm:"column:actor_user_id;type:char(36);not null"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (CaseStatusTransition) TableName() string { return "case_status_transitions" }
