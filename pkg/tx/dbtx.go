package tx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DBTX abstracts database operations common to *sqlx.DB, *sqlx.Tx, and mocks.
type DBTX interface {
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
}
