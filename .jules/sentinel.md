## .jules/sentinel.md

## 2025-10-27 - Cross-Site WebSocket Hijacking (CSWSH) Fix
**Vulnerability:** The WebSocket controller was permitting connections from all origins by default, making it vulnerable to CSWSH.
**Learning:** In Go/Gin with Gorilla WebSocket, the `CheckOrigin` function in `websocket.Upgrader` must be explicitly configured. The default behavior is restrictive (same origin), but the boilerplate code had explicitly set it to return `true` (allow all) for development convenience, which was left active.
**Prevention:** Always validate the `Origin` header against a whitelist of allowed domains in production. Used `config.CORS.AllowedOrigins` to enforce this policy.
