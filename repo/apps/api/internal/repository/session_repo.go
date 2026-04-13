package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/model"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *model.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SessionRepository) FindValidByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error) {
	var s model.Session
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, now).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, sessionID string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Session{}).
		Where("id = ?", sessionID).
		Update("revoked_at", at).Error
}
