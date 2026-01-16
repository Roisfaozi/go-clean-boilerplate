# Task: Implement Data Export (Audit Logs to CSV)

## 🎯 Objective
Allow admins to export audit logs to CSV format for compliance reporting.

## 🛠 Specifications

### 1. New Endpoint
`GET /api/v1/audit-logs/export`
- **Role**: `superadmin`
- **QueryParams**: `from_date`, `to_date`

### 2. Implementation (`internal/modules/audit/delivery/http/audit_controller.go`)
Use **Streaming Response** to avoid loading all records into RAM.

```go
c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")
c.Header("Content-Type", "text/csv")

writer := csv.NewWriter(c.Writer)
// 1. Write Header
// 2. Query DB in batches (using limit/offset or cursor)
// 3. Write rows
// 4. Flush writer
```

### 3. Repository
Add `FindAllInBatches` or similar method to `AuditRepository` that accepts a callback function to process rows as they come.

### 4. Testing
- E2E Test: Call the endpoint and verify the response `Content-Type` and body content format.
