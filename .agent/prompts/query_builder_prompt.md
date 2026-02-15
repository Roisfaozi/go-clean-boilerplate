# AI Prompt: Secure Dynamic Query Builder for Go (GORM + Postgres)

This file contains a ready-to-use prompt to generate a secure, parameterized dynamic query builder for Go projects that use GORM and PostgreSQL. Copy and paste the entire content into your generative AI model prompt (or hand this to a developer).

---

## Tujuan

Buat sebuah query builder reusable dan secure untuk aplikasi Go (Go 1.20+) yang menggunakan GORM dan PostgreSQL. Query builder akan:

- Menghasilkan WHERE clause yang dinamis berdasarkan struktur `filter` JSON/objek.
- Mengembalikan SQL string dengan placeholder (`?`) dan `args` slice untuk parameterized queries (aman dari SQL injection).
- Mendukung dynamic sorting, preloads (GORM Preload), dan soft-delete default: `deleted_by is null`.
- Menggunakan Generics (Generics-based API) dan reflection untuk mengambil field names dari model struct.
- Menggunakan `gorm:"column:..."` tag bila tersedia, jika tidak fallback ke `snake_case` dari field name.
- Menghasilkan unit tests (table-driven tests) dan contoh penggunaan repository.

---

## Fitur Operator (harus didukung)

- String: `contains`, `notContains`, `startsWith`, `endsWith`, `equals`, `notEqual` (menggunakan `ILIKE` untuk PostgreSQL)
- Numeric/Date: `lessThan`, `lessThanOrEqual`, `greaterThan`, `greaterThanOrEqual`
- Range: `inRange` (two-sided range; numeric/date); for strings optionally support a numeric-like range or pattern-based
- List operators: `in`, `notIn` (SQL `IN` and `NOT IN`, pass slice args)
- Null checks: `isNull`, `notNull` (no arg)
- `custom` (only with explicit args - advanced, use carefully)

Opsional namun direkomendasikan:

- `inList`: accept CSV or array, and support max size checks to avoid exploding SQL
- `containsAny`: for OR logic across multiple values
- `notIn`/`in`: expand to `IN (?)` and pass slice arg to GORM

---

## API yang diharapkan (format fungsi)

- `func GenerateDynamicQuery[T any](filter *filter.DynamicFilter) (string, []interface{}, []string, error)`

    - Returns: `query`, `args`, `warnings` (fields/ops ignored), `error`.

- `func GenerateDynamicSort[T any](filter *filter.DynamicFilter) (string, error)`

    - Returns: `sort` string plus optional `error`.

- `func Preload(db *gorm.DB, preloads []PreloadEntity) *gorm.DB` — applies GORM `.Preload` calls.

- `func GetDBFieldName(field reflect.StructField) string` — helper to return SQL column name, prefer `gorm:"column:..."` tag, fallback to `snake_case`.

---

## Contoh struktur domain/filter (tambahkan sebagai file `domain/filter/types.go`)

```go
package filter

type Filter struct {
    Type string      `json:"type"`          // e.g., "contains", "equals" ...
    From interface{} `json:"from,omitempty"`
    To   interface{} `json:"to,omitempty"`
}

type SortModel struct {
    ColId string `json:"colId"`
    Sort  string `json:"sort"` // "asc" | "desc"
}

type DynamicFilter struct {
    Filter map[string]Filter `json:"filter,omitempty"`
    Sort   *[]SortModel      `json:"sort,omitempty"`
}
```

---

## Contoh `GenerateDynamicQuery` behaviour

1. Example: contains string

Model:

```go
// Country model example
type Country struct {
    Id int
    Name string `gorm:"column:name"`
    DeletedBy *sql.NullInt64
}
```

Filter:

```json
{ "filter": { "Name": { "type": "contains", "from": "Iran" } } }
```

Expect:

- query: `"deleted_by is null AND name ILIKE ?"`
- args: `[]interface{}{"%Iran%"}`
- usage: `db.Where(query, args...).Find(&countries)`

2. Example: inRange numeric (year)

Filter: `{"filter":{"Year":{"type":"inRange","from":2015,"to":2020}}}`

Expect:

- `query`: `"deleted_by is null AND year >= ? AND year <= ?"`
- `args`: `[]interface{}{2015, 2020}`

3. Example: in operator

Filter: `{"filter":{"Color":{"type":"in","from":["black","white"]}}}`

Expect:

- `query`: `"deleted_by is null AND color IN (?)"`
- `args`: `[]interface{}{[]interface{}{"black","white"}}`

GORM will expand the `IN (?)` with a slice.

---

## Implementation requirements & constraints

- Implement strong input validation and skip unknown fields without crashing.
- Always return `query` and `args` instead of building interpolated SQL strings.
- For numeric comparisons, attempt to parse from string to number if necessary; if parsing fails, either skip or return a clear error.
- `in`/`notIn` must accept slices, and optionally CSV strings.
- Limit size of IN lists with a `MaxParams` guard to avoid huge SQL queries; return an error or fallback if limit exceeded.
- For `contains` and related string ops, wrap with `%` operator in args and use `ILIKE` for Postgres.
- `inRange` for strings: prefer pattern usage if asked or numeric range only for numbers/dates.
- `custom` operator: supports a fully parameterized raw expression, but must pass args safely.

---

## Helper functions recommended

- `getDBFieldName(structField reflect.StructField) string` — inspects `gorm:"column:..."` and fallback to `snake_case`.
- `toSnakeCase(s string) string` — implement a basic snake case helper or use a small helper lib.
- `parseToNumberOrNull(interface{}) (interface{}, error)` — try parse string to numeric or time.

---

## Tests (file `database/query_builder_test.go`)

Write table-driven tests for:

- `contains` => query+args
- `greaterThan`/`lessThan` numeric => query+args
- `in`/`notIn` => query+args
- `inRange` => query+args
- invalid field => warnings returned
- SQL injection: ensure input is passed as arg and not interpolated
- sort generation => `GenerateDynamicSort` output
- Preload function chaining (optional with sqlite in-memory GORM for integration test)

Minimal test example:

```go
func TestGenerateDynamicQuery_Contains(t *testing.T) {
    f := &filter.DynamicFilter{Filter: map[string]filter.Filter{"Name": {Type:"contains", From:"Iran"}}}
    q, args, warns, err := GenerateDynamicQuery[models.Country](f)
    require.NoError(t, err)
    require.Empty(t, warns)
    assert.Equal(t, "deleted_by is null AND name ILIKE ?", q)
    assert.Equal(t, []interface{}{"%Iran%"}, args)
}
```

Integration test example (sqlite memory):

```go
func TestIntegrationWithGorm(t *testing.T) {
    // Setup in-memory gorm db
    db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    // Migrate and seed
    db.AutoMigrate(&models.Country{})
    db.Create(&models.Country{Name: "Iran"})

    // Use query builder
    f := &filter.DynamicFilter{Filter: map[string]filter.Filter{"Name": {Type: "contains", From: "Iran"}}}
    q, args, _, _ := GenerateDynamicQuery[models.Country](f)
    var res []models.Country
    db.Where(q, args...).Find(&res)
    assert.Len(t, res, 1)
}
```

---

## README snippet for usage

````md
Usage (README):

1. Build filter struct and call generator:

```go
f := &filter.DynamicFilter{Filter: map[string]filter.Filter{"Name": {Type: "contains", From: "Iran"}}}
query, args, warnings, err := database.GenerateDynamicQuery[models.Country](f)
if err != nil { /* handle */ }
sort, _ := database.GenerateDynamicSort[models.Country](f)

db := database.Preload(db, preloads)
db.Where(query, args...).Order(sort).Find(&countries)
```
````

2. Running tests:

```bash
go test ./... -run TestGenerateDynamicQuery_
```

---

## Security & Notes

- ALWAYS pass args as `[]interface{}` and use placeholders `?` to avoid SQL injection.
- Avoid `fmt.Sprintf` to interpolate user data directly into SQL strings.
- For `custom` raw SQL, restrict usage and require the caller to provide their own args array.
- Provide clear logging for skipped/ignored fields/operators.

---

## Next steps / Optional improvements

- Add a JSON Schema to validate incoming filter payloads.
- Support nested filters and logical operators (`AND`/`OR`) if needed for complex queries.
- Add query caching / memoization for frequently used queries.

---

## License & Attribution

This prompt and associated example code are provided as a template to be adapted into your project; feel free to adjust types, operators, and naming conventions as required.
