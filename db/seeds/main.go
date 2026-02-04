package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
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

	log.Println("Database connected. Starting Go seeder...")

	superAdminRoleName := "role:superadmin"
	var count int64
	db.Model(&entity.Role{}).Where("name = ?", superAdminRoleName).Count(&count)
	if count == 0 {
		newRoleID, err := uuid.NewV7()
		if err != nil {
			log.Fatalf("Failed to generate UUID for role: %v", err)
		}
		newRole := entity.Role{
			ID:             newRoleID.String(),
			Name:           superAdminRoleName,
			OrganizationID: "global",
			Description:    "Super Administrator with full access",
			CreatedAt:      time.Now().UnixMilli(),
			UpdatedAt:      time.Now().UnixMilli(),
		}
		if err := db.Create(&newRole).Error; err != nil {
			log.Printf("Error creating role %s: %v", superAdminRoleName, err)
		} else {
			log.Printf("Role '%s' created.", superAdminRoleName)
		}
	} else {
		log.Printf("Role '%s' already exists.", superAdminRoleName)
	}

	var policyCount int64
	db.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ?", "p", superAdminRoleName, "*", "*", "*").Count(&policyCount)
	if policyCount == 0 {
		if err := db.Table("casbin_rule").Create(map[string]interface{}{
			"ptype": "p",
			"v0":    superAdminRoleName,
			"v1":    "*",
			"v2":    "*",
			"v3":    "*",
		}).Error; err != nil {
			log.Printf("Error creating policy for superadmin: %v", err)
		} else {
			log.Println("Policy 'p, role:superadmin, *, *, *' created.")
		}
	}

	adminUsername := "superadmin"
	adminEmail := "superadmin@example.com"

	// Check for environment variable for password
	adminPassword := os.Getenv("SUPERADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("SUPERADMIN_PASSWORD environment variable must be set. Exiting.")
	}

	var user userEntity.User
	result := db.Where("username = ?", adminUsername).First(&user)

	var userID string

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			newUserID, err := uuid.NewV7()
			if err != nil {
				log.Fatalf("Failed to generate UUID for user: %v", err)
			}
			userID = newUserID.String()

			newUser := userEntity.User{
				ID:        userID,
				Username:  adminUsername,
				Email:     adminEmail,
				Password:  string(hashedPwd),
				Name:      "Super Admin",
				Token:     "",
				CreatedAt: time.Now().UnixMilli(),
				UpdatedAt: time.Now().UnixMilli(),
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

	var gCount int64
	db.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "g", userID, superAdminRoleName, "global").Count(&gCount)
	if gCount == 0 {
		if err := db.Table("casbin_rule").Create(map[string]interface{}{
			"ptype": "g",
			"v0":    userID,
			"v1":    superAdminRoleName,
			"v2":    "global",
		}).Error; err != nil {
			log.Printf("Error assigning role: %v", err)
		} else {
			log.Printf("Role '%s' assigned to user '%s' in domain 'global' (ID: %s)", superAdminRoleName, adminUsername, userID)
		}
	} else {
		log.Printf("User '%s' already has role '%s' in domain 'global'", adminUsername, superAdminRoleName)
	}

	log.Println("Seeding process completed successfully.")
}
