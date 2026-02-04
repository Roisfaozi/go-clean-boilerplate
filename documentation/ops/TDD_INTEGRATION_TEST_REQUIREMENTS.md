# TDD Requirements: Integration Tests (Multi-tenancy)

**Target:** `tests/integration/modules` & `tests/e2e/api`
**Infrastructure:** Real MySQL (Docker), Real Redis (Docker).
**Goal:** Verify data isolation, transaction atomicity, and full user journeys.

---

## 1. Repository Integration Tests (`organization_repository_test.go`)

**Focus:** Verifying GORM Scopes and Data Persistence.

### ✅ Scenario: Data Isolation (The "Leak" Test)
*   **Setup:**
    *   Insert `Org A` and `Org B`.
    *   Insert `Role A` linked to `Org A`.
    *   Insert `Role B` linked to `Org B`.
*   **Action:**
    *   Create context with `organization_id = Org A`.
    *   Call `repo.FindAll(ctx)`.
*   **Assertion:**
    *   Result must contain **ONLY** `Role A`.
    *   Result must **NOT** contain `Role B`.
    *   *Critical Pass Criteria:* Count == 1.

### ✅ Scenario: Transactional Create (Atomicity)
*   **Setup:** Prepare `Organization` struct with valid data.
*   **Action:** Call `repo.Create(ctx, org)`.
*   **Assertion:**
    *   Query `organizations` table -> Record exists.
    *   Query `organization_members` table -> Owner is added as member.
    *   Query `organization_settings` table -> Default settings created.
    *   *Verification:* All 3 tables must be populated, or none (if error occurred).

### 🛡️ Scenario: Cross-Tenant Access (Security)
*   **Setup:** Context with `Org A`. Known ID of `Role B` (from Org B).
*   **Action:** Call `repo.FindByID(ctx, ID_of_Role_B)`.
*   **Assertion:**
    *   Must return `ErrNotFound` (GORM `RecordNotFound`).
    *   Database must behave as if `Role B` does not exist for this user.

---

## 2. API End-to-End (E2E) Tests (`organization_e2e_test.go`)

**Focus:** Verifying HTTP Middleware, Headers, and JSON Response.

### 🚀 Scenario: Full Onboarding Flow
1.  **Register:** `POST /auth/register` (User: "Alice").
    *   *Check:* Auto-login, Token received.
2.  **Create Org:** `POST /organizations` (Name: "Alice Corp").
    *   *Check:* 201 Created, ID returned.
3.  **List Orgs:** `GET /organizations/me`.
    *   *Check:* Response contains "Alice Corp" and "Personal Workspace".
4.  **Get Context:** `GET /organization` with header `X-Org-ID: <Alice_Corp_ID>`.
    *   *Check:* 200 OK, returns details.

### 🚫 Scenario: Unauthorized Access (Header Spoofing)
1.  **Setup:**
    *   Alice creates "Alice Corp" (ID: `org_alice`).
    *   Bob creates "Bob Inc" (ID: `org_bob`).
2.  **Attack:**
    *   Bob sends `GET /organization` with header `X-Org-ID: <org_alice>`.
3.  **Assertion:**
    *   Response Code: `403 Forbidden`.
    *   Body: `{"error": "Access denied"}`.

### 👥 Scenario: Member Invitation Flow
1.  **Invite:** Admin calls `POST /members/invite` (Email: "new@user.com").
2.  **Verify DB:** "Shadow User" created in `users` table (status: invited).
3.  **Verify Redis:** Background task for email enqueued.
4.  **Public Access:** "new@user.com" tries to login.
    *   *Check:* Login fails (password not set).

---

## 3. Middleware Integration Tests (`middleware_integration_test.go`)

**Focus:** Verifying Redis Caching behavior.

### ⚡ Scenario: Membership Caching
1.  **First Request:** User accesses Org Endpoint.
    *   *Observe:* DB Query executed (Log/Trace check). Redis key `org:member:user:org` created.
2.  **Second Request:** User accesses same endpoint immediately.
    *   *Observe:* **NO** DB Query executed. Redis HIT.
3.  **Cache Invalidation:**
    *   Action: `DELETE /members/:id` (Kick user).
    *   Check: Redis key deleted.
4.  **Third Request:** User accesses endpoint again.
    *   *Observe:* DB Query executed -> Returns `false` (Not member) -> 403 Forbidden.

---

## 4. Helper Implementation (`tests/helpers/multi_tenancy.go`)

We need strict helpers to make these tests readable.

```go
// Helper: Create Org and return ID
func CreateTestOrg(t *testing.T, db *gorm.DB, ownerID string) string {
    // ... insert logic ...
    return org.ID
}

// Helper: Create Context with Tenant
func CtxWithTenant(orgID string) context.Context {
    return context.WithValue(context.Background(), "organization_id", orgID)
}

// Helper: HTTP Request with Tenant Header
func RequestWithTenant(method, url, token, orgID string, body interface{}) *http.Request {
    req := // ... create request ...
    req.Header.Set("Authorization", "Bearer " + token)
    req.Header.Set("X-Org-ID", orgID)
    return req
}
```

---

**Approval:** This document defines the success criteria for the "Multi-tenancy" milestone.
