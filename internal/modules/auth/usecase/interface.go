package usecase

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
)

type AuthUseCase interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	ValidateRefreshToken(token string) (Claims, error)
}
