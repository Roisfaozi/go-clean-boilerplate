## 2024-05-23 - WebSocket Origin Validation
**Vulnerability:** The WebSocket controller was allowing connections from any origin (`CheckOrigin: func(r *http.Request) bool { return true }`). This enables Cross-Site WebSocket Hijacking (CSWSH), where a malicious site can connect to the WebSocket endpoint using the victim's credentials (cookies/auth headers).
**Learning:** `gorilla/websocket`'s `CheckOrigin` function is the primary defense against CSWSH. Returning `true` indiscriminately bypasses this protection. The safe default (checking Origin == Host) is disabled when `CheckOrigin` is set to a custom function that returns `true`.
**Prevention:** Always validate the `Origin` header against a whitelist of allowed origins. Use the same `AllowedOrigins` configuration as the REST API CORS settings. If the list is empty, default to strict same-origin checks.
## 2024-05-23 - [Information Disclosure Prevention in Dynamic Queries]
**Vulnerability:** The dynamic query builder allowed sorting and filtering by any field in the entity struct, including sensitive fields like `Password` or `Token`. This could allow side-channel attacks (blind SQLi) to infer sensitive data values.
**Learning:** Generic query builders based on reflection must strictly whitelist or blacklist fields to prevent exposing internal or sensitive data.
**Prevention:** Added a blacklist in `pkg/querybuilder/query_builder.go` to block access to fields named "Password", "Token", "Secret", "Key", "Salt".
## 2024-05-23 - Broken Access Control in User Module
**Vulnerability:** Administrative endpoints in the User module (e.g., `GetAllUsers`, `DeleteUser`, `UpdateUserStatus`) were grouped together with self-service endpoints (`/me`) and only required a valid JWT token. This meant any logged-in user (even with a basic role) could access sensitive admin functions, leading to Privilege Escalation.
**Learning:** Grouping routes solely by "module" can be dangerous if the module mixes public/user/admin functionality.
**Prevention:** Explicitly split route registration functions based on the required privilege level (e.g., `RegisterAuthenticatedRoutes` vs `RegisterAuthorizedRoutes`) and apply different middleware chains (Authentication vs Authorization/RBAC) in the main router setup.
