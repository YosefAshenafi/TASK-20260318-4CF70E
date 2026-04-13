package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"pharmaops/api/internal/access"
)

func Test_duplicateContentHash_stable(t *testing.T) {
	inst := "10000000-0000-4000-8000-000000000001"
	h1 := duplicateContentHash(inst, "ADR", "Title", "Desc", "2026-04-13T12:00:00Z")
	h2 := duplicateContentHash(inst, "ADR", "Title", "Desc", "2026-04-13T12:00:00Z")
	if h1 != h2 || h1 == "" {
		t.Fatalf("expected stable non-empty hash, got %q %q", h1, h2)
	}
	h3 := duplicateContentHash(inst, "ADR", "Title", "Different", "2026-04-13T12:00:00Z")
	if h3 == h1 {
		t.Fatal("expected hash to change when description changes")
	}
}

func Test_allowedTransition_rules(t *testing.T) {
	if !allowedTransition("submitted", "assigned") {
		t.Fatal("submitted -> assigned")
	}
	if allowedTransition("submitted", "submitted") {
		t.Fatal("same status should not transition")
	}
	if allowedTransition("closed", "in_progress") {
		t.Fatal("closed is terminal")
	}
}

func TestCaseService_CreateCase_objectScopeForbidden(t *testing.T) {
	s := NewCaseService(nil, NewAuditService(nil))
	inst := "20000000-0000-4000-8000-000000000002"
	p := &access.Principal{
		Scopes: []access.Scope{{InstitutionID: "30000000-0000-4000-8000-000000000003"}},
	}
	_, err := s.CreateCase(context.Background(), p, CreateCaseInput{
		InstitutionID: inst,
		CaseType:      "ADR",
		Title:         "T",
		Description:   "D",
		ReportedAt:    time.Now().UTC(),
	}, AuditRequestMeta{OperatorUserID: "op-1"})
	if !errors.Is(err, ErrForbiddenScope) {
		t.Fatalf("expected ErrForbiddenScope, got %v", err)
	}
}

func TestCaseService_CreateCase_noScopes(t *testing.T) {
	s := NewCaseService(nil, NewAuditService(nil))
	p := &access.Principal{Scopes: nil}
	_, err := s.CreateCase(context.Background(), p, CreateCaseInput{
		InstitutionID: "20000000-0000-4000-8000-000000000002",
		CaseType:      "ADR",
		Title:         "T",
		Description:   "D",
		ReportedAt:    time.Now().UTC(),
	}, AuditRequestMeta{})
	if !errors.Is(err, ErrForbiddenScope) {
		t.Fatalf("expected ErrForbiddenScope for empty scopes, got %v", err)
	}
}
