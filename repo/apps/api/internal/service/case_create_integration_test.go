package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

func TestCaseService_CreateCase_serialAndDuplicateWindow(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE institutions (id TEXT PRIMARY KEY, code TEXT NOT NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO institutions (id, code) VALUES (?, ?)`, "inst-1", "INST").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.CaseNumberSequence{}, &model.CaseRecord{}); err != nil {
		t.Fatal(err)
	}

	repo := repository.NewCaseRepository(db)
	svc := NewCaseService(repo, NewAuditService(nil))
	pr := &access.Principal{Scopes: []access.Scope{{InstitutionID: "inst-1"}}}
	ctx := context.Background()
	reported := time.Now().UTC()

	first, err := svc.CreateCase(ctx, pr, CreateCaseInput{
		InstitutionID: "inst-1",
		CaseType:      "quality",
		Title:         "Batch discrepancy",
		Description:   "Line check mismatch",
		ReportedAt:    reported,
	}, AuditRequestMeta{OperatorUserID: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(first.CaseNumber, "000001") {
		t.Fatalf("expected first serial 000001, got %s", first.CaseNumber)
	}

	_, err = svc.CreateCase(ctx, pr, CreateCaseInput{
		InstitutionID: "inst-1",
		CaseType:      "quality",
		Title:         "Batch discrepancy",
		Description:   "Line check mismatch",
		ReportedAt:    reported,
	}, AuditRequestMeta{OperatorUserID: "u1"})
	if !errors.Is(err, ErrDuplicateCaseSubmission) {
		t.Fatalf("expected duplicate guard error, got %v", err)
	}

	second, err := svc.CreateCase(ctx, pr, CreateCaseInput{
		InstitutionID: "inst-1",
		CaseType:      "quality",
		Title:         "Different title",
		Description:   "Line check mismatch",
		ReportedAt:    reported,
	}, AuditRequestMeta{OperatorUserID: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(second.CaseNumber, "000002") {
		t.Fatalf("expected second serial 000002, got %s", second.CaseNumber)
	}
}
