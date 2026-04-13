package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

func TestCheckPurchase_prescriptionEnforcedFromServerRule(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.PurchaseRestriction{}, &model.RestrictionViolationRecord{}, &model.CompliancePurchaseRecord{}, &model.FileObject{}); err != nil {
		t.Fatal(err)
	}

	inst := "inst-1"
	now := time.Now().UTC()
	med := "med-controlled"
	rule := []byte(`{"requiresPrescription":true,"frequencyDays":0}`)
	if err := db.Create(&model.PurchaseRestriction{
		ID:            "r1",
		InstitutionID: inst,
		ClientID:      "client-1",
		MedicationID:  &med,
		RuleJSON:      rule,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatal(err)
	}

	repo := repository.NewComplianceRepository(db)
	fileRepo := repository.NewFileRepository(db)
	svc := NewComplianceService(repo, NewAuditService(nil), WithFileRepository(fileRepo))
	pr := &access.Principal{Scopes: []access.Scope{{InstitutionID: inst}}}

	out, err := svc.CheckPurchase(context.Background(), pr, CheckPurchaseInput{
		InstitutionID: inst,
		ClientID:      "client-1",
		MedicationID:  med,
		PurchaseAt:    now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Allowed {
		t.Fatal("purchase should be denied without a prescription attachment")
	}

	rxID := "file-rx-1"
	out, err = svc.CheckPurchase(context.Background(), pr, CheckPurchaseInput{
		InstitutionID:            inst,
		ClientID:                 "client-1",
		MedicationID:             med,
		PrescriptionAttachmentID: &rxID,
		PurchaseAt:               now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Allowed {
		t.Fatal("purchase should be denied when attachment id does not exist")
	}

	if err := db.Create(&model.FileObject{
		ID:          rxID,
		SHA256:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SizeBytes:   42,
		StoragePath: "objects/rx.bin",
		CreatedAt:   now,
	}).Error; err != nil {
		t.Fatal(err)
	}
	out, err = svc.CheckPurchase(context.Background(), pr, CheckPurchaseInput{
		InstitutionID:            inst,
		ClientID:                 "client-1",
		MedicationID:             med,
		PrescriptionAttachmentID: &rxID,
		PurchaseAt:               now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Allowed {
		t.Fatalf("expected allowed purchase once valid attachment exists, reasons=%v", out.Reasons)
	}
}

func TestCheckPurchase_frequencyCountsAcrossPartitionsForInstitutionLevelRule(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.PurchaseRestriction{}, &model.RestrictionViolationRecord{}, &model.CompliancePurchaseRecord{}); err != nil {
		t.Fatal(err)
	}

	inst := "inst-1"
	now := time.Now().UTC()
	med := "med-a"
	dept := "dept-a"
	team := "team-a"
	rule := []byte(`{"frequencyDays":7}`)
	if err := db.Create(&model.PurchaseRestriction{
		ID:            "r-inst-level",
		InstitutionID: inst,
		ClientID:      "client-1",
		MedicationID:  &med,
		RuleJSON:      rule,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.CompliancePurchaseRecord{
		ID:            "pr-1",
		InstitutionID: inst,
		DepartmentID:  &dept,
		TeamID:        &team,
		ClientID:      "client-1",
		MedicationID:  &med,
		RecordedAt:    now.Add(-24 * time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}

	repo := repository.NewComplianceRepository(db)
	svc := NewComplianceService(repo, NewAuditService(nil))
	pr := &access.Principal{Scopes: []access.Scope{{InstitutionID: inst}}}
	out, err := svc.CheckPurchase(context.Background(), pr, CheckPurchaseInput{
		InstitutionID: inst,
		ClientID:      "client-1",
		MedicationID:  med,
		PurchaseAt:    now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Allowed {
		t.Fatalf("institution-level rule should count prior purchase across dept/team partitions: %+v", out)
	}
}

func TestCheckPurchase_frequencyForScopedRuleOnlyCountsMatchingScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.PurchaseRestriction{}, &model.RestrictionViolationRecord{}, &model.CompliancePurchaseRecord{}); err != nil {
		t.Fatal(err)
	}

	inst := "inst-1"
	now := time.Now().UTC()
	med := "med-a"
	deptA := "dept-a"
	teamA := "team-a"
	deptB := "dept-b"
	teamB := "team-b"
	rule := []byte(`{"frequencyDays":7}`)
	if err := db.Create(&model.PurchaseRestriction{
		ID:            "r-scoped",
		InstitutionID: inst,
		DepartmentID:  &deptA,
		TeamID:        &teamA,
		ClientID:      "client-1",
		MedicationID:  &med,
		RuleJSON:      rule,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.CompliancePurchaseRecord{
		ID:            "pr-2",
		InstitutionID: inst,
		DepartmentID:  &deptB,
		TeamID:        &teamB,
		ClientID:      "client-1",
		MedicationID:  &med,
		RecordedAt:    now.Add(-24 * time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}

	repo := repository.NewComplianceRepository(db)
	svc := NewComplianceService(repo, NewAuditService(nil))
	pr := &access.Principal{Scopes: []access.Scope{{InstitutionID: inst, DepartmentID: &deptA, TeamID: &teamA}}}
	out, err := svc.CheckPurchase(context.Background(), pr, CheckPurchaseInput{
		InstitutionID: inst,
		ClientID:      "client-1",
		MedicationID:  med,
		PurchaseAt:    now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Allowed {
		t.Fatalf("scoped rule should ignore purchases from other dept/team partitions: %+v", out)
	}
}
