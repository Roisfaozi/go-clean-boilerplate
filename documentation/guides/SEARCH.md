# Dynamic Search Guide

This guide explains how to use the secure and flexible search engine implemented in this project.

## 1. Concepts: GET vs POST Search

There are two main ways to retrieve resource lists in this API:

### A. Static Search (HTTP GET)
- **Method**: `GET`
- **Filters**: Static, defined in query parameters (e.g., `?username=john`).
- **Use Case**: Simple listing, basic pagination.
- **Limitation**: Not flexible for complex logic (OR, Range, etc.) and limited by URL length.

### B. Dynamic Search (HTTP POST /search)
- **Method**: `POST`
- **Endpoint**: Ends with `/search` (e.g., `/api/v1/users/search`).
- **Filters**: JSON body with operators and sorting.
- **Use Case**: Advanced filtering, complex UI tables, multi-column sorting.
- **Benefits**: Secure (no data in URL), supports complex operators, consistent naming.

---

## 2. Payload Structure

All `/search` endpoints accept the following JSON structure:

```json
{
  "filter": {
    "field_name": { "type": "operator", "from": "value", "to": "value_optional" }
  },
  "sort": [
    { "colId": "field_name", "sort": "asc" }
  ],
  "page": 1,
  "page_size": 10
}
```

### Supported Operators
- **String**: `contains`, `equals`, `not_equal`, `starts_with`, `ends_with`.
- **Numeric/Date**: `gt`, `gte`, `lt`, `lte`, `between`.
- **List**: `in`, `ne`.

---

## 3. Practical Examples (Curl)

### Find Users by Name
```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -H "Content-Type: application/json" \
  -d '{
    "filter": {
      "name": { "type": "contains", "from": "Admin" }
    }
  }'
```

### Complex Filtering and Sorting
```bash
curl -X POST http://localhost:8080/api/v1/users/search \
  -d '{
    "filter": {
      "email": { "type": "ends_with", "from": "@gmail.com" },
      "created_at": { "type": "gt", "from": 1700000000000 }
    },
    "sort": [
      { "colId": "name", "sort": "asc" }
    ],
    "page": 1,
    "page_size": 5
  }'
```

