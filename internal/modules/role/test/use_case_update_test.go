package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestRoleUseCase_Update(t *testing.T) {
	roleID := "role-123"

	t.Run("Success - Role Updated", func(t *testing.T) {
		deps, uc := setupRoleTest()
		req := &model.UpdateRoleRequest{Description: "Updated Description"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			})

		deps.Repo.On("FindByID", mock.Anything, roleID).Return(&entity.Role{ID: roleID, Name: "role_name", Description: "Old Desc"}, nil)
		deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(r *entity.Role) bool {
			return r.ID == roleID && r.Description == "Updated Description"
		})).Return(nil)

		res, err := uc.Update(context.Background(), roleID, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Updated Description", res.Description)
		deps.Repo.AssertExpectations(t)
		deps.TM.AssertExpectations(t)
	})

	t.Run("Error - Role Not Found", func(t *testing.T) {
		deps, uc := setupRoleTest()
		req := &model.UpdateRoleRequest{Description: "Updated Description"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrNotFound).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrNotFound, err)
			})

		deps.Repo.On("FindByID", mock.Anything, roleID).Return((*entity.Role)(nil), gorm.ErrRecordNotFound)

		res, err := uc.Update(context.Background(), roleID, req)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
		deps.TM.AssertExpectations(t)
	})

	t.Run("Error - Database Error During Update", func(t *testing.T) {
		deps, uc := setupRoleTest()
		req := &model.UpdateRoleRequest{Description: "Updated Description"}
		dbError := errors.New("database error")

		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrInternalServer).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrInternalServer, err)
			})

		deps.Repo.On("FindByID", mock.Anything, roleID).Return(&entity.Role{ID: roleID, Name: "role_name"}, nil)
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(dbError)

		res, err := uc.Update(context.Background(), roleID, req)

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
		deps.TM.AssertExpectations(t)
	})
}
