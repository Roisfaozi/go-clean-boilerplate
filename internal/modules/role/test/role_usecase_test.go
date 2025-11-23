package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/casbin-db/internal/mocking"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestRoleUseCase_Create_Success(t *testing.T) {
	mockRepo := new(mocks.RoleRepository)
	mockTM := new(mocking.MockTransactionManager)

	tests := []struct {
		name     string
		req      *model.CreateRoleRequest
		mockFunc func()
	}{
		{
			name: "Success - Basic Role Creation",
			req:  &model.CreateRoleRequest{Name: "new_role", Description: "A new role for testing"},
			mockFunc: func() {
				mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					})

				mockRepo.On("FindByName", mock.Anything, "new_role").Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *entity.Role) bool {
					return r.Name == "new_role" && r.Description == "A new role for testing"
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			defer func() {
				mockRepo.ExpectedCalls = nil
				mockTM.ExpectedCalls = nil
			}()

			uc := usecase.NewRoleUseCase(logrus.New(), mockTM, mockRepo)
			res, err := uc.Create(context.Background(), tt.req)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.Name, res.Name)
			mockRepo.AssertExpectations(t)
			mockTM.AssertExpectations(t)
		})
	}
}

func TestRoleUseCase_Create_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		req         *model.CreateRoleRequest
		setupMocks  func(*mocks.RoleRepository, *mocking.MockTransactionManager)
		expectedErr error
	}{
		{
			name: "Error - Role Already Exists",
			req:  &model.CreateRoleRequest{Name: "existing_role"},
			setupMocks: func(r *mocks.RoleRepository, tm *mocking.MockTransactionManager) {
				tm.On("WithinTransaction", mock.Anything, mock.Anything).
					Return(exception.ErrConflict).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						err := fn(context.Background())
						assert.Error(t, err)
					})
				r.On("FindByName", mock.Anything, "existing_role").
					Return(&entity.Role{Name: "existing_role"}, nil)
			},
			expectedErr: exception.ErrConflict,
		},
		{
			name: "Error - Database Error on FindByName",
			req:  &model.CreateRoleRequest{Name: "db_error_role"},
			setupMocks: func(r *mocks.RoleRepository, tm *mocking.MockTransactionManager) {
				tm.On("WithinTransaction", mock.Anything, mock.Anything).
					Return(exception.ErrInternalServer).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						err := fn(context.Background())
						assert.Error(t, err)
					})
				r.On("FindByName", mock.Anything, "db_error_role").
					Return(nil, errors.New("database error"))
			},
			expectedErr: exception.ErrInternalServer,
		},
		{
			name: "Error - Database Error on Create",
			req:  &model.CreateRoleRequest{Name: "create_error_role"},
			setupMocks: func(r *mocks.RoleRepository, tm *mocking.MockTransactionManager) {
				tm.On("WithinTransaction", mock.Anything, mock.Anything).
					Return(exception.ErrInternalServer).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						err := fn(context.Background())
						assert.Error(t, err)
					})
				r.On("FindByName", mock.Anything, "create_error_role").
					Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
				r.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("database error"))
			},
			expectedErr: exception.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.RoleRepository)
			mockTM := new(mocking.MockTransactionManager)

			tt.setupMocks(mockRepo, mockTM)

			uc := usecase.NewRoleUseCase(logrus.New(), mockTM, mockRepo)
			res, err := uc.Create(context.Background(), tt.req)

			assert.Error(t, err)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tt.expectedErr)

			mockRepo.AssertExpectations(t)
			mockTM.AssertExpectations(t)
		})
	}
}

func TestRoleUseCase_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mocks.RoleRepository, *mocking.MockTransactionManager)
		expectedLen int
		expectError bool
	}{
		{
			name: "Success - Get All Roles",
			setupMocks: func(r *mocks.RoleRepository, tm *mocking.MockTransactionManager) {
				tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					})

				r.On("FindAll", mock.Anything).
					Return([]*entity.Role{
						{ID: "1", Name: "admin"},
						{ID: "2", Name: "user"},
					}, nil)
			},
			expectedLen: 2,
			expectError: false,
		},
		{
			name: "Success - No Roles Found",
			setupMocks: func(r *mocks.RoleRepository, tm *mocking.MockTransactionManager) {
				tm.On("WithinTransaction", mock.Anything, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					})

				r.On("FindAll", mock.Anything).Return([]*entity.Role{}, nil)
			},
			expectedLen: 0,
			expectError: false,
		},
		{
			name: "Error - Database Error",
			setupMocks: func(r *mocks.RoleRepository, tm *mocking.MockTransactionManager) {
				tm.On("WithinTransaction", mock.Anything, mock.Anything).
					Return(exception.ErrInternalServer).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						err := fn(context.Background())
						assert.Error(t, err)
					})

				r.On("FindAll", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.RoleRepository)
			mockTM := new(mocking.MockTransactionManager)

			tt.setupMocks(mockRepo, mockTM)

			uc := usecase.NewRoleUseCase(logrus.New(), mockTM, mockRepo)
			result, err := uc.GetAll(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedLen)
			}

			mockRepo.AssertExpectations(t)
			mockTM.AssertExpectations(t)
		})
	}
}

func TestRoleUseCase_ContextCancellation(t *testing.T) {
	tests := []struct {
		name        string
		req         *model.CreateRoleRequest
		setupMocks  func(*mocks.RoleRepository, *mocking.MockTransactionManager)
		expectError bool
	}{
		{
			name: "Context Cancelled",
			req:  &model.CreateRoleRequest{Name: "test_role", Description: "Test"},
			setupMocks: func(r *mocks.RoleRepository, tm *mocking.MockTransactionManager) {
				tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(context.Canceled)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.RoleRepository)
			mockTM := new(mocking.MockTransactionManager)

			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo, mockTM)
			}

			uc := usecase.NewRoleUseCase(logrus.New(), mockTM, mockRepo)
			result, err := uc.Create(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			}

			mockTM.AssertExpectations(t)
		})
	}
}
