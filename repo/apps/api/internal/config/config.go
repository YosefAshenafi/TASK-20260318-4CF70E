package config

import (
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	HTTPAddr         string
	DSN              string
	SessionTTL       time.Duration
	Environment      string
	FileStorageRoot  string
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
		HTTPAddr:        addr,
		DSN:             dsn,
		SessionTTL:      8 * time.Hour,
		Environment:     env,
		FileStorageRoot: root,
	}
}
