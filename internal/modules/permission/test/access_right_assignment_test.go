package test

import (
	"context"
	"errors"
	"testing"

	accessEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRoleAccessRights_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	accessRights := []*accessEntity.AccessRight{
		{
			ID:   "ar1",
			Name: "Role Management",
			Endpoints: []accessEntity.Endpoint{
				{Method: "GET", Path: "/api/roles"},
				{Method: "POST", Path: "/api/roles"},
			},
		},
		{
			ID:   "ar2",
			Name: "User Management",
			Endpoints: []accessEntity.Endpoint{
				{Method: "GET", Path: "/api/users"},
				{Method: "DELETE", Path: "/api/users/:id"},
			},
		},
		{
			ID:   "ar3",
			Name: "Stats",
			Endpoints: []accessEntity.Endpoint{
				{Method: "GET", Path: "/api/stats"},
			},
		},
	}

	deps.AccessRepo.On("GetAccessRights", mock.Anything).Return(accessRights, nil)

	deps.Enforcer.On("Enforce", "admin", "global", "/api/roles", "GET").Return(true, nil)
	deps.Enforcer.On("Enforce", "admin", "global", "/api/roles", "POST").Return(true, nil)

	deps.Enforcer.On("Enforce", "admin", "global", "/api/users", "GET").Return(true, nil)
	deps.Enforcer.On("Enforce", "admin", "global", "/api/users/:id", "DELETE").Return(false, nil)

	deps.Enforcer.On("Enforce", "admin", "global", "/api/stats", "GET").Return(false, nil)

	res, err := uc.GetRoleAccessRights(context.Background(), "admin", "")

	assert.NoError(t, err)
	assert.Len(t, res, 3)

	assert.Equal(t, "Role Management", res[0].Name)
	assert.True(t, res[0].Assigned)
	assert.False(t, res[0].Partial)

	assert.Equal(t, "User Management", res[1].Name)
	assert.False(t, res[1].Assigned)
	assert.True(t, res[1].Partial)

	assert.Equal(t, "Stats", res[2].Name)
	assert.False(t, res[2].Assigned)
	assert.False(t, res[2].Partial)
}

func TestGetRoleAccessRights_RepoError(t *testing.T) {
	deps, uc := setupPermissionTest()

	deps.AccessRepo.On("GetAccessRights", mock.Anything).Return(nil, errors.New("db error"))

	res, err := uc.GetRoleAccessRights(context.Background(), "admin", "")

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestAssignAccessRight_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	ar := &accessEntity.AccessRight{
		ID:   "ar1",
		Name: "Role Management",
		Endpoints: []accessEntity.Endpoint{
			{Method: "GET", Path: "/api/roles"},
			{Method: "POST", Path: "/api/roles"},
		},
	}

	deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(ar, nil)

	// One already granted, one not granted
	deps.Enforcer.On("Enforce", "admin", "global", "/api/roles", "GET").Return(true, nil)
	deps.Enforcer.On("Enforce", "admin", "global", "/api/roles", "POST").Return(false, nil)
	deps.Enforcer.On("AddPolicy", "admin", "global", "/api/roles", "POST").Return(true, nil)

	req := model.AssignAccessRightRequest{
		AccessRightID: "ar1",
		Role:          "admin",
	}

	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	err := uc.AssignAccessRight(context.Background(), req)

	assert.NoError(t, err)
}

func TestAssignAccessRight_NotFound(t *testing.T) {
	deps, uc := setupPermissionTest()

	deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(nil, exception.ErrNotFound)

	req := model.AssignAccessRightRequest{
		AccessRightID: "ar1",
		Role:          "admin",
	}

	err := uc.AssignAccessRight(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)
}

func TestAssignAccessRight_NoEndpoints(t *testing.T) {
	deps, uc := setupPermissionTest()

	ar := &accessEntity.AccessRight{
		ID:        "ar1",
		Name:      "Empty Access Right",
		Endpoints: []accessEntity.Endpoint{},
	}

	deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(ar, nil)

	req := model.AssignAccessRightRequest{
		AccessRightID: "ar1",
		Role:          "admin",
	}

	err := uc.AssignAccessRight(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no endpoints configured")
}

func TestRevokeAccessRight_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	ar := &accessEntity.AccessRight{
		ID:   "ar1",
		Name: "Role Management",
		Endpoints: []accessEntity.Endpoint{
			{Method: "GET", Path: "/api/roles"},
			{Method: "POST", Path: "/api/roles"},
		},
	}

	deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(ar, nil)

	deps.Enforcer.On("RemovePolicy", "admin", "global", "/api/roles", "GET").Return(true, nil)
	deps.Enforcer.On("RemovePolicy", "admin", "global", "/api/roles", "POST").Return(true, nil)

	req := model.AssignAccessRightRequest{
		AccessRightID: "ar1",
		Role:          "admin",
	}

	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	err := uc.RevokeAccessRight(context.Background(), req)

	assert.NoError(t, err)
}

func TestRevokeAccessRight_NotFound(t *testing.T) {
	deps, uc := setupPermissionTest()

	deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(nil, exception.ErrNotFound)

	req := model.AssignAccessRightRequest{
		AccessRightID: "ar1",
		Role:          "admin",
	}

	err := uc.RevokeAccessRight(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)
}
