package tx

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type sqlxTxKey struct{}

// SQLXTransactionManager manages transactions using *sqlx.DB.
type SQLXTransactionManager struct {
	DB  *sqlx.DB
	Log *logrus.Logger
}

func NewSQLXTransactionManager(db *sqlx.DB, log *logrus.Logger) WithTransactionManager {
	return &SQLXTransactionManager{
		DB:  db,
		Log: log,
	}
}

func (tm *SQLXTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, sqlxTxKey{}, tx)

	defer func() {
		if r := recover(); r != nil {
			if tm.Log != nil {
				tm.Log.Errorf("panic recovered in transaction: %v", r)
			}
			if rbErr := tx.Rollback(); rbErr != nil && tm.Log != nil {
				tm.Log.Errorf("failed to rollback transaction after panic: %v", rbErr)
			}
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction rollback error: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DBTXFromContext returns the DBTX interface stored in context (if in transaction),
// or false if not in a transaction.
func DBTXFromContext(ctx context.Context) (DBTX, bool) {
	tx, ok := ctx.Value(sqlxTxKey{}).(*sqlx.Tx)
	if ok && tx != nil {
		return tx, true
	}
	return nil, false
}
