package model

import "time"

type AuditLog struct {
	ID             string    `gorm:"column:id;type:char(36);primaryKey"`
	Module         string    `gorm:"column:module;not null;index"`
	Operation      string    `gorm:"column:operation;not null"`
	OperatorUserID string    `gorm:"column:operator_user_id;type:char(36);not null;index"`
	InstitutionID  *string   `gorm:"column:institution_id;type:char(36);index"`
	DepartmentID   *string   `gorm:"column:department_id;type:char(36)"`
	TeamID         *string   `gorm:"column:team_id;type:char(36)"`
	RequestSource  *string   `gorm:"column:request_source"`
	RequestID      *string   `gorm:"column:request_id"`
	TargetType     string    `gorm:"column:target_type;not null"`
	TargetID       string    `gorm:"column:target_id;type:char(36);not null"`
	BeforeJSON     []byte    `gorm:"column:before_json;type:json"`
	AfterJSON      []byte    `gorm:"column:after_json;type:json"`
	CreatedAt      time.Time `gorm:"column:created_at;index"`
}

func (AuditLog) TableName() string { return "audit_logs" }

type AuditExport struct {
	ID                 string    `gorm:"column:id;type:char(36);primaryKey"`
	RequestedByUserID  string    `gorm:"column:requested_by_user_id;type:char(36);not null;index"`
	FilterJSON         []byte    `gorm:"column:filter_json;type:json"`
	Status             string    `gorm:"column:status;not null;default:pending"`
	OutputFilePath     *string   `gorm:"column:output_file_path"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	CompletedAt        *time.Time `gorm:"column:completed_at"`
}

func (AuditExport) TableName() string { return "audit_exports" }
