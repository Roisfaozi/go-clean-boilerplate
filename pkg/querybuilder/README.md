# Secure Dynamic Query Builder

This package provides a secure, reusable dynamic query builder for GORM and PostgreSQL.

## Features

- **Secure**: Uses parameterized queries to prevent SQL injection.
- **Dynamic Filtering**: Supports `contains`, `equals`, `in`, `inRange`, `startsWith`, `endsWith`, `lessThan`, `greaterThan`, `isNull`, etc.
- **Dynamic Sorting**: Supports sorting by multiple columns.
- **Soft Delete Support**: Automatically handles `DeletedAt` or `DeletedBy`.
- **Reflection-based**: Automatically maps struct fields to database columns using `gorm` tags or snake_case fallback.

## Usage

### 1. Define Filter Structure

You can use the `DynamicFilter` struct directly or embed it in your request models.

```go
import "github.com/Roisfaozi/casbin-db/internal/utils/querybuilder"

// Example request payload
filterPayload := &querybuilder.DynamicFilter{
    Filter: map[string]querybuilder.Filter{
        "Name": {Type: "contains", From: "Admin"},
        "Age":  {Type: "greaterThan", From: 18},
    },
    Sort: &[]querybuilder.SortModel{
        {ColId: "Name", Sort: "asc"},
    },
}
```

### 2. Use in Repository

See `internal/modules/user/repository/user_repository.go` for a live example.

```go
func (r *userRepositoryData) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.User, error) {
	var users []*entity.User
	query := r.db.WithContext(ctx)

	// 1. Generate WHERE clause and Args
	where, args, warnings, err := querybuilder.GenerateDynamicQuery[entity.User](filter)
	if err != nil {
		return nil, err
	}
	
    // Log warnings if any field was ignored
    if len(warnings) > 0 {
        r.log.Warnf("QueryBuilder warnings: %v", warnings)
    }

	if where != "" {
		query = query.Where(where, args...)
	}

	// 2. Generate ORDER BY clause
	sort, err := querybuilder.GenerateDynamicSort[entity.User](filter)
	if err != nil {
		return nil, err
	}
	if sort != "" {
		query = query.Order(sort)
	}

	// 3. Execute
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
```

### 3. Supported Operators

| Operator | SQL Example | Notes |
| :--- | :--- | :--- |
| `contains` | `col ILIKE ?` | `%value%` |
| `notContains` | `col NOT ILIKE ?` | `%value%` |
| `startsWith` | `col ILIKE ?` | `value%` |
| `endsWith` | `col ILIKE ?` | `%value` |
| `equals` | `col = ?` | |
| `notEqual` | `col <> ?` | |
| `in` | `col IN (?)` | Pass slice as `From` |
| `notIn` | `col NOT IN (?)` | Pass slice as `From` |
| `inRange` | `col >= ? AND col <= ?` | Uses `From` and `To` |
| `lessThan` | `col < ?` | |
| `greaterThan` | `col > ?` | |
| `isNull` | `col IS NULL` | |
| `notNull` | `col IS NOT NULL` | |

## Testing

Run unit tests:

```bash
go test ./internal/utils/querybuilder/...
```
