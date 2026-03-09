# API Key Management - Implementation Plan

## Overview
API Key Management allows enterprise clients to integrate with the REST API programmatically using Machine-to-Machine (M2M) authentication. This feature provides a secure way to access protected resources without requiring a user session (JWT).

## 1. Database Schema
Create a new table `api_keys` with the following fields:
- `id`: UUID (Primary Key)
- `name`: String (Human-readable label)
- `key_hash`: String (SHA-256 hash of the API key)
- `organization_id`: UUID (FK to organizations, for multi-tenancy)
- `user_id`: UUID (FK to users, the owner/creator)
- `scopes`: JSON/Text (Optional specific permissions for the key)
- `expires_at`: Timestamp (Optional expiration date)
- `last_used_at`: Timestamp
- `is_active`: Boolean
- `created_at`, `updated_at`, `deleted_at`: Standard GORM fields

## 2. Authentication Middleware
Implement a new middleware `APIKeyAuthMiddleware`:
- Extract key from `X-API-Key` header.
- Hash the incoming key and look it up in the database.
- If valid and active:
    - Inject the associated User and Organization into the context.
    - Set a flag `is_api_key_auth = true`.
- Support caching using Redis to avoid DB hits on every request.

## 3. Integration with Casbin
Ensure the `CasbinMiddleware` can handle API Key authentication:
- Since the API key is associated with a User, Casbin will use that User's roles for authorization.
- Alternatively, we can support specific "Key Scopes" that map directly to Casbin permissions.

## 4. Module Implementation (`internal/modules/api_key`)
- **Entity**: `ApiKey` model.
- **Repository**: Database operations (Create, FindByHash, ListByOrg, Revoke).
- **UseCase**: 
    - `GenerateKey`: Securely generate a high-entropy string, hash it, and save.
    - `ListKeys`: Fetch active keys for an organization.
    - `RevokeKey`: Deactivate or delete a key.
- **Delivery (HTTP)**:
    - `POST /api/v1/api-keys`: Create a key (Return raw key only once).
    - `GET /api/v1/api-keys`: List keys.
    - `DELETE /api/v1/api-keys/:id`: Revoke key.

## 5. Security Best Practices
- Never store the raw API key in the database (always hash).
- Use a prefix for keys (e.g., `sk_live_...`) for easier identification.
- Enforce Rate Limiting specifically for API Keys.
- Provide a "Last Used" timestamp for auditability.

## 6. Implementation Steps
1.  **Migration**: Create `xxxxxx_create_api_keys_table.up.sql`.
2.  **Domain Layer**: Define Entity, Request/Response DTOs.
3.  **Repository Layer**: Implement GORM-based persistence.
4.  **UseCase Layer**: Implement key generation logic (using secure random).
5.  **Middleware**: Implement header-based authentication logic.
6.  **HTTP Layer**: Implement Controllers and register Routes.
7.  **Configuration**: Wire the module in `internal/config/app.go`.
8.  **Testing**: Unit tests for UseCase and Integration tests for Middleware.
