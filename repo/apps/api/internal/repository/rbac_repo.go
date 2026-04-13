package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pharmaops/api/internal/model"
)

type RbacRepository struct {
	db *gorm.DB
}

func NewRbacRepository(db *gorm.DB) *RbacRepository {
	return &RbacRepository{db: db}
}

// GetDB exposes the DB handle for transactions that span repositories.
func (r *RbacRepository) GetDB() *gorm.DB {
	return r.db
}

// RoleSlugsForUser returns role slugs assigned to the user.
func (r *RbacRepository) RoleSlugsForUser(ctx context.Context, userID string) ([]string, error) {
	var slugs []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT r.slug FROM user_roles ur INNER JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = ? ORDER BY r.slug
	`, userID).Scan(&slugs).Error
	return slugs, err
}

type UserListRow struct {
	ID          string    `gorm:"column:id"`
	Username    string    `gorm:"column:username"`
	DisplayName string    `gorm:"column:display_name"`
	IsActive    bool      `gorm:"column:is_active"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	RoleSlugs   string    `gorm:"column:role_slugs"`
}

func (r *RbacRepository) ListUsers(ctx context.Context) ([]UserListRow, error) {
	var rows []UserListRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT u.id, u.username, u.display_name, u.is_active, u.created_at,
			IFNULL(GROUP_CONCAT(DISTINCT r.slug ORDER BY r.slug SEPARATOR ','), '') AS role_slugs
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		GROUP BY u.id, u.username, u.display_name, u.is_active, u.created_at
		ORDER BY u.username
	`).Scan(&rows).Error
	return rows, err
}

func (r *RbacRepository) ListRoles(ctx context.Context) ([]model.Role, error) {
	var rows []model.Role
	err := r.db.WithContext(ctx).Order("slug").Find(&rows).Error
	return rows, err
}

func (r *RbacRepository) GetRole(ctx context.Context, id string) (*model.Role, error) {
	var row model.Role
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RbacRepository) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	var rows []model.Permission
	err := r.db.WithContext(ctx).Order("code").Find(&rows).Error
	return rows, err
}

func (r *RbacRepository) PermissionIDsForRole(ctx context.Context, roleID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT permission_id FROM role_permissions WHERE role_id = ? ORDER BY permission_id
	`, roleID).Scan(&ids).Error
	return ids, err
}

func (r *RbacRepository) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM role_permissions WHERE role_id = ?`, roleID).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, pid := range permissionIDs {
			if err := tx.Exec(`
				INSERT INTO role_permissions (role_id, permission_id, created_at)
				VALUES (?, ?, ?)
			`, roleID, pid, now).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RbacRepository) UpdateRole(ctx context.Context, role *model.Role) error {
	res := r.db.WithContext(ctx).Model(&model.Role{}).
		Where("id = ?", role.ID).
		Updates(map[string]interface{}{
			"name":        role.Name,
			"description": role.Description,
			"updated_at":  time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *RbacRepository) ListDataScopes(ctx context.Context) ([]model.DataScope, error) {
	var rows []model.DataScope
	err := r.db.WithContext(ctx).Order("scope_key").Find(&rows).Error
	return rows, err
}

func (r *RbacRepository) CountPermissionsByIDs(ctx context.Context, ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("id IN ?", ids).Count(&n).Error
	return n, err
}

// CountRolesByIDs returns how many of the given role ids exist.
func (r *RbacRepository) CountRolesByIDs(ctx context.Context, ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Role{}).Where("id IN ?", ids).Count(&n).Error
	return n, err
}

// CountDataScopesByIDs returns how many of the given data scope ids exist.
func (r *RbacRepository) CountDataScopesByIDs(ctx context.Context, ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var n int64
	err := r.db.WithContext(ctx).Model(&model.DataScope{}).Where("id IN ?", ids).Count(&n).Error
	return n, err
}

func (r *RbacRepository) RoleIDsForUser(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Raw(`SELECT role_id FROM user_roles WHERE user_id = ? ORDER BY role_id`, userID).Scan(&ids).Error
	return ids, err
}

func (r *RbacRepository) ScopeIDsForUser(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Raw(`SELECT data_scope_id FROM user_data_scopes WHERE user_id = ? ORDER BY data_scope_id`, userID).Scan(&ids).Error
	return ids, err
}

func (r *RbacRepository) ReplaceUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM user_roles WHERE user_id = ?`, userID).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, rid := range roleIDs {
			if err := tx.Exec(`INSERT INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, ?)`, userID, rid, now).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RbacRepository) ReplaceUserScopes(ctx context.Context, userID string, scopeIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM user_data_scopes WHERE user_id = ?`, userID).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, sid := range scopeIDs {
			if err := tx.Exec(`INSERT INTO user_data_scopes (user_id, data_scope_id, created_at) VALUES (?, ?, ?)`, userID, sid, now).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateRole inserts a role row.
func (r *RbacRepository) CreateRole(ctx context.Context, row *model.Role) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// RoleSlugExists reports whether slug is taken.
func (r *RbacRepository) RoleSlugExists(ctx context.Context, slug string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Role{}).Where("slug = ?", slug).Count(&n).Error
	return n > 0, err
}

// CreateDataScope inserts a data_scopes row.
func (r *RbacRepository) CreateDataScope(ctx context.Context, row *model.DataScope) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// DataScopeKeyExists reports whether scope_key is taken.
func (r *RbacRepository) DataScopeKeyExists(ctx context.Context, key string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.DataScope{}).Where("scope_key = ?", key).Count(&n).Error
	return n > 0, err
}

// InstitutionExists returns whether an institution id exists.
func (r *RbacRepository) InstitutionExists(ctx context.Context, id string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM institutions WHERE id = ?`, id).Scan(&n).Error
	return n > 0, err
}

// NormalizeScopeHierarchy validates institution and optional department/team; returns final dept/team ids for storage.
// Team, if set, must belong to institution; the team's department is used when team is set.
func (r *RbacRepository) NormalizeScopeHierarchy(ctx context.Context, institutionID string, departmentID *string, teamID *string) (deptOut *string, teamOut *string, err error) {
	ok, err := r.InstitutionExists(ctx, institutionID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, gorm.ErrRecordNotFound
	}
	if teamID != nil && *teamID != "" {
		var rw struct {
			DeptID string `gorm:"column:dept_id"`
			InstID string `gorm:"column:inst_id"`
		}
		err := r.db.WithContext(ctx).Raw(`
			SELECT t.department_id AS dept_id, d.institution_id AS inst_id
			FROM teams t
			INNER JOIN departments d ON d.id = t.department_id
			WHERE t.id = ?
		`, *teamID).Scan(&rw).Error
		if err != nil {
			return nil, nil, err
		}
		if rw.DeptID == "" || rw.InstID != institutionID {
			return nil, nil, gorm.ErrRecordNotFound
		}
		if departmentID != nil && *departmentID != "" && *departmentID != rw.DeptID {
			return nil, nil, gorm.ErrRecordNotFound
		}
		d := rw.DeptID
		t := *teamID
		return &d, &t, nil
	}
	if departmentID != nil && *departmentID != "" {
		var inst string
		err := r.db.WithContext(ctx).Raw(`SELECT institution_id FROM departments WHERE id = ?`, *departmentID).Scan(&inst).Error
		if err != nil {
			return nil, nil, err
		}
		if inst != institutionID {
			return nil, nil, gorm.ErrRecordNotFound
		}
		d := *departmentID
		return &d, nil, nil
	}
	return nil, nil, nil
}
