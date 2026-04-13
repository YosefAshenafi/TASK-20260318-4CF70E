package service

import (
	"context"
	"testing"
)

func TestDTOToAuditMap_roundTrip(t *testing.T) {
	m := DTOToAuditMap(QualificationDTO{ID: "q1", ClientID: "c1", DisplayName: "x", Status: "active", CreatedAt: "t"})
	if m["id"] != "q1" || m["displayName"] != "x" {
		t.Fatalf("map: %#v", m)
	}
}

func TestAuditService_LogMutation_nilRepo_noPanic(t *testing.T) {
	s := &AuditService{repo: nil}
	err := s.LogMutation(context.Background(), AuditMutationInput{
		Module:     "cases",
		Operation:  "case.create",
		TargetType: "case",
		TargetID:   "id1",
		After:      map[string]any{"ok": true},
		Meta:       AuditRequestMeta{OperatorUserID: "u1"},
	})
	if err != nil {
		t.Fatal(err)
	}
}
