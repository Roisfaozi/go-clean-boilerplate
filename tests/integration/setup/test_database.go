package setup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

const migrationsDir = "../../../db/migrations"

// migrationFiles returns the sorted absolute paths of migration files of the
// given direction ("up" or "down"), optionally capped at maxVersion (>0).
func migrationFiles(t *testing.T, maxVersion int, direction string) []string {
	entries, err := os.ReadDir(migrationsDir)
	require.NoError(t, err)

	suffix := "." + direction + ".sql"
	var paths []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		var version int
		_, err := fmt.Sscanf(name, "%d_", &version)
		if err != nil || version == 0 {
			continue
		}
		if maxVersion > 0 && version > maxVersion {
			continue
		}
		paths = append(paths, filepath.Join(migrationsDir, name))
	}
	sort.Strings(paths)
	return paths
}

// FreshMigrationDB creates a dedicated database on the same MySQL server for
// applying raw SQL migrations in isolation from the shared schema.
func FreshMigrationDB(t *testing.T, name string) *gorm.DB {
	require.NotNil(t, globalDB, "shared DB must be initialized first")

	// The testcontainers MySQL module mirrors the password to MYSQL_ROOT_PASSWORD,
	// so root uses the same password as the app user.
	rootDSN := strings.Replace(strings.SplitN(mysqlAddr, "?", 2)[0], "test:test@tcp", "root:test@tcp", 1)
	rootDB, err := connectWithRetry(rootDSN, 5)
	require.NoError(t, err)

	require.NoError(t, rootDB.Exec("DROP DATABASE IF EXISTS `"+name+"`").Error)
	require.NoError(t, rootDB.Exec("CREATE DATABASE `"+name+"`").Error)
	require.NoError(t, rootDB.Exec("GRANT ALL PRIVILEGES ON `"+name+"`.* TO 'test'@'%'").Error)
	require.NoError(t, rootDB.Exec("FLUSH PRIVILEGES").Error)

	sqlDB, err := rootDB.DB()
	require.NoError(t, err)
	_ = sqlDB.Close()

	base := strings.SplitN(mysqlAddr, "?", 2)[0]
	base = strings.TrimSuffix(base, "test_db")
	dsn := base + name + "?parseTime=true&multiStatements=true&charset=utf8mb4"

	db, err := connectWithRetry(dsn, 5)
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		rootDB, err := connectWithRetry(rootDSN, 3)
		if err == nil {
			_ = rootDB.Exec("DROP DATABASE IF EXISTS `" + name + "`").Error
			sqlDB, _ := rootDB.DB()
			if sqlDB != nil {
				_ = sqlDB.Close()
			}
		}
	})

	return db
}

// ApplyMigrationsUpTo applies every up migration with version <= maxVersion.
func ApplyMigrationsUpTo(t *testing.T, db *gorm.DB, maxVersion int) {
	for _, f := range migrationFiles(t, maxVersion, "up") {
		content, err := os.ReadFile(f)
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(content)).Error, "failed to apply migration %s", f)
	}
}

// ApplyMigrationFile applies exactly one migration file (direction: "up"/"down").
func ApplyMigrationFile(t *testing.T, db *gorm.DB, version int, direction string) {
	for _, f := range migrationFiles(t, 0, direction) {
		var v int
		if _, err := fmt.Sscanf(filepath.Base(f), "%d_", &v); err != nil || v != version {
			continue
		}
		content, err := os.ReadFile(f)
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(content)).Error, "failed to apply migration %s", f)
		return
	}
	require.FailNow(t, "migration file %06d.%s.sql not found", version, direction)
}

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
		&userEntity.UserSSOIdentity{},
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
	globalOrgRecord := orgEntity.Organization{
		ID:      globalOrg,
		Name:    "Global Organization",
		Slug:    "global",
		OwnerID: "system",
		Status:  orgEntity.OrgStatusActive,
	}
	db.FirstOrCreate(&globalOrgRecord, orgEntity.Organization{ID: globalOrg})

	roles := []roleEntity.Role{
		{ID: uuid.NewString(), Name: "role:superadmin", OrganizationID: &globalOrg, Description: "Super Administrator role"},
		{ID: uuid.NewString(), Name: "role:admin", OrganizationID: &globalOrg, Description: "Administrator role"},
		{ID: uuid.NewString(), Name: "role:user", OrganizationID: &globalOrg, Description: "Regular user role"},
		{ID: uuid.NewString(), Name: "role:org-owner", OrganizationID: &globalOrg, Description: "Organization owner role"},
		{ID: uuid.NewString(), Name: "role:moderator", OrganizationID: &globalOrg, Description: "Moderator role"},
	}

	for _, role := range roles {
		db.FirstOrCreate(&role, roleEntity.Role{Name: role.Name})
	}

	policies := [][]string{
		{"role:user", "global", "/api/v1/users/me", "GET"},
		{"role:user", "global", "/api/v1/users/me", "PUT"},
		{"role:user", "global", "/api/v1/auth/logout", "POST"},
		{"role:user", "global", "/api/v1/organizations/:id", "GET"},
		{"role:user", "global", "/api/v1/organizations/slug/:slug", "GET"},
		{"role:user", "global", "/api/v1/organizations/:id/presence", "GET"},
		{"role:user", "global", "/api/v1/projects", "GET"},
		{"role:user", "global", "/api/v1/projects/:id", "GET"},
		{"role:admin", "global", "/api/v1/organizations/:id", "GET"},
		{"role:admin", "global", "/api/v1/organizations/slug/:slug", "GET"},
		{"role:admin", "global", "/api/v1/organizations/:id", "PUT"},
		{"role:admin", "global", "/api/v1/organizations/:id", "DELETE"},
		{"role:admin", "global", "/api/v1/organizations/:id/members/invite", "POST"},
		{"role:admin", "global", "/api/v1/organizations/:id/members", "GET"},
		{"role:admin", "global", "/api/v1/organizations/:id/members/:userId", "PATCH"},
		{"role:admin", "global", "/api/v1/organizations/:id/members/:userId", "DELETE"},
		{"role:admin", "global", "/api/v1/organizations/:id/presence", "GET"},
		{"role:admin", "global", "/api/v1/organizations/:id/roles", "GET"},
		{"role:admin", "global", "/api/v1/organizations/:id/roles", "POST"},
		{"role:admin", "global", "/api/v1/organizations/:id/roles/:roleId", "PUT"},
		{"role:admin", "global", "/api/v1/organizations/:id/roles/:roleId", "DELETE"},
		{"role:user", "global", "/api/v1/organizations/:id/roles", "GET"},
		{"role:admin", "global", "/api/v1/projects", "GET"},
		{"role:admin", "global", "/api/v1/projects/:id", "GET"},
		{"role:admin", "global", "/api/v1/projects", "POST"},
		{"role:admin", "global", "/api/v1/projects/:id", "PUT"},
		{"role:admin", "global", "/api/v1/projects/:id", "DELETE"},
		// Superadmin permissions for E2E
		{"role:superadmin", "global", "*", "*"},
		{"role:superadmin", "global", "/api/v1/webhooks", "POST"},
		{"role:superadmin", "global", "/api/v1/webhooks", "GET"},
		{"role:superadmin", "global", "/api/v1/webhooks/:id", "GET"},
		{"role:superadmin", "global", "/api/v1/webhooks/:id", "PUT"},
		{"role:superadmin", "global", "/api/v1/webhooks/:id", "DELETE"},
		{"role:superadmin", "global", "/api/v1/webhooks/:id/logs", "GET"},
		// API Keys permissions
		{"role:superadmin", "global", "/api/v1/api-keys", "POST"},
		{"role:superadmin", "global", "/api/v1/api-keys", "GET"},
		{"role:superadmin", "global", "/api/v1/api-keys/:id", "DELETE"},
	}

	for _, p := range policies {
		db.Exec("INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES (?, ?, ?, ?, ?)", "p", p[0], p[1], p[2], p[3])
	}
}

func RoleIDByName(t *testing.T, db *gorm.DB, name string) string {
	var r roleEntity.Role
	err := db.Where("name = ?", name).First(&r).Error
	require.NoError(t, err)
	return r.ID
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
		"user_sso_identities",
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

func CreateTestOrganization(t *testing.T, db *gorm.DB, ownerID, name, slug string) *orgEntity.Organization {
	org := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    name,
		Slug:    slug,
		OwnerID: ownerID,
		Status:  orgEntity.OrgStatusActive,
	}

	err := db.Create(org).Error
	require.NoError(t, err, "Failed to create test organization")

	return org
}

func CreateTestRole(t *testing.T, db *gorm.DB, name string) *roleEntity.Role {
	globalOrg := "global"
	db.FirstOrCreate(&orgEntity.Organization{}, orgEntity.Organization{
		ID:      globalOrg,
		Name:    "Global Organization",
		Slug:    "global",
		OwnerID: "system",
		Status:  orgEntity.OrgStatusActive,
	})

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

func HashSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
