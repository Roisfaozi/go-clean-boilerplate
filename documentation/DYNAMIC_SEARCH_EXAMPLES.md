# Dynamic Search API Examples (Curl)

This document provides comprehensive `curl` examples for the dynamic search endpoints implemented in the User, Role, and Access modules. These endpoints accept a JSON body with flexible filter and sort criteria.

The backend supports **snake_case**, **camelCase**, or **PascalCase** field names in the filter keys. It automatically maps them to the corresponding Go struct fields and Database columns.

## Base URL
Assuming the API is running locally:
`http://localhost:8080/api/v1`

## Common Payload Structure

All `/search` endpoints accept the same JSON structure:

```json
{
  "filter": {
    "field_name": { "type": "operator", "from": "value", "to": "value_optional" }
  },
  "sort": [
    { "colId": "field_name", "sort": "asc" }
  ]
}
```

**Supported Operators (snake_case preferred):**
- String: `contains`, `not_contains`, `starts_with`, `ends_with`, `equals`, `not_equal`
- List: `in`, `not_in`
- Range: `in_range`
- Comparison: `less_than`, `less_than_or_equal`, `greater_than`, `greater_than_or_equal`
- Null: `is_null`, `not_null`

*(Note: `camelCase` operators like `startsWith` are also supported for backward compatibility)*

---

## 1. User Module (`POST /users/search`)

### A. Simple Search (Contains)
Find users whose name contains "Admin".

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "name": { "type": "contains", "from": "Admin" }
    }
  }'
```

### B. Exact Match (Equals)
Find a user with specific username "johndoe".

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "username": { "type": "equals", "from": "johndoe" }
    }
  }'
```

### C. Multiple Filters & Sorting (AND Logic)
Find users whose email ends with `@gmail.com` **AND** whose name starts with `R`.
Sort results by `created_at` descending (newest first).

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "email": { "type": "ends_with", "from": "@gmail.com" },
      "name": { "type": "starts_with", "from": "R" }
    },
    "sort": [
      { "colId": "created_at", "sort": "desc" }
    ]
  }'
```

### D. List Filtering (IN Operator)
Find users with specific IDs or Usernames. Useful for bulk selection.

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "username": { "type": "in", "from": ["alice", "bob", "charlie"] }
    }
  }'
```

### E. Exclusion (NOT IN)
Find all users EXCEPT specific ones.

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "username": { "type": "not_in", "from": ["admin", "superadmin"] }
    }
  }'
```

---

## 2. Role Module (`POST /roles/search`)

### A. Search by Name (In List)
Find roles that are either "admin" or "editor".

```bash
curl -X POST http://localhost:8080/api/v1/roles/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "name": { "type": "in", "from": ["admin", "editor"] }
    }
  }'
```

### B. Search by Description (Contains)
Find roles related to "management".

```bash
curl -X POST http://localhost:8080/api/v1/roles/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "description": { "type": "contains", "from": "management" }
    }
  }'
```

---

## 3. Access Module - Endpoints (`POST /endpoints/search`)

### A. Search by HTTP Method (Equals)
Find all `DELETE` endpoints.

```bash
curl -X POST http://localhost:8080/api/v1/endpoints/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "method": { "type": "equals", "from": "DELETE" }
    }
  }'
```

### B. Search by Path (Starts With)
Find all endpoints starting with `/api/v1/users`.

```bash
curl -X POST http://localhost:8080/api/v1/endpoints/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "path": { "type": "starts_with", "from": "/api/v1/users" }
    }
  }'
```

---

## 4. Access Module - Access Rights (`POST /access-rights/search`)

### A. Complex Search (Multiple Filters)
Find Access Rights where `name` contains "User" AND `description` starts with "Allow". Sorted by name.

```bash
curl -X POST http://localhost:8080/api/v1/access-rights/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "name": { "type": "contains", "from": "User" },
      "description": { "type": "starts_with", "from": "Allow" }
    },
    "sort": [
      { "colId": "name", "sort": "asc" }
    ]
  }'
```

---

## 5. Advanced & Generic Examples

### A. Date Range Search (InRange)
Assuming any entity has a `created_at` (timestamp/int64) field. Find records created between two timestamps (Unix milliseconds).

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "created_at": { "type": "in_range", "from": 1700000000000, "to": 1733000000000 }
    }
  }'
```

### B. Null Check (IsNull / NotNull)
Find users who have NOT been soft-deleted (though the system does this by default, this shows explicit usage).

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "deleted_at": { "type": "is_null" }
    }
  }'
```

### C. Multi-Column Sorting
Sort primarily by `role` (asc), and then by `name` (desc) for users within the same role.

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "sort": [
      { "colId": "role", "sort": "asc" },
      { "colId": "name", "sort": "desc" }
    ]
  }'
```