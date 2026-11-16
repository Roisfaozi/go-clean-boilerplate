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

type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *NoOpWriter) Levels() []logrus.Level {
	return logrus.AllLevels
}

func TestTransactionManager_WithinTransaction_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	tm := tx.NewTransactionManager(sqlxDB, logger)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").WithArgs("Alice").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
		txDB, ok := tx.DBFromContext(ctx)
		assert.True(t, ok)
		assert.NotNil(t, txDB)

		_, err := txDB.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "Alice")
		return err
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_WithinTransaction_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	tm := tx.NewTransactionManager(sqlxDB, logger)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").WithArgs("Bob").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	err = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
		txDB, ok := tx.DBFromContext(ctx)
		assert.True(t, ok)
		assert.NotNil(t, txDB)

		_, _ = txDB.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "Bob")

		return errors.New("simulated error")
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_WithinTransaction_PanicRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	tm := tx.NewTransactionManager(sqlxDB, logger)

	mock.ExpectBegin()
	mock.ExpectRollback()

	assert.PanicsWithValue(t, "panic inside transaction", func() {
		_ = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
			panic("panic inside transaction")
		})
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
