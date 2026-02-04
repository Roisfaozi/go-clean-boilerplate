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

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.getDB(ctx).Model(role).Omit("ID", "Name", "CreatedAt").Updates(role).Error
}

func (r *roleRepository) FindByID(ctx context.Context, id string) (*entity.Role, error) {
	var role entity.Role
	if err := r.getDB(ctx).First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	if err := r.getDB(ctx).First(&role, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindAll(ctx context.Context) ([]*entity.Role, error) {
	var roles []*entity.Role
	result := r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).Find(&roles)
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
	query := r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).Model(&entity.Role{})

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

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&entity.Role{}, "id = ?", id).Error
}
