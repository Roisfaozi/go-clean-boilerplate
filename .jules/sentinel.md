## 2025-02-18 - WebSocket Origin Validation
**Vulnerability:** Cross-Site WebSocket Hijacking (CSWSH) was possible because the WebSocket upgrader allowed all origins (`CheckOrigin` returned `true` unconditionally).
**Learning:** In Go `gorilla/websocket`, the default `CheckOrigin` is nil (safe for same origin), but often developers override it to `return true` to avoid CORS issues during dev, forgetting to secure it for production.
**Prevention:** Always implement an origin check that validates the `Origin` header against a whitelist (e.g., from CORS config).
