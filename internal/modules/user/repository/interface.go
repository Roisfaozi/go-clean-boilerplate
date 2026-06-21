package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	UpdateStatus(ctx context.Context, userID, status string) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByToken(ctx context.Context, token string) (*entity.User, error)
	Delete(ctx context.Context, id string) error
	FindAll(ctx context.Context, filter *model.GetUserListRequest) ([]*entity.User, int64, error)
	FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.User, int64, error)
	HardDeleteSoftDeletedUsers(ctx context.Context, retentionDays int) error
	GetByOrganization(ctx context.Context, orgID string) ([]*entity.User, error)
	FindBySSOIdentity(ctx context.Context, provider, providerID string) (*entity.UserSSOIdentity, error)
	CreateSSOIdentity(ctx context.Context, identity *entity.UserSSOIdentity) error
}
