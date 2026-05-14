# Product Requirements Document (PRD): Multi-tenancy Architecture

| Metadata | Detail |
| :--- | :--- |
| **Project Name** | NexusOS Enterprise Platform |
| **Feature Name** | Multi-tenancy Isolation Layer (MTIL) |
| **Version** | 2.1 (Global Identity Model) |
| **Status** | **Approved for Development** |
| **Priority** | **Critical** (Architectural Foundation) |

---

## 1. Executive Summary
NexusOS is transitioning to a **Multi-tenant SaaS** architecture. The goal is to allow multiple independent organizations to coexist within a single database while maintaining absolute data isolation. Every request is scoped to an organization context, ensuring that data is never leaked across tenants.

---

## 2. Identity & Context Philosophy (Global User, Local Member)
NexusOS follows the "GitHub/Slack" model of identity:

1.  **Global User:** A user has one identity (email/password) across the entire platform. Credentials and profile data are global.
2.  **Local Member:** Access rights are organization-specific. A user can be an `Owner` in Org A and a `Viewer` in Org B. Status (Active/Banned) can also be scoped per organization.

---

## 3. Core Features & Business Rules

### 3.1 Organization Onboarding
*   **Create Organization:** Users can create multiple "Workspaces".
*   **Auto-Provisioning:** Upon user registration, the system creates a "Personal Organization" (e.g., "John Doe's Workspace") by default.
*   **Slug Integrity:** Slugs (used for routing) must be globally unique and follow URL-safe naming conventions.

### 3.2 Member Management (Shadow User Pattern)
*   **Invite Existing User:** If an invited email already exists in NexusOS, they are added to the organization immediately.
*   **Invite New User (Shadow User):** If the email is unknown, the system creates a "Shadow User" (status: `invited`, no password) and sends an activation link. Membership is established immediately but remains `pending` until activation.

### 3.3 Data Segregation
*   **Row-Level Enforcement:** All domain entities (Roles, Logs, Files) are owned by an `organization_id`.
*   **Super Admin Bypass:** System-wide administrators can bypass tenant filters for maintenance and auditing.

---

## 4. Database Schema (Normalized 3NF)

### 4.1 Organizations (Identity)
```sql
CREATE TABLE organizations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    owner_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at BIGINT,
    updated_at BIGINT
);
```

### 4.2 Organization Members (Pivot)
```sql
CREATE TABLE organization_members (
    id VARCHAR(36) PRIMARY KEY,
    organization_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    UNIQUE(organization_id, user_id)
);
```

---

## 5. Security & Isolation Mechanism
1.  **Context Resolution:** The frontend must include the `X-Org-ID` header for all tenant-scoped API calls.
2.  **Validation Middleware:** Every request is checked against the `organization_members` table (result cached in Redis for 5 minutes).
3.  **Automatic Scoping:** GORM Global Scopes automatically append `WHERE organization_id = ?` to prevent developer oversight.

---

## 6. Acceptance Criteria
*   [ ] User A cannot access Org B's data via ID manipulation (API returns 404/403).
*   [ ] Membership check is performed via Redis cache to maintain <10ms overhead.
*   [ ] Audit logs correctly record the organization context for every action.