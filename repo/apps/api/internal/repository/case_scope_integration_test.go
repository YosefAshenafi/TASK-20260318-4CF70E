package repository

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
)

func TestCaseRepository_ListCases_enforcesInstitutionScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.CaseRecord{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	rows := []model.CaseRecord{
		{ID: "c1", CaseNumber: "N1", InstitutionID: "inst-a", CaseType: "quality", Title: "A1", Description: "A1", Status: "submitted", ReportedAt: now, CreatedAt: now, UpdatedAt: now},
		{ID: "c2", CaseNumber: "N2", InstitutionID: "inst-b", CaseType: "quality", Title: "B1", Description: "B1", Status: "submitted", ReportedAt: now, CreatedAt: now, UpdatedAt: now},
	}
	for _, row := range rows {
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := NewCaseRepository(db)
	pr := &access.Principal{Scopes: []access.Scope{{InstitutionID: "inst-a"}}}
	out, _, err := repo.ListCases(context.Background(), pr, 0, 20, "created_at DESC", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].InstitutionID != "inst-a" {
		t.Fatalf("expected only inst-a records, got %+v", out)
	}
}
