package tx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLXTransactionManager_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	tm := tx.NewSQLXTransactionManager(sqlxDB, logger)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
		dbtx, ok := tx.DBTXFromContext(ctx)
		assert.True(t, ok)
		assert.NotNil(t, dbtx)
		return nil
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLXTransactionManager_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	tm := tx.NewSQLXTransactionManager(sqlxDB, logger)

	mock.ExpectBegin()
	mock.ExpectRollback()

	expectedErr := errors.New("business logic error")
	err = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
		dbtx, ok := tx.DBTXFromContext(ctx)
		assert.True(t, ok)
		assert.NotNil(t, dbtx)
		return expectedErr
	})

	assert.ErrorIs(t, err, expectedErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLXTransactionManager_Panic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	tm := tx.NewSQLXTransactionManager(sqlxDB, logger)

	mock.ExpectBegin()
	mock.ExpectRollback()

	assert.PanicsWithValue(t, "something went terribly wrong", func() {
		_ = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
			panic("something went terribly wrong")
		})
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
