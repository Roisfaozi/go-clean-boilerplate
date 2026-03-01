package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
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

	// 3. Authenticate as Superadmin to get the token
	token, err := authenticateSuperadmin(cfg)
	if err != nil {
		log.Fatalf("Failed to authenticate as superadmin during seeding: %v", err)
	}

	// 4. Seed Access Rights, Endpoints, and Tiered Policies via API
	seedAccessRightsAndPoliciesViaAPI(cfg, token)

	// 5. Seed Default Organization, Users, and Projects via API
	seedOrganizationsUsersAndProjects(cfg, token)

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

func authenticateSuperadmin(cfg *config.AppConfig) (string, error) {
	adminPassword := os.Getenv("SUPERADMIN_PASSWORD")
	if adminPassword == "" {
		return "", fmt.Errorf("SUPERADMIN_PASSWORD environment variable is missing in .env")
	}

	apiBaseURL := fmt.Sprintf("http://localhost:%d/api/v1", cfg.Server.Port)
	loginURL := fmt.Sprintf("%s/auth/login", apiBaseURL)

	payload := map[string]string{
		"username": "superadmin",
		"password": adminPassword,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", loginURL, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return "", fmt.Errorf("login failed with status %d: %v", resp.StatusCode, errResp)
	}

	var loginResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("failed to decode login response: %v", err)
	}

	return loginResp.Data.AccessToken, nil
}

func doJSONRequest(method, url string, payload interface{}, token string, expectedStatus int) (map[string]interface{}, error) {
	var bodyReader *bytes.Buffer
	if payload != nil {
		payloadBytes, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(payloadBytes)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != expectedStatus {
		return result, fmt.Errorf("API request %s %s failed with status %d: %v", method, url, resp.StatusCode, result)
	}

	return result, nil
}

func seedAccessRightsAndPoliciesViaAPI(cfg *config.AppConfig, token string) {
	apiBaseURL := fmt.Sprintf("http://localhost:%d/api/v1", cfg.Server.Port)

	accessMap := map[string][]map[string]string{
		"dashboard:view": {
			{"Path": "/api/v1/stats/summary", "Method": "GET"},
			{"Path": "/api/v1/stats/activity", "Method": "GET"},
			{"Path": "/api/v1/stats/insights", "Method": "GET"},
		},
		"user:view": {
			{"Path": "/api/v1/users/", "Method": "GET"},
			{"Path": "/api/v1/users/search", "Method": "POST"},
			{"Path": "/api/v1/users/:id", "Method": "GET"},
		},
		"user:manage": {
			{"Path": "/api/v1/users/:id/status", "Method": "PATCH"},
			{"Path": "/api/v1/users/:id", "Method": "DELETE"},
		},
		"org:view": {
			{"Path": "/api/v1/organizations/:id", "Method": "GET"},
			{"Path": "/api/v1/organizations/me", "Method": "GET"},
			{"Path": "/api/v1/organizations/slug/:slug", "Method": "GET"},
		},
		"org:manage": {
			{"Path": "/api/v1/organizations", "Method": "POST"},
			{"Path": "/api/v1/organizations/:id", "Method": "PUT"},
			{"Path": "/api/v1/organizations/:id", "Method": "DELETE"},
		},
		"member:manage": {
			{"Path": "/api/v1/organizations/:id/members/invite", "Method": "POST"},
			{"Path": "/api/v1/organizations/:id/members", "Method": "GET"},
			{"Path": "/api/v1/organizations/:id/members/:userId", "Method": "PATCH"},
			{"Path": "/api/v1/organizations/:id/members/:userId", "Method": "DELETE"},
		},
		"presence:view": {
			{"Path": "/api/v1/organizations/:id/presence", "Method": "GET"},
		},
		"project:view": {
			{"Path": "/api/v1/projects", "Method": "GET"},
			{"Path": "/api/v1/projects/:id", "Method": "GET"},
		},
		"project:manage": {
			{"Path": "/api/v1/projects", "Method": "POST"},
			{"Path": "/api/v1/projects/:id", "Method": "PUT"},
			{"Path": "/api/v1/projects/:id", "Method": "DELETE"},
		},
		"role:view": {
			{"Path": "/api/v1/roles", "Method": "GET"},
			{"Path": "/api/v1/roles/search", "Method": "POST"},
		},
		"role:manage": {
			{"Path": "/api/v1/roles", "Method": "POST"},
			{"Path": "/api/v1/roles/:id", "Method": "PUT"},
			{"Path": "/api/v1/roles/:id", "Method": "DELETE"},
		},
		"permission:view": {
			{"Path": "/api/v1/permissions", "Method": "GET"},
			{"Path": "/api/v1/permissions/:role", "Method": "GET"},
			{"Path": "/api/v1/permissions/roles/:role/users", "Method": "GET"},
			{"Path": "/api/v1/permissions/:role/parents", "Method": "GET"},
			{"Path": "/api/v1/permissions/resources", "Method": "GET"},
			{"Path": "/api/v1/permissions/inheritance-tree", "Method": "GET"},
		},
		"permission:manage": {
			{"Path": "/api/v1/permissions/assign-role", "Method": "POST"},
			{"Path": "/api/v1/permissions/revoke-role", "Method": "DELETE"},
			{"Path": "/api/v1/permissions/grant", "Method": "POST"},
			{"Path": "/api/v1/permissions", "Method": "PUT"},
			{"Path": "/api/v1/permissions/revoke", "Method": "DELETE"},
			{"Path": "/api/v1/permissions/inheritance", "Method": "POST"},
			{"Path": "/api/v1/permissions/inheritance", "Method": "DELETE"},
			// Bulk Access Right assignment (new)
			{"Path": "/api/v1/permissions/assign-access-right", "Method": "POST"},
			{"Path": "/api/v1/permissions/revoke-access-right", "Method": "DELETE"},
			{"Path": "/api/v1/permissions/roles/:role/access-rights", "Method": "GET"},
		},
		"access:view": {
			{"Path": "/api/v1/access-rights", "Method": "GET"},
			{"Path": "/api/v1/access-rights/search", "Method": "POST"},
			{"Path": "/api/v1/endpoints/search", "Method": "POST"},
		},
		"access:manage": {
			{"Path": "/api/v1/access-rights", "Method": "POST"},
			{"Path": "/api/v1/access-rights/:id", "Method": "DELETE"},
			{"Path": "/api/v1/access-rights/link", "Method": "POST"},
			{"Path": "/api/v1/access-rights/unlink", "Method": "POST"},
			{"Path": "/api/v1/endpoints", "Method": "POST"},
			{"Path": "/api/v1/endpoints/:id", "Method": "DELETE"},
		},
		"audit:view": {
			{"Path": "/api/v1/audit-logs/search", "Method": "POST"},
			{"Path": "/api/v1/audit-logs/export", "Method": "GET"},
		},
	}

	roleToRights := map[string][]string{
		"role:admin": {
			"dashboard:view",
			"user:view", "user:manage",
			"role:view", "role:manage",
			"project:view", "project:manage",
			"org:view", "org:manage",
			"member:manage", "presence:view",
			"audit:view",
			"permission:view", "permission:manage",
			"access:view", "access:manage",
		},
		"role:user": {
			"dashboard:view", "project:view", "org:view", "presence:view",
		},
	}

	// 1. Seed Endpoints and AccessRights into DB
	for arName, eps := range accessMap {

		// Search for access right
		searchARPayload := map[string]interface{}{
			"filter": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "equals",
					"from": arName,
				},
			},
		}

		resp, err := doJSONRequest("POST", fmt.Sprintf("%s/access-rights/search", apiBaseURL), searchARPayload, token, http.StatusOK)

		var arID string
		if err == nil && resp["data"] != nil {
			if data, ok := resp["data"].([]interface{}); ok && len(data) > 0 {
				if item, ok := data[0].(map[string]interface{}); ok {
					if id, ok := item["id"].(string); ok {
						arID = id
					}
				}
			}
		} else if err != nil {
			log.Printf("Warning: search access right failed: %v", err)
		}

		if arID == "" {
			// Create it
			createARPayload := map[string]interface{}{
				"name": arName,
			}
			resp, err := doJSONRequest("POST", fmt.Sprintf("%s/access-rights", apiBaseURL), createARPayload, token, http.StatusCreated)
			if err != nil {
				log.Printf("Failed to create access right %s: %v", arName, err)
				continue
			}
			arID = ""
			if resp["data"] != nil {
				if item, ok := resp["data"].(map[string]interface{}); ok {
					if id, ok := item["id"].(string); ok {
						arID = id
					}
				}
			}
		}

		for _, ep := range eps {
			// Search for endpoint
			searchEpPayload := map[string]interface{}{
				"filter": map[string]interface{}{
					"path": map[string]interface{}{
						"type": "equals",
						"from": ep["Path"],
					},
					"method": map[string]interface{}{
						"type": "equals",
						"from": ep["Method"],
					},
				},
			}

			resp, err := doJSONRequest("POST", fmt.Sprintf("%s/endpoints/search", apiBaseURL), searchEpPayload, token, http.StatusOK)

			var epID string
			if err == nil && resp["data"] != nil {
				if data, ok := resp["data"].([]interface{}); ok && len(data) > 0 {
					if item, ok := data[0].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							epID = id
						}
					}
				}
			}

			if epID == "" {
				// Create endpoint
				createEpPayload := map[string]interface{}{
					"path":   ep["Path"],
					"method": ep["Method"],
				}
				resp, err := doJSONRequest("POST", fmt.Sprintf("%s/endpoints", apiBaseURL), createEpPayload, token, http.StatusCreated)
				if err != nil {
					log.Printf("Failed to create endpoint %s %s: %v", ep["Method"], ep["Path"], err)
					continue
				}
				epID = ""
				if resp["data"] != nil {
					if item, ok := resp["data"].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							epID = id
						}
					}
				}
			}

			// Link in DB
			linkPayload := map[string]interface{}{
				"access_right_id": arID,
				"endpoint_id":     epID,
			}
			_, _ = doJSONRequest("POST", fmt.Sprintf("%s/access-rights/link", apiBaseURL), linkPayload, token, http.StatusOK)
			// Ignore error on link if it's already linked
		}
	}

	// 3. Seed Casbin Inheritance: g, role, accessRight, global
	for roleName, rights := range roleToRights {
		for _, arName := range rights {
			// Need to fetch AR ID again, or rely assigning access right
			// find AR ID
			searchARPayload := map[string]interface{}{
				"filter": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "equals",
						"from": arName,
					},
				},
			}
			resp, err := doJSONRequest("POST", fmt.Sprintf("%s/access-rights/search", apiBaseURL), searchARPayload, token, http.StatusOK)
			var arID string
			if err == nil && resp["data"] != nil {
				if data, ok := resp["data"].([]interface{}); ok && len(data) > 0 {
					if item, ok := data[0].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							arID = id
						}
					}
				}
			}

			if arID != "" {
				assignPayload := map[string]interface{}{
					"role":            roleName,
					"access_right_id": arID,
					"domain":          "global",
				}
				_, err := doJSONRequest("POST", fmt.Sprintf("%s/permissions/assign-access-right", apiBaseURL), assignPayload, token, http.StatusOK)
				if err != nil {
					log.Printf("Warning: failed to assign %s to %s: %v", arName, roleName, err)
				}
			}
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

func seedOrganizationsUsersAndProjects(cfg *config.AppConfig, token string) {
	apiBaseURL := fmt.Sprintf("http://localhost:%d/api/v1", cfg.Server.Port)

	// 1. Create Default Organization
	orgPayload := map[string]interface{}{
		"name": "Default Organization",
		"slug": "default-org",
	}

	resp, err := doJSONRequest("POST", fmt.Sprintf("%s/organizations", apiBaseURL), orgPayload, token, http.StatusCreated)
	var orgID string
	if err != nil {
		log.Printf("Failed to create default organization (or it already exists): %v", err)
		// Try to fetch it
		resp, err = doJSONRequest("GET", fmt.Sprintf("%s/organizations/slug/default-org", apiBaseURL), nil, token, http.StatusOK)
		if err == nil && resp["data"] != nil {
			if item, ok := resp["data"].(map[string]interface{}); ok {
				if id, ok := item["id"].(string); ok {
					orgID = id
				}
			}
		}
	} else if resp["data"] != nil {
		if item, ok := resp["data"].(map[string]interface{}); ok {
			if id, ok := item["id"].(string); ok {
				orgID = id
			}
		}
	}

	if orgID == "" {
		log.Println("Could not determine default organization ID. Skipping user and project seeding.")
		return
	}
	log.Printf("Default Organization seeded/found with ID: %s", orgID)

	// 2. Assign Superadmin as member/owner of default org (often handled by creation, but let's ensure)
	// We might need to invite or just rely on the API auto-assigning the creator.
	// The access control model uses organizations/:id/members.
	// But first let's create the other users.

	usersToSeed := []map[string]string{
		{"username": "adminuser", "fullname": "Admin User", "email": "admin@example.com", "role": "role:admin"},
		{"username": "regularuser", "fullname": "Regular User", "email": "user@example.com", "role": "role:user"},
	}

	for _, u := range usersToSeed {
		userPayload := map[string]interface{}{
			"username": u["username"],
			"password": "Password0!",
			"fullname": u["fullname"],
			"email":    u["email"],
		}

		resp, err := doJSONRequest("POST", fmt.Sprintf("%s/users", apiBaseURL), userPayload, token, http.StatusCreated)
		var userID string
		if err != nil {
			log.Printf("User %s might already exist. Attempting to fetch...", u["username"])
			// Search user by username
			searchPayload := map[string]interface{}{
				"filter": map[string]interface{}{
					"username": map[string]interface{}{
						"type": "equals",
						"from": u["username"],
					},
				},
			}
			searchResp, searchErr := doJSONRequest("POST", fmt.Sprintf("%s/users/search", apiBaseURL), searchPayload, token, http.StatusOK)
			if searchErr == nil && searchResp["data"] != nil {
				if dataList, ok := searchResp["data"].([]interface{}); ok && len(dataList) > 0 {
					if item, ok := dataList[0].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							userID = id
						}
					}
				}
			}
		} else if resp["data"] != nil {
			if item, ok := resp["data"].(map[string]interface{}); ok {
				if id, ok := item["id"].(string); ok {
					userID = id
				}
			}
		}

		if userID != "" {
			log.Printf("User %s seeded/found with ID: %s", u["username"], userID)

			// Get Role ID
			roleSearchPayload := map[string]interface{}{
				"filter": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "equals",
						"from": u["role"],
					},
				},
			}
			roleResp, roleErr := doJSONRequest("POST", fmt.Sprintf("%s/roles/search", apiBaseURL), roleSearchPayload, token, http.StatusOK)
			var roleID string
			if roleErr == nil && roleResp["data"] != nil {
				if list, ok := roleResp["data"].([]interface{}); ok && len(list) > 0 {
					if item, ok := list[0].(map[string]interface{}); ok {
						if id, ok := item["id"].(string); ok {
							roleID = id
						}
					}
				}
			}

			if roleID != "" {
				invitePayload := map[string]interface{}{
					"email":   u["email"],
					"user_id": userID,
					"role_id": roleID,
				}
				_, err = doJSONRequest("POST", fmt.Sprintf("%s/organizations/%s/members/invite", apiBaseURL, orgID), invitePayload, token, http.StatusOK)
				if err != nil {
					log.Printf("Warning: Could not invite user %s to org %s: %v", u["username"], orgID, err)
				} else {
					log.Printf("User %s added to organization.", u["username"])
				}
			} else {
				log.Printf("Could not find role %s for user %s", u["role"], u["username"])
			}
		}
	}

	// 3. Create Default Project
	projectPayload := map[string]interface{}{
		"name":   "Sample E-Commerce App",
		"domain": "e-commerce",
	}

	// Create project requests typically go to /organizations/:id/projects if nested,
	// but this boilerplate routes Projects at /api/v1/projects and relies on the user's Context (which includes the Active Organization if set).
	// Because Superadmin is making the request, we need to ensure they are operating within an organization context.
	// Since X-Organization-Id might be required by middleware, let's inject it into `doJSONRequest` if needed,
	// or assume the generic `doJSONRequest` doesn't pass it and see if it falls back.
	// For now, let's just trace the path.

	// To reliably create a project for an org, we might need a modified request that passes the org ID header,
	// but let's try the standard POST and see if the controller handles it.
	doJSONRequestWithOrg(apiBaseURL, orgID, projectPayload, token)
}

func doJSONRequestWithOrg(apiBaseURL, orgID string, payload interface{}, token string) {
	var bodyReader *bytes.Buffer
	if payload != nil {
		payloadBytes, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(payloadBytes)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects", apiBaseURL), bodyReader)
	if err != nil {
		log.Printf("Failed to create project request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("X-Organization-Id", orgID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to execute project creation request: %v", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		log.Printf("API request to create project failed with status %d: %v", resp.StatusCode, errResp)
	} else {
		log.Println("Default Project created successfully.")
	}
}
