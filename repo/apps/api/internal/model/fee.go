package model

import "time"

type FeeRecord struct {
	ID              string    `gorm:"column:id;type:char(36);primaryKey"`
	InstitutionID   string    `gorm:"column:institution_id;type:char(36);not null"`
	DepartmentID    *string   `gorm:"column:department_id;type:char(36)"`
	TeamID          *string   `gorm:"column:team_id;type:char(36)"`
	CaseID          *string   `gorm:"column:case_id;type:char(36)"`
	CandidateID     *string   `gorm:"column:candidate_id;type:char(36)"`
	FeeType         string    `gorm:"column:fee_type;size:64;not null"`
	Amount          float64   `gorm:"column:amount;type:decimal(12,2);not null"`
	Currency        string    `gorm:"column:currency;size:8;not null"`
	Note            *string   `gorm:"column:note"`
	CreatedByUserID string    `gorm:"column:created_by_user_id;type:char(36);not null"`
	UpdatedByUserID *string   `gorm:"column:updated_by_user_id;type:char(36)"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (FeeRecord) TableName() string { return "fees" }
