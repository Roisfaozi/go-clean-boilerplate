package casbinadapter_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/casbinadapter"
	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testModelText = `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act
`

func setupAdapterMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *casbinadapter.SQLXCasbinAdapter) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	adapter := casbinadapter.NewSQLXCasbinAdapter(sqlxDB)
	return sqlxDB, mock, adapter
}

func TestSQLXCasbinAdapter_LoadPolicy(t *testing.T) {
	db, mock, adapter := setupAdapterMock(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "ptype", "v0", "v1", "v2", "v3", "v4", "v5"}).
		AddRow(1, "p", "admin", "org1", "/api/v1/projects", "GET", nil, nil).
		AddRow(2, "g", "alice", "admin", "org1", nil, nil, nil)

	mock.ExpectQuery("SELECT (.+) FROM casbin_rule").
		WillReturnRows(rows)

	m, err := model.NewModelFromString(testModelText)
	require.NoError(t, err)

	err = adapter.LoadPolicy(m)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLXCasbinAdapter_LoadPolicyIgnoresTrailingEmptyColumns(t *testing.T) {
	db, mock, adapter := setupAdapterMock(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "ptype", "v0", "v1", "v2", "v3", "v4", "v5"}).
		AddRow(1, "g", "alice", "admin", "org1", "", "", nil)
	mock.ExpectQuery("SELECT (.+) FROM casbin_rule").WillReturnRows(rows)

	m, err := model.NewModelFromString(testModelText)
	require.NoError(t, err)

	require.NoError(t, adapter.LoadPolicy(m))
	e, err := casbin.NewEnforcer(m)
	require.NoError(t, err)

	allowed, err := e.Enforce("alice", "org1", "/api/v1/projects", "GET")
	require.NoError(t, err)
	assert.False(t, allowed)
	policy, err := e.GetGroupingPolicy()
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"alice", "admin", "org1"}}, policy)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLXCasbinAdapter_AddPolicy(t *testing.T) {
	db, mock, adapter := setupAdapterMock(t)
	defer db.Close()

	mock.ExpectExec("INSERT INTO casbin_rule").
		WithArgs("p", "admin", "org1", "/api/v1/projects", "GET").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := adapter.AddPolicy("p", "p", []string{"admin", "org1", "/api/v1/projects", "GET"})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLXCasbinAdapter_RemovePolicy(t *testing.T) {
	db, mock, adapter := setupAdapterMock(t)
	defer db.Close()

	mock.ExpectExec("DELETE FROM casbin_rule WHERE").
		WithArgs("p", "admin", "org1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := adapter.RemovePolicy("p", "p", []string{"admin", "org1"})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLXCasbinAdapter_Enforcement(t *testing.T) {
	db, mock, adapter := setupAdapterMock(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "ptype", "v0", "v1", "v2", "v3", "v4", "v5"}).
		AddRow(1, "p", "admin", "org1", "/api/v1/projects", "GET", nil, nil).
		AddRow(2, "g", "alice", "admin", "org1", nil, nil, nil)

	mock.ExpectQuery("SELECT (.+) FROM casbin_rule").
		WillReturnRows(rows)

	m, err := model.NewModelFromString(testModelText)
	require.NoError(t, err)

	e, err := casbin.NewEnforcer(m, adapter)
	require.NoError(t, err)

	ok, err := e.Enforce("alice", "org1", "/api/v1/projects", "GET")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = e.Enforce("bob", "org1", "/api/v1/projects", "GET")
	require.NoError(t, err)
	assert.False(t, ok)

	assert.NoError(t, mock.ExpectationsWereMet())
}
