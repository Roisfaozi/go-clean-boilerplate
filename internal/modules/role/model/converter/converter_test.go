package converter_test

import (
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model/converter"
	"github.com/stretchr/testify/assert"
)

func TestRoleToResponse(t *testing.T) {
	now := time.Now().Unix()
	role := &entity.Role{
		ID:          "role-1",
		Name:        "Admin",
		Description: "Administrator",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	res := converter.RoleToResponse(role)

	assert.NotNil(t, res)
	assert.Equal(t, role.ID, res.ID)
	assert.Equal(t, role.Name, res.Name)
	assert.Equal(t, role.Description, res.Description)
}

func TestRolesToResponse(t *testing.T) {
	now := time.Now().Unix()
	roles := []*entity.Role{
		{
			ID:          "role-1",
			Name:        "Admin",
			Description: "Administrator",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "role-2",
			Name:        "User",
			Description: "Standard User",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	res := converter.RolesToResponse(roles)

	assert.NotNil(t, res)
	assert.Len(t, res, 2)
	assert.Equal(t, roles[0].ID, res[0].ID)
	assert.Equal(t, roles[1].ID, res[1].ID)
}

func TestRolesToResponse_Empty(t *testing.T) {
	var roles []*entity.Role

	res := converter.RolesToResponse(roles)

	assert.Nil(t, res)
	assert.Len(t, res, 0)
}
