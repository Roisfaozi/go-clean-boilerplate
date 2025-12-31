## 2024-05-23 - [Information Disclosure Prevention in Dynamic Queries]
**Vulnerability:** The dynamic query builder allowed sorting and filtering by any field in the entity struct, including sensitive fields like `Password` or `Token`. This could allow side-channel attacks (blind SQLi) to infer sensitive data values.
**Learning:** Generic query builders based on reflection must strictly whitelist or blacklist fields to prevent exposing internal or sensitive data.
**Prevention:** Added a blacklist in `pkg/querybuilder/query_builder.go` to block access to fields named "Password", "Token", "Secret", "Key", "Salt".
