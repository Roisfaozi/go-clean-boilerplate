package tx

import (
	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

// ExtractSQLX extracts or wraps *sqlx.DB from *sqlx.DB or *gorm.DB for transition compatibility.
func ExtractSQLX(db any) *sqlx.DB {
	switch d := db.(type) {
	case *sqlx.DB:
		return d
	case *gorm.DB:
		sqlDB, err := d.DB()
		if err != nil {
			panic(err)
		}
		driverName := "mysql"
		if d.Name() != "" {
			driverName = d.Name()
		}
		return sqlx.NewDb(sqlDB, driverName)
	default:
		return nil
	}
}
