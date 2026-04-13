package model

import (
	"time"
)

type Session struct {
	ID        string     `gorm:"column:id;type:char(36);primaryKey"`
	UserID    string     `gorm:"column:user_id;type:char(36);index;not null"`
	TokenHash string     `gorm:"column:token_hash;size:64;uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	ClientIP  *string    `gorm:"column:client_ip;size:45"`
	UserAgent *string    `gorm:"column:user_agent;size:512"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (Session) TableName() string { return "sessions" }
