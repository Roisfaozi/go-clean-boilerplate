package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestProjectRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewProjectRepository(db)

	p := &entity.Project{
		OrganizationID: "org-1",
		UserID:         "user-1",
		Name:           "My Project",
		Domain:         "example.com",
	}

	mock.ExpectExec("INSERT INTO projects").
		WithArgs(sqlmock.AnyArg(), "org-1", "user-1", "My Project", "example.com", "active", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), p)
	assert.NoError(t, err)
	assert.NotEmpty(t, p.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_GetByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewProjectRepository(db)

	rows := sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "domain", "status", "created_at", "updated_at", "deleted_at"}).
		AddRow("proj-1", "org-1", "user-1", "My Project", "example.com", "active", time.Now().UnixMilli(), time.Now().UnixMilli(), 0)

	mock.ExpectQuery("SELECT (.+) FROM projects WHERE (.+)").
		WithArgs("proj-1").
		WillReturnRows(rows)

	res, err := repo.GetByID(context.Background(), "proj-1")
	require.NoError(t, err)
	assert.Equal(t, "proj-1", res.ID)
	assert.Equal(t, "My Project", res.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewProjectRepository(db)

	rows := sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "domain", "status", "created_at", "updated_at", "deleted_at"})

	mock.ExpectQuery("SELECT (.+) FROM projects WHERE (.+)").
		WithArgs("proj-notfound").
		WillReturnRows(rows)

	res, err := repo.GetByID(context.Background(), "proj-notfound")
	assert.ErrorIs(t, err, exception.ErrNotFound)
	assert.Nil(t, res)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_GetByID_TenantScoped(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewProjectRepository(db)
	ctx := database.SetOrganizationContext(context.Background(), "org-1")

	rows := sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "domain", "status", "created_at", "updated_at", "deleted_at"}).
		AddRow("proj-1", "org-1", "user-1", "My Project", "example.com", "active", time.Now().UnixMilli(), time.Now().UnixMilli(), 0)

	mock.ExpectQuery("SELECT (.+) FROM projects WHERE (.+)").
		WithArgs("proj-1", "org-1").
		WillReturnRows(rows)

	res, err := repo.GetByID(ctx, "proj-1")
	require.NoError(t, err)
	assert.Equal(t, "proj-1", res.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_GetByOrgID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewProjectRepository(db)

	rows := sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "domain", "status", "created_at", "updated_at", "deleted_at"}).
		AddRow("proj-1", "org-1", "user-1", "Project 1", "example.com", "active", 1000, 1000, 0).
		AddRow("proj-2", "org-1", "user-2", "Project 2", "test.com", "active", 2000, 2000, 0)

	mock.ExpectQuery("SELECT (.+) FROM projects WHERE (.+)").
		WithArgs("org-1").
		WillReturnRows(rows)

	list, err := repo.GetByOrgID(context.Background(), "org-1")
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewProjectRepository(db)

	p := &entity.Project{
		ID:     "proj-1",
		Name:   "Updated Project",
		Domain: "newdomain.com",
		Status: "inactive",
	}

	mock.ExpectExec("UPDATE projects SET (.+) WHERE (.+)").
		WithArgs("Updated Project", "newdomain.com", "inactive", sqlmock.AnyArg(), "proj-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), p)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewProjectRepository(db)

	mock.ExpectExec("UPDATE projects SET deleted_at = \\? WHERE (.+)").
		WithArgs(sqlmock.AnyArg(), "proj-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), "proj-1")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepository_CountByUserID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewProjectRepository(db)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(3)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM projects WHERE user_id = \\? AND deleted_at = 0").
		WithArgs("user-1").
		WillReturnRows(rows)

	count, err := repo.CountByUserID(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}
