# Module Improvement Specification

This document details specific feature enhancements required for each domain module to reach enterprise-grade functionality.

## 🔐 1. Auth Module
**Goal:** Provide comprehensive account security and self-service capabilities.

### 1.1 Password Reset Flow (Forgot Password)
- **Priority:** High
- **Problem:** Users currently cannot recover accounts if they forget passwords.
- **Solution:**
    - Create `password_reset_tokens` table (email, token, expires_at).
    - Endpoint `POST /auth/forgot-password`: Generates token & sends email.
    - Endpoint `POST /auth/reset-password`: Validates token & updates password.
- **Dependencies:** Email Service (SMTP/SendGrid).

### 1.2 Account Verification
- **Priority:** Medium
- **Problem:** Fake emails can register freely.
- **Solution:**
    - Add `email_verified_at` (timestamp) column to `users` table.
    - Block login if `email_verified_at` is NULL (configurable via ENV).
    - Send verification link upon registration.

### 1.3 Social Login (OAuth2)
- **Priority:** Low
- **Problem:** Registration friction is high.
- **Solution:**
    - Integrate `golang.org/x/oauth2`.
    - Support Google and GitHub providers initially.
    - Link social ID to existing user account by email.

---

## 👤 2. User Module
**Goal:** Enhance user profile management and administrative controls.

### 2.1 Profile Avatar
- **Priority:** Medium
- **Problem:** Users have no visual identity.
- **Solution:**
    - Add `avatar_url` column to `users`.
    - Integrate with **Storage Module** (S3/MinIO) for upload.
    - Endpoint `POST /users/avatar` for multipart upload.

### 2.2 User Status (Ban/Suspend)
- **Priority:** High
- **Problem:** Admins can only delete users, not temporarily suspend them.
- **Solution:**
    - Add `status` column (enum: `active`, `suspended`, `banned`).
    - Middleware check: Reject requests from non-active users immediately.

### 2.3 Bulk Operations
- **Priority:** Low
- **Problem:** Managing thousands of users one-by-one is inefficient.
- **Solution:**
    - Endpoint `DELETE /users/bulk` accepts `ids: []string`.
    - Endpoint `PATCH /users/bulk/status` to ban multiple users.

---

## 🛡️ 3. Role & Permission Module
**Goal:** simplify complex permission management.

### 3.1 Role Hierarchy (Inheritance)
- **Priority:** High
- **Problem:** Duplicate permission assignment for upper-level roles.
- **Solution:**
    - Leverage Casbin's `g` (grouping) policy for roles.
    - Example: `role:superadmin` inherits `role:admin`.

### 3.2 Frontend Permission Batch Check
- **Priority:** Medium
- **Problem:** Frontend logic is complex when determining UI element visibility.
- **Solution:**
    - Endpoint `POST /permissions/check-batch`.
    - Input: List of `{resource, action}`.
    - Output: Map of `{resource_action: boolean}`.
    - Allows frontend to query all rights in one go upon page load.

---

## 🌐 4. Access & Endpoint Module
**Goal:** Automate configuration to reduce human error.

### 4.1 Route Auto-Discovery
- **Priority:** Low
- **Problem:** Developers forget to register new endpoints in the database.
- **Solution:**
    - Create a startup hook that iterates `gin.Engine.Routes()`.
    - Automatically `UPSERT` discovered routes into the `endpoints` table.
    - Mark vanished routes as `deprecated` or delete them.

---

## 📜 5. Audit Module
**Goal:** Compliance and long-term data management.

### 5.1 Export to CSV/Excel
- **Priority:** Medium
- **Problem:** Auditors need offline reports.
- **Solution:**
    - Endpoint `GET /audit-logs/export`.
    - Stream data directly to CSV writer (low memory usage) for download.

### 5.2 Retention Policy
- **Priority:** Low
- **Problem:** `audit_logs` table grows indefinitely.
- **Solution:**
    - Implement a Background Job (Scheduler).
    - Logic: "Delete logs older than X days" (Configurable).
    - Option to archive to cold storage (S3) before deletion.