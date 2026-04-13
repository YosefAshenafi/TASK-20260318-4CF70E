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

func TestQualificationService_supplierQualificationCRUD(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QualificationProfile{}); err != nil {
		t.Fatal(err)
	}

	inst := "inst-supplier-1"
	repo := repository.NewComplianceRepository(db)
	auditSvc := NewAuditService(nil)
	svc := NewComplianceService(repo, auditSvc)
	pr := &access.Principal{Scopes: []access.Scope{{InstitutionID: inst}}}

	t.Run("create client qualification", func(t *testing.T) {
		dto, err := svc.CreateQualification(context.Background(), pr, CreateQualificationInput{
			InstitutionID: inst,
			ClientID:      "client-1",
			PartyType:     "client",
			DisplayName:   "Client QF",
			ExpiresOn:     nil,
		}, AuditRequestMeta{})
		if err != nil {
			t.Fatalf("create client qualification: %v", err)
		}
		if dto.ClientID != "client-1" {
			t.Fatalf("expected client-1, got %s", dto.ClientID)
		}
		if dto.PartyType != "client" {
			t.Fatalf("expected partyType client, got %s", dto.PartyType)
		}
	})

	t.Run("create supplier qualification", func(t *testing.T) {
		supplierID := "sup-001"
		dto, err := svc.CreateQualification(context.Background(), pr, CreateQualificationInput{
			InstitutionID: inst,
			ClientID:      "client-1",
			PartyType:     "supplier",
			SupplierID:    &supplierID,
			DisplayName:   "Supplier QF",
			ExpiresOn:     nil,
		}, AuditRequestMeta{})
		if err != nil {
			t.Fatalf("create supplier qualification: %v", err)
		}
		if dto.PartyType != "supplier" {
			t.Fatalf("expected partyType supplier, got %s", dto.PartyType)
		}
		if dto.SupplierID == nil || *dto.SupplierID != "sup-001" {
			t.Fatalf("expected supplierID sup-001, got %v", dto.SupplierID)
		}
	})

	t.Run("list qualifications includes both types", func(t *testing.T) {
		rows, total, _, _, err := svc.ListQualifications(context.Background(), pr, 1, 10, 0, "created_at", "desc")
		if err != nil {
			t.Fatalf("list qualifications: %v", err)
		}
		if total != 2 {
			t.Fatalf("expected 2 qualifications, got %d", total)
		}
		clientFound, supplierFound := false, false
		for _, r := range rows {
			if r.PartyType == "client" {
				clientFound = true
			}
			if r.PartyType == "supplier" {
				supplierFound = true
			}
		}
		if !clientFound {
			t.Fatal("client qualification not found in list")
		}
		if !supplierFound {
			t.Fatal("supplier qualification not found in list")
		}
	})

	t.Run("get supplier qualification by id", func(t *testing.T) {
		allRows, _, _, _, err := svc.ListQualifications(context.Background(), pr, 1, 10, 0, "created_at", "desc")
		if err != nil {
			t.Fatal(err)
		}
		var supplierID string
		for _, r := range allRows {
			if r.PartyType == "supplier" {
				supplierID = r.ID
				break
			}
		}
		if supplierID == "" {
			t.Fatal("supplier qualification not found")
		}
		dto, err := svc.GetQualification(context.Background(), pr, supplierID)
		if err != nil {
			t.Fatalf("get qualification: %v", err)
		}
		if dto.PartyType != "supplier" {
			t.Fatalf("expected supplier, got %s", dto.PartyType)
		}
	})

	t.Run("deactivate and activate supplier qualification", func(t *testing.T) {
		allRows, _, _, _, err := svc.ListQualifications(context.Background(), pr, 1, 10, 0, "created_at", "desc")
		if err != nil {
			t.Fatal(err)
		}
		var supplierID string
		for _, r := range allRows {
			if r.PartyType == "supplier" {
				supplierID = r.ID
				break
			}
		}
		dto, err := svc.DeactivateQualification(context.Background(), pr, supplierID, AuditRequestMeta{})
		if err != nil {
			t.Fatalf("deactivate: %v", err)
		}
		if dto.Status != "inactive" {
			t.Fatalf("expected status inactive, got %s", dto.Status)
		}
		dto, err = svc.ActivateQualification(context.Background(), pr, supplierID, AuditRequestMeta{})
		if err != nil {
			t.Fatalf("activate: %v", err)
		}
		if dto.Status != "active" {
			t.Fatalf("expected status active, got %s", dto.Status)
		}
	})
}

func TestQualificationService_supplierScopeIsolation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QualificationProfile{}); err != nil {
		t.Fatal(err)
	}

	instA, instB := "inst-a", "inst-b"
	now := time.Now().UTC()

	if err := db.Create(&model.QualificationProfile{
		ID:            "qf-a",
		InstitutionID: instA,
		ClientID:      "client-a",
		PartyType:     "supplier",
		SupplierID:    func() *string { s := "sup-a"; return &s }(),
		DisplayName:   "Supplier A",
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.QualificationProfile{
		ID:            "qf-b",
		InstitutionID: instB,
		ClientID:      "client-b",
		PartyType:     "supplier",
		SupplierID:    func() *string { s := "sup-b"; return &s }(),
		DisplayName:   "Supplier B",
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatal(err)
	}

	repo := repository.NewComplianceRepository(db)
	svc := NewComplianceService(repo, NewAuditService(nil))

	prA := &access.Principal{Scopes: []access.Scope{{InstitutionID: instA}}}

	rowsA, totalA, _, _, err := svc.ListQualifications(context.Background(), prA, 1, 10, 0, "created_at", "desc")
	if err != nil {
		t.Fatalf("list for A: %v", err)
	}
	if totalA != 1 {
		t.Fatalf("inst A should see 1 qualification, got %d", totalA)
	}
	if rowsA[0].ID != "qf-a" {
		t.Fatalf("expected qf-a, got %s", rowsA[0].ID)
	}

	prB := &access.Principal{Scopes: []access.Scope{{InstitutionID: instB}}}

	rowsB, totalB, _, _, err := svc.ListQualifications(context.Background(), prB, 1, 10, 0, "created_at", "desc")
	if err != nil {
		t.Fatalf("list for B: %v", err)
	}
	if totalB != 1 {
		t.Fatalf("inst B should see 1 qualification, got %d", totalB)
	}
	if rowsB[0].ID != "qf-b" {
		t.Fatalf("expected qf-b, got %s", rowsB[0].ID)
	}
}
