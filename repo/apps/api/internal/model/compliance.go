package model

import (
	"time"
)

type QualificationProfile struct {
	ID             string     `gorm:"column:id;type:char(36);primaryKey"`
	InstitutionID  string     `gorm:"column:institution_id;type:char(36);not null"`
	DepartmentID   *string    `gorm:"column:department_id;type:char(36)"`
	TeamID         *string    `gorm:"column:team_id;type:char(36)"`
	ClientID       string     `gorm:"column:client_id;not null"`
	DisplayName    string     `gorm:"column:display_name;not null"`
	Status         string     `gorm:"column:status;not null;default:active"`
	ExpiresOn      *time.Time `gorm:"column:expires_on;type:date"`
	DeactivatedAt  *time.Time `gorm:"column:deactivated_at"`
	MetadataJSON   []byte     `gorm:"column:metadata_json;type:json"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (QualificationProfile) TableName() string { return "qualification_profiles" }

type PurchaseRestriction struct {
	ID             string    `gorm:"column:id;type:char(36);primaryKey"`
	InstitutionID  string    `gorm:"column:institution_id;type:char(36);not null"`
	DepartmentID   *string   `gorm:"column:department_id;type:char(36)"`
	TeamID         *string   `gorm:"column:team_id;type:char(36)"`
	ClientID       string    `gorm:"column:client_id;not null"`
	MedicationID   *string   `gorm:"column:medication_id"`
	RuleJSON       []byte    `gorm:"column:rule_json;type:json;not null"`
	IsActive       bool      `gorm:"column:is_active;not null"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (PurchaseRestriction) TableName() string { return "purchase_restrictions" }

type RestrictionViolationRecord struct {
	ID             string    `gorm:"column:id;type:char(36);primaryKey"`
	RestrictionID  *string   `gorm:"column:restriction_id;type:char(36)"`
	InstitutionID  string    `gorm:"column:institution_id;type:char(36);not null"`
	DepartmentID   *string   `gorm:"column:department_id;type:char(36)"`
	TeamID         *string   `gorm:"column:team_id;type:char(36)"`
	ClientID       string    `gorm:"column:client_id;not null"`
	MedicationID   *string   `gorm:"column:medication_id"`
	CaseID         *string   `gorm:"column:case_id;type:char(36)"`
	DetailsJSON    []byte    `gorm:"column:details_json;type:json"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (RestrictionViolationRecord) TableName() string { return "restriction_violation_records" }

type CompliancePurchaseRecord struct {
	ID             string    `gorm:"column:id;type:char(36);primaryKey"`
	InstitutionID  string    `gorm:"column:institution_id;type:char(36);not null"`
	DepartmentID   *string   `gorm:"column:department_id;type:char(36)"`
	TeamID         *string   `gorm:"column:team_id;type:char(36)"`
	ClientID       string    `gorm:"column:client_id;not null"`
	MedicationID   *string   `gorm:"column:medication_id"`
	RecordedAt     time.Time `gorm:"column:recorded_at"`
}

func (CompliancePurchaseRecord) TableName() string { return "compliance_purchase_records" }
