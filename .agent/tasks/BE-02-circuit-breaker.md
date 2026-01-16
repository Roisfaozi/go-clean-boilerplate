# Task: Implement Circuit Breaker

## 🎯 Objective
Prevent cascading failures when external services (Storage S3, Email, etc.) are down.

## 🛠 Specifications

### 1. Library
Use `github.com/sony/gobreaker` or a similar thread-safe implementation.

### 2. Wrapper Implementation (`pkg/circuitbreaker/`)
Create a generic wrapper helper.
- `Execute(name string, fn func() error) error`

### 3. Integration Points

**A. Storage Provider (`pkg/storage/s3/s3.go`)**
- Wrap `UploadFile` calls.
- If S3 is down (timeouts/500s), the breaker should trip to "Open" state and fail fast for subsequent requests.

**B. Email Service (Future)**
- Prepare the wrapper to be used in the Email worker.

### 4. Configuration
Add settings to `.env`:
- `CB_MAX_REQUESTS`: 5
- `CB_INTERVAL`: 60s
- `CB_TIMEOUT`: 30s

### 5. Testing
- Mock the S3 client to return errors.
- Verify that after N errors, the circuit opens and subsequent calls return `ErrCircuitOpen` immediately without calling S3.
