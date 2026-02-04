# Multi-Tenancy Architecture Documentation

## Overview

This document describes the multi-tenancy implementation in the Go Clean Boilerplate project, enabling organization-based data isolation and member management.

## Architecture Components

### 1. Database Schema

#### Organizations Table

```sql
CREATE TABLE organizations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    owner_id VARCHAR(36) NOT NULL,
    settings JSON,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_organizations_slug (slug),
    INDEX idx_organizations_owner_id (owner_id)
);
```

#### Organization Members Table

```sql
CREATE TABLE organization_members (
    id VARCHAR(36) PRIMARY KEY,
    organization_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    role ENUM('owner', 'admin', 'member') NOT NULL DEFAULT 'member',
    joined_at BIGINT NOT NULL,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_org_user (organization_id, user_id),
    INDEX idx_org_members_user_id (user_id),
    INDEX idx_org_members_org_id (organization_id)
);
```

### 2. Module Structure

```
internal/modules/organization/
├── delivery/http/
│   ├── organization_controller.go    # HTTP handlers
│   └── organization_routes.go        # Route registration
├── entity/
│   ├── organization_entity.go        # Organization entity
│   └── organization_member_entity.go # Member entity
├── model/
│   ├── organization_model.go         # DTOs
│   └── converter/
│       └── organization_converter.go # Entity ↔ DTO conversion
├── repository/
│   ├── interfaces.go                 # Repository contracts
│   ├── organization_repository.go    # Organization data access
│   └── organization_member_repository.go
├── usecase/
│   ├── interfaces.go                 # Usecase contracts
│   ├── organization_usecase.go       # Business logic
│   └── organization_member_usecase.go
├── test/
│   └── mocks/                        # Generated mocks
└── module.go                         # Module initialization
```

### Module Wiring

**Router Integration** (`internal/router/router.go`):

```go
func SetupRouter(
    // ... existing params
    organizationModule *organization.OrganizationModule,
    tenantMiddleware *middleware.TenantMiddleware,
    // ...
) *gin.Engine {
    // Organization routes (no tenant middleware)
    apiV1.POST("/organizations", organizationModule.Controller.CreateOrganization)
    apiV1.GET("/organizations/me", organizationModule.Controller.GetMyOrganizations)
    apiV1.GET("/organizations/:id", organizationModule.Controller.GetOrganizationByID)
    apiV1.GET("/organizations/slug/:slug", organizationModule.Controller.GetOrganizationBySlug)

    // Tenant-scoped routes
    tenant := apiV1.Group("")
    tenant.Use(authMiddleware.ValidateToken())
    tenant.Use(tenantMiddleware.RequireOrganization())
    tenant.PATCH("/organizations/:id", organizationModule.Controller.UpdateOrganization)
    tenant.DELETE("/organizations/:id", organizationModule.Controller.DeleteOrganization)
}
```

**App Initialization** (`internal/config/app.go`):

```go
// Initialize organization module
orgModule := organization.NewOrganizationModule(db, redis, logger, validate, txManager)

// Initialize tenant middleware
tenantMiddleware := middleware.NewTenantMiddleware(orgModule.MemberUseCase, logger)

// Wire to router
router := router.SetupRouter(..., orgModule, tenantMiddleware, ...)
```

### 3. Middleware: TenantMiddleware

**Purpose:** Automatically inject organization context into requests.

**Location:** `internal/middleware/tenant_middleware.go`

**Functionality:**

- Extracts organization ID from request header (`X-Organization-ID`)
- Validates user membership in the organization
- Injects `organization_id` into Gin context
- Returns 403 Forbidden if user is not a member

**Usage:**

```go
tenant := apiV1.Group("")
tenant.Use(authMiddleware.ValidateToken())
tenant.Use(tenantMiddleware.RequireOrganization())
{
    // All routes here require organization context
    tenant.GET("/tenant-scoped-resource", handler)
}
```

### 4. GORM Scopes for Tenant Isolation

**Location:** `pkg/database/scopes.go`

**Available Scopes:**

```go
// Filter by organization ID
db.Scopes(scopes.ByOrganization(orgID)).Find(&resources)

// Filter by organization member
db.Scopes(scopes.ByMember(userID)).Find(&organizations)

// Filter by organization owner
db.Scopes(scopes.ByOwner(userID)).Find(&organizations)

// Filter by organization slug
db.Scopes(scopes.BySlug(slug)).First(&organization)
```

**Example Usage in Repository:**

```go
func (r *ResourceRepository) GetByOrganization(ctx context.Context, orgID string) ([]entity.Resource, error) {
    var resources []entity.Resource
    err := r.db.WithContext(ctx).
        Scopes(scopes.ByOrganization(orgID)).
        Find(&resources).Error
    return resources, err
}
```

## API Endpoints

### Organization Management

| Method | Endpoint                    | Description              | Auth | Tenant |
| ------ | --------------------------- | ------------------------ | ---- | ------ |
| POST   | `/organizations`            | Create organization      | ✅   | ❌     |
| GET    | `/organizations/me`         | Get user's organizations | ✅   | ❌     |
| GET    | `/organizations/:id`        | Get organization by ID   | ✅   | ❌     |
| GET    | `/organizations/slug/:slug` | Get organization by slug | ✅   | ❌     |
| PATCH  | `/organizations/:id`        | Update organization      | ✅   | ✅     |
| DELETE | `/organizations/:id`        | Delete organization      | ✅   | ✅     |

### Request/Response Examples

#### Create Organization

```bash
POST /api/v1/organizations
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Acme Corporation",
  "slug": "acme-corp"
}
```

**Response:**

```json
{
  "data": {
    "id": "uuid-here",
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "owner_id": "user-uuid",
    "created_at": 1234567890
  }
}
```

#### Get User's Organizations

```bash
GET /api/v1/organizations/me
Authorization: Bearer <token>
```

**Response:**

```json
{
  "data": {
    "owned": [
      {
        "id": "org-1",
        "name": "My Company",
        "slug": "my-company",
        "role": "owner"
      }
    ],
    "member_of": [
      {
        "id": "org-2",
        "name": "Partner Org",
        "slug": "partner-org",
        "role": "admin"
      }
    ]
  }
}
```

## Swagger Documentation

All organization endpoints are fully documented with Swagger annotations.

### Coverage

- **Total API endpoints documented:** 39 endpoints across all modules
- **Organization endpoints:** 6 endpoints (Create, Get My Orgs, Get by ID, Get by Slug, Update, Delete)
- **Swagger UI:** `http://localhost:8080/api/docs/index.html`

### Swagger Wrapper Types

Created in `pkg/response/swagger_types.go`:

- `SwaggerSuccessResponseWrapper` - Generic success responses
- `SwaggerOrganizationResponseWrapper` - Single organization response
- `SwaggerOrganizationListResponseWrapper` - User's organizations list
- `SwaggerAuditLogListResponseWrapper` - Audit logs (also added)

### Regenerate Swagger Documentation

```bash
swag init -g cmd/api/main.go -o docs --pd --parseInternal
```

### Example Swagger Annotation

```go
// @Summary      Create organization
// @Description  Creates a new organization with the current user as owner
// @Tags         organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      model.CreateOrganizationRequest  true  "Organization creation request"
// @Success      201      {object}  response.SwaggerSuccessResponseWrapper{data=model.OrganizationResponse}
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper
// @Router       /organizations [post]
func (ctrl *OrganizationController) CreateOrganization(c *gin.Context) { ... }
```

## Custom Validators

### Slug Validator

**Location:** `pkg/validation/custom_validators.go`

**Purpose:** Validate organization slugs are URL-safe.

**Rules:**

- Lowercase letters, numbers, hyphens only
- Must start and end with alphanumeric
- Length: 2-100 characters

**Usage:**

```go
type CreateOrganizationRequest struct {
    Name string `json:"name" validate:"required,min=2,max=255"`
    Slug string `json:"slug" validate:"required,min=2,max=100,slug"`
}
```

## Testing

### E2E Tests

**Location:** `tests/e2e/api/organization_e2e_test.go`

**Coverage:**

- Create organization
- Get user's organizations
- Get organization by ID
- Get organization by slug
- Update organization
- Delete organization
- Membership validation

### Integration Tests

**Location:** `tests/integration/modules/organization_integration_test.go`

**Coverage:**

- Repository CRUD operations
- GORM scopes functionality
- Database constraints

### Unit Tests

**Location:** `internal/modules/organization/usecase/organization_usecase_test.go`

**Coverage:**

- Business logic validation
- Error handling
- Mock interactions

### Running Tests

```bash
# Unit tests only (fast, no Docker required)
make test-unit

# Integration tests (requires Docker)
make test-integration

# E2E tests (requires Docker)
make test-e2e

# Run all tests
make test-all

# Generate coverage report
make test-coverage
```

**Test Coverage:**

- Unit tests: Organization usecase business logic
- Integration tests: Repository layer with real database
- E2E tests: Full HTTP request/response lifecycle

## Migration Guide

### Phase 1: Organization Module (✅ Completed)

**Migration Files:**

- Up: `db/migrations/000011_create_organizations_tables.up.sql`
- Down: `db/migrations/000011_create_organizations_tables.down.sql`

**Run Migration:**

```bash
make migrate-up
```

**Completed Items:**

- [x] Create organizations and organization_members tables
- [x] Implement organization CRUD (entity, repository, usecase, controller)
- [x] Add TenantMiddleware for automatic context injection
- [x] Create GORM scopes (ByOrganization, ByMember, ByOwner, BySlug)
- [x] Add custom 'slug' validator
- [x] Add complete Swagger documentation (39 total endpoints)
- [x] Wire module to router and app initialization
- [x] Add E2E, integration, and unit tests

### Phase 2: Tenant Data Isolation (⏳ Pending)

#### Step 1: Database Migration

```sql
-- Add organization_id to existing tables
ALTER TABLE users ADD COLUMN organization_id VARCHAR(36);
ALTER TABLE roles ADD COLUMN organization_id VARCHAR(36);
ALTER TABLE access_rights ADD COLUMN organization_id VARCHAR(36);
ALTER TABLE audit_logs ADD COLUMN organization_id VARCHAR(36);

-- Add foreign keys
ALTER TABLE users ADD FOREIGN KEY (organization_id) REFERENCES organizations(id);
ALTER TABLE roles ADD FOREIGN KEY (organization_id) REFERENCES organizations(id);
-- ... etc
```

#### Step 2: Update Entities

```go
type User struct {
    ID             string `gorm:"primaryKey"`
    OrganizationID string `gorm:"type:varchar(36);index"`
    // ... other fields
}
```

#### Step 3: Update Repositories

```go
func (r *UserRepository) GetByOrganization(ctx context.Context, orgID string) ([]entity.User, error) {
    var users []entity.User
    err := r.db.WithContext(ctx).
        Scopes(scopes.ByOrganization(orgID)).
        Find(&users).Error
    return users, err
}
```

#### Step 4: Update UseCases

```go
func (uc *UserUseCase) GetUsers(ctx context.Context) ([]model.UserResponse, error) {
    orgID := ctx.Value("organization_id").(string)
    users, err := uc.repo.GetByOrganization(ctx, orgID)
    // ...
}
```

### Phase 3: Organization Features (⏳ Pending)

#### Invitation System

- Create `organization_invitations` table
- Add invitation endpoints
- Implement email notifications
- Add acceptance flow

#### Settings Management

- Add `settings` JSONB column to organizations
- Implement settings CRUD
- Add validation for settings schema

## Security Considerations

### 1. Authorization Checks

Always verify user has permission to access organization resources:

```go
// In usecase
func (uc *OrganizationUseCase) UpdateOrganization(ctx context.Context, id string, req model.UpdateOrganizationRequest) error {
    userID := ctx.Value("user_id").(string)

    // Check if user is owner or admin
    member, err := uc.memberRepo.GetByUserAndOrg(ctx, userID, id)
    if err != nil || (member.Role != "owner" && member.Role != "admin") {
        return exception.ErrForbidden
    }

    // Proceed with update
}
```

### 2. Slug Uniqueness

Slugs are globally unique to prevent organization impersonation.

### 3. Cascade Deletion

When an organization is deleted:

- All organization_members are deleted (CASCADE)
- Organization owner must be the one deleting
- Consider soft delete for audit trail

### 4. Rate Limiting

Organization creation should be rate-limited per user to prevent abuse.

## Best Practices

### 1. Always Use Scopes

```go
// ❌ Bad - No tenant isolation
db.Find(&users)

// ✅ Good - Tenant-scoped
db.Scopes(scopes.ByOrganization(orgID)).Find(&users)
```

### 2. Validate Organization Context

```go
// In middleware or usecase
orgID, exists := c.Get("organization_id")
if !exists {
    return exception.ErrOrganizationRequired
}
```

### 3. Use Transactions for Multi-Table Operations

```go
err := uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
    // Create organization
    org, err := uc.orgRepo.Create(txCtx, orgEntity)
    if err != nil {
        return err
    }

    // Add owner as member
    member := &entity.OrganizationMember{
        OrganizationID: org.ID,
        UserID: ownerID,
        Role: "owner",
    }
    return uc.memberRepo.Create(txCtx, member)
})
```

## Troubleshooting

### Issue: 403 Forbidden on tenant-scoped endpoints

**Solution:** Ensure `X-Organization-ID` header is set in request.

### Issue: User sees data from other organizations

**Solution:** Verify all queries use `ByOrganization` scope.

### Issue: Cannot create organization with existing slug

**Solution:** Slugs must be globally unique. Choose a different slug.

## Future Enhancements

1. **Multi-Organization Support per User**
   - Allow users to switch between organizations
   - Store active organization in session

2. **Organization Hierarchy**
   - Parent-child organization relationships
   - Inherited permissions and settings

3. **Resource Quotas**
   - Limit resources per organization (users, storage, API calls)
   - Implement quota enforcement

4. **Audit Trail**
   - Track all organization-level changes
   - Member join/leave events
   - Settings modifications

## References

- [ROADMAP.md](./ops/ROADMAP.md) - Overall project roadmap
- [MODULE_IMPROVEMENTS.md](./ops/MODULE_IMPROVEMENTS.md) - Module-specific improvements
- [API_ACCESS_WORKFLOW.md](./API_ACCESS_WORKFLOW.md) - RBAC workflow
- [Swagger Documentation](http://localhost:8080/api/docs/index.html) - Live API docs
- [Testing Strategy](./guides/TESTING.md) - Comprehensive testing guide
