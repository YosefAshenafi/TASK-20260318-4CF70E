package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

func TestAuditService_LogMutation_recordsChangedFieldsAndSanitizesPII(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	svc := NewAuditService(repository.NewAuditRepository(db))
	err = svc.LogMutation(context.Background(), AuditMutationInput{
		Module:     "recruitment",
		Operation:  "candidate.update",
		TargetType: "candidate",
		TargetID:   "cand-1",
		Before: map[string]any{
			"institutionId": "inst-1",
			"name":          "Alice",
			"phone":         "18888888888",
		},
		After: map[string]any{
			"institutionId": "inst-1",
			"name":          "Alice Q",
			"phone":         "19999999999",
		},
		Meta: AuditRequestMeta{OperatorUserID: "u1", InstitutionID: strPtr("inst-1")},
	})
	if err != nil {
		t.Fatal(err)
	}
	var row model.AuditLog
	if err := db.First(&row).Error; err != nil {
		t.Fatal(err)
	}
	var after map[string]any
	if err := json.Unmarshal(row.AfterJSON, &after); err != nil {
		t.Fatal(err)
	}
	if v, ok := after["phone"]; ok && v != "[REDACTED]" {
		t.Fatalf("expected phone to be redacted when present, got %#v", v)
	}
	changed, ok := after["_changedFields"].([]any)
	if !ok || len(changed) == 0 {
		t.Fatalf("expected _changedFields in audit payload, got %#v", after["_changedFields"])
	}
}

func TestAuditService_LogMutation_dropsBusinessEventWithoutScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	svc := NewAuditService(repository.NewAuditRepository(db))
	err = svc.LogMutation(context.Background(), AuditMutationInput{
		Module:     "cases",
		Operation:  "case.update",
		TargetType: "case",
		TargetID:   "case-1",
		After:      map[string]any{"status": "assigned"},
		Meta: AuditRequestMeta{
			OperatorUserID: "u1",
			RequestID:      time.Now().UTC().Format(time.RFC3339Nano),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := db.Model(&model.AuditLog{}).Where("target_id = ?", "case-1").Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected no persisted row without scope metadata, got %d", n)
	}
}
