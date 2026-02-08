package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	accessEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
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

	log.Println("Database connected. Starting Tiered Authorization Seeder...")

	// 1. Seed Roles
	seedRoles(db)

	// 2. Seed Superadmin User
	seedSuperAdmin(db)

	// 3. Seed Access Rights, Endpoints, and Tiered Policies
	seedAccessRightsAndPolicies(db)

	log.Println("Seeding process completed successfully.")
}

func seedRoles(db *gorm.DB) {
	roles := []roleEntity.Role{
		{Name: "role:superadmin", Description: "Full Access", OrganizationID: ptrString("global")},
		{Name: "role:admin", Description: "Org Administrator", OrganizationID: ptrString("global")},
		{Name: "role:user", Description: "Org User", OrganizationID: ptrString("global")},
	}

	for _, r := range roles {
		var count int64
		db.Model(&roleEntity.Role{}).Where("name = ?", r.Name).Count(&count)
		if count == 0 {
			r.ID = uuid.NewString()
			r.CreatedAt = time.Now().UnixMilli()
			r.UpdatedAt = time.Now().UnixMilli()
			db.Create(&r)
			log.Printf("Role '%s' created.", r.Name)
		} else {
			// Update existing role description just in case
			db.Model(&roleEntity.Role{}).Where("name = ?", r.Name).Update("description", r.Description)
		}
	}
}

func seedSuperAdmin(db *gorm.DB) {
	adminUsername := "superadmin"
	adminPassword := os.Getenv("SUPERADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("SUPERADMIN_PASSWORD environment variable is missing in .env")
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	var user userEntity.User
	if err := db.Where("username = ?", adminUsername).First(&user).Error; err != nil {
		now := time.Now().UnixMilli()
		userID := uuid.NewString()

		// Use Map to avoid "Unknown column 'avatar_url'" errors if DB schema is not fully up to date
		userData := map[string]interface{}{
			"id":         userID,
			"username":   adminUsername,
			"email":      "superadmin@example.com",
			"password":   string(hashedPwd),
			"name":       "Super Admin",
			"created_at": now,
			"updated_at": now,
		}

		if err := db.Table("users").Create(userData).Error; err != nil {
			log.Fatalf("Failed to create superadmin: %v", err)
		}
		user.ID = userID
		log.Printf("Superadmin user '%s' created.", adminUsername)
	} else {
		// ALWAYS reset superadmin password to ensure login works with current .env
		db.Table("users").Where("id = ?", user.ID).Update("password", string(hashedPwd))
		log.Printf("Superadmin user '%s' password reset.", adminUsername)
	}

	// Policy: superadmin USER has superadmin ROLE
	ensurePolicy(db, "g", user.ID, "role:superadmin", "global", "", "")
	// Policy: superadmin ROLE has all permission
	ensurePolicy(db, "p", "role:superadmin", "global", "*", "*", "")
}

func seedAccessRightsAndPolicies(db *gorm.DB) {
	accessMap := map[string][]accessEntity.Endpoint{
		"dashboard:view": {
			{Path: "/api/v1/stats/summary", Method: "GET"},
			{Path: "/api/v1/stats/activity", Method: "GET"},
			{Path: "/api/v1/stats/insights", Method: "GET"},
		},
		"user:view": {
			{Path: "/api/v1/users/", Method: "GET"},
			{Path: "/api/v1/users/search", Method: "POST"},
			{Path: "/api/v1/users/:id", Method: "GET"},
		},
		"user:manage": {
			{Path: "/api/v1/users/:id/status", Method: "PATCH"},
			{Path: "/api/v1/users/:id", Method: "DELETE"},
		},
		"org:view": {
			{Path: "/api/v1/organizations/:id", Method: "GET"},
			{Path: "/api/v1/organizations/me", Method: "GET"},
			{Path: "/api/v1/organizations/slug/:slug", Method: "GET"},
		},
		"org:manage": {
			{Path: "/api/v1/organizations", Method: "POST"},
			{Path: "/api/v1/organizations/:id", Method: "PUT"},
			{Path: "/api/v1/organizations/:id", Method: "DELETE"},
		},
		"member:manage": {
			{Path: "/api/v1/organizations/:id/members/invite", Method: "POST"},
			{Path: "/api/v1/organizations/:id/members", Method: "GET"},
			{Path: "/api/v1/organizations/:id/members/:userId", Method: "PATCH"},
			{Path: "/api/v1/organizations/:id/members/:userId", Method: "DELETE"},
		},
		"presence:view": {
			{Path: "/api/v1/organizations/:id/presence", Method: "GET"},
		},
		"project:view": {
			{Path: "/api/v1/projects", Method: "GET"},
			{Path: "/api/v1/projects/:id", Method: "GET"},
		},
		"project:manage": {
			{Path: "/api/v1/projects", Method: "POST"},
			{Path: "/api/v1/projects/:id", Method: "PUT"},
			{Path: "/api/v1/projects/:id", Method: "DELETE"},
		},
		"role:view": {
			{Path: "/api/v1/roles", Method: "GET"},
			{Path: "/api/v1/roles/search", Method: "POST"},
		},
		"role:manage": {
			{Path: "/api/v1/roles", Method: "POST"},
			{Path: "/api/v1/roles/:id", Method: "PUT"},
			{Path: "/api/v1/roles/:id", Method: "DELETE"},
		},
		"permission:view": {
			{Path: "/api/v1/permissions", Method: "GET"},
			{Path: "/api/v1/permissions/:role", Method: "GET"},
			{Path: "/api/v1/permissions/roles/:role/users", Method: "GET"},
			{Path: "/api/v1/permissions/:role/parents", Method: "GET"},
		},
		"permission:manage": {
			{Path: "/api/v1/permissions/assign-role", Method: "POST"},
			{Path: "/api/v1/permissions/revoke-role", Method: "DELETE"},
			{Path: "/api/v1/permissions/grant", Method: "POST"},
			{Path: "/api/v1/permissions", Method: "PUT"},
			{Path: "/api/v1/permissions/revoke", Method: "DELETE"},
			{Path: "/api/v1/permissions/inheritance", Method: "POST"},
			{Path: "/api/v1/permissions/inheritance", Method: "DELETE"},
		},
		"access:view": {
			{Path: "/api/v1/access-rights", Method: "GET"},
			{Path: "/api/v1/access-rights/search", Method: "POST"},
			{Path: "/api/v1/endpoints/search", Method: "POST"},
		},
		"access:manage": {
			{Path: "/api/v1/access-rights", Method: "POST"},
			{Path: "/api/v1/access-rights/:id", Method: "DELETE"},
			{Path: "/api/v1/access-rights/link", Method: "POST"},
			{Path: "/api/v1/endpoints", Method: "POST"},
			{Path: "/api/v1/endpoints/:id", Method: "DELETE"},
		},
		"audit:view": {
			{Path: "/api/v1/audit-logs/search", Method: "POST"},
			{Path: "/api/v1/audit-logs/export", Method: "GET"},
		},
	}

	roleToRights := map[string][]string{
		"role:admin": {
			"dashboard:view", "user:view", "role:view", "role:manage",
			"project:view", "project:manage", "org:view", "org:manage",
			"member:manage", "presence:view", "audit:view",
		},
		"role:user": { // FIXED: Mapping changed from member to user
			"dashboard:view", "project:view", "org:view", "presence:view",
		},
	}

	// 1. Seed Endpoints and AccessRights into DB
	for arName, eps := range accessMap {
		var ar accessEntity.AccessRight
		if err := db.Where("name = ?", arName).First(&ar).Error; err != nil {
			ar = accessEntity.AccessRight{ID: uuid.NewString(), Name: arName}
			db.Create(&ar)
		}

		for _, ep := range eps {
			var endpoint accessEntity.Endpoint
			if err := db.Where("path = ? AND method = ?", ep.Path, ep.Method).First(&endpoint).Error; err != nil {
				endpoint = accessEntity.Endpoint{ID: uuid.NewString(), Path: ep.Path, Method: ep.Method}
				db.Create(&endpoint)
			}
			// Link in DB
			db.Exec("INSERT IGNORE INTO access_right_endpoints (access_right_id, endpoint_id) VALUES (?, ?)", ar.ID, endpoint.ID)

			// 2. Seed Casbin Policy: p, accessRight, global, path, method
			ensurePolicy(db, "p", arName, "global", ep.Path, ep.Method, "")
		}
	}

	// 3. Seed Casbin Inheritance: g, role, accessRight, global
	for roleName, rights := range roleToRights {
		for _, arName := range rights {
			ensurePolicy(db, "g", roleName, arName, "global", "", "")
		}
	}
}

func ensurePolicy(db *gorm.DB, ptype, v0, v1, v2, v3, v4 string) {
	var count int64
	query := db.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", ptype, v0, v1, v2)
	if v3 != "" {
		query = query.Where("v3 = ?", v3)
	}
	if v4 != "" {
		query = query.Where("v4 = ?", v4)
	}
	query.Count(&count)

	if count == 0 {
		db.Table("casbin_rule").Create(map[string]interface{}{
			"ptype": ptype, "v0": v0, "v1": v1, "v2": v2, "v3": v3, "v4": v4,
		})
	}
}

func ptrString(s string) *string {
	return &s
}
