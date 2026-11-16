package fixtures

import (
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserFactory struct {
	db *sqlx.DB
}

func NewUserFactory(db *sqlx.DB) *UserFactory {
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
		Status:   entity.UserStatusActive,
	}

	for _, override := range overrides {
		override(user)
	}

	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.Status == "" {
		user.Status = entity.UserStatusActive
	}
	now := time.Now().UTC().UnixMilli()
	if user.CreatedAt == 0 {
		user.CreatedAt = now
	}
	if user.UpdatedAt == 0 {
		user.UpdatedAt = now
	}

	_, err := f.db.Exec(
		`INSERT INTO users (id, organization_id, username, email, name, password, status, avatar_url, token, email_verified_at, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.OrganizationID, user.Username, user.Email, user.Name, user.Password,
		user.Status, user.AvatarURL, user.Token, user.EmailVerifiedAt, user.CreatedAt, user.UpdatedAt, user.DeletedAt,
	)
	if err != nil {
		panic(err)
	}
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
