//go:build integration
// +build integration

package scenarios

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration000029_BackfillsLegacyRoleNames verifies that migration 000029
// converts legacy role-name representations in organization_members and
// invitation_tokens into role IDs (and reverts on down).
func TestMigration000029_BackfillsLegacyRoleNames(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	migDB := setup.FreshMigrationDB(t, "test_db_migration_000029")
	setup.ApplyMigrationsUpTo(t, migDB, 28)

	orgID := uuid.NewString()
	userID := uuid.NewString()

	require.NoError(t, migDB.Exec(
		"INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, 'active', 0, 0, 0)",
		orgID, "Migration Org", "migration-org", "system").Error)

	require.NoError(t, migDB.Exec(
		"INSERT INTO users (id, username, email, password, name, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, 0, 0, 0)",
		userID, "miguser", "mig@example.com", "x", "Mig User").Error)

	// Legacy representation: role_id stores the role NAME, invitation_tokens uses column `role`.
	require.NoError(t, migDB.Exec(
		"INSERT INTO organization_members (id, organization_id, user_id, role_id, status) VALUES (?, ?, ?, ?, 'active')",
		uuid.NewString(), orgID, userID, "role:admin").Error)

	require.NoError(t, migDB.Exec(
		"INSERT INTO invitation_tokens (id, organization_id, email, token, role, expires_at, created_at) VALUES (?, ?, ?, ?, ?, 9999999999999, 1)",
		uuid.NewString(), orgID, "mig@example.com", "legacy_token", "role:admin").Error)

	var adminRoleID string
	require.NoError(t, migDB.Raw("SELECT id FROM roles WHERE name = 'role:admin'").Scan(&adminRoleID).Error)
	require.NotEmpty(t, adminRoleID, "role:admin seed must exist before migration 000029")

	// Apply up migration
	setup.ApplyMigrationFile(t, migDB, 29, "up")

	var memberRoleID string
	require.NoError(t, migDB.Raw("SELECT role_id FROM organization_members WHERE user_id = ?", userID).Scan(&memberRoleID).Error)
	assert.Equal(t, adminRoleID, memberRoleID, "legacy role name must be backfilled to role UUID in organization_members")

	var invitationRoleID string
	require.NoError(t, migDB.Raw("SELECT role_id FROM invitation_tokens WHERE token = 'legacy_token'").Scan(&invitationRoleID).Error)
	assert.Equal(t, adminRoleID, invitationRoleID, "legacy role name must be backfilled to role UUID in invitation_tokens")

	var fkCount int64
	require.NoError(t, migDB.Raw(
		`SELECT COUNT(*) FROM information_schema.REFERENTIAL_CONSTRAINTS
		 WHERE CONSTRAINT_SCHEMA = DATABASE()
		 AND CONSTRAINT_NAME IN ('fk_organization_members_role', 'fk_invitation_tokens_role')`).
		Scan(&fkCount).Error)
	assert.Equal(t, int64(2), fkCount, "both role foreign keys must be created by migration 000029")

	// Apply down migration: restore legacy representation
	setup.ApplyMigrationFile(t, migDB, 29, "down")

	var memberRoleReverted string
	require.NoError(t, migDB.Raw("SELECT role_id FROM organization_members WHERE user_id = ?", userID).Scan(&memberRoleReverted).Error)
	assert.Equal(t, "role:admin", memberRoleReverted, "down migration must restore role name in organization_members")

	var invitationRoleReverted string
	require.NoError(t, migDB.Raw("SELECT `role` FROM invitation_tokens WHERE token = 'legacy_token'").Scan(&invitationRoleReverted).Error)
	assert.Equal(t, "role:admin", invitationRoleReverted, "down migration must restore role name column and value in invitation_tokens")

	var fkReverted int64
	require.NoError(t, migDB.Raw(
		`SELECT COUNT(*) FROM information_schema.REFERENTIAL_CONSTRAINTS
		 WHERE CONSTRAINT_SCHEMA = DATABASE()
		 AND CONSTRAINT_NAME IN ('fk_organization_members_role', 'fk_invitation_tokens_role')`).
		Scan(&fkReverted).Error)
	assert.Equal(t, int64(0), fkReverted, "down migration must drop both role FKs")
}

// TestMigration000029_PreservesUUIDRoleIDs ensures already-correct rows (role IDs)
// are untouched by the up migration backfill.
func TestMigration000029_PreservesUUIDRoleIDs(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	migDB := setup.FreshMigrationDB(t, "test_db_migration_000029_preserve")
	setup.ApplyMigrationsUpTo(t, migDB, 28)

	orgID := uuid.NewString()
	userID := uuid.NewString()

	require.NoError(t, migDB.Exec(
		"INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, 'active', 0, 0, 0)",
		orgID, "Preserve Org", "preserve-org", "system").Error)
	require.NoError(t, migDB.Exec(
		"INSERT INTO users (id, username, email, password, name, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, 0, 0, 0)",
		userID, "preserveuser", "preserve@example.com", "Password", "Preserve User").Error)

	var adminRoleID string
	require.NoError(t, migDB.Raw("SELECT id FROM roles WHERE name = 'role:admin'").Scan(&adminRoleID).Error)
	require.NotEmpty(t, adminRoleID)

	// Row already using the role UUID (production representation).
	require.NoError(t, migDB.Exec(
		"INSERT INTO organization_members (id, organization_id, user_id, role_id, status) VALUES (?, ?, ?, ?, 'active')",
		uuid.NewString(), orgID, userID, adminRoleID).Error)

	setup.ApplyMigrationFile(t, migDB, 29, "up")

	var memberRoleID string
	require.NoError(t, migDB.Raw("SELECT role_id FROM organization_members WHERE user_id = ?", userID).Scan(&memberRoleID).Error)
	assert.Equal(t, adminRoleID, memberRoleID, "already-UUID role_id must not change during backfill")
}
