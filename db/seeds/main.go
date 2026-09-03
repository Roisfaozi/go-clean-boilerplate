package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	accessEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	projectEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected. Starting Tiered Authorization Seeder (direct DB)...")

	adminPassword := os.Getenv("SUPERADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("SUPERADMIN_PASSWORD environment variable is missing in .env")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := seedRoles(tx); err != nil {
			return fmt.Errorf("seed roles: %w", err)
		}

		superAdminID, err := seedSuperAdmin(tx, adminPassword)
		if err != nil {
			return fmt.Errorf("seed superadmin: %w", err)
		}

		if err := seedAccessRightsAndPolicies(tx); err != nil {
			return fmt.Errorf("seed access rights: %w", err)
		}

		orgID, err := seedDefaultOrganization(tx, superAdminID)
		if err != nil {
			return fmt.Errorf("seed organization: %w", err)
		}

		if err := seedOrganizationUsers(tx, orgID); err != nil {
			return fmt.Errorf("seed organization users: %w", err)
		}

		if err := seedDefaultProject(tx, orgID, superAdminID); err != nil {
			return fmt.Errorf("seed project: %w", err)
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Seeding failed, transaction rolled back: %v", err)
	}

	log.Println("Seeding process completed successfully.")
}

func seedRoles(db *gorm.DB) error {
	// organization_id stays NULL for global roles: the column has an FK to
	// organizations(id), and OrganizationScope treats NULL as global.
	roles := []roleEntity.Role{
		{Name: superAdminRole, Description: "Full Access"},
		{Name: adminRoleName, Description: "Org Administrator"},
		{Name: userRoleName, Description: "Org User"},
		{ID: ownerRoleName, Name: ownerRoleName, Description: "Organization Owner"},
	}

	for _, r := range roles {
		var existing roleEntity.Role
		err := db.Where("name = ?", r.Name).First(&existing).Error
		if err == nil {
			if err := db.Model(&roleEntity.Role{}).Where("id = ?", existing.ID).Update("description", r.Description).Error; err != nil {
				return err
			}
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		if r.ID == "" {
			r.ID = uuid.NewString()
		}
		if err := db.Create(&r).Error; err != nil {
			return err
		}
		log.Printf("Role '%s' created.", r.Name)
	}

	return nil
}

func seedSuperAdmin(db *gorm.DB, adminPassword string) (string, error) {
	adminUsername := "superadmin"

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	var user userEntity.User
	err = db.Where("username = ?", adminUsername).First(&user).Error
	switch err {
	case nil:
		if err := db.Table("users").Where("id = ?", user.ID).Update("password", string(hashedPwd)).Error; err != nil {
			return "", err
		}
		log.Printf("Superadmin user '%s' password reset.", adminUsername)
	case gorm.ErrRecordNotFound:
		now := time.Now().UnixMilli()
		user.ID = uuid.NewString()

		// Map insert keeps seeder resilient to optional columns such as avatar_url.
		userData := map[string]interface{}{
			"id":         user.ID,
			"username":   adminUsername,
			"email":      "superadmin@example.com",
			"password":   string(hashedPwd),
			"name":       "Super Admin",
			"created_at": now,
			"updated_at": now,
		}
		if err := db.Table("users").Create(userData).Error; err != nil {
			return "", err
		}
		log.Printf("Superadmin user '%s' created.", adminUsername)
	default:
		return "", err
	}

	// superadmin user holds superadmin role in global domain
	if err := ensureGroupingPolicy(db, user.ID, superAdminRole, globalDomain); err != nil {
		return "", err
	}
	// superadmin role has unrestricted policy
	if err := ensurePolicy(db, superAdminRole, globalDomain, "*", "*"); err != nil {
		return "", err
	}

	return user.ID, nil
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
			{"/api/v1/users/:id/status", "PATCH"},
			{"/api/v1/users/:id", "DELETE"},
		},
		"org:view": {
			{"/api/v1/organizations/:id", "GET"},
			{"/api/v1/organizations/me", "GET"},
			{"/api/v1/organizations/slug/:slug", "GET"},
		},
		"org:manage": {
			{"/api/v1/organizations", "POST"},
			{"/api/v1/organizations/:id", "PUT"},
			{"/api/v1/organizations/:id", "DELETE"},
		},
		"member:manage": {
			{"/api/v1/organizations/:id/members/invite", "POST"},
			{"/api/v1/organizations/:id/members", "GET"},
			{"/api/v1/organizations/:id/members/:userId", "PATCH"},
			{"/api/v1/organizations/:id/members/:userId", "DELETE"},
		},
		"presence:view": {
			{"/api/v1/organizations/:id/presence", "GET"},
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
		"role:view": {
			{"/api/v1/roles", "GET"},
			{"/api/v1/roles/search", "POST"},
		},
		"role:manage": {
			{"/api/v1/roles", "POST"},
			{"/api/v1/roles/:id", "PUT"},
			{"/api/v1/roles/:id", "DELETE"},
		},
		"permission:view": {
			{"/api/v1/permissions", "GET"},
			{"/api/v1/permissions/:role", "GET"},
			{"/api/v1/permissions/roles/:role/users", "GET"},
			{"/api/v1/permissions/:role/parents", "GET"},
			{"/api/v1/permissions/resources", "GET"},
			{"/api/v1/permissions/inheritance-tree", "GET"},
		},
		"permission:manage": {
			{"/api/v1/permissions/assign-role", "POST"},
			{"/api/v1/permissions/revoke-role", "DELETE"},
			{"/api/v1/permissions/grant", "POST"},
			{"/api/v1/permissions", "PUT"},
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

func seedAccessRightsAndPolicies(db *gorm.DB) error {
	accessMap := accessRightSpec()

	for arName, endpoints := range accessMap {
		arID, err := ensureAccessRight(db, arName)
		if err != nil {
			return err
		}

		for _, ep := range endpoints {
			epID, err := ensureEndpoint(db, ep)
			if err != nil {
				return err
			}
			if err := ensureAccessRightEndpointLink(db, arID, epID); err != nil {
				return err
			}
		}
	}

	// Expand access rights into Casbin policies: p, role, domain, path, method
	for roleName, rights := range roleAccessRightSpec() {
		for _, arName := range rights {
			for _, ep := range accessMap[arName] {
				if err := ensurePolicy(db, roleName, globalDomain, ep.Path, ep.Method); err != nil {
					return err
				}
			}
		}
		log.Printf("Access rights expanded into policies for role '%s'.", roleName)
	}

	return nil
}

func ensureAccessRight(db *gorm.DB, name string) (string, error) {
	var ar accessEntity.AccessRight
	err := db.Where("name = ?", name).First(&ar).Error
	if err == nil {
		return ar.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	// organization_id stays NULL for global access rights: the column has an FK to
	// organizations(id), and OrganizationScope treats NULL as global.
	ar = accessEntity.AccessRight{
		ID:   uuid.NewString(),
		Name: name,
	}
	if err := db.Create(&ar).Error; err != nil {
		return "", err
	}
	log.Printf("Access right '%s' created.", name)
	return ar.ID, nil
}

func ensureEndpoint(db *gorm.DB, ep endpointSpec) (string, error) {
	var existing accessEntity.Endpoint
	err := db.Where("path = ? AND method = ?", ep.Path, ep.Method).First(&existing).Error
	if err == nil {
		return existing.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	created := accessEntity.Endpoint{
		ID:     uuid.NewString(),
		Path:   ep.Path,
		Method: ep.Method,
	}
	if err := db.Create(&created).Error; err != nil {
		return "", err
	}
	return created.ID, nil
}

func ensureAccessRightEndpointLink(db *gorm.DB, accessRightID, endpointID string) error {
	var count int64
	if err := db.Table("access_right_endpoints").
		Where("access_right_id = ? AND endpoint_id = ?", accessRightID, endpointID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return db.Create(&accessEntity.AccessRightEndpoint{
		AccessRightID: accessRightID,
		EndpointID:    endpointID,
	}).Error
}

func seedDefaultOrganization(db *gorm.DB, ownerUserID string) (string, error) {
	ownerRoleID, err := roleIDByName(db, ownerRoleName)
	if err != nil {
		return "", err
	}

	var org orgEntity.Organization
	err = db.Where("slug = ?", defaultOrgSlug).First(&org).Error
	switch err {
	case nil:
		log.Printf("Default organization found with ID: %s", org.ID)
	case gorm.ErrRecordNotFound:
		org = orgEntity.Organization{
			ID:      uuid.NewString(),
			Name:    defaultOrgName,
			Slug:    defaultOrgSlug,
			OwnerID: ownerUserID,
			Status:  orgEntity.OrgStatusActive,
		}
		if err := db.Create(&org).Error; err != nil {
			return "", err
		}
		log.Printf("Default organization created with ID: %s", org.ID)
	default:
		return "", err
	}

	// Owner membership row
	if err := ensureMember(db, org.ID, ownerUserID, ownerRoleID, orgEntity.MemberStatusActive); err != nil {
		return "", err
	}

	// Mirror global role policies into org domain, matching bootstrapOrganizationPolicies
	accessMap := accessRightSpec()
	for roleName, rights := range roleAccessRightSpec() {
		for _, arName := range rights {
			for _, ep := range accessMap[arName] {
				if err := ensurePolicy(db, roleName, org.ID, ep.Path, ep.Method); err != nil {
					return "", err
				}
			}
		}
	}

	// org-owner inherits admin inside org domain
	if err := ensureGroupingPolicy(db, ownerRoleName, adminRoleName, org.ID); err != nil {
		return "", err
	}
	// owner user holds org-owner role in org domain
	if err := ensureGroupingPolicy(db, ownerUserID, ownerRoleName, org.ID); err != nil {
		return "", err
	}

	return org.ID, nil
}

func seedOrganizationUsers(db *gorm.DB, orgID string) error {
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

	for _, u := range usersToSeed {
		roleID, err := roleIDByName(db, u.RoleName)
		if err != nil {
			return err
		}

		var user userEntity.User
		err = db.Where("username = ?", u.Username).First(&user).Error
		switch err {
		case nil:
			log.Printf("User '%s' found with ID: %s", u.Username, user.ID)
		case gorm.ErrRecordNotFound:
			now := time.Now().UnixMilli()
			user.ID = uuid.NewString()
			userData := map[string]interface{}{
				"id":         user.ID,
				"username":   u.Username,
				"email":      u.Email,
				"password":   string(hashedPwd),
				"name":       u.Name,
				"created_at": now,
				"updated_at": now,
			}
			if err := db.Table("users").Create(userData).Error; err != nil {
				return err
			}
			log.Printf("User '%s' created with ID: %s", u.Username, user.ID)
		default:
			return err
		}

		if err := ensureMember(db, orgID, user.ID, roleID, orgEntity.MemberStatusActive); err != nil {
			return err
		}
		if err := ensureGroupingPolicy(db, user.ID, u.RoleName, orgID); err != nil {
			return err
		}
	}

	return nil
}

func seedDefaultProject(db *gorm.DB, orgID, userID string) error {
	const projectName = "Sample E-Commerce App"

	var count int64
	if err := db.Model(&projectEntity.Project{}).
		Where("organization_id = ? AND name = ?", orgID, projectName).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	project := projectEntity.Project{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		UserID:         userID,
		Name:           projectName,
		Domain:         "e-commerce",
		Status:         "active",
	}
	if err := db.Create(&project).Error; err != nil {
		return err
	}

	log.Println("Default project created successfully.")
	return nil
}

func ensureMember(db *gorm.DB, orgID, userID, roleID, status string) error {
	var count int64
	if err := db.Model(&orgEntity.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return db.Create(&orgEntity.OrganizationMember{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		UserID:         userID,
		RoleID:         roleID,
		Status:         status,
	}).Error
}

func roleIDByName(db *gorm.DB, name string) (string, error) {
	var role roleEntity.Role
	if err := db.Where("name = ?", name).First(&role).Error; err != nil {
		return "", fmt.Errorf("role %q not found: %w", name, err)
	}
	return role.ID, nil
}

// ensurePolicy inserts a Casbin permission rule: p, subject, domain, object, action
func ensurePolicy(db *gorm.DB, subject, domain, object, action string) error {
	return ensureCasbinRule(db, "p", subject, domain, object, action)
}

// ensureGroupingPolicy inserts a Casbin grouping rule: g, member, role, domain
func ensureGroupingPolicy(db *gorm.DB, member, role, domain string) error {
	return ensureCasbinRule(db, "g", member, role, domain, "")
}

func ensureCasbinRule(db *gorm.DB, ptype, v0, v1, v2, v3 string) error {
	query := db.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", ptype, v0, v1, v2)
	if v3 != "" {
		query = query.Where("v3 = ?", v3)
	} else {
		query = query.Where("v3 IS NULL OR v3 = ?", "")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return db.Table("casbin_rule").Create(map[string]interface{}{
		"ptype": ptype, "v0": v0, "v1": v1, "v2": v2, "v3": v3, "v4": "",
	}).Error
}
