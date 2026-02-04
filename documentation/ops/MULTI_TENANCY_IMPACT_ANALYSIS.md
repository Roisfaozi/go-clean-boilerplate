# Detailed Impact Analysis: Multi-tenancy Refactoring

This document outlines the specific code-level impacts of transitioning from a Single-tenant to a Multi-tenant architecture using the **Global User & Local Member** model.

---

## 1. Auth Module (`internal/modules/auth`)

| File | Change | Impact Level |
| :--- | :--- | :--- |
| `usecase/auth_usecase.go` | **Login Function:** After password verification, must query all organizations where the user is a member. | **High** |
| `model/auth_model.go` | **LoginResponse DTO:** Add `Organizations []OrgResponse` field so the frontend can show the workspace switcher. | **Medium** |
| `delivery/http/auth_controller.go` | Map the new organization list data to the final JSON response. | **Low** |

---

## 2. Role Module (`internal/modules/role`)

| File | Change | Impact Level |
| :--- | :--- | :--- |
| `entity/role_entity.go` | Add `OrganizationID *string` column. Nullable for system-wide roles, populated for tenant-specific roles. | **High** |
| `repository/role_repository.go` | **FindAll / FindByID:** Must use `db.Scopes(database.OrganizationScope(ctx))` and also allow `organization_id IS NULL`. | **Critical** |
| `usecase/role_usecase.go` | **CreateRole:** Validate name uniqueness within the specific organization scope only. | **Medium** |

---

## 3. Permission & Middleware Layer

| Component | Change | Impact Level |
| :--- | :--- | :--- |
| `casbin_model.conf` | Change model definition from standard RBAC to **RBAC with Domains**. <br>`r = sub, dom, obj, act` | **Critical** |
| `middleware/casbin_middleware.go` | Retrieve `org_id` from context (set by TenantMiddleware) and pass it to `enforcer.Enforce(sub, dom, obj, act)`. | **High** |
| `middleware/tenant_middleware.go` | **New Component:** Validates header `X-Org-ID` against the user's memberships in Redis/DB. | **High** |

---

## 4. Audit Module (`internal/modules/audit`)

| File | Change | Impact Level |
| :--- | :--- | :--- |
| `entity/audit_log.go` | Add `OrganizationID string` field to record the context of every action. | **Medium** |
| `usecase/audit_usecase.go` | Capture `org_id` from context during `LogActivity` calls. | **Low** |
| `delivery/http/audit_controller.go` | Ensure all log listing and export requests are filtered by the current tenant ID. | **Medium** |

---

## 5. Infrastructure (`pkg/database`)

| Component | Change | Impact Level |
| :--- | :--- | :--- |
| `scopes.go` | **New Utility:** Implement `OrganizationScope(ctx)` to automatically append `WHERE organization_id = ?` to queries. | **High** |

---

## 6. Effort & Risk Summary

### 📊 Effort Estimation
*   **Total Refactor Points:** 85/100
*   **Most Complex Task:** Updating Casbin Policy Matchers and Role Scoping logic.
*   **Time Allocation:** 40% Logic, 40% Testing/Validation, 20% Data Migration.

### ⚠️ Security Risk Checklist
1.  **Data Leakage:** If a developer forgets `.Scopes(database.OrganizationScope(ctx))` in a new repository method, data will leak.
    *   *Mitigation:* Global Unit Test suite that tests every repository with a "Foreign" Org ID.
2.  **Orphaned Policies:** Deleting an Organization must delete all associated Casbin policies.
    *   *Mitigation:* Use GORM hooks or cascading database deletes.
3.  **Super Admin Bypass:** Ensure that only genuine System Admins can access data across all tenants.

---

**Warning:** This refactor touches the security core of the application. Proceed with full integration test coverage.
