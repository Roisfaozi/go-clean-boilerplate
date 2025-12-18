package repository_test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTestDB in-memory for audit tests
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&entity.AuditLog{})
	require.NoError(t, err)
	return db
}

func TestAuditRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	repo := repository.NewAuditRepository(db, logger)

	t.Run("Success - Create Full Audit Log", func(t *testing.T) {
		ctx := context.Background()

		oldVal := `{"name": "Old Name"}`
		newVal := `{"name": "New Name"}`

		log := &entity.AuditLog{
			UserID:    "user-123",
			Action:    "UPDATE",
			Entity:    "User",
			EntityID:  "target-user-id",
			OldValues: oldVal,
			NewValues: newVal,
			IPAddress: "127.0.0.1",
			UserAgent: "Mozilla/5.0",
		}

		err := repo.Create(ctx, log)
		assert.NoError(t, err)
		assert.NotEmpty(t, log.ID, "ID should be auto-generated")
		assert.NotZero(t, log.CreatedAt, "CreatedAt should be set")

		// Verify in DB
		var storedLog entity.AuditLog
		err = db.First(&storedLog, "id = ?", log.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "user-123", storedLog.UserID)
		assert.Equal(t, oldVal, storedLog.OldValues)
		assert.Equal(t, "UPDATE", storedLog.Action)
	})

	t.Run("Success - Create Partial Log (No JSON values)", func(t *testing.T) {
		ctx := context.Background()
		log := &entity.AuditLog{
			UserID:   "user-456",
			Action:   "LOGIN",
			Entity:   "Auth",
			EntityID: "session-id",
		}

		err := repo.Create(ctx, log)
		assert.NoError(t, err)

		var storedLog entity.AuditLog
		db.First(&storedLog, "id = ?", log.ID)
		assert.Empty(t, storedLog.OldValues)
		assert.Equal(t, "LOGIN", storedLog.Action)
	})

	t.Run("Edge - JSON Injection Safety", func(t *testing.T) {
		// Ensure storing complex/malformed JSON strings doesn't break SQL
		// Since we treat it as a string field in GORM (type:json is just a hint for some DBs),
		// it should be safe from SQL injection due to GORM parameterization.
		ctx := context.Background()
		malformedJSON := `{"key": "value" -- broken query'; DROP TABLE users; --`

		log := &entity.AuditLog{
			UserID:    "hacker",
			Action:    "HACK",
			Entity:    "System",
			EntityID:  "1",
			OldValues: malformedJSON, // Should be stored literally
		}

		err := repo.Create(ctx, log)
		assert.NoError(t, err)

		var storedLog entity.AuditLog
		db.First(&storedLog, "id = ?", log.ID)
		assert.Equal(t, malformedJSON, storedLog.OldValues)
	})

	t.Run("Edge - Very Long UserAgent", func(t *testing.T) {
		ctx := context.Background()
		// Entity definition has UserAgent varchar(255)
		// If we send more, it depends on DB strict mode. SQLite might allow it, but we want to check behavior.
		longUA := ""
		for i := 0; i < 300; i++ {
			longUA += "a"
		}

		log := &entity.AuditLog{
			UserID:    "user-long",
			Action:    "TEST",
			Entity:    "Test",
			EntityID:  "1",
			UserAgent: longUA,
		}

		err := repo.Create(ctx, log)
		assert.NoError(t, err) // SQLite usually allows this unless STRICT table.
		
		var storedLog entity.AuditLog
		db.First(&storedLog, "id = ?", log.ID)
		// Verifying it stored the long value (showing GORM/SQLite flexibility)
		assert.Equal(t, longUA, storedLog.UserAgent)
	})
}
