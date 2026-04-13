package service

import (
	"context"

	"pharmaops/api/internal/access"
	"pharmaops/api/internal/repository"
)

type AccessService struct {
	repo *repository.AccessRepository
}

func NewAccessService(repo *repository.AccessRepository) *AccessService {
	return &AccessService{repo: repo}
}

func (s *AccessService) LoadPrincipal(ctx context.Context, userID string) (*access.Principal, error) {
	codes, err := s.repo.PermissionCodesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	slugs, err := s.repo.RoleSlugsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.DataScopesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	pset := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		pset[c] = struct{}{}
	}
	scopes := make([]access.Scope, 0, len(rows))
	for _, r := range rows {
		s := access.Scope{
			ID:            r.ID,
			ScopeKey:      r.ScopeKey,
			InstitutionID: r.InstitutionID,
		}
		if r.DepartmentID.Valid {
			v := r.DepartmentID.String
			s.DepartmentID = &v
		}
		if r.TeamID.Valid {
			v := r.TeamID.String
			s.TeamID = &v
		}
		scopes = append(scopes, s)
	}
	return &access.Principal{
		PermissionSet: pset,
		RoleSlugs:     slugs,
		Scopes:        scopes,
	}, nil
}
