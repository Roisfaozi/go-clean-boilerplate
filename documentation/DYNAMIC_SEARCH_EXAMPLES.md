# Dynamic Search API Examples (Curl)

This document provides comprehensive `curl` examples for the dynamic search endpoints implemented in the User, Role, and Access modules. These endpoints accept a JSON body with flexible filter and sort criteria.

## Base URL
Assuming the API is running locally:
`http://localhost:8080/api/v1`

## Common Payload Structure

All `/search` endpoints accept the same JSON structure:

```json
{
  "filter": {
    "FieldName": { "type": "operator", "from": "value", "to": "value_optional" }
  },
  "sort": [
    { "colId": "FieldName", "sort": "asc" }
  ]
}
```

**Supported Operators:** `contains`, `equals`, `startsWith`, `endsWith`, `in`, `notIn`, `inRange`, `lessThan`, `greaterThan`, `isNull`, `notNull`.

---

## 1. User Module (`POST /users/search`)

### A. Search by Name (Contains)
Find users whose name contains "Admin".

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "Name": { "type": "contains", "from": "Admin" }
    }
  }'
```

### B. Search by Exact Username (Equals)
Find a user with specific username "johndoe".

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "Username": { "type": "equals", "from": "johndoe" }
    }
  }'
```

### C. Search by Email Domain (EndsWith) & Sort
Find all users with gmail accounts, sorted by CreatedAt descending (newest first).

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "Email": { "type": "endsWith", "from": "@gmail.com" }
    },
    "sort": [
      { "colId": "CreatedAt", "sort": "desc" }
    ]
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
      "Name": { "type": "in", "from": ["admin", "editor"] }
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
      "Description": { "type": "contains", "from": "management" }
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
      "Method": { "type": "equals", "from": "DELETE" }
    }
  }'
```

### B. Search by Path (StartsWith)
Find all endpoints starting with `/api/v1/users`.

```bash
curl -X POST http://localhost:8080/api/v1/endpoints/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "Path": { "type": "startsWith", "from": "/api/v1/users" }
    }
  }'
```

---

## 4. Access Module - Access Rights (`POST /access-rights/search`)

### A. Complex Search (Multiple Filters)
Find Access Rights where `Name` contains "User" AND `Description` starts with "Allow".

```bash
curl -X POST http://localhost:8080/api/v1/access-rights/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "Name": { "type": "contains", "from": "User" },
      "Description": { "type": "startsWith", "from": "Allow" }
    },
    "sort": [
      { "colId": "Name", "sort": "asc" }
    ]
  }'
```

---

## 5. Advanced Examples

### A. Date Range Search (Generic)
Assuming any entity has a `CreatedAt` (timestamp/int64) field. Find records created between two timestamps.

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "CreatedAt": { "type": "inRange", "from": 1700000000000, "to": 1733000000000 }
    }
  }'
```

### B. Null Check
Find users who have NOT been soft-deleted (though the system does this by default, this shows explicit usage). Or finding records with optional fields empty.

```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "filter": {
      "DeletedAt": { "type": "isNull" }
    }
  }'
```
