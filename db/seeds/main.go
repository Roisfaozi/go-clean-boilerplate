package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/config"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. Load Config
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to DB
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

	log.Println("Database connected. Starting Go seeder...")

	// --- SEED ROLES ---
	// Ensure 'role:superadmin' exists
	superAdminRoleName := "role:superadmin"
	var count int64
	db.Model(&entity.Role{}).Where("name = ?", superAdminRoleName).Count(&count)
	if count == 0 {
		newRole := entity.Role{
			ID:          uuid.NewString(),
			Name:        superAdminRoleName,
			Description: "Super Administrator with full access",
			CreatedAt:   time.Now().UnixMilli(), // Fix: Use int64
			UpdatedAt:   time.Now().UnixMilli(), // Fix: Use int64
		}
		if err := db.Create(&newRole).Error; err != nil {
			log.Printf("Error creating role %s: %v", superAdminRoleName, err)
		} else {
			log.Printf("Role '%s' created.", superAdminRoleName)
		}
	} else {
		log.Printf("Role '%s' already exists.", superAdminRoleName)
	}

	// --- SEED CASBIN POLICIES ---
	// Ensure role:superadmin has full access (*)
	var policyCount int64
	db.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", superAdminRoleName, "*", "*").Count(&policyCount)
	if policyCount == 0 {
		if err := db.Table("casbin_rule").Create(map[string]interface{}{
			"ptype": "p",
			"v0":    superAdminRoleName,
			"v1":    "*",
			"v2":    "*",
		}).Error; err != nil {
			log.Printf("Error creating policy for superadmin: %v", err)
		} else {
			log.Println("Policy 'p, role:superadmin, *, *' created.")
		}
	}

	// --- SEED SUPER ADMIN USER ---
	adminUsername := "superadmin"
	adminEmail := "superadmin@example.com"
	adminPassword := "password123"

	// Check if user exists
	var user userEntity.User
	result := db.Where("username = ?", adminUsername).First(&user)

	var userID string

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new user
			hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			userID = uuid.NewString()
			newUser := userEntity.User{
				ID:       userID,
				Username: adminUsername,
				Email:    adminEmail,
				Password: string(hashedPwd),
				Name:     "Super Admin",
				Token:    "",
			}
			if err := db.Create(&newUser).Error; err != nil {
				log.Fatalf("Failed to create user: %v", err)
			}
			log.Printf("User '%s' created with ID: %s", adminUsername, userID)
		} else {
			log.Fatalf("Error checking user: %v", result.Error)
		}
	} else {
		log.Printf("User '%s' already exists with ID: %s", adminUsername, user.ID)
		userID = user.ID
	}

	// --- ASSIGN ROLE TO USER (Grouping Policy) ---
	// Check if grouping policy exists
	var gCount int64
	db.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ?", "g", userID, superAdminRoleName).Count(&gCount)
	if gCount == 0 {
		if err := db.Table("casbin_rule").Create(map[string]interface{}{
			"ptype": "g",
			"v0":    userID,
			"v1":    superAdminRoleName,
		}).Error; err != nil {
			log.Printf("Error assigning role: %v", err)
		} else {
			log.Printf("Role '%s' assigned to user '%s' (ID: %s)", superAdminRoleName, adminUsername, userID)
		}
	} else {
		log.Printf("User '%s' already has role '%s'", adminUsername, superAdminRoleName)
	}

	log.Println("Seeding process completed successfully.")
}
