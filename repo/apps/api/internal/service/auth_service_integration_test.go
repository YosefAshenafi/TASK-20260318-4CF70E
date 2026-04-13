package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pharmaops/api/internal/config"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

func setupAuthService(t *testing.T, ttl time.Duration) (*AuthService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Session{}); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{SessionTTL: ttl}
	return NewAuthService(cfg, repository.NewUserRepository(db), repository.NewSessionRepository(db)), db
}

func seedActiveUser(t *testing.T, db *gorm.DB, id, username, password string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := db.Create(&model.User{
		ID:           id,
		Username:     username,
		PasswordHash: string(hash),
		DisplayName:  "Test User",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error; err != nil {
		t.Fatal(err)
	}
}

func TestAuthService_LoginLogoutSessionLifecycle(t *testing.T) {
	svc, db := setupAuthService(t, 30*time.Minute)
	seedActiveUser(t, db, "u-1", "alice", "StrongPassword123")

	ctx := context.Background()
	login, err := svc.Login(ctx, "alice", "StrongPassword123", nil, nil)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if login.Token == "" {
		t.Fatal("expected opaque token")
	}

	uid, err := svc.SessionUserID(ctx, login.Token)
	if err != nil {
		t.Fatalf("session lookup failed: %v", err)
	}
	if uid != "u-1" {
		t.Fatalf("unexpected user id: %s", uid)
	}

	if err := svc.Logout(ctx, login.Token); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if _, err := svc.SessionUserID(ctx, login.Token); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected revoked session to be invalid, got %v", err)
	}
}

func TestAuthService_LoginValidationAndExpiry(t *testing.T) {
	svc, db := setupAuthService(t, 1*time.Minute)
	seedActiveUser(t, db, "u-2", "bob", "AnotherPass123")
	ctx := context.Background()

	if _, err := svc.Login(ctx, "bob", "short", nil, nil); !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
	if _, err := svc.Login(ctx, "bob", "wrongpassword", nil, nil); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	okLogin, err := svc.Login(ctx, "bob", "AnotherPass123", nil, nil)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	hash := tokenSHA256Hex(okLogin.Token)
	past := time.Now().UTC().Add(-2 * time.Minute)
	if err := db.Model(&model.Session{}).Where("token_hash = ?", hash).Update("expires_at", past).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SessionUserID(ctx, okLogin.Token); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected expired session to be invalid, got %v", err)
	}
}

