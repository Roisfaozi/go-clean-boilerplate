# TDD Requirements: Advanced Unit Tests (Multi-tenancy)

**Target:** `internal/modules/organization` & `internal/middleware`
**Goal:** 100% Logic Coverage including Security, Edge Cases, and Negative paths.

---

## 1. Organization UseCase (`CreateOrganization`)

**Test File:** `internal/modules/organization/usecase/organization_usecase_test.go`

### ✅ Positive Cases
1.  **Standard Creation:**
    *   *Input:* `Name="Acme Corp"`, `User="u1"`
    *   *Expect:* Success, Slug="acme-corp", Owner="u1", Default Plan="free".
2.  **Unicode Name Support:**
    *   *Input:* `Name="🚀 StartUp!"`, `User="u1"`
    *   *Expect:* Success, Slug="startup" (sanitized).

### ❌ Negative Cases
1.  **Empty Name:**
    *   *Input:* `Name=""`
    *   *Expect:* `ErrValidation` (Name required).
2.  **Name Too Long:**
    *   *Input:* `Name="A" * 256`
    *   *Expect:* `ErrValidation` (Max 255 chars).
3.  **Database Failure:**
    *   *Mock:* `Repo.Create` returns `db connection error`.
    *   *Expect:* `ErrInternalServer`.

### 🛡️ Security Cases
1.  **XSS Injection in Name:**
    *   *Input:* `Name="<script>alert(1)</script>"`
    *   *Expect:* Success, but name is sanitized to `alert(1)` OR encoded. (Depends on sanitization policy).
2.  **SQL Injection in Input:**
    *   *Input:* `Name="Acme'; DROP TABLE users;--"`
    *   *Expect:* Success, saved as literal string (Parameterized query validation).

### ⚡ Edge Cases (Race Conditions)
1.  **Slug Collision Retry:**
    *   *Scenario:* "Acme" exists. "Acme 1" exists.
    *   *Mock:*
        *   `SlugExists("acme")` -> true
        *   `SlugExists("acme-1")` -> true
        *   `SlugExists("acme-2")` -> false
    *   *Expect:* Success, Slug="acme-2".

#### 💻 Implementation Example
```go
func TestCreateOrganization_SlugCollision(t *testing.T) {
    mockRepo := new(mocks.OrganizationRepository)
    uc := NewOrganizationUseCase(mockRepo, ...)

    // Setup Mock for collision
    mockRepo.On("SlugExists", "acme").Return(true)
    mockRepo.On("SlugExists", "acme-1").Return(true)
    mockRepo.On("SlugExists", "acme-2").Return(false)
    mockRepo.On("Create", mock.Anything).Return(nil)

    res, err := uc.Create(ctx, CreateRequest{Name: "Acme"})

    assert.NoError(t, err)
    assert.Equal(t, "acme-2", res.Slug)
}
```

---

## 2. Organization Member UseCase (`InviteMember`)

**Test File:** `internal/modules/organization/usecase/member_usecase_test.go`

### ✅ Positive Cases
1.  **Invite Existing User:**
    *   *Input:* Email exists in Global Users.
    *   *Expect:* Added to `organization_members`. Status="Active". No email verification needed.
2.  **Invite New User:**
    *   *Input:* Email not found.
    *   *Expect:* Shadow User created. Added to `organization_members`. Status="Invited". Activation email sent.

### ❌ Negative Cases
1.  **Already Member:**
    *   *Input:* User is already in `organization_members`.
    *   *Expect:* `ErrConflict` ("User already in organization").
2.  **Invalid Role:**
    *   *Input:* Role="super-god-admin" (non-existent).
    *   *Expect:* `ErrInvalidRole`.
3.  **Self-Invite:**
    *   *Input:* Inviter email == Invitee email.
    *   *Expect:* `ErrBadRequest`.

### 🛡️ Security Cases
1.  **Privilege Escalation:**
    *   *Scenario:* Inviter is "Editor", tries to invite "Admin".
    *   *Expect:* `ErrForbidden` (Cannot invite with role higher than self).
2.  **Organization Traversal:**
    *   *Scenario:* User tries to invite member to `org-2` while logged into `org-1`.
    *   *Expect:* `ErrForbidden` (Context mismatch).

#### 💻 Implementation Example
```go
func TestInviteMember_ShadowUser(t *testing.T) {
    mockUserRepo := new(mocks.UserRepository)
    mockMemberRepo := new(mocks.MemberRepository)
    uc := NewMemberUseCase(mockUserRepo, mockMemberRepo, ...)

    // Mock: Email not found globally
    mockUserRepo.On("FindByEmail", "new@mail.com").Return(nil, ErrNotFound)
    // Mock: Create Shadow User
    mockUserRepo.On("Create", mock.MatchedBy(func(u *entity.User) bool {
        return u.Status == "invited" && u.Password == ""
    })).Return(nil)
    // Mock: Add to Org
    mockMemberRepo.On("AddMember", "org-1", mock.Anything, "viewer").Return(nil)

    err := uc.InviteMember(ctx, "org-1", "new@mail.com", "viewer")
    assert.NoError(t, err)
}
```

---

## 3. Middleware (`TenantResolver`)

**Test File:** `internal/middleware/tenant_middleware_test.go`

### ✅ Positive Cases
1.  **Valid Membership (Cache Hit):** Redis has "1". Pass through.
2.  **Valid Membership (DB Hit):** Redis miss, DB finds member. Pass through & Cache set.

### ❌ Negative Cases
1.  **Missing Header:** `X-Org-ID` is missing.
    *   *Expect:* Pass through (Global Context), `c.Get("organization_id")` is nil.
2.  **Malformed Header:** `X-Org-ID` is not a UUID ("abc").
    *   *Expect:* `400 Bad Request`.
3.  **Not a Member:** Redis/DB confirms no membership.
    *   *Expect:* `403 Forbidden`.

### 🛡️ Security Cases
1.  **Header Spoofing:**
    *   *Scenario:* User sends `X-Org-ID` of a competitor.
    *   *Expect:* `403 Forbidden` (Validation ensures user_id + org_id pair exists).
2.  **Banned Member:**
    *   *Scenario:* User exists in `organization_members` but status="banned".
    *   *Expect:* `403 Forbidden` ("Your access has been revoked").

#### 💻 Implementation Example
```go
func TestTenantMiddleware_AccessDenied(t *testing.T) {
    mockReader := new(mocks.OrganizationReader)
    mockReader.On("ValidateMembership", mock.Anything, "org-X", "user-1").Return(false, nil)

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/", nil)
    c.Request.Header.Set("X-Org-ID", "org-X")
    c.Set("user_id", "user-1") // Simulated Auth

    middleware := TenantMiddleware(mockReader, logrus.New())
    middleware(c)

    assert.Equal(t, 403, w.Code)
    assert.JSONEq(t, `{"error": "Access denied"}`, w.Body.String())
}
```

---

## 4. Repository Scopes (`OrganizationScope`)

**Test File:** `pkg/database/scopes_test.go`

### ✅ Positive Cases
1.  **Scope Injection:**
    *   *Context:* `organization_id="org-123"`
    *   *SQL Output:* `... WHERE "organization_id" = 'org-123'`

### 🛡️ Security Cases
1.  **Scope Bypass Attempt:**
    *   *Scenario:* Developer manually adds `Where("organization_id = ?", "other-org")` alongside scope.
    *   *Expect:* SQL should contain `WHERE ... AND organization_id = 'org-123' AND organization_id = 'other-org'`. (Result: Empty, Safe).
2.  **Empty Context Safety:**
    *   *Scenario:* Context has empty string `""` for org ID.
    *   *Expect:* Scope should NOT apply filter (System Admin mode) OR Fail safe (depending on strict config). *Decision: Default to strictly ignore empty strings to avoid `WHERE id = ""` issues.*

#### 💻 Implementation Example
```go
func TestOrganizationScope_Applied(t *testing.T) {
    db, mock, _ := sqlmock.New()
    gormDB, _ := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{})

    ctx := context.WithValue(context.Background(), "organization_id", "org-123")

    // We expect the query to contain the WHERE clause
    mock.ExpectQuery("SELECT .* FROM `users` WHERE `organization_id` = ?").
        WithArgs("org-123").
        WillReturnRows(sqlmock.NewRows([]string{"id"}))

    gormDB.Scopes(OrganizationScope(ctx)).Find(&User{})
}
```

---

**Approval:** This document serves as the "Definition of Done" for Unit Testing.