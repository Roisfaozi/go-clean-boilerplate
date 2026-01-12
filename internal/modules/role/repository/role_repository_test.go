package repository_test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	querybuilder2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupRoleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&entity.Role{})
	assert.NoError(t, err)
	return db
}

func TestRoleRepository_Create(t *testing.T) {
	db := setupRoleTestDB(t)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewRoleRepository(db, logger)

	role := &entity.Role{
		ID:          "role-123",
		Name:        "admin",
		Description: "Administrator",
	}

	err := repo.Create(context.Background(), role)
	assert.NoError(t, err)

	var stored entity.Role
	err = db.First(&stored, "id = ?", role.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, role.Name, stored.Name)
}

func TestRoleRepository_FindByID(t *testing.T) {
	db := setupRoleTestDB(t)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewRoleRepository(db, logger)

	role := &entity.Role{
		ID:          "role-123",
		Name:        "admin",
		Description: "Administrator",
	}
	db.Create(role)

	found, err := repo.FindByID(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, role.Name, found.Name)

	found, err = repo.FindByID(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestRoleRepository_FindByName(t *testing.T) {
	db := setupRoleTestDB(t)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewRoleRepository(db, logger)

	role := &entity.Role{
		ID:          "role-123",
		Name:        "admin",
		Description: "Administrator",
	}
	db.Create(role)

	found, err := repo.FindByName(context.Background(), role.Name)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, role.ID, found.ID)

	found, err = repo.FindByName(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestRoleRepository_FindAll(t *testing.T) {
	db := setupRoleTestDB(t)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewRoleRepository(db, logger)

	role1 := &entity.Role{ID: "r1", Name: "admin"}
	role2 := &entity.Role{ID: "r2", Name: "user"}
	db.Create(role1)
	db.Create(role2)

	roles, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
}

func TestRoleRepository_Delete(t *testing.T) {
	db := setupRoleTestDB(t)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewRoleRepository(db, logger)

	role := &entity.Role{ID: "r1", Name: "admin"}
	db.Create(role)

	err := repo.Delete(context.Background(), role.ID)
	assert.NoError(t, err)

	var stored entity.Role
	err = db.First(&stored, "id = ?", role.ID).Error
	assert.Error(t, err)
}

func TestRoleRepository_FindAllDynamic(t *testing.T) {
	db := setupRoleTestDB(t)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewRoleRepository(db, logger)

	role1 := &entity.Role{ID: "r1", Name: "admin"}
	role2 := &entity.Role{ID: "r2", Name: "user"}
	db.Create(role1)
	db.Create(role2)

	// Filter by name
	filter := &querybuilder2.DynamicFilter{
		Filter: querybuilder2.Filter{
			Field:    "name",
			Operator: "eq",
			Value:    "admin",
		},
	}
	roles, err := repo.FindAllDynamic(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, "admin", roles[0].Name)

	// Sort desc
	filter = &querybuilder2.DynamicFilter{
		Sort: []querybuilder2.SortModel{
			{ColId: "name", Sort: "desc"},
		},
	}
	roles, err = repo.FindAllDynamic(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Equal(t, "user", roles[0].Name)
}

type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) { return 0, nil }
func (w *NoOpWriter) Levels() []logrus.Level    { return logrus.AllLevels }
