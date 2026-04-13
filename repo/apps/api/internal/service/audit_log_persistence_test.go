package service

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

// LogMutation persists one append-only row (non-repudiation contract).
func TestAuditService_LogMutation_persistsRow(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewAuditRepository(db)
	svc := NewAuditService(repo)
	meta := AuditRequestMeta{OperatorUserID: "u1", RequestID: "req-1"}
	err = svc.LogMutation(context.Background(), AuditMutationInput{
		Module:     "cases",
		Operation:  "case.create",
		TargetType: "case",
		TargetID:   "c1",
		Before:     map[string]any{"status": "submitted"},
		After:      map[string]any{"status": "assigned"},
		Meta:       meta,
	})
	if err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := db.Model(&model.AuditLog{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 audit row, got %d", n)
	}
}
