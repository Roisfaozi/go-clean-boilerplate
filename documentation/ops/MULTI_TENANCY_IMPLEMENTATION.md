# Technical Design: Multi-tenancy Implementation

**Architecture Style:** Clean Architecture + Redis Caching Middleware
**Identity Model:** Global User with Scoped Membership
**Status:** Ready for Implementation

---

## 1. Database Schema Evolution

**Migration:** `db/migrations/20260128120000_create_organizations_tables.up.sql`

```sql
-- Core Organization Identity
CREATE TABLE organizations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    owner_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    INDEX idx_org_slug (slug)
);

-- Organization Members (Pivot)
-- Connects Global Users to specific Organizations.
CREATE TABLE organization_members (
    id VARCHAR(36) PRIMARY KEY,
    organization_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'active', -- active, invited, suspended
    joined_at BIGINT,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(organization_id, user_id)
);

-- Business & Branding (Normalized Tables)
-- organization_details (Profile) and organization_settings (Config).
```

---

## 2. Business Logic Patterns

### 2.1 The "Shadow User" Pattern (Invitations)
When inviting a new user by email:
1.  **Repository Check:** Check if `email` exists in `users` table.
2.  **Existing User:** Create `organization_members` record linking existing `user_id`.
3.  **New User:**
    *   Insert into `users` with `status="invited"`, empty password.
    *   Insert into `organization_members` linking the new `user_id`.
    *   Enqueue `SendInvitationEmail` task via worker.

### 2.2 Auto-Provisioning (Registration)
When a user registers:
1.  **User Create:** Normal registration flow.
2.  **Org Create:** Create a "Default Workspace" (Name: `[User Name]'s Workspace`).
3.  **Member Link:** Set user as `owner` of that workspace.

---

## 3. High-Performance Middleware Layer

### 3.1 Organization Reader (Clean Architecture)
Middleware MUST NOT hit GORM directly. It uses the `OrganizationReader` interface.

**Interface:** `internal/modules/organization/usecase/interfaces.go`
```go
type OrganizationReader interface {
    // ValidateMembership checks if user is active in org. Uses Redis cache.
    ValidateMembership(ctx context.Context, orgID, userID string) (bool, error)
}
```

**Implementation:** `internal/modules/organization/usecase/reader.go`
```go
func (r *cachedOrgReader) ValidateMembership(ctx context.Context, orgID, userID string) (bool, error) {
    cacheKey := fmt.Sprintf("org:member:%s:%s", orgID, userID)

    // 1. Redis check
    if val, err := r.redis.Get(ctx, cacheKey).Result(); err == nil {
        return val == "1", nil
    }

    // 2. DB fallback (via Repo)
    exists, err := r.repo.CheckMembership(ctx, orgID, userID)
    if err != nil { return false, err }

    // 3. Populate Cache
    r.redis.Set(ctx, cacheKey, "1", 5*time.Minute)
    return exists, nil
}
```

---

## 4. Routing Strategy (Centralized Grouping)

**File:** `internal/router/router.go`

Modules must register routes into two distinct groups to prevent context leakage.

```go
func SetupRouter(...) {
    // 1. AUTHENTICATED GLOBAL (No X-Org-ID required)
    // Profile, Workspace Switcher, Create New Org
    global := apiV1.Group("").Use(authMiddleware)
    userHttp.RegisterGlobalRoutes(global, userCtrl)
    orgHttp.RegisterGlobalRoutes(global, orgCtrl)

    // 2. TENANT SCOPED (X-Org-ID REQUIRED)
    // All routes here automatically restricted to org context
    tenant := apiV1.Group("").Use(authMiddleware, tenantMiddleware)
    {
        userHttp.RegisterTenantRoutes(tenant, userCtrl) // List Team Members
        roleHttp.RegisterTenantRoutes(tenant, roleCtrl)
        auditHttp.RegisterTenantRoutes(tenant, auditCtrl)
    }
}
```

---

## 5. Row-Level Security (GORM Scopes)

**Implementation:** `pkg/database/scopes.go`

```go
func OrganizationScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        orgID, ok := ctx.Value("organization_id").(string)
        if ok && orgID != "" {
            return db.Where("organization_id = ?", orgID)
        }
        return db // Fallback for Super Admin
    }
}
```
**Usage in Repo:**
## 8. Strategic Note: Member Management Implementation

To maintain clean architecture and avoid polluting the existing `user` module, "Member Management" will be implemented as a distinct sub-module or controller within the Organization domain.

**Endpoint:** `/api/v1/organization/members`

**Components:**
1.  **Repository (`organization_member_repository.go`):** Handles queries to `organization_members` table with JOINs to `users`.
2.  **UseCase (`organization_member_usecase.go`):** Handles the "Invite" logic (Shadow User creation vs Existing User linking).
3.  **Controller (`organization_member_controller.go`):** Manages the HTTP interface for member operations.

**Benefits:**
*   Keeps the global `user` module focused on Identity/Profile.
*   Isolates "Team Management" logic to the Tenant scope.
*   Non-destructive to existing User CRUD operations.