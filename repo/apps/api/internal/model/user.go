package model

import "time"

type User struct {
	ID           string    `gorm:"column:id;type:char(36);primaryKey"`
	Username     string    `gorm:"column:username;size:64;uniqueIndex;not null"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null"`
	DisplayName  string    `gorm:"column:display_name;size:255;not null"`
	IsActive     bool      `gorm:"column:is_active;not null"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string { return "users" }
