# Task: Implement Advanced Rate Limiter

## 🎯 Objective
Upgrade the simple IP-based rate limiter to a multi-tiered, route-specific limiter.

## 🛠 Specifications

### 1. Strategy
Modify `internal/middleware/rate_limit_redis.go`.

### 2. Logic
Use **Redis Lua Script** (already in place or upgrade it) to support multiple keys.

**Key Structure:** `rate_limit:{type}:{identifier}`

**Tiers:**
1.  **Public API**: IP-based. Low limit (e.g., 10 RPS).
2.  **Authenticated User**: UserID-based. High limit (e.g., 100 RPS).
3.  **Critical Endpoints** (e.g., `/auth/login`): IP-based. Very low limit (5 RPM).

### 3. Implementation
Refactor the middleware middleware to accept options:
```go
func RateLimitMiddleware(type LimiterType, limit int, window time.Duration) gin.HandlerFunc
```

Apply different instances of middleware to different Router Groups in `internal/router/router.go`.

### 4. Testing
- Integration Test: Verify that limits are independent (User A getting rate limited doesn't affect User B).
