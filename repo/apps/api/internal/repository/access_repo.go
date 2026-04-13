package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type AccessRepository struct {
	db *gorm.DB
}

func NewAccessRepository(db *gorm.DB) *AccessRepository {
	return &AccessRepository{db: db}
}

func (r *AccessRepository) PermissionCodesForUser(ctx context.Context, userID string) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT p.code FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		INNER JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = ?
	`, userID).Scan(&codes).Error
	return codes, err
}

func (r *AccessRepository) RoleSlugsForUser(ctx context.Context, userID string) ([]string, error) {
	var slugs []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT r.slug FROM roles r
		INNER JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = ?
	`, userID).Scan(&slugs).Error
	return slugs, err
}

type scopeRow struct {
	ID            string
	ScopeKey      string
	InstitutionID string
	DepartmentID  sql.NullString
	TeamID        sql.NullString
}

func (r *AccessRepository) DataScopesForUser(ctx context.Context, userID string) ([]scopeRow, error) {
	var rows []scopeRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT ds.id, ds.scope_key, ds.institution_id, ds.department_id, ds.team_id
		FROM data_scopes ds
		INNER JOIN user_data_scopes uds ON uds.data_scope_id = ds.id
		WHERE uds.user_id = ?
	`, userID).Scan(&rows).Error
	return rows, err
}
