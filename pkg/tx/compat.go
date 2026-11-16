package tx

import (
	"github.com/jmoiron/sqlx"
)

// ExtractSQLX extracts *sqlx.DB if db is a *sqlx.DB, and returns nil otherwise.
func ExtractSQLX(db any) *sqlx.DB {
	if d, ok := db.(*sqlx.DB); ok {
		return d
	}
	return nil
}
