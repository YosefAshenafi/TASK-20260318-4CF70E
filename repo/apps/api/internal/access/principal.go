package access

import "sort"

// PermissionFullAccess grants all permission-gated operations (route-level).
const PermissionFullAccess = "system.full_access"

// Principal holds RBAC and data-scope facts for the authenticated user.
type Principal struct {
	PermissionSet map[string]struct{}
	RoleSlugs       []string
	Scopes          []Scope
}

type Scope struct {
	ID             string
	ScopeKey       string
	InstitutionID  string
	DepartmentID   *string
	TeamID         *string
}

// Has returns true if the principal holds the permission code or full access.
func (p *Principal) Has(code string) bool {
	if p == nil {
		return false
	}
	if _, ok := p.PermissionSet[PermissionFullAccess]; ok {
		return true
	}
	_, ok := p.PermissionSet[code]
	return ok
}

// AllowsInstitution returns true if any assigned data scope includes this institution
// (institution-wide, or narrower department/team rows under that institution).
func (p *Principal) AllowsInstitution(institutionID string) bool {
	if p == nil || institutionID == "" {
		return false
	}
	for _, s := range p.Scopes {
		if s.InstitutionID == institutionID {
			return true
		}
	}
	return false
}

// AllowedInstitutionIDs returns distinct institution IDs from data scopes (for query filtering).
func (p *Principal) AllowedInstitutionIDs() []string {
	if p == nil {
		return nil
	}
	seen := make(map[string]struct{})
	for _, s := range p.Scopes {
		if s.InstitutionID != "" {
			seen[s.InstitutionID] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
