package test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// TestRoleUseCase_Edge_MaxNameLength tests that extremely long role names are handled (either rejected or processed).
// Validation is usually done in the handler/model level, but UseCase should handle it if passed.
func TestRoleUseCase_Edge_MaxNameLength(t *testing.T) {
	deps, uc := setupRoleTest()
	longName := strings.Repeat("r", 255)
	req := &model.CreateRoleRequest{Name: longName, Description: "Valid description"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)

	// Assume repository handles it or DB constraint
	deps.Repo.On("FindByName", mock.Anything, longName).Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
	deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(r interface{}) bool {
		// Just ensure it reaches here
		return true
	})).Return(nil)

	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	if res != nil {
		assert.Equal(t, longName, res.Name)
	}
}

// TestRoleUseCase_Vulnerability_SQLInjectionInName tests that SQL injection strings are treated literally.
func TestRoleUseCase_Vulnerability_SQLInjectionInName(t *testing.T) {
	deps, uc := setupRoleTest()
	sqliName := "role'; DROP TABLE roles; --"
	req := &model.CreateRoleRequest{Name: sqliName}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)

	deps.Repo.On("FindByName", mock.Anything, sqliName).Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
	deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(r interface{}) bool {
		return true
	})).Return(nil)

	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	if res != nil {
		assert.Equal(t, sqliName, res.Name)
	}
}

// TestRoleUseCase_Negative_CreateWithEmptyName should ideally be caught by validation,
// but if it reaches UseCase, DB might reject it or we might allow it (business rule dependent).
// Assuming 'Name' is required.
func TestRoleUseCase_Negative_CreateWithEmptyName(t *testing.T) {
	deps, uc := setupRoleTest()
	req := &model.CreateRoleRequest{Name: ""}

	// If UseCase relies on Handler validation, it might proceed to Repo.
	// Let's assume Repo returns error for empty name if it violates DB constraint.

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(exception.ErrInternalServer)

	deps.Repo.On("FindByName", mock.Anything, "").Return(nil, errors.New("record not found"))
	deps.Repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db constraint violation"))

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
}
