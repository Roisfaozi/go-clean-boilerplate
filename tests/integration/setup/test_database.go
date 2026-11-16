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
	"time"

	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
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
func FreshMigrationDB(t *testing.T, name string) *sqlx.DB {
	require.NotNil(t, globalDB, "shared DB must be initialized first")

	// The testcontainers MySQL module mirrors the password to MYSQL_ROOT_PASSWORD,
	// so root uses the same password as the app user.
	rootDSN := strings.Replace(strings.SplitN(mysqlAddr, "?", 2)[0], "test:test@tcp", "root:test@tcp", 1)
	rootDB, err := connectWithRetry(rootDSN, 5)
	require.NoError(t, err)

	_, err = rootDB.Exec("DROP DATABASE IF EXISTS `" + name + "`")
	require.NoError(t, err)
	_, err = rootDB.Exec("CREATE DATABASE `" + name + "`")
	require.NoError(t, err)
	_, err = rootDB.Exec("GRANT ALL PRIVILEGES ON `" + name + "`.* TO 'test'@'%'")
	require.NoError(t, err)
	_, err = rootDB.Exec("FLUSH PRIVILEGES")
	require.NoError(t, err)

	_ = rootDB.Close()

	base := strings.SplitN(mysqlAddr, "?", 2)[0]
	base = strings.TrimSuffix(base, "test_db")
	dsn := base + name + "?parseTime=true&multiStatements=true&charset=utf8mb4"

	db, err := connectWithRetry(dsn, 5)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
		rootDB, err := connectWithRetry(rootDSN, 3)
		if err == nil {
			_, _ = rootDB.Exec("DROP DATABASE IF EXISTS `" + name + "`")
			_ = rootDB.Close()
		}
	})

	return db
}

// ApplyMigrationsUpTo applies every up migration with version <= maxVersion.
func ApplyMigrationsUpTo(t *testing.T, db *sqlx.DB, maxVersion int) {
	for _, f := range migrationFiles(t, maxVersion, "up") {
		content, err := os.ReadFile(f)
		require.NoError(t, err)
		_, err = db.Exec(string(content))
		require.NoError(t, err, "failed to apply migration %s", f)
	}
}

// ApplyMigrationFile applies exactly one migration file (direction: "up"/"down").
func ApplyMigrationFile(t *testing.T, db *sqlx.DB, version int, direction string) {
	for _, f := range migrationFiles(t, 0, direction) {
		var v int
		if _, err := fmt.Sscanf(filepath.Base(f), "%d_", &v); err != nil || v != version {
			continue
		}
		content, err := os.ReadFile(f)
		require.NoError(t, err)
		_, err = db.Exec(string(content))
		require.NoError(t, err, "failed to apply migration %s", f)
		return
	}
	require.FailNow(t, "migration file %06d.%s.sql not found", version, direction)
}

func RunMigrations(t *testing.T, db *sqlx.DB) {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS organizations (
			id varchar(100) NOT NULL,
			name varchar(191) NOT NULL,
			slug varchar(191) NOT NULL,
			owner_id varchar(100) NOT NULL,
			status varchar(50) NOT NULL DEFAULT 'active',
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			deleted_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_organizations_slug (slug, deleted_at)
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id varchar(100) NOT NULL,
			organization_id varchar(100) DEFAULT NULL,
			password varchar(191) NOT NULL,
			email varchar(191) NOT NULL,
			username varchar(191) NOT NULL,
			name varchar(191) NOT NULL,
			avatar_url text,
			token varchar(255) DEFAULT '',
			status varchar(50) NOT NULL DEFAULT 'active',
			email_verified_at bigint DEFAULT NULL,
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			deleted_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_users_email (email, deleted_at),
			UNIQUE KEY idx_users_username (username, deleted_at)
		)`,
		`CREATE TABLE IF NOT EXISTS roles (
			id varchar(100) NOT NULL,
			name varchar(50) NOT NULL,
			organization_id varchar(100) DEFAULT NULL,
			description text,
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			deleted_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_roles_name_org (name, organization_id, deleted_at)
		)`,
		`CREATE TABLE IF NOT EXISTS organization_members (
			id varchar(100) NOT NULL,
			organization_id varchar(100) NOT NULL,
			user_id varchar(100) NOT NULL,
			role_id varchar(100) NOT NULL,
			status varchar(50) NOT NULL DEFAULT 'active',
			joined_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_org_member (organization_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS endpoints (
			id varchar(100) NOT NULL,
			path varchar(191) NOT NULL,
			method varchar(20) NOT NULL,
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			deleted_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_endpoint_path_method (path, method, deleted_at)
		)`,
		`CREATE TABLE IF NOT EXISTS access_rights (
			id varchar(100) NOT NULL,
			organization_id varchar(100) DEFAULT NULL,
			name varchar(191) NOT NULL,
			description text,
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			deleted_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_access_rights_name (name),
			UNIQUE KEY idx_access_rights_name_org (name, organization_id, deleted_at)
		)`,
		`CREATE TABLE IF NOT EXISTS access_right_endpoints (
			access_right_id varchar(100) NOT NULL,
			endpoint_id varchar(100) NOT NULL,
			PRIMARY KEY (access_right_id, endpoint_id)
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id varchar(100) NOT NULL,
			organization_id varchar(100) DEFAULT NULL,
			user_id varchar(100) NOT NULL DEFAULT '',
			action varchar(100) NOT NULL,
			entity varchar(100) NOT NULL,
			entity_id varchar(100) NOT NULL DEFAULT '',
			old_values text,
			new_values text,
			ip_address varchar(100) DEFAULT '',
			user_agent text,
			created_at bigint NOT NULL DEFAULT 0,
			deleted_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id)
		)`,
		`CREATE TABLE IF NOT EXISTS audit_outbox (
			id varchar(100) NOT NULL,
			organization_id varchar(100) DEFAULT NULL,
			user_id varchar(100) NOT NULL DEFAULT '',
			action varchar(100) NOT NULL,
			entity varchar(100) NOT NULL,
			entity_id varchar(100) NOT NULL DEFAULT '',
			old_values text,
			new_values text,
			ip_address varchar(100) DEFAULT '',
			user_agent text,
			status varchar(50) NOT NULL DEFAULT 'pending',
			retry_count int NOT NULL DEFAULT 0,
			last_error text,
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id)
		)`,
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
			email varchar(191) NOT NULL,
			token varchar(191) NOT NULL,
			expires_at bigint NOT NULL,
			created_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (email),
			KEY idx_prt_token (token)
		)`,
		`CREATE TABLE IF NOT EXISTS email_verification_tokens (
			email varchar(191) NOT NULL,
			token varchar(191) NOT NULL,
			expires_at bigint NOT NULL,
			created_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (email),
			KEY idx_evt_token (token)
		)`,
		`CREATE TABLE IF NOT EXISTS user_sso_identities (
			id varchar(100) NOT NULL,
			user_id varchar(100) NOT NULL,
			provider varchar(50) NOT NULL,
			provider_id varchar(191) NOT NULL,
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_provider_id (provider, provider_id)
		)`,
		`CREATE TABLE IF NOT EXISTS invitation_tokens (
			id varchar(100) NOT NULL,
			organization_id varchar(100) NOT NULL,
			email varchar(191) NOT NULL,
			token varchar(191) NOT NULL,
			role_id varchar(100) NOT NULL,
			expires_at bigint NOT NULL,
			created_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_invitation_token (token)
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id varchar(100) NOT NULL,
			organization_id varchar(100) NOT NULL,
			user_id varchar(100) NOT NULL,
			name varchar(191) NOT NULL,
			domain varchar(191) NOT NULL,
			status varchar(50) NOT NULL DEFAULT 'active',
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			deleted_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id)
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id varchar(100) NOT NULL,
			name varchar(191) NOT NULL,
			key_hash varchar(191) NOT NULL,
			organization_id varchar(100) NOT NULL,
			user_id varchar(100) NOT NULL,
			scopes text,
			expires_at bigint DEFAULT NULL,
			last_used_at bigint DEFAULT NULL,
			is_active tinyint(1) NOT NULL DEFAULT 1,
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			deleted_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_api_keys_hash (key_hash)
		)`,
		`CREATE TABLE IF NOT EXISTS webhooks (
			id varchar(100) NOT NULL,
			name varchar(191) NOT NULL,
			organization_id varchar(100) NOT NULL,
			url text NOT NULL,
			events text,
			secret varchar(255) DEFAULT '',
			is_active tinyint(1) NOT NULL DEFAULT 1,
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0,
			deleted_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id)
		)`,
		`CREATE TABLE IF NOT EXISTS webhook_logs (
			id varchar(100) NOT NULL,
			webhook_id varchar(100) NOT NULL,
			event_type varchar(100) NOT NULL,
			payload text,
			response_status_code int NOT NULL DEFAULT 0,
			response_body text,
			execution_time bigint NOT NULL DEFAULT 0,
			error_message text,
			retry_count int NOT NULL DEFAULT 0,
			created_at bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (id)
		)`,
		`CREATE TABLE IF NOT EXISTS casbin_rule (
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
		)`,
	}

	for _, stmt := range statements {
		_, err := db.Exec(stmt)
		if t != nil {
			require.NoError(t, err, "Failed to run migrations")
		} else if err != nil {
			panic("Failed to run migrations: " + err.Error())
		}
	}
}

func SeedTestData(t *testing.T, db *sqlx.DB) {
	globalOrg := "global"
	now := time.Now().UTC().UnixMilli()
	_, _ = db.Exec(
		`INSERT IGNORE INTO organizations (id, name, slug, owner_id, status, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
		globalOrg, "Global Organization", "global", "system", orgEntity.OrgStatusActive, now, now,
	)

	roles := []roleEntity.Role{
		{ID: uuid.NewString(), Name: "role:superadmin", OrganizationID: &globalOrg, Description: "Super Administrator role"},
		{ID: uuid.NewString(), Name: "role:admin", OrganizationID: &globalOrg, Description: "Administrator role"},
		{ID: uuid.NewString(), Name: "role:user", OrganizationID: &globalOrg, Description: "Regular user role"},
		{ID: uuid.NewString(), Name: "role:org-owner", OrganizationID: &globalOrg, Description: "Organization owner role"},
		{ID: uuid.NewString(), Name: "role:moderator", OrganizationID: &globalOrg, Description: "Moderator role"},
	}

	for _, role := range roles {
		var existingID string
		err := db.Get(&existingID, "SELECT id FROM roles WHERE name = ? AND deleted_at = 0 LIMIT 1", role.Name)
		if err != nil {
			_, _ = db.Exec(
				`INSERT INTO roles (id, organization_id, name, description, created_at, updated_at, deleted_at)
				 VALUES (?, ?, ?, ?, ?, ?, 0)`,
				role.ID, role.OrganizationID, role.Name, role.Description, now, now,
			)
		}
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
		_, _ = db.Exec("INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES (?, ?, ?, ?, ?)", "p", p[0], p[1], p[2], p[3])
	}
}

func RoleIDByName(t *testing.T, db *sqlx.DB, name string) string {
	var id string
	err := db.Get(&id, "SELECT id FROM roles WHERE name = ? AND deleted_at = 0 LIMIT 1", name)
	require.NoError(t, err)
	return id
}

func CleanupDatabase(t *testing.T, db *sqlx.DB) {
	tables := []string{
		"projects",
		"organization_members",
		"organizations",
		"audit_logs",
		"audit_outbox",
		"access_right_endpoints",
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

	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	for _, table := range tables {
		_, _ = db.Exec("TRUNCATE TABLE " + table)
	}
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
}

func CreateTestUser(t *testing.T, db *sqlx.DB, username, email, password string, orgIDs ...string) *userEntity.User {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "Failed to hash password")

	now := time.Now().UTC().UnixMilli()
	user := &userEntity.User{
		ID:        uuid.New().String(),
		Username:  username,
		Email:     email,
		Name:      username,
		Password:  string(hashedPassword),
		Status:    userEntity.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = db.Exec(
		`INSERT INTO users (id, organization_id, username, email, name, password, status, avatar_url, token, email_verified_at, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.OrganizationID, user.Username, user.Email, user.Name, user.Password,
		user.Status, user.AvatarURL, user.Token, user.EmailVerifiedAt, user.CreatedAt, user.UpdatedAt, user.DeletedAt,
	)
	require.NoError(t, err, "Failed to create test user")

	// Add memberships if orgIDs provided
	for _, orgID := range orgIDs {
		member := &orgEntity.OrganizationMember{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			UserID:         user.ID,
			RoleID:         "role:user",
			Status:         "active",
			JoinedAt:       now,
		}
		_, err = db.Exec(
			`INSERT INTO organization_members (id, organization_id, user_id, role_id, status, joined_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			member.ID, member.OrganizationID, member.UserID, member.RoleID, member.Status, member.JoinedAt,
		)
		require.NoError(t, err, "Failed to create member record")
	}

	return user
}

func CreateTestOrganization(t *testing.T, db *sqlx.DB, ownerID, name, slug string) *orgEntity.Organization {
	now := time.Now().UTC().UnixMilli()
	org := &orgEntity.Organization{
		ID:        uuid.New().String(),
		Name:      name,
		Slug:      slug,
		OwnerID:   ownerID,
		Status:    orgEntity.OrgStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := db.Exec(
		`INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		org.ID, org.Name, org.Slug, org.OwnerID, org.Status, org.CreatedAt, org.UpdatedAt, org.DeletedAt,
	)
	require.NoError(t, err, "Failed to create test organization")

	return org
}

func CreateTestRole(t *testing.T, db *sqlx.DB, name string) *roleEntity.Role {
	globalOrg := "global"
	now := time.Now().UTC().UnixMilli()
	_, _ = db.Exec(
		`INSERT IGNORE INTO organizations (id, name, slug, owner_id, status, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
		globalOrg, "Global Organization", "global", "system", orgEntity.OrgStatusActive, now, now,
	)

	role := &roleEntity.Role{
		ID:             uuid.New().String(),
		Name:           name,
		OrganizationID: &globalOrg,
		Description:    "Test role " + name,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err := db.Exec(
		`INSERT INTO roles (id, organization_id, name, description, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		role.ID, role.OrganizationID, role.Name, role.Description, role.CreatedAt, role.UpdatedAt, role.DeletedAt,
	)
	if t != nil {
		require.NoError(t, err, "Failed to create test role")
	}

	return role
}

func HashSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
