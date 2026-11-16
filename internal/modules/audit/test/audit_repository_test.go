package test_test

import (
	"context"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id TEXT PRIMARY KEY,
		organization_id TEXT,
		user_id TEXT,
		action TEXT,
		entity TEXT,
		entity_id TEXT,
		old_values TEXT,
		new_values TEXT,
		ip_address TEXT,
		user_agent TEXT,
		created_at INTEGER DEFAULT 0,
		deleted_at INTEGER DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS audit_outbox (
		id TEXT PRIMARY KEY,
		organization_id TEXT,
		user_id TEXT,
		action TEXT,
		entity TEXT,
		entity_id TEXT,
		old_values TEXT,
		new_values TEXT,
		ip_address TEXT,
		user_agent TEXT,
		status TEXT,
		retry_count INTEGER DEFAULT 0,
		last_error TEXT,
		created_at INTEGER DEFAULT 0,
		updated_at INTEGER DEFAULT 0
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)
	return db
}

func TestAuditRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
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
		err = db.Get(&storedLog, "SELECT id, organization_id, user_id, action, entity, entity_id, old_values, new_values, ip_address, user_agent, created_at, deleted_at FROM audit_logs WHERE id = ?", log.ID)
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
		err = db.Get(&storedLog, "SELECT id, organization_id, user_id, action, entity, entity_id, old_values, new_values, ip_address, user_agent, created_at, deleted_at FROM audit_logs WHERE id = ?", log.ID)
		assert.NoError(t, err)
		assert.Empty(t, storedLog.OldValues)
		assert.Equal(t, "LOGIN", storedLog.Action)
	})

	t.Run("Edge - JSON Injection Safety", func(t *testing.T) {
		// Ensure storing complex/malformed JSON strings doesn't break SQL
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
		err = db.Get(&storedLog, "SELECT id, organization_id, user_id, action, entity, entity_id, old_values, new_values, ip_address, user_agent, created_at, deleted_at FROM audit_logs WHERE id = ?", log.ID)
		assert.NoError(t, err)
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
		err = db.Get(&storedLog, "SELECT id, organization_id, user_id, action, entity, entity_id, old_values, new_values, ip_address, user_agent, created_at, deleted_at FROM audit_logs WHERE id = ?", log.ID)
		assert.NoError(t, err)
		// Verifying it stored the long value
		assert.Equal(t, longUA, storedLog.UserAgent)
	})
}
