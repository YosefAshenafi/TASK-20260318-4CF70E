package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr         string
	DSN              string
	SessionTTL       time.Duration
	Environment      string
	FileStorageRoot  string
	// PIIAESKeyHex is 64 hex chars (32 bytes) for AES-256 candidate PII at rest (PII_AES_KEY_HEX).
	PIIAESKeyHex string
	// HealthCheckToken when non-empty requires X-Internal-Health-Token on GET /api/v1/health (HEALTH_CHECK_TOKEN).
	HealthCheckToken string
}

func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "pharmaops:pharmaops@tcp(127.0.0.1:3306)/pharmaops?parseTime=true&loc=UTC&charset=utf8mb4"
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	root := os.Getenv("FILE_STORAGE_ROOT")
	if root == "" {
		root = filepath.Join(os.TempDir(), "pharmaops-uploads")
	}
	return Config{
		HTTPAddr:         addr,
		DSN:              dsn,
		SessionTTL:       8 * time.Hour,
		Environment:      env,
		FileStorageRoot:  root,
		PIIAESKeyHex:     strings.TrimSpace(os.Getenv("PII_AES_KEY_HEX")),
		HealthCheckToken: strings.TrimSpace(os.Getenv("HEALTH_CHECK_TOKEN")),
	}
}
