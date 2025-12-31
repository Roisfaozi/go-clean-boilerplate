package fixtures

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleFactory struct {
	db *gorm.DB
}

func NewRoleFactory(db *gorm.DB) *RoleFactory {
	return &RoleFactory{db: db}
}

func (f *RoleFactory) Create(overrides ...func(*entity.Role)) *entity.Role {
	uniqueID := uuid.New().String()[:8]

	role := &entity.Role{
		ID:          "role:" + uniqueID,
		Name:        "Test Role " + uniqueID,
		Description: "Test role description",
	}

	for _, override := range overrides {
		override(role)
	}

	f.db.Create(role)
	return role
}

func (f *RoleFactory) CreateWithID(id string) *entity.Role {
	return f.Create(func(r *entity.Role) {
		r.ID = id
		r.Name = id
	})
}

func (f *RoleFactory) CreateMany(count int) []*entity.Role {
	roles := make([]*entity.Role, count)
	for i := 0; i < count; i++ {
		roles[i] = f.Create()
	}
	return roles
}
