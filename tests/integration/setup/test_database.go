package setup

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	apiKeyEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/entity"
	auditEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	projectEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	webhookEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RunMigrations(t *testing.T, db *gorm.DB) {
	err := db.AutoMigrate(
		&userEntity.User{},
		&roleEntity.Role{},
		&entity.Endpoint{},
		&entity.AccessRight{},
		&auditEntity.AuditLog{},
		&auditEntity.AuditOutbox{},
		&authEntity.PasswordResetToken{},
		&authEntity.EmailVerificationToken{},
		&orgEntity.Organization{},
		&orgEntity.OrganizationMember{},
		&orgEntity.InvitationToken{},
		&projectEntity.Project{},
		&apiKeyEntity.ApiKey{},
		&webhookEntity.Webhook{},
		&webhookEntity.WebhookLog{},
	)
	if t != nil {
		require.NoError(t, err, "Failed to run migrations")
	} else if err != nil {
		panic("Failed to run migrations: " + err.Error())
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS casbin_rule (
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		ptype varchar(100) DEFAULT NULL,
		v0 varchar(100) DEFAULT NULL,
		v1 varchar(100) DEFAULT NULL,
		v2 varchar(100) DEFAULT NULL,
		v3 varchar(100) DEFAULT NULL,
		v4 varchar(100) DEFAULT NULL,
		v5 varchar(100) DEFAULT NULL,
		PRIMARY KEY (id),
		UNIQUE KEY idx_casbin_rule (ptype,v0,v1,v2,v3,v4,v5)
	)`)
}

func SeedTestData(t *testing.T, db *gorm.DB) {
	globalOrg := "global"
	roles := []roleEntity.Role{
		{ID: "role:superadmin", Name: "role:superadmin", OrganizationID: &globalOrg, Description: "Super Administrator role"},
		{ID: "role:admin", Name: "role:admin", OrganizationID: &globalOrg, Description: "Administrator role"},
		{ID: "role:user", Name: "role:user", OrganizationID: &globalOrg, Description: "Regular user role"},
		{ID: "role:moderator", Name: "role:moderator", OrganizationID: &globalOrg, Description: "Moderator role"},
	}

	for _, role := range roles {
		db.FirstOrCreate(&role, roleEntity.Role{ID: role.ID})
	}

	policies := [][]string{
		{"role:user", "global", "/api/v1/users/me", "GET"},
		{"role:user", "global", "/api/v1/users/me", "PUT"},
		{"role:user", "global", "/api/v1/auth/logout", "POST"},
		// Superadmin permissions for E2E
		{"role:superadmin", "global", "/api/v1/webhooks", "POST"},
		{"role:superadmin", "global", "/api/v1/webhooks", "GET"},
		{"role:superadmin", "global", "/api/v1/webhooks/:id", "GET"},
		{"role:superadmin", "global", "/api/v1/webhooks/:id", "PUT"},
		{"role:superadmin", "global", "/api/v1/webhooks/:id", "DELETE"},
		{"role:superadmin", "global", "/api/v1/webhooks/:id/logs", "GET"},
	}

	for _, p := range policies {
		db.Exec("INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES (?, ?, ?, ?, ?)", "p", p[0], p[1], p[2], p[3])
	}
}

func CleanupDatabase(t *testing.T, db *gorm.DB) {
	tables := []string{
		"projects",
		"organization_members",
		"organizations",
		"audit_logs",
		"audit_outbox",
		"access_rights",
		"endpoints",
		"casbin_rule",
		"users",
		"roles",
		"password_reset_tokens",
		"email_verification_tokens",
		"invitation_tokens",
		"api_keys",
		"webhooks",
		"webhook_logs",
	}

	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	for _, table := range tables {
		db.Exec("TRUNCATE TABLE " + table)
	}
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
}

func CreateTestUser(t *testing.T, db *gorm.DB, username, email, password string, orgIDs ...string) *userEntity.User {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "Failed to hash password")

	user := &userEntity.User{
		ID:       uuid.New().String(),
		Username: username,
		Email:    email,
		Name:     username,
		Password: string(hashedPassword),
	}

	err = db.Create(user).Error
	require.NoError(t, err, "Failed to create test user")

	// Add memberships if orgIDs provided
	for _, orgID := range orgIDs {
		member := &orgEntity.OrganizationMember{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			UserID:         user.ID,
			RoleID:         "role:user",
			Status:         "active",
		}
		err = db.Create(member).Error
		require.NoError(t, err, "Failed to create member record")
	}

	return user
}

func CreateTestRole(t *testing.T, db *gorm.DB, name string) *roleEntity.Role {
	globalOrg := "global"
	role := &roleEntity.Role{
		ID:             uuid.New().String(),
		Name:           name,
		OrganizationID: &globalOrg,
		Description:    "Test role " + name,
	}

	err := db.Create(role).Error
	if t != nil {
		require.NoError(t, err, "Failed to create test role")
	}

	return role
}
