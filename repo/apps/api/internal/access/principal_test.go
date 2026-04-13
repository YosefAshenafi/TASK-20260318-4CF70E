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
