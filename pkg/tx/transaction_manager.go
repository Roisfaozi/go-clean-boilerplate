package tx

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WithTransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type TransactionManager struct {
	DB  *gorm.DB
	Log *logrus.Logger
}

func NewTransactionManager(db *gorm.DB, log *logrus.Logger) WithTransactionManager {
	return &TransactionManager{
		DB:  db,
		Log: log,
	}
}

type txKey struct{}

func (tm *TransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx := tm.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	defer func() {
		if r := recover(); r != nil {
			tm.Log.Errorf("panic recovered: %v", r)
			if err := tx.Rollback().Error; err != nil {
				tm.Log.Errorf("failed to rollback transaction: %v", err)
			}
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return errors.New("rollback error")
		}
		return err
	}

	return tx.Commit().Error
}

func DBFromContext(ctx context.Context) (*gorm.DB, bool) {
	db, ok := ctx.Value(txKey{}).(*gorm.DB)
	return db, ok
}
