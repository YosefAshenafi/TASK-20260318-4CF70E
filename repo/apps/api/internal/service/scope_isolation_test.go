package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"pharmaops/api/internal/access"
)

func TestRecruitmentService_CreateCandidate_scopeDeptTeam(t *testing.T) {
	s := NewRecruitmentService(nil, nil, NewAuditService(nil))
	inst := "10000000-0000-4000-8000-000000000001"
	dept := "dept-1"
	team := "team-1"

	t.Run("dept_team_scope_rejects_wrong_dept", func(t *testing.T) {
		p := &access.Principal{
			Scopes: []access.Scope{{
				InstitutionID: inst,
				DepartmentID:  &dept,
				TeamID:        &team,
			}},
		}
		wrongDept := "other-dept"
		_, err := s.CreateCandidate(context.Background(), p, CreateCandidateInput{
			InstitutionID: inst,
			DepartmentID:  &wrongDept,
			TeamID:        &team,
		}, GetCandidateOpts{})
		if !errors.Is(err, ErrForbiddenScope) {
			t.Fatalf("expected ErrForbiddenScope, got %v", err)
		}
	})

	t.Run("dept_team_scope_rejects_wrong_team", func(t *testing.T) {
		p := &access.Principal{
			Scopes: []access.Scope{{
				InstitutionID: inst,
				DepartmentID:  &dept,
				TeamID:        &team,
			}},
		}
		wrongTeam := "other-team"
		_, err := s.CreateCandidate(context.Background(), p, CreateCandidateInput{
			InstitutionID: inst,
			DepartmentID:  &dept,
			TeamID:        &wrongTeam,
		}, GetCandidateOpts{})
		if !errors.Is(err, ErrForbiddenScope) {
			t.Fatalf("expected ErrForbiddenScope, got %v", err)
		}
	})

	t.Run("institution_wide_scope_allows_any_dept_team", func(t *testing.T) {
		p := &access.Principal{
			Scopes: []access.Scope{{InstitutionID: inst}},
		}
		if !p.RowVisible(inst, &dept, &team) {
			t.Fatal("institution-wide scope should allow any dept/team via RowVisible")
		}
	})
}

func TestCaseService_CreateCase_scopeDeptTeam(t *testing.T) {
	s := NewCaseService(nil, NewAuditService(nil))
	inst := "10000000-0000-4000-8000-000000000001"
	dept := "dept-1"

	t.Run("dept_scope_rejects_wrong_dept", func(t *testing.T) {
		p := &access.Principal{
			Scopes: []access.Scope{{
				InstitutionID: inst,
				DepartmentID:  &dept,
			}},
		}
		wrongDept := "other-dept"
		_, err := s.CreateCase(context.Background(), p, CreateCaseInput{
			InstitutionID: inst,
			DepartmentID:  &wrongDept,
			CaseType:      "ADR",
			Title:         "Test",
			Description:   "Test desc",
			ReportedAt:    time.Now().UTC(),
		}, AuditRequestMeta{OperatorUserID: "op-1"})
		if !errors.Is(err, ErrForbiddenScope) {
			t.Fatalf("expected ErrForbiddenScope, got %v", err)
		}
	})
}

func TestComplianceService_CreateQualification_scopeCheck(t *testing.T) {
	s := NewComplianceService(nil, NewAuditService(nil))
	inst := "10000000-0000-4000-8000-000000000001"
	dept := "dept-1"

	t.Run("wrong_institution", func(t *testing.T) {
		p := &access.Principal{
			Scopes: []access.Scope{{InstitutionID: "other-inst"}},
		}
		_, err := s.CreateQualification(context.Background(), p, CreateQualificationInput{
			InstitutionID: inst,
			ClientID:      "c1",
			DisplayName:   "Test",
		}, AuditRequestMeta{OperatorUserID: "op-1"})
		if !errors.Is(err, ErrForbiddenScope) {
			t.Fatalf("expected ErrForbiddenScope, got %v", err)
		}
	})

	t.Run("dept_scope_rejects_wrong_dept_explicit", func(t *testing.T) {
		p := &access.Principal{
			Scopes: []access.Scope{{
				InstitutionID: inst,
				DepartmentID:  &dept,
			}},
		}
		wrongDept := "wrong-dept"
		_, err := s.CreateQualification(context.Background(), p, CreateQualificationInput{
			InstitutionID: inst,
			DepartmentID:  &wrongDept,
			ClientID:      "c1",
			DisplayName:   "Test",
		}, AuditRequestMeta{OperatorUserID: "op-1"})
		if !errors.Is(err, ErrForbiddenScope) {
			t.Fatalf("expected ErrForbiddenScope, got %v", err)
		}
	})
}

func TestRecruitmentService_CreatePosition_scopeCheck(t *testing.T) {
	s := NewRecruitmentService(nil, nil, NewAuditService(nil))
	inst := "10000000-0000-4000-8000-000000000001"

	t.Run("wrong_institution", func(t *testing.T) {
		p := &access.Principal{
			Scopes: []access.Scope{{InstitutionID: "other-inst"}},
		}
		_, err := s.CreatePosition(context.Background(), p, CreatePositionInput{
			InstitutionID: inst,
			Title:         "Tester",
		}, AuditRequestMeta{OperatorUserID: "op-1"})
		if !errors.Is(err, ErrForbiddenScope) {
			t.Fatalf("expected ErrForbiddenScope, got %v", err)
		}
	})

	t.Run("nil_principal", func(t *testing.T) {
		_, err := s.CreatePosition(context.Background(), nil, CreatePositionInput{
			InstitutionID: inst,
			Title:         "Tester",
		}, AuditRequestMeta{})
		if !errors.Is(err, ErrForbiddenScope) {
			t.Fatalf("expected ErrForbiddenScope for nil principal, got %v", err)
		}
	})
}
