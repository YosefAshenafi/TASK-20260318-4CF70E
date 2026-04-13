package service

import (
	"context"
	"errors"
	"testing"

	"pharmaops/api/internal/access"
)

func TestRecruitmentService_CreateCandidate_forbiddenScope(t *testing.T) {
	s := NewRecruitmentService(nil, nil, NewAuditService(nil))
	inst := "40000000-0000-4000-8000-000000000004"
	p := &access.Principal{
		Scopes: []access.Scope{{InstitutionID: "50000000-0000-4000-8000-000000000005"}},
	}
	_, err := s.CreateCandidate(context.Background(), p, CreateCandidateInput{
		Name:          "N",
		InstitutionID: inst,
	}, GetCandidateOpts{OperatorUserID: "op"})
	if !errors.Is(err, ErrForbiddenScope) {
		t.Fatalf("expected ErrForbiddenScope, got %v", err)
	}
}
