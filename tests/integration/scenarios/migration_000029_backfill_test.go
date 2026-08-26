//go:build integration
// +build integration

package scenarios

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const migration000029FKQuery = `SELECT COUNT(*) FROM information_schema.REFERENTIAL_CONSTRAINTS
	 WHERE CONSTRAINT_SCHEMA = DATABASE()
	 AND CONSTRAINT_NAME IN ('fk_organization_members_role', 'fk_invitation_tokens_role')`

type migrationSeedResult struct {
	orgID       string
	userID      string
	adminRoleID string
}

// seedLegacyMigrationData applies migrations up to 28 and inserts a member and
// invitation row using the given role reference (name for legacy, id for modern).
func seedLegacyMigrationData(t *testing.T, db *gorm.DB, roleRefIsName bool) migrationSeedResult {
	t.Helper()

	setup.ApplyMigrationsUpTo(t, db, 28)

	orgID := uuid.NewString()
	userID := uuid.NewString()

	require.NoError(t, db.Exec(
		"INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, 'active', 0, 0, 0)",
		orgID, "Migration Org", "migration-org-"+orgID[:8], "system").Error)

	require.NoError(t, db.Exec(
		"INSERT INTO users (id, username, email, password, name, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, 0, 0, 0)",
		userID, "miguser-"+userID[:8], "mig-"+userID[:8]+"@example.com", "x", "Mig User").Error)

	var adminRoleID string
	require.NoError(t, db.Raw("SELECT id FROM roles WHERE name = 'role:admin'").Scan(&adminRoleID).Error)
	require.NotEmpty(t, adminRoleID, "role:admin seed must exist before migration 000029")

	roleRef := adminRoleID
	if roleRefIsName {
		roleRef = "role:admin"
	}

	require.NoError(t, db.Exec(
		"INSERT INTO organization_members (id, organization_id, user_id, role_id, status) VALUES (?, ?, ?, ?, 'active')",
		uuid.NewString(), orgID, userID, roleRef).Error)

	require.NoError(t, db.Exec(
		"INSERT INTO invitation_tokens (id, organization_id, email, token, role, expires_at, created_at) VALUES (?, ?, ?, ?, ?, 9999999999999, 1)",
		uuid.NewString(), orgID, "mig-"+userID[:8]+"@example.com", "legacy_token_"+userID[:8], roleRef).Error)

	return migrationSeedResult{orgID: orgID, userID: userID, adminRoleID: adminRoleID}
}

// TestMigration000029_Backfill verifies the up migration converts legacy role
// names into role IDs while leaving already-correct rows untouched.
func TestMigration000029_Backfill(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tests := []struct {
		name          string
		dbName        string
		roleRefIsName bool
		description   string
	}{
		{
			name:          "backfills legacy role names to role ids",
			dbName:        "test_db_migration_000029_legacy",
			roleRefIsName: true,
			description:   "legacy role name must be backfilled to role UUID",
		},
		{
			name:          "preserves rows that already store role ids",
			dbName:        "test_db_migration_000029_modern",
			roleRefIsName: false,
			description:   "already-UUID role_id must not change during backfill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migDB := setup.FreshMigrationDB(t, tt.dbName)
			seed := seedLegacyMigrationData(t, migDB, tt.roleRefIsName)

			setup.ApplyMigrationFile(t, migDB, 29, "up")

			var memberRoleID string
			require.NoError(t, migDB.Raw("SELECT role_id FROM organization_members WHERE user_id = ?", seed.userID).Scan(&memberRoleID).Error)
			assert.Equal(t, seed.adminRoleID, memberRoleID, tt.description)

			var invitationRoleID string
			require.NoError(t, migDB.Raw("SELECT role_id FROM invitation_tokens WHERE organization_id = ?", seed.orgID).Scan(&invitationRoleID).Error)
			assert.Equal(t, seed.adminRoleID, invitationRoleID, "invitation_tokens.role_id must hold the role UUID after migration")

			var fkCount int64
			require.NoError(t, migDB.Raw(migration000029FKQuery).Scan(&fkCount).Error)
			assert.Equal(t, int64(2), fkCount, "both role foreign keys must be created by migration 000029")
		})
	}
}

// TestMigration000029_Down verifies the down migration restores the legacy
// column name and role-name representation, and drops the foreign keys.
func TestMigration000029_Down(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	migDB := setup.FreshMigrationDB(t, "test_db_migration_000029_down")
	seed := seedLegacyMigrationData(t, migDB, true)

	setup.ApplyMigrationFile(t, migDB, 29, "up")
	setup.ApplyMigrationFile(t, migDB, 29, "down")

	tests := []struct {
		name     string
		query    string
		args     []interface{}
		expected string
		message  string
	}{
		{
			name:     "organization_members role_id reverts to role name",
			query:    "SELECT role_id FROM organization_members WHERE user_id = ?",
			args:     []interface{}{seed.userID},
			expected: "role:admin",
			message:  "down migration must restore role name in organization_members",
		},
		{
			name:     "invitation_tokens role column reverts to role name",
			query:    "SELECT `role` FROM invitation_tokens WHERE organization_id = ?",
			args:     []interface{}{seed.orgID},
			expected: "role:admin",
			message:  "down migration must restore role column and name in invitation_tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			require.NoError(t, migDB.Raw(tt.query, tt.args...).Scan(&got).Error)
			assert.Equal(t, tt.expected, got, tt.message)
		})
	}

	t.Run("both role foreign keys are dropped", func(t *testing.T) {
		var fkCount int64
		require.NoError(t, migDB.Raw(migration000029FKQuery).Scan(&fkCount).Error)
		assert.Equal(t, int64(0), fkCount, "down migration must drop both role FKs")
	})
}
