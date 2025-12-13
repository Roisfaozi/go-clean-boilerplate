package tx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/glebarez/sqlite" // Pure Go SQLite
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// NoOpWriter is a logrus.Hook that discards all log entries.
type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *NoOpWriter) Levels() []logrus.Level {
	return logrus.AllLevels
}

type User struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func setupTestDB(t *testing.T) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)
	return db, nil
}

func TestTransactionManager_WithinTransaction_Commit(t *testing.T) {
	db, err := setupTestDB(t)
	assert.NoError(t, err)

	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	tm := tx.NewTransactionManager(db, logger)

	err = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
		txDB, ok := tx.DBFromContext(ctx)
		assert.True(t, ok)

		// Create user within transaction
		if err := txDB.Create(&User{Name: "Alice"}).Error; err != nil {
			return err
		}
		return nil
	})

	assert.NoError(t, err)

	// Verify data is committed
	var count int64
	db.Model(&User{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestTransactionManager_WithinTransaction_Rollback(t *testing.T) {
	db, err := setupTestDB(t)
	assert.NoError(t, err)

	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	tm := tx.NewTransactionManager(db, logger)

	err = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
		txDB, ok := tx.DBFromContext(ctx)
		assert.True(t, ok)

		txDB.Create(&User{Name: "Bob"})

		// Return error to trigger rollback
		return errors.New("simulated error")
	})

	assert.Error(t, err)

	// Verify data is NOT committed
	var count int64
	db.Model(&User{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestTransactionManager_WithinTransaction_PanicRollback(t *testing.T) {
	db, err := setupTestDB(t)
	assert.NoError(t, err)

	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	tm := tx.NewTransactionManager(db, logger)

	assert.PanicsWithValue(t, "panic inside transaction", func() {
		_ = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
			panic("panic inside transaction")
		})
	})

	// Verify data is NOT committed
	var count int64
	db.Model(&User{}).Count(&count)
	assert.Equal(t, int64(0), count)
}
