package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

const (
	globalDomain       = "global"
	ownerRoleName      = "role:org-owner"
	adminRoleName      = "role:admin"
	userRoleName       = "role:user"
	superAdminRole     = "role:superadmin"
	defaultOrgSlug     = "default-org"
	defaultOrgName     = "Default Organization"
	seededUserPassword = "Password0!"
)

type endpointSpec struct {
	Path   string
	Method string
}

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		cfg.Mysql.User,
		cfg.Mysql.Password,
		cfg.Mysql.Host,
		cfg.Mysql.Port,
		cfg.Mysql.DBName,
	)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected. Starting Tiered Authorization Seeder (direct SQLX)...")

	adminPassword := os.Getenv("SUPERADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("SUPERADMIN_PASSWORD environment variable is missing in .env")
	}

	ctx := context.Background()
	txx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to begin seed transaction: %v", err)
	}
	defer txx.Rollback()

	if err := seedRoles(ctx, txx); err != nil {
		log.Fatalf("seed roles: %v", err)
	}

	superAdminID, err := seedSuperAdmin(ctx, txx, adminPassword)
	if err != nil {
		log.Fatalf("seed superadmin: %v", err)
	}

	if err := seedAccessRightsAndPolicies(ctx, txx); err != nil {
		log.Fatalf("seed access rights: %v", err)
	}

	orgID, err := seedDefaultOrganization(ctx, txx, superAdminID)
	if err != nil {
		log.Fatalf("seed organization: %v", err)
	}

	if err := seedOrganizationUsers(ctx, txx, orgID); err != nil {
		log.Fatalf("seed organization users: %v", err)
	}

	if err := seedDefaultProject(ctx, txx, orgID, superAdminID); err != nil {
		log.Fatalf("seed project: %v", err)
	}

	if err := txx.Commit(); err != nil {
		log.Fatalf("Seeding failed, transaction rolled back: %v", err)
	}

	log.Println("Seeding process completed successfully.")
}

func seedRoles(ctx context.Context, dbtx tx.DBTX) error {
	roles := []struct {
		ID          string
		Name        string
		Description string
	}{
		{Name: superAdminRole, Description: "Full Access"},
		{Name: adminRoleName, Description: "Org Administrator"},
		{Name: userRoleName, Description: "Org User"},
		{ID: ownerRoleName, Name: ownerRoleName, Description: "Organization Owner"},
	}

	for _, r := range roles {
		id := r.ID
		if id == "" {
			id = uuid.NewString()
		}
		now := time.Now().UnixMilli()

		query := `
			INSERT INTO roles (id, name, organization_id, description, created_at, updated_at, deleted_at)
			VALUES (?, ?, NULL, ?, ?, ?, 0)
			ON DUPLICATE KEY UPDATE description = VALUES(description), updated_at = VALUES(updated_at)
		`
		if _, err := dbtx.ExecContext(ctx, query, id, r.Name, r.Description, now, now); err != nil {
			return err
		}
		log.Printf("Role '%s' seeded.", r.Name)
	}
	return nil
}

func seedSuperAdmin(ctx context.Context, dbtx tx.DBTX, adminPassword string) (string, error) {
	adminUsername := "superadmin"
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	var userID string
	err = dbtx.GetContext(ctx, &userID, `SELECT id FROM users WHERE username = ? AND deleted_at = 0 LIMIT 1`, adminUsername)
	now := time.Now().UnixMilli()

	if err == nil {
		// Update password
		query := `UPDATE users SET password = ?, updated_at = ? WHERE id = ?`
		if _, err := dbtx.ExecContext(ctx, query, string(hashedPwd), now, userID); err != nil {
			return "", err
		}
		log.Printf("Superadmin user '%s' password reset.", adminUsername)
	} else if errors.Is(err, sql.ErrNoRows) {
		userID = uuid.NewString()
		query := `
			INSERT INTO users (id, username, email, password, name, status, created_at, updated_at, deleted_at)
			VALUES (?, ?, ?, ?, ?, 'active', ?, ?, 0)
		`
		if _, err := dbtx.ExecContext(ctx, query, userID, adminUsername, "superadmin@example.com", string(hashedPwd), "Super Admin", now, now); err != nil {
			return "", err
		}
		log.Printf("Superadmin user '%s' created.", adminUsername)
	} else {
		return "", err
	}

	if err := ensureGroupingPolicy(ctx, dbtx, userID, superAdminRole, globalDomain); err != nil {
		return "", err
	}
	if err := ensurePolicy(ctx, dbtx, superAdminRole, globalDomain, "*", "*"); err != nil {
		return "", err
	}

	return userID, nil
}

func accessRightSpec() map[string][]endpointSpec {
	return map[string][]endpointSpec{
		"dashboard:view": {
			{"/api/v1/stats/summary", "GET"},
			{"/api/v1/stats/activity", "GET"},
			{"/api/v1/stats/insights", "GET"},
		},
		"user:view": {
			{"/api/v1/users/", "GET"},
			{"/api/v1/users/search", "POST"},
			{"/api/v1/users/:id", "GET"},
		},
		"user:manage": {
			{"/api/v1/users/:id", "DELETE"},
			{"/api/v1/users/:id/status", "PATCH"},
		},
		"role:view": {
			{"/api/v1/roles", "GET"},
			{"/api/v1/roles/search", "POST"},
		},
		"role:manage": {
			{"/api/v1/roles", "POST"},
			{"/api/v1/roles/:id", "PUT"},
			{"/api/v1/roles/:id", "DELETE"},
		},
		"project:view": {
			{"/api/v1/projects", "GET"},
			{"/api/v1/projects/:id", "GET"},
		},
		"project:manage": {
			{"/api/v1/projects", "POST"},
			{"/api/v1/projects/:id", "PUT"},
			{"/api/v1/projects/:id", "DELETE"},
		},
		"org:view": {
			{"/api/v1/organizations/:id", "GET"},
			{"/api/v1/organizations/slug/:slug", "GET"},
		},
		"org:manage": {
			{"/api/v1/organizations/:id", "PUT"},
			{"/api/v1/organizations/:id", "DELETE"},
			{"/api/v1/organizations/:id/restore", "POST"},
			{"/api/v1/organizations/:id/hard", "DELETE"},
		},
		"member:manage": {
			{"/api/v1/organizations/:id/members", "GET"},
			{"/api/v1/organizations/:id/members/invite", "POST"},
			{"/api/v1/organizations/:id/members/:userId/role", "PUT"},
			{"/api/v1/organizations/:id/members/:userId", "DELETE"},
		},
		"presence:view": {
			{"/api/v1/organizations/:id/presence", "GET"},
		},
		"permission:view": {
			{"/api/v1/permissions", "GET"},
			{"/api/v1/permissions/roles/:role", "GET"},
			{"/api/v1/permissions/roles/:role/users", "GET"},
			{"/api/v1/permissions/roles/:role/parents", "GET"},
			{"/api/v1/permissions/aggregation", "GET"},
			{"/api/v1/permissions/inheritance-tree", "GET"},
		},
		"permission:manage": {
			{"/api/v1/permissions/assign", "POST"},
			{"/api/v1/permissions/revoke-role", "POST"},
			{"/api/v1/permissions/grant", "POST"},
			{"/api/v1/permissions/update", "PUT"},
			{"/api/v1/permissions/revoke", "DELETE"},
			{"/api/v1/permissions/inheritance", "POST"},
			{"/api/v1/permissions/inheritance", "DELETE"},
			{"/api/v1/permissions/assign-access-right", "POST"},
			{"/api/v1/permissions/revoke-access-right", "DELETE"},
			{"/api/v1/permissions/roles/:role/access-rights", "GET"},
		},
		"access:view": {
			{"/api/v1/access-rights", "GET"},
			{"/api/v1/access-rights/search", "POST"},
			{"/api/v1/endpoints/search", "POST"},
		},
		"access:manage": {
			{"/api/v1/access-rights", "POST"},
			{"/api/v1/access-rights/:id", "DELETE"},
			{"/api/v1/access-rights/link", "POST"},
			{"/api/v1/access-rights/unlink", "POST"},
			{"/api/v1/endpoints", "POST"},
			{"/api/v1/endpoints/:id", "DELETE"},
		},
		"audit:view": {
			{"/api/v1/audit-logs/search", "POST"},
			{"/api/v1/audit-logs/export", "GET"},
			{"/api/v1/audit-logs/export-async", "POST"},
		},
		"webhook:manage": {
			{"/api/v1/webhooks", "POST"},
			{"/api/v1/webhooks", "GET"},
			{"/api/v1/webhooks/:id", "GET"},
			{"/api/v1/webhooks/:id", "PUT"},
			{"/api/v1/webhooks/:id", "DELETE"},
			{"/api/v1/webhooks/:id/logs", "GET"},
		},
	}
}

func roleAccessRightSpec() map[string][]string {
	return map[string][]string{
		adminRoleName: {
			"dashboard:view",
			"user:view", "user:manage",
			"role:view", "role:manage",
			"project:view", "project:manage",
			"org:view", "org:manage",
			"member:manage", "presence:view",
			"audit:view",
			"permission:view", "permission:manage",
			"access:view", "access:manage",
			"webhook:manage",
		},
		userRoleName: {
			"dashboard:view", "project:view", "org:view", "presence:view",
		},
	}
}

func seedAccessRightsAndPolicies(ctx context.Context, dbtx tx.DBTX) error {
	accessMap := accessRightSpec()

	for arName, endpoints := range accessMap {
		arID, err := ensureAccessRight(ctx, dbtx, arName)
		if err != nil {
			return err
		}

		for _, ep := range endpoints {
			epID, err := ensureEndpoint(ctx, dbtx, ep)
			if err != nil {
				return err
			}
			if err := ensureAccessRightEndpointLink(ctx, dbtx, arID, epID); err != nil {
				return err
			}
		}
	}

	for roleName, rights := range roleAccessRightSpec() {
		for _, arName := range rights {
			for _, ep := range accessMap[arName] {
				if err := ensurePolicy(ctx, dbtx, roleName, globalDomain, ep.Path, ep.Method); err != nil {
					return err
				}
			}
		}
		log.Printf("Access rights expanded into policies for role '%s'.", roleName)
	}

	return nil
}

func ensureAccessRight(ctx context.Context, dbtx tx.DBTX, name string) (string, error) {
	var arID string
	err := dbtx.GetContext(ctx, &arID, `SELECT id FROM access_rights WHERE name = ? AND deleted_at = 0 LIMIT 1`, name)
	if err == nil {
		return arID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	arID = uuid.NewString()
	now := time.Now().UnixMilli()
	query := `
		INSERT INTO access_rights (id, name, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, 0)
	`
	if _, err := dbtx.ExecContext(ctx, query, arID, name, now, now); err != nil {
		return "", err
	}
	log.Printf("Access right '%s' created.", name)
	return arID, nil
}

func ensureEndpoint(ctx context.Context, dbtx tx.DBTX, ep endpointSpec) (string, error) {
	var epID string
	err := dbtx.GetContext(ctx, &epID, `SELECT id FROM endpoints WHERE path = ? AND method = ? AND deleted_at = 0 LIMIT 1`, ep.Path, ep.Method)
	if err == nil {
		return epID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	epID = uuid.NewString()
	now := time.Now().UnixMilli()
	query := `
		INSERT INTO endpoints (id, path, method, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, 0)
	`
	if _, err := dbtx.ExecContext(ctx, query, epID, ep.Path, ep.Method, now, now); err != nil {
		return "", err
	}
	return epID, nil
}

func ensureAccessRightEndpointLink(ctx context.Context, dbtx tx.DBTX, accessRightID, endpointID string) error {
	query := `
		INSERT IGNORE INTO access_right_endpoints (access_right_id, endpoint_id)
		VALUES (?, ?)
	`
	_, err := dbtx.ExecContext(ctx, query, accessRightID, endpointID)
	return err
}

func seedDefaultOrganization(ctx context.Context, dbtx tx.DBTX, ownerUserID string) (string, error) {
	ownerRoleID, err := roleIDByName(ctx, dbtx, ownerRoleName)
	if err != nil {
		return "", err
	}
	adminRoleID, err := roleIDByName(ctx, dbtx, adminRoleName)
	if err != nil {
		return "", err
	}

	var orgID string
	err = dbtx.GetContext(ctx, &orgID, `SELECT id FROM organizations WHERE slug = ? AND deleted_at = 0 LIMIT 1`, defaultOrgSlug)
	now := time.Now().UnixMilli()

	if err == nil {
		log.Printf("Default organization found with ID: %s", orgID)
	} else if errors.Is(err, sql.ErrNoRows) {
		orgID = uuid.NewString()
		query := `
			INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at, deleted_at)
			VALUES (?, ?, ?, ?, 'active', ?, ?, 0)
		`
		if _, err := dbtx.ExecContext(ctx, query, orgID, defaultOrgName, defaultOrgSlug, ownerUserID, now, now); err != nil {
			return "", err
		}
		log.Printf("Default organization created with ID: %s", orgID)
	} else {
		return "", err
	}

	if err := ensureMember(ctx, dbtx, orgID, ownerUserID, ownerRoleID, "active"); err != nil {
		return "", err
	}

	accessMap := accessRightSpec()
	for roleName, rights := range roleAccessRightSpec() {
		for _, arName := range rights {
			for _, ep := range accessMap[arName] {
				if err := ensurePolicy(ctx, dbtx, roleName, orgID, ep.Path, ep.Method); err != nil {
					return "", err
				}
			}
		}
	}

	if err := ensureGroupingPolicy(ctx, dbtx, ownerRoleName, adminRoleName, orgID); err != nil {
		return "", err
	}
	if err := ensureGroupingPolicy(ctx, dbtx, ownerUserID, ownerRoleName, orgID); err != nil {
		return "", err
	}

	_ = adminRoleID
	return orgID, nil
}

func seedOrganizationUsers(ctx context.Context, dbtx tx.DBTX, orgID string) error {
	usersToSeed := []struct {
		Username string
		Name     string
		Email    string
		RoleName string
	}{
		{"adminuser", "Admin User", "admin@example.com", adminRoleName},
		{"regularuser", "Regular User", "user@example.com", userRoleName},
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(seededUserPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now().UnixMilli()
	for _, u := range usersToSeed {
		roleID, err := roleIDByName(ctx, dbtx, u.RoleName)
		if err != nil {
			return err
		}

		var userID string
		err = dbtx.GetContext(ctx, &userID, `SELECT id FROM users WHERE username = ? AND deleted_at = 0 LIMIT 1`, u.Username)
		if err == nil {
			log.Printf("User '%s' found with ID: %s", u.Username, userID)
		} else if errors.Is(err, sql.ErrNoRows) {
			userID = uuid.NewString()
			query := `
				INSERT INTO users (id, username, email, password, name, status, created_at, updated_at, deleted_at)
				VALUES (?, ?, ?, ?, ?, 'active', ?, ?, 0)
			`
			if _, err := dbtx.ExecContext(ctx, query, userID, u.Username, u.Email, string(hashedPwd), u.Name, now, now); err != nil {
				return err
			}
			log.Printf("User '%s' created with ID: %s", u.Username, userID)
		} else {
			return err
		}

		if err := ensureMember(ctx, dbtx, orgID, userID, roleID, "active"); err != nil {
			return err
		}
		if err := ensureGroupingPolicy(ctx, dbtx, userID, u.RoleName, orgID); err != nil {
			return err
		}
	}

	return nil
}

func seedDefaultProject(ctx context.Context, dbtx tx.DBTX, orgID, userID string) error {
	const projectName = "Sample E-Commerce App"
	var count int64
	err := dbtx.GetContext(ctx, &count, `SELECT COUNT(*) FROM projects WHERE organization_id = ? AND name = ? AND deleted_at = 0`, orgID, projectName)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	query := `
		INSERT INTO projects (id, organization_id, user_id, name, domain, status, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, 'e-commerce', 'active', ?, ?, 0)
	`
	_, err = dbtx.ExecContext(ctx, query, uuid.NewString(), orgID, userID, projectName, now, now)
	if err != nil {
		return err
	}

	log.Println("Default project created successfully.")
	return nil
}

func ensureMember(ctx context.Context, dbtx tx.DBTX, orgID, userID, roleID, status string) error {
	var count int64
	err := dbtx.GetContext(ctx, &count, `SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ?`, orgID, userID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role_id, status, joined_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = dbtx.ExecContext(ctx, query, uuid.NewString(), orgID, userID, roleID, status, now)
	return err
}

func roleIDByName(ctx context.Context, dbtx tx.DBTX, name string) (string, error) {
	var id string
	err := dbtx.GetContext(ctx, &id, `SELECT id FROM roles WHERE name = ? AND deleted_at = 0 LIMIT 1`, name)
	if err != nil {
		return "", fmt.Errorf("role %q not found: %w", name, err)
	}
	return id, nil
}

func ensurePolicy(ctx context.Context, dbtx tx.DBTX, subject, domain, object, action string) error {
	return ensureCasbinRule(ctx, dbtx, "p", subject, domain, object, action)
}

func ensureGroupingPolicy(ctx context.Context, dbtx tx.DBTX, member, role, domain string) error {
	return ensureCasbinRule(ctx, dbtx, "g", member, role, domain, "")
}

func ensureCasbinRule(ctx context.Context, dbtx tx.DBTX, ptype, v0, v1, v2, v3 string) error {
	var count int64
	queryCount := `SELECT COUNT(*) FROM casbin_rule WHERE ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND (v3 = ? OR (v3 IS NULL AND ? = ''))`
	err := dbtx.GetContext(ctx, &count, queryCount, ptype, v0, v1, v2, v3, v3)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	queryInsert := `INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4) VALUES (?, ?, ?, ?, ?, '')`
	_, err = dbtx.ExecContext(ctx, queryInsert, ptype, v0, v1, v2, v3)
	return err
}
