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

// LoginSuccess is returned after a successful password check and session creation.
type LoginSuccess struct {
	Token     string
	ExpiresAt time.Time
	UserID    string
	Username  string
}

func (s *AuthService) Login(ctx context.Context, username, password string, clientIP, userAgent *string) (LoginSuccess, error) {
	var out LoginSuccess
	if len(password) < 8 {
		return out, ErrPasswordTooShort
	}
	u, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, ErrInvalidCredentials
		}
		return out, err
	}
	if !u.IsActive {
		return out, ErrAccountDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return out, ErrInvalidCredentials
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return out, err
	}
	opaque := hex.EncodeToString(raw)
	expiresAt := time.Now().UTC().Add(s.cfg.SessionTTL)
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
		return out, err
	}
	out.Token = opaque
	out.ExpiresAt = expiresAt
	out.UserID = u.ID
	out.Username = u.Username
	return out, nil
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
