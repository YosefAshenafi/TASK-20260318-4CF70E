package service

import (
	"testing"

	"pharmaops/api/internal/config"
)

func TestConstructorPointers_nonNil(t *testing.T) {
	if NewAccessService(nil) == nil {
		t.Fatal("NewAccessService")
	}
	cfg := config.Config{}
	if NewAuthService(cfg, nil, nil) == nil {
		t.Fatal("NewAuthService")
	}
}
