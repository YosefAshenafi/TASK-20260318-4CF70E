package access

// DefaultOrgAssignment returns department/team pointers for new rows when the caller omits them.
// Institution-wide scope for this institution yields nil, nil; otherwise the first narrow scope for that institution is used.
func DefaultOrgAssignment(p *Principal, institutionID string) (deptID, teamID *string) {
	if p == nil {
		return nil, nil
	}
	var narrow *Scope
	for i := range p.Scopes {
		s := &p.Scopes[i]
		if s.InstitutionID != institutionID {
			continue
		}
		if s.DepartmentID == nil && s.TeamID == nil {
			return nil, nil
		}
		if narrow == nil {
			narrow = s
		}
	}
	if narrow == nil {
		return nil, nil
	}
	return narrow.DepartmentID, narrow.TeamID
}
