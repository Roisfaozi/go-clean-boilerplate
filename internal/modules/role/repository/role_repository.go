package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	querybuilder2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
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

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) FindByID(ctx context.Context, id string) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).First(&role, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindAll(ctx context.Context) ([]*entity.Role, error) {
	var roles []*entity.Role
	if err := r.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) FindAllDynamic(ctx context.Context, filter *querybuilder2.DynamicFilter) ([]*entity.Role, error) {
	var roles []*entity.Role
	query := r.db.WithContext(ctx)

	where, args, _, err := querybuilder2.GenerateDynamicQuery[entity.Role](filter)
	if err != nil {
		return nil, err
	}

	if where != "" {
		query = query.Where(where, args...)
	}

	sort, err := querybuilder2.GenerateDynamicSort[entity.Role](filter)
	if err != nil {
		return nil, err
	}
	if sort != "" {
		query = query.Order(sort)
	}

	if err := query.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, "id = ?", id).Error
}
