package test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	roleMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	txMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type roleTestDeps struct {
	Repo *roleMocks.MockRoleRepository
	TM   *txMocks.MockWithTransactionManager
}

func setupRoleTest() (*roleTestDeps, usecase.RoleUseCase) {
	deps := &roleTestDeps{
		Repo: new(roleMocks.MockRoleRepository),
		TM:   new(txMocks.MockWithTransactionManager),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo)
	return deps, uc
}

func TestRoleUseCase_Create_Success(t *testing.T) {
	deps, uc := setupRoleTest()
	req := &model.CreateRoleRequest{Name: "new-role", Description: "New Role"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)

		deps.Repo.On("FindByName", mock.Anything, "new-role").Return(nil, gorm.ErrRecordNotFound)
		deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(r *entity.Role) bool {
			return r.Name == "new-role" && r.Description == "New Role"
		})).Return(nil)

		_ = fn(context.Background())
	})

	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "new-role", res.Name)
	deps.Repo.AssertExpectations(t)
}

func TestRoleUseCase_Create_Conflict(t *testing.T) {
	deps, uc := setupRoleTest()
	req := &model.CreateRoleRequest{Name: "existing-role"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(exception.ErrConflict).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		deps.Repo.On("FindByName", mock.Anything, "existing-role").Return(&entity.Role{}, nil)
		_ = fn(context.Background())
	})

	res, err := uc.Create(context.Background(), req)

	assert.ErrorIs(t, err, exception.ErrConflict)
	assert.Nil(t, res)
}

func TestRoleUseCase_GetAll_Success(t *testing.T) {
	deps, uc := setupRoleTest()
	roles := []*entity.Role{
		{ID: "1", Name: "r1"},
		{ID: "2", Name: "r2"},
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		deps.Repo.On("FindAll", mock.Anything).Return(roles, nil)
		_ = fn(context.Background())
	})

	res, err := uc.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestRoleUseCase_Delete_Success(t *testing.T) {
	deps, uc := setupRoleTest()
	id := "r1"

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)

		deps.Repo.On("FindByID", mock.Anything, id).Return(&entity.Role{ID: id, Name: "role:user"}, nil)
		deps.Repo.On("Delete", mock.Anything, id).Return(nil)

		_ = fn(context.Background())
	})

	err := uc.Delete(context.Background(), id)

	assert.NoError(t, err)
}

func TestRoleUseCase_Delete_NotFound(t *testing.T) {
	deps, uc := setupRoleTest()
	id := "unknown"

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(exception.ErrNotFound).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		deps.Repo.On("FindByID", mock.Anything, id).Return(nil, gorm.ErrRecordNotFound)
		_ = fn(context.Background())
	})

	err := uc.Delete(context.Background(), id)

	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestRoleUseCase_Delete_Superadmin(t *testing.T) {
	deps, uc := setupRoleTest()
	id := "sa"

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(exception.ErrForbidden).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		deps.Repo.On("FindByID", mock.Anything, id).Return(&entity.Role{ID: id, Name: "role:superadmin"}, nil)
		_ = fn(context.Background())
	})

	err := uc.Delete(context.Background(), id)

	assert.ErrorIs(t, err, exception.ErrForbidden)
}

func TestRoleUseCase_Delete_InternalError(t *testing.T) {
	deps, uc := setupRoleTest()
	id := "r1"

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(exception.ErrInternalServer).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)

		deps.Repo.On("FindByID", mock.Anything, id).Return(&entity.Role{ID: id, Name: "role:user"}, nil)
		deps.Repo.On("Delete", mock.Anything, id).Return(errors.New("db error"))

		_ = fn(context.Background())
	})

	err := uc.Delete(context.Background(), id)

	assert.ErrorIs(t, err, exception.ErrInternalServer)
}
