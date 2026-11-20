package tx

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type WithTransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewTransactionManager(db *sqlx.DB, log *logrus.Logger) WithTransactionManager {
	return NewSQLXTransactionManager(db, log)
}

func DBFromContext(ctx context.Context) (DBTX, bool) {
	return DBTXFromContext(ctx)
}
