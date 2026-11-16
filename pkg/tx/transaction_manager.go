package tx

import (
	"context"

	"github.com/sirupsen/logrus"
)

type WithTransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewTransactionManager(db any, log *logrus.Logger) WithTransactionManager {
	return NewSQLXTransactionManager(ExtractSQLX(db), log)
}

func DBFromContext(ctx context.Context) (DBTX, bool) {
	return DBTXFromContext(ctx)
}
