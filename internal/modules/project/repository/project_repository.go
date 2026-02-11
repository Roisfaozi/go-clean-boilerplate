package repository

import (
	"context"
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	GetByID(ctx context.Context, id string) (*entity.Project, error)
	GetByOrgID(ctx context.Context, orgID string) ([]*entity.Project, error)
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, id string) error
	CountByUserID(ctx context.Context, userID string) (int64, error)
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := tx.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *projectRepository) Create(ctx context.Context, project *entity.Project) error {
	return r.getDB(ctx).Create(project).Error
}

func (r *projectRepository) GetByID(ctx context.Context, id string) (*entity.Project, error) {
	var project entity.Project
	if err := r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).First(&project, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) GetByOrgID(ctx context.Context, orgID string) ([]*entity.Project, error) {
	var projects []*entity.Project
	if err := r.getDB(ctx).Where("organization_id = ?", orgID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) Update(ctx context.Context, project *entity.Project) error {
	return r.getDB(ctx).Save(project).Error
}

func (r *projectRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).Delete(&entity.Project{}, "id = ?", id).Error
}

func (r *projectRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&entity.Project{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
