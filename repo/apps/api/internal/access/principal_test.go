package access

import "testing"

func TestPrincipal_Has_fullAccess(t *testing.T) {
	p := &Principal{
		PermissionSet: map[string]struct{}{PermissionFullAccess: {}},
	}
	if !p.Has("recruitment.candidates.read") {
		t.Fatal("expected full access to allow any permission code")
	}
}

func TestPrincipal_Has_specific(t *testing.T) {
	p := &Principal{
		PermissionSet: map[string]struct{}{"a.b": {}},
	}
	if !p.Has("a.b") || p.Has("c.d") {
		t.Fatal("specific permission mismatch")
	}
}

func TestPrincipal_AllowsInstitution(t *testing.T) {
	inst := "10000000-0000-4000-8000-000000000001"
	p := &Principal{
		Scopes: []Scope{{InstitutionID: inst}},
	}
	if !p.AllowsInstitution(inst) || p.AllowsInstitution("other") {
		t.Fatal("institution scope mismatch")
	}
}

func TestPrincipal_AllowedInstitutionIDs(t *testing.T) {
	p := &Principal{
		Scopes: []Scope{
			{InstitutionID: "b"},
			{InstitutionID: "a"},
			{InstitutionID: "a"},
		},
	}
	got := p.AllowedInstitutionIDs()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected: %v", got)
	}
	if (*Principal)(nil).AllowedInstitutionIDs() != nil {
		t.Fatal("nil principal")
	}
}

func TestPrincipal_RowVisible(t *testing.T) {
	inst := "10000000-0000-4000-8000-000000000001"
	dept := "d1"
	team := "t1"
	deptPtr := &dept
	teamPtr := &team

	t.Run("institution_wide", func(t *testing.T) {
		p := &Principal{Scopes: []Scope{{InstitutionID: inst}}}
		if !p.RowVisible(inst, nil, nil) || !p.RowVisible(inst, deptPtr, nil) {
			t.Fatal("expected institution-wide rows to see any org assignment")
		}
	})

	t.Run("department_only", func(t *testing.T) {
		p := &Principal{Scopes: []Scope{{InstitutionID: inst, DepartmentID: deptPtr}}}
		if p.RowVisible(inst, nil, nil) {
			t.Fatal("expected NULL department to be hidden")
		}
		if !p.RowVisible(inst, deptPtr, nil) || !p.RowVisible(inst, deptPtr, teamPtr) {
			t.Fatal("expected matching department (team on row does not exclude)")
		}
	})

	t.Run("department_and_team", func(t *testing.T) {
		p := &Principal{Scopes: []Scope{{InstitutionID: inst, DepartmentID: deptPtr, TeamID: teamPtr}}}
		if p.RowVisible(inst, deptPtr, nil) || !p.RowVisible(inst, deptPtr, teamPtr) {
			t.Fatal("expected team match required")
		}
	})
}

func TestDefaultOrgAssignment(t *testing.T) {
	inst := "10000000-0000-4000-8000-000000000001"
	dept := "d1"
	deptPtr := &dept
	p := &Principal{Scopes: []Scope{{InstitutionID: inst, DepartmentID: deptPtr}}}
	d, tm := DefaultOrgAssignment(p, inst)
	if d != deptPtr || tm != nil {
		t.Fatalf("narrow scope: got dept=%v team=%v", d, tm)
	}
	p2 := &Principal{Scopes: []Scope{{InstitutionID: inst}}}
	d2, tm2 := DefaultOrgAssignment(p2, inst)
	if d2 != nil || tm2 != nil {
		t.Fatalf("institution-wide scope")
	}
}
