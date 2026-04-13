package config

import (
	"testing"
)

func TestLoad_DefaultHTTPAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("MYSQL_DSN", "")
	t.Setenv("APP_ENV", "")
	c := Load()
	if c.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr: got %q", c.HTTPAddr)
	}
	if c.DSN == "" {
		t.Fatal("expected default DSN")
	}
	if c.Environment != "development" {
		t.Fatalf("Environment: %q", c.Environment)
	}
	if c.FileStorageRoot == "" {
		t.Fatal("expected default FileStorageRoot")
	}
}
