package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"pharmaops/api/internal/access"
)

func TestCheckPurchase_noScopes(t *testing.T) {
	s := NewComplianceService(nil, NewAuditService(nil))
	p := &access.Principal{Scopes: nil}
	_, err := s.CheckPurchase(context.Background(), p, CheckPurchaseInput{
		InstitutionID: "inst-1",
		ClientID:      "c1",
		MedicationID:  "m1",
		PurchaseAt:    time.Now().UTC(),
	})
	if !errors.Is(err, ErrForbiddenScope) {
		t.Fatalf("expected ErrForbiddenScope for empty scopes, got %v", err)
	}
}

func TestCheckPurchase_wrongInstitution(t *testing.T) {
	s := NewComplianceService(nil, NewAuditService(nil))
	p := &access.Principal{
		Scopes: []access.Scope{{InstitutionID: "other-inst"}},
	}
	_, err := s.CheckPurchase(context.Background(), p, CheckPurchaseInput{
		InstitutionID: "inst-1",
		ClientID:      "c1",
		MedicationID:  "m1",
		PurchaseAt:    time.Now().UTC(),
	})
	if !errors.Is(err, ErrForbiddenScope) {
		t.Fatalf("expected ErrForbiddenScope for mismatched institution, got %v", err)
	}
}

func TestRestrictionRuleBranching(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		wantRx        bool
		wantFreqDays  int
	}{
		{"rx_only", `{"requiresPrescription":true}`, true, 0},
		{"freq_only", `{"frequencyDays":7}`, false, 7},
		{"both", `{"requiresPrescription":true,"frequencyDays":14}`, true, 14},
		{"empty", `{}`, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := parseRestrictionRule([]byte(tt.json))
			if err != nil {
				t.Fatal(err)
			}
			if r.RequiresPrescription != tt.wantRx {
				t.Errorf("requiresPrescription: got %v want %v", r.RequiresPrescription, tt.wantRx)
			}
			if r.FrequencyDays != tt.wantFreqDays {
				t.Errorf("frequencyDays: got %d want %d", r.FrequencyDays, tt.wantFreqDays)
			}
		})
	}
}
