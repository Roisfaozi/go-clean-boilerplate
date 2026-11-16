package fixtures

import (
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RoleFactory struct {
	db *sqlx.DB
}

func NewRoleFactory(db *sqlx.DB) *RoleFactory {
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

	if role.ID == "" {
		role.ID = uuid.New().String()
	}
	now := time.Now().UTC().UnixMilli()
	if role.CreatedAt == 0 {
		role.CreatedAt = now
	}
	if role.UpdatedAt == 0 {
		role.UpdatedAt = now
	}

	_, err := f.db.Exec(
		`INSERT INTO roles (id, organization_id, name, description, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		role.ID, role.OrganizationID, role.Name, role.Description, role.CreatedAt, role.UpdatedAt, role.DeletedAt,
	)
	if err != nil {
		panic(err)
	}
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
