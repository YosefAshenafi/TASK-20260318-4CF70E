package model

import "time"

type Role struct {
	ID          string    `gorm:"column:id;type:char(36);primaryKey"`
	Slug        string    `gorm:"column:slug;not null;uniqueIndex"`
	Name        string    `gorm:"column:name;not null"`
	Description *string   `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID          string    `gorm:"column:id;type:char(36);primaryKey"`
	Code        string    `gorm:"column:code;not null;uniqueIndex"`
	Description *string   `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (Permission) TableName() string { return "permissions" }

type DataScope struct {
	ID             string    `gorm:"column:id;type:char(36);primaryKey"`
	ScopeKey       string    `gorm:"column:scope_key;not null;uniqueIndex"`
	InstitutionID  string    `gorm:"column:institution_id;type:char(36);not null"`
	DepartmentID   *string   `gorm:"column:department_id;type:char(36)"`
	TeamID         *string   `gorm:"column:team_id;type:char(36)"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (DataScope) TableName() string { return "data_scopes" }
