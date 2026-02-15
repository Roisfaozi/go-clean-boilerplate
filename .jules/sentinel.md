## 2024-05-23 - WebSocket Origin Validation
**Vulnerability:** The WebSocket controller was allowing connections from any origin (`CheckOrigin: func(r *http.Request) bool { return true }`). This enables Cross-Site WebSocket Hijacking (CSWSH), where a malicious site can connect to the WebSocket endpoint using the victim's credentials (cookies/auth headers).
**Learning:** `gorilla/websocket`'s `CheckOrigin` function is the primary defense against CSWSH. Returning `true` indiscriminately bypasses this protection. The safe default (checking Origin == Host) is disabled when `CheckOrigin` is set to a custom function that returns `true`.
**Prevention:** Always validate the `Origin` header against a whitelist of allowed origins. Use the same `AllowedOrigins` configuration as the REST API CORS settings. If the list is empty, default to strict same-origin checks.
=======
## 2024-05-23 - [Information Disclosure Prevention in Dynamic Queries]
**Vulnerability:** The dynamic query builder allowed sorting and filtering by any field in the entity struct, including sensitive fields like `Password` or `Token`. This could allow side-channel attacks (blind SQLi) to infer sensitive data values.
**Learning:** Generic query builders based on reflection must strictly whitelist or blacklist fields to prevent exposing internal or sensitive data.
**Prevention:** Added a blacklist in `pkg/querybuilder/query_builder.go` to block access to fields named "Password", "Token", "Secret", "Key", "Salt".
