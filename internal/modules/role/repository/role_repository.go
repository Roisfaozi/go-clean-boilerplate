package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	querybuilder2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type roleRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewRoleRepository(db *gorm.DB, log *logrus.Logger) RoleRepository {
	return &roleRepository{
		db:  db,
		log: log,
	}
}

func (r *roleRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := tx.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.getDB(ctx).Create(role).Error
}

// Update persists the mutable fields of a role. Only description is writable
// through the API (see model.UpdateRoleRequest); name and organization_id are
// immutable because Casbin policies key off role.Name.
func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.getDB(ctx).Model(&entity.Role{}).Where("id = ?", role.ID).
		Updates(map[string]interface{}{"description": role.Description}).Error
}

func (r *roleRepository) FindByID(ctx context.Context, id string) (*entity.Role, error) {
	var role entity.Role
	if err := r.getDB(ctx).
		Scopes(database.OrganizationScope(ctx), database.OrganizationVisibilityScope(ctx, "roles.organization_id")).
		First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindOrganizationRoleByID(ctx context.Context, organizationID, roleID string) (*entity.Role, error) {
	var role entity.Role
	if err := r.getDB(ctx).
		Scopes(database.OrganizationVisibilityScope(ctx, "roles.organization_id")).
		Where("id = ? AND organization_id = ?", roleID, organizationID).
		First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindOrganizationRoles(ctx context.Context, organizationID string) ([]*entity.Role, error) {
	var roles []*entity.Role
	if err := r.getDB(ctx).
		Scopes(database.OrganizationVisibilityScope(ctx, "roles.organization_id")).
		Where("organization_id = ?", organizationID).
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	if err := r.getDB(ctx).
		Scopes(database.OrganizationScope(ctx), database.OrganizationVisibilityScope(ctx, "roles.organization_id")).
		First(&role, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByNameInScope(ctx context.Context, name string, orgID *string) (*entity.Role, error) {
	var role entity.Role
	q := r.getDB(ctx).Where("name = ?", name)
	if orgID == nil || *orgID == "" {
		q = q.Where("organization_id IS NULL")
	} else {
		q = q.Where("organization_id = ?", *orgID)
	}
	if err := q.First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindAll(ctx context.Context) ([]*entity.Role, error) {
	var roles []*entity.Role
	result := r.getDB(ctx).
		Scopes(database.OrganizationScope(ctx), database.OrganizationVisibilityScope(ctx, "roles.organization_id")).
		Find(&roles)
	if result.Error != nil {
		r.log.WithError(result.Error).Error("Error in FindAll")
		return nil, result.Error
	}

	r.log.WithFields(logrus.Fields{
		"roles_found": len(roles),
		"query": r.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Find(&entity.Role{})
		}),
	}).Info("Roles query executed")
	return roles, nil
}

func (r *roleRepository) FindAllDynamic(ctx context.Context, filter *querybuilder2.DynamicFilter) ([]*entity.Role, error) {
	var roles []*entity.Role
	query := r.getDB(ctx).
		Scopes(database.OrganizationScope(ctx), database.OrganizationVisibilityScope(ctx, "roles.organization_id")).
		Model(&entity.Role{})

	// Apply Dynamic Filter
	query, err := querybuilder2.GenerateDynamicQuery(query, &entity.Role{}, filter)
	if err != nil {
		return nil, err
	}

	query, err = querybuilder2.GenerateDynamicSort(query, &entity.Role{}, filter)
	if err != nil {
		return nil, err
	}

	if err := query.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// Delete removes a role by ID without tenant filtering.
// Callers MUST authorize ownership first (see DeleteInOrg for tenant-scoped deletes).
func (r *roleRepository) Delete(ctx context.Context, id string) error {
	res := r.getDB(ctx).Delete(&entity.Role{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteInOrg removes a role only when it belongs to the given organization.
// The tenant comes from the argument, never from context, so a mismatched or
// absent org context cannot silently turn this into a no-op.
func (r *roleRepository) DeleteInOrg(ctx context.Context, orgID, roleID string) error {
	if orgID == "" || roleID == "" {
		return gorm.ErrRecordNotFound
	}

	res := r.getDB(ctx).
		Where("organization_id = ?", orgID).
		Delete(&entity.Role{}, "id = ?", roleID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
