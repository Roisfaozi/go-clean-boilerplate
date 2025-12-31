package fixtures

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserFactory struct {
	db *gorm.DB
}

func NewUserFactory(db *gorm.DB) *UserFactory {
	return &UserFactory{db: db}
}

func (f *UserFactory) Create(overrides ...func(*entity.User)) *entity.User {
	uniqueID := uuid.New().String()[:8]
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &entity.User{
		ID:       uuid.New().String(),
		Username: "testuser_" + uniqueID,
		Email:    "test_" + uniqueID + "@example.com",
		Name:     "Test User",
		Password: string(hashedPassword),
	}

	for _, override := range overrides {
		override(user)
	}

	f.db.Create(user)
	return user
}

func (f *UserFactory) CreateAdmin() *entity.User {
	return f.Create(func(u *entity.User) {
		u.Username = "admin_" + uuid.New().String()[:8]
		u.Email = "admin_" + uuid.New().String()[:8] + "@example.com"
		u.Name = "Admin User"
	})
}

func (f *UserFactory) CreateWithUsername(username string) *entity.User {
	return f.Create(func(u *entity.User) {
		u.Username = username
		u.Email = username + "@example.com"
	})
}

func (f *UserFactory) CreateWithEmail(email string) *entity.User {
	return f.Create(func(u *entity.User) {
		u.Email = email
	})
}

func (f *UserFactory) CreateMany(count int) []*entity.User {
	users := make([]*entity.User, count)
	for i := 0; i < count; i++ {
		users[i] = f.Create()
	}
	return users
}
