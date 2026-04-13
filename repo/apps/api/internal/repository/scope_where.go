package repository

import (
	"strings"

	"gorm.io/gorm"

	"pharmaops/api/internal/access"
)

// applyDataScope adds OR predicates for institution + optional department + team (design §10.2).
func applyDataScope(db *gorm.DB, p *access.Principal, instCol, deptCol, teamCol string) *gorm.DB {
	expr, args, ok := buildDataScopeExpr(p, instCol, deptCol, teamCol)
	if !ok {
		return db.Where("1 = 0")
	}
	return db.Where(expr, args...)
}

// applyInstitutionScope filters by institution_id for entities without department/team columns.
func applyInstitutionScope(db *gorm.DB, p *access.Principal, instCol string) *gorm.DB {
	if p == nil || len(p.Scopes) == 0 {
		return db.Where("1 = 0")
	}
	seen := make(map[string]struct{})
	var parts []string
	var args []any
	for _, s := range p.Scopes {
		if s.InstitutionID == "" {
			continue
		}
		if _, ok := seen[s.InstitutionID]; ok {
			continue
		}
		seen[s.InstitutionID] = struct{}{}
		parts = append(parts, instCol+" = ?")
		args = append(args, s.InstitutionID)
	}
	if len(parts) == 0 {
		return db.Where("1 = 0")
	}
	return db.Where("("+strings.Join(parts, " OR ")+")", args...)
}

func buildDataScopeExpr(p *access.Principal, instCol, deptCol, teamCol string) (sql string, args []any, ok bool) {
	if p == nil || len(p.Scopes) == 0 {
		return "", nil, false
	}
	var parts []string
	for _, s := range p.Scopes {
		if s.InstitutionID == "" {
			continue
		}
		if s.DepartmentID == nil && s.TeamID == nil {
			parts = append(parts, instCol+" = ?")
			args = append(args, s.InstitutionID)
			continue
		}
		if s.TeamID != nil {
			if s.DepartmentID != nil {
				parts = append(parts, "("+instCol+" = ? AND "+deptCol+" = ? AND "+teamCol+" = ?)")
				args = append(args, s.InstitutionID, *s.DepartmentID, *s.TeamID)
			} else {
				parts = append(parts, "("+instCol+" = ? AND "+teamCol+" = ?)")
				args = append(args, s.InstitutionID, *s.TeamID)
			}
			continue
		}
		parts = append(parts, "("+instCol+" = ? AND "+deptCol+" = ?)")
		args = append(args, s.InstitutionID, *s.DepartmentID)
	}
	if len(parts) == 0 {
		return "", nil, false
	}
	return "(" + strings.Join(parts, " OR ") + ")", args, true
}
