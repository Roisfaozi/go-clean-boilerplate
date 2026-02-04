# TDD Requirements: Unit Tests (Multi-tenancy)

**Target:** `internal/modules/organization` & `internal/middleware`
**Strategy:** Mock-heavy testing to verify logic correctness without DB dependencies.

---

## 1. Organization UseCase (`CreateOrganization`)

**File:** `internal/modules/organization/usecase/organization_usecase_test.go`
**Function:** `Create(ctx, req)`

### Scenario 1.1: Successful Creation
*   **Input:** `req = { Name: "Acme Corp" }`, `ctx.UserID = "user-123"`
*   **Mocks:**
    *   `Repo.SlugExists("acme-corp")` -> Returns `false`.
    *   `Repo.Create(AnyOrgStruct)` -> Returns `nil` (Success).
*   **Assertions:**
    *   Expect error: `nil`.
    *   Expect Result: `Organization` object with `ID` (uuid), `Slug` ("acme-corp"), `OwnerID` ("user-123").
    *   Verify `Repo.Create` was called exactly once.

### Scenario 1.2: Slug Collision (Auto-retry logic or Error)
*   **Input:** `req = { Name: "Acme Corp" }`
*   **Mocks:**
    *   `Repo.SlugExists("acme-corp")` -> Returns `true` (Exists).
    *   `Repo.SlugExists("acme-corp-1")` -> Returns `false` (Available).
    *   `Repo.Create(...)` -> Returns `nil`.
*   **Assertions:**
    *   Expect Result: `Organization` object with `Slug` **"acme-corp-1"**.
    *   *(Note: This tests the slug generation retry logic).*

---

## 2. Organization Member UseCase (`InviteMember`)

**File:** `internal/modules/organization/usecase/member_usecase_test.go`
**Function:** `InviteMember(ctx, orgID, email, role)`

### Scenario 2.1: Invite Existing User
*   **Input:** `email="john@exists.com"`, `orgID="org-1"`.
*   **Mocks:**
    *   `UserRepo.FindByEmail("john@exists.com")` -> Returns `User{ID: "user-999"}`.
    *   `MemberRepo.AddMember("org-1", "user-999", "role-editor")` -> Returns `nil`.
*   **Assertions:**
    *   `MemberRepo.AddMember` called with correct UserID.
    *   `TaskDistributor.SendEmail` called (Notification "You have been added").

### Scenario 2.2: Invite New User (Shadow User)
*   **Input:** `email="new@new.com"`, `orgID="org-1"`.
*   **Mocks:**
    *   `UserRepo.FindByEmail` -> Returns `ErrNotFound`.
    *   `UserRepo.Create(User{Status: "invited", Password: ""})` -> Returns `User{ID: "user-new"}`.
    *   `MemberRepo.AddMember("org-1", "user-new", "role-editor")` -> Returns `nil`.
*   **Assertions:**
    *   `UserRepo.Create` called with Invited status.
    *   `TaskDistributor.SendEmail` called (Verification/Setup Password link).

---

## 3. Middleware Unit Tests (`TenantResolver`)

**File:** `internal/middleware/tenant_middleware_test.go`
**Function:** `TenantResolver(reader)`

### Scenario 3.1: Valid Access (Cache Hit)
*   **Setup:**
    *   Header: `X-Org-ID: org-1`
    *   Context: `userID: user-1`
    *   **Mock Reader:** `ValidateMembership(ctx, "org-1", "user-1")` -> Returns `true, nil`.
*   **Execution:** Run Gin context through middleware.
*   **Assertions:**
    *   Response Code: `200` (Next handler called).
    *   Context Value: `c.Get("organization_id")` == `"org-1"`.

### Scenario 3.2: Invalid Access (Not a Member)
*   **Setup:**
    *   Header: `X-Org-ID: org-bad`
    *   Context: `userID: user-1`
    *   **Mock Reader:** `ValidateMembership(ctx, "org-bad", "user-1")` -> Returns `false, nil`.
*   **Execution:** Run Gin context.
*   **Assertions:**
    *   Response Code: `403 Forbidden`.
    *   Response Body: `{"error": "Access denied..."}`.
    *   Next handler **NOT** called.

### Scenario 3.3: Missing Header
*   **Setup:** Header `X-Org-ID` is empty.
*   **Execution:** Run Gin context.
*   **Assertions:**
    *   Response Code: `200` (Passthrough).
    *   Context Value: `c.Get("organization_id")` is nil/empty.
    *   *(Note: This verifies that the middleware doesn't crash on global routes).*

---

## 4. Repository Unit Tests (Query Scopes)

**File:** `pkg/database/scopes_test.go`
**Function:** `OrganizationScope(ctx)`

### Scenario 4.1: Scope Applied
*   **Setup:** Context with `organization_id = "org-1"`.
*   **Action:** `db.Scopes(OrganizationScope(ctx)).ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&User{}) })`
*   **Assertions:**
    *   Generated SQL string must contain: `WHERE "organization_id" = 'org-1'`.

### Scenario 4.2: Scope Skipped (System Admin / Global)
*   **Setup:** Context with `organization_id` empty.
*   **Action:** Same as above.
*   **Assertions:**
    *   Generated SQL string must **NOT** contain `organization_id`.

---

## 5. Mock Requirements (GoMock)

We need to generate mocks for these interfaces before writing tests:

```go
// internal/modules/organization/repository/mocks/mock_repo.go
type MockOrganizationRepository interface {
    Create(...)
    CheckMembership(...)
    // ...
}

// internal/modules/organization/usecase/mocks/mock_reader.go
type MockOrganizationReader interface {
    ValidateMembership(...)
}
```

**Instruction:** Use `mockery` to generate these mocks automatically.
