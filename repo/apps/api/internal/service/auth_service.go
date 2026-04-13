package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pharmaops/api/internal/config"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDisabled    = errors.New("account disabled")
	ErrPasswordTooShort   = errors.New("password too short")
)

type AuthService struct {
	cfg        config.Config
	users      *repository.UserRepository
	sessions   *repository.SessionRepository
}

func NewAuthService(cfg config.Config, users *repository.UserRepository, sessions *repository.SessionRepository) *AuthService {
	return &AuthService{cfg: cfg, users: users, sessions: sessions}
}

func tokenSHA256Hex(opaque string) string {
	sum := sha256.Sum256([]byte(opaque))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) Login(ctx context.Context, username, password string, clientIP, userAgent *string) (token string, expiresAt time.Time, err error) {
	if len(password) < 8 {
		return "", time.Time{}, ErrPasswordTooShort
	}
	u, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", time.Time{}, ErrInvalidCredentials
		}
		return "", time.Time{}, err
	}
	if !u.IsActive {
		return "", time.Time{}, ErrAccountDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", time.Time{}, ErrInvalidCredentials
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}
	opaque := hex.EncodeToString(raw)
	expiresAt = time.Now().UTC().Add(s.cfg.SessionTTL)
	sess := &model.Session{
		ID:        uuid.NewString(),
		UserID:    u.ID,
		TokenHash: tokenSHA256Hex(opaque),
		ExpiresAt: expiresAt,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.sessions.Create(ctx, sess); err != nil {
		return "", time.Time{}, err
	}
	return opaque, expiresAt, nil
}

func (s *AuthService) Logout(ctx context.Context, opaqueToken string) error {
	hash := tokenSHA256Hex(opaqueToken)
	sess, err := s.sessions.FindValidByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	now := time.Now().UTC()
	return s.sessions.Revoke(ctx, sess.ID, now)
}

func (s *AuthService) SessionUserID(ctx context.Context, opaqueToken string) (userID string, err error) {
	hash := tokenSHA256Hex(opaqueToken)
	sess, err := s.sessions.FindValidByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", gorm.ErrRecordNotFound
		}
		return "", err
	}
	return sess.UserID, nil
}
