package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pharmaops/api/internal/model"
	"pharmaops/api/internal/repository"
)

// UserSummaryDTO for GET /users.
type UserSummaryDTO struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	IsActive    bool     `json:"isActive"`
	Roles       []string `json:"roles"`
	CreatedAt   string   `json:"createdAt"`
}

const minUserPasswordLen = 8

// UserDetailDTO is GET /users/:id — includes ids for role/scope assignment UIs.
type UserDetailDTO struct {
	UserSummaryDTO
	RoleIDs  []string `json:"roleIds"`
	ScopeIDs []string `json:"scopeIds"`
}

// RoleDTO for GET /roles.
type RoleDTO struct {
	ID          string  `json:"id"`
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// RoleDetailDTO includes permission ids.
type RoleDetailDTO struct {
	RoleDTO
	PermissionIDs []string `json:"permissionIds"`
}

// PermissionDTO for GET /permissions.
type PermissionDTO struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Description *string `json:"description,omitempty"`
	CreatedAt   string  `json:"createdAt"`
}

// DataScopeDTO for GET /scopes.
type DataScopeDTO struct {
	ID            string  `json:"id"`
	ScopeKey      string  `json:"scopeKey"`
	InstitutionID string  `json:"institutionId"`
	DepartmentID  *string `json:"departmentId,omitempty"`
	TeamID        *string `json:"teamId,omitempty"`
	CreatedAt     string  `json:"createdAt"`
}

type RbacService struct {
	users *repository.UserRepository
	repo  *repository.RbacRepository
}

func NewRbacService(users *repository.UserRepository, rbac *repository.RbacRepository) *RbacService {
	return &RbacService{users: users, repo: rbac}
}

func (s *RbacService) ListUsers(ctx context.Context) ([]UserSummaryDTO, error) {
	rows, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]UserSummaryDTO, 0, len(rows))
	for _, r := range rows {
		var roles []string
		if r.RoleSlugs != "" {
			for _, p := range strings.Split(r.RoleSlugs, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					roles = append(roles, p)
				}
			}
		}
		out = append(out, UserSummaryDTO{
			ID:          r.ID,
			Username:    r.Username,
			DisplayName: r.DisplayName,
			IsActive:    r.IsActive,
			Roles:       roles,
			CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

func toRoleDTO(r *model.Role) RoleDTO {
	return RoleDTO{
		ID:          r.ID,
		Slug:        r.Slug,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (s *RbacService) ListRoles(ctx context.Context) ([]RoleDTO, error) {
	rows, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]RoleDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toRoleDTO(&rows[i]))
	}
	return out, nil
}

func (s *RbacService) GetRole(ctx context.Context, id string) (*RoleDetailDTO, error) {
	r, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	permIDs, err := s.repo.PermissionIDsForRole(ctx, id)
	if err != nil {
		return nil, err
	}
	d := RoleDetailDTO{
		RoleDTO:       toRoleDTO(r),
		PermissionIDs: permIDs,
	}
	return &d, nil
}

func (s *RbacService) ListPermissions(ctx context.Context) ([]PermissionDTO, error) {
	rows, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PermissionDTO, 0, len(rows))
	for i := range rows {
		p := rows[i]
		out = append(out, PermissionDTO{
			ID:          p.ID,
			Code:        p.Code,
			Description: p.Description,
			CreatedAt:   p.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

// UpdateRoleInput for PATCH /roles/:id.
type UpdateRoleInput struct {
	Name        *string
	Description *string
}

func (s *RbacService) UpdateRole(ctx context.Context, id string, in UpdateRoleInput) (*RoleDTO, error) {
	r, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		r.Name = strings.TrimSpace(*in.Name)
		if r.Name == "" {
			return nil, ErrRbacValidation
		}
	}
	if in.Description != nil {
		r.Description = in.Description
	}
	if err := s.repo.UpdateRole(ctx, r); err != nil {
		return nil, err
	}
	loaded, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toRoleDTO(loaded)
	return &dto, nil
}

// ErrRbacValidation for invalid RBAC input.
var ErrRbacValidation = errors.New("rbac validation failed")

// SetRolePermissions replaces role_permissions for a role.
func (s *RbacService) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	if _, err := s.repo.GetRole(ctx, roleID); err != nil {
		return err
	}
	if len(permissionIDs) > 0 {
		n, err := s.repo.CountPermissionsByIDs(ctx, permissionIDs)
		if err != nil {
			return err
		}
		if int(n) != len(permissionIDs) {
			return ErrRbacValidation
		}
	}
	return s.repo.ReplaceRolePermissions(ctx, roleID, permissionIDs)
}

func (s *RbacService) ListScopes(ctx context.Context) ([]DataScopeDTO, error) {
	rows, err := s.repo.ListDataScopes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]DataScopeDTO, 0, len(rows))
	for i := range rows {
		ds := rows[i]
		out = append(out, DataScopeDTO{
			ID:            ds.ID,
			ScopeKey:      ds.ScopeKey,
			InstitutionID: ds.InstitutionID,
			DepartmentID:  ds.DepartmentID,
			TeamID:        ds.TeamID,
			CreatedAt:     ds.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

// GetUser returns one user with role and scope ids for admin UIs.
func (s *RbacService) GetUser(ctx context.Context, id string) (*UserDetailDTO, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	slugs, err := s.repo.RoleSlugsForUser(ctx, id)
	if err != nil {
		return nil, err
	}
	rids, err := s.repo.RoleIDsForUser(ctx, id)
	if err != nil {
		return nil, err
	}
	sids, err := s.repo.ScopeIDsForUser(ctx, id)
	if err != nil {
		return nil, err
	}
	summary := UserSummaryDTO{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		IsActive:    u.IsActive,
		Roles:       slugs,
		CreatedAt:   u.CreatedAt.UTC().Format(time.RFC3339),
	}
	return &UserDetailDTO{
		UserSummaryDTO: summary,
		RoleIDs:        rids,
		ScopeIDs:       sids,
	}, nil
}

// CreateUserInput for POST /users.
type CreateUserInput struct {
	Username    string
	Password    string
	DisplayName string
	IsActive    bool
	RoleIDs     []string
}

// CreateUser provisions a login account and role links.
func (s *RbacService) CreateUser(ctx context.Context, in CreateUserInput) (*UserDetailDTO, error) {
	username := strings.TrimSpace(in.Username)
	display := strings.TrimSpace(in.DisplayName)
	if username == "" || display == "" {
		return nil, ErrRbacValidation
	}
	if len(in.Password) < minUserPasswordLen {
		return nil, ErrRbacValidation
	}
	_, err := s.users.FindByUsername(ctx, username)
	if err == nil {
		return nil, ErrRbacValidation
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if len(in.RoleIDs) > 0 {
		n, err := s.repo.CountRolesByIDs(ctx, in.RoleIDs)
		if err != nil {
			return nil, err
		}
		if int(n) != len(in.RoleIDs) {
			return nil, ErrRbacValidation
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	uid := uuid.NewString()
	now := time.Now().UTC()
	u := &model.User{
		ID:           uid,
		Username:     username,
		PasswordHash: string(hash),
		DisplayName:  display,
		IsActive:     in.IsActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	err = s.users.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ur := repository.NewUserRepository(tx)
		rr := repository.NewRbacRepository(tx)
		if err := ur.Create(ctx, u); err != nil {
			return err
		}
		return rr.ReplaceUserRoles(ctx, uid, in.RoleIDs)
	})
	if err != nil {
		return nil, err
	}
	return s.GetUser(ctx, uid)
}

// UpdateUserInput for PATCH /users/:id.
type UpdateUserInput struct {
	DisplayName *string
	IsActive    *bool
	Password    *string
	RoleIDs     *[]string
}

// UpdateUser patches profile, optional password, and optionally replaces roles.
func (s *RbacService) UpdateUser(ctx context.Context, id string, in UpdateUserInput) (*UserDetailDTO, error) {
	if _, err := s.users.FindByID(ctx, id); err != nil {
		return nil, err
	}
	patch := map[string]any{}
	if in.DisplayName != nil {
		d := strings.TrimSpace(*in.DisplayName)
		if d == "" {
			return nil, ErrRbacValidation
		}
		patch["display_name"] = d
	}
	if in.IsActive != nil {
		patch["is_active"] = *in.IsActive
	}
	if in.Password != nil && *in.Password != "" {
		if len(*in.Password) < minUserPasswordLen {
			return nil, ErrRbacValidation
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		patch["password_hash"] = string(hash)
	}
	if len(patch) > 0 {
		if err := s.users.PatchUser(ctx, id, patch); err != nil {
			return nil, err
		}
	}
	if in.RoleIDs != nil {
		rids := *in.RoleIDs
		if len(rids) > 0 {
			n, err := s.repo.CountRolesByIDs(ctx, rids)
			if err != nil {
				return nil, err
			}
			if int(n) != len(rids) {
				return nil, ErrRbacValidation
			}
		}
		if err := s.repo.ReplaceUserRoles(ctx, id, rids); err != nil {
			return nil, err
		}
	}
	return s.GetUser(ctx, id)
}

// SetUserScopes replaces data scopes for a user (POST /users/:id/scopes).
func (s *RbacService) SetUserScopes(ctx context.Context, userID string, scopeIDs []string) error {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return err
	}
	if len(scopeIDs) > 0 {
		n, err := s.repo.CountDataScopesByIDs(ctx, scopeIDs)
		if err != nil {
			return err
		}
		if int(n) != len(scopeIDs) {
			return ErrRbacValidation
		}
	}
	return s.repo.ReplaceUserScopes(ctx, userID, scopeIDs)
}

var roleSlugRe = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}$`)

// CreateRoleInput for POST /roles.
type CreateRoleInput struct {
	Slug        string
	Name        string
	Description *string
}

// CreateRole inserts a new role (slug unique).
func (s *RbacService) CreateRole(ctx context.Context, in CreateRoleInput) (*RoleDTO, error) {
	slug := strings.ToLower(strings.TrimSpace(in.Slug))
	name := strings.TrimSpace(in.Name)
	if slug == "" || name == "" || !roleSlugRe.MatchString(slug) {
		return nil, ErrRbacValidation
	}
	taken, err := s.repo.RoleSlugExists(ctx, slug)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrRbacValidation
	}
	now := time.Now().UTC()
	row := &model.Role{
		ID:          uuid.NewString(),
		Slug:        slug,
		Name:        name,
		Description: in.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateRole(ctx, row); err != nil {
		return nil, err
	}
	dto := toRoleDTO(row)
	return &dto, nil
}

// CreateDataScopeInput for POST /scopes.
type CreateDataScopeInput struct {
	ScopeKey      string
	InstitutionID string
	DepartmentID  *string
	TeamID        *string
}

func emptyStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

// CreateDataScope inserts a data scope row; validates institution/dept/team chain.
func (s *RbacService) CreateDataScope(ctx context.Context, in CreateDataScopeInput) (*DataScopeDTO, error) {
	in.DepartmentID = emptyStringPtr(in.DepartmentID)
	in.TeamID = emptyStringPtr(in.TeamID)
	key := strings.TrimSpace(in.ScopeKey)
	inst := strings.TrimSpace(in.InstitutionID)
	if key == "" || len(inst) != 36 {
		return nil, ErrRbacValidation
	}
	taken, err := s.repo.DataScopeKeyExists(ctx, key)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrRbacValidation
	}
	dept, team, err := s.repo.NormalizeScopeHierarchy(ctx, inst, in.DepartmentID, in.TeamID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRbacValidation
	}
	if err != nil {
		return nil, err
	}
	row := &model.DataScope{
		ID:            uuid.NewString(),
		ScopeKey:      key,
		InstitutionID: inst,
		DepartmentID:  dept,
		TeamID:        team,
		CreatedAt:     time.Now().UTC(),
	}
	if err := s.repo.CreateDataScope(ctx, row); err != nil {
		return nil, err
	}
	return &DataScopeDTO{
		ID:            row.ID,
		ScopeKey:      row.ScopeKey,
		InstitutionID: row.InstitutionID,
		DepartmentID:  row.DepartmentID,
		TeamID:        row.TeamID,
		CreatedAt:     row.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}
