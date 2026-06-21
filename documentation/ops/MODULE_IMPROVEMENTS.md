# Module Improvement Specification

This document details specific feature enhancements required for each domain module to reach enterprise-grade functionality.

## 🔐 1. Auth Module

**Goal:** Provide comprehensive account security and self-service capabilities.

### 1.1 Password Reset Flow (Forgot Password) - DONE

- **Status:** [x] Completed
- **Features:**
  - `password_reset_tokens` table.
  - `POST /auth/forgot-password`.
  - `POST /auth/reset-password`.
  - Async email simulation.

## 👤 2. User Module

**Goal:** Enhance user profile management and administrative controls.

### 2.1 Profile Avatar - DONE

- **Status:** [x] Completed
- **Implementation:**
  - `avatar_url` column in `users`.
  - Integrated with **Storage Module**.
  - Endpoint `PATCH /users/me/avatar`.

### 2.2 User Status (Ban/Suspend) - DONE

- **Status:** [x] Completed
- **Implementation:**
  - `status` column (active, suspended, banned).
  - Session revocation on ban.

## 🛡️ 3. Role & Permission Module

**Goal:** simplify complex permission management.

### 3.1 Role Hierarchy (Inheritance) - DONE

- **Status:** [x] Completed
- **Implementation:**
  - Casbin `g` policy support.
  - Endpoint for managing inheritance.

### 3.2 Frontend Permission Batch Check - DONE

- **Status:** [x] Completed
- **Implementation:**
  - Endpoint `POST /permissions/check-batch`.

## 📜 5. Audit Module

**Goal:** Compliance and long-term data management.

### 5.2 Retention Policy - DONE

- **Status:** [x] Completed
- **Implementation:**
  - Background Job (Scheduler) deletes old logs.

## 🏢 6. Organization Module (Multi-Tenancy)

**Goal:** Enable multi-tenant architecture with organization-level data isolation and member management.

### 6.1 Organization Management - DONE

- **Status:** [x] Completed
- **Implementation:**
  - Organizations table with unique slug-based routing
  - Organization members table with role-based access (owner, admin, member)
  - TenantMiddleware for automatic organization context injection
  - GORM scopes for tenant-scoped queries (ByOrganization, ByMember, ByOwner)
  - Complete CRUD API endpoints:
    - `POST /organizations` - Create organization
    - `GET /organizations/me` - Get user's organizations
    - `GET /organizations/:id` - Get by ID
    - `GET /organizations/slug/:slug` - Get by slug
    - `PATCH /organizations/:id` - Update organization
    - `DELETE /organizations/:id` - Delete organization
  - Custom 'slug' validator for URL-safe organization identifiers
  - Full Swagger documentation (39 total API endpoints)

### 6.2 Tenant Data Isolation - PENDING

- **Status:** [ ] In Progress
- **Goal:** Extend multi-tenancy to all existing modules
- **Implementation:**
  - Add `organization_id` foreign key to:
    - `users` table
    - `roles` table
    - `access_rights` table
    - `audit_logs` table
  - Create migration for existing data
  - Update all repository queries to use tenant scopes
  - Update all usecases to enforce organization context
  - Add organization-level permission checks

### 6.3 Organization Invitation System - PENDING

- **Status:** [ ] Pending
- **Goal:** Allow organization owners/admins to invite new members
- **Features:**
  - `organization_invitations` table with token and expiry
  - `POST /organizations/:id/invitations` - Send invitation
  - `POST /organizations/invitations/accept` - Accept invitation
  - Email notification with invitation link
  - Configurable invitation expiry (default 7 days)
  - Role assignment on acceptance

### 6.4 Organization Settings & Configuration - PENDING

- **Status:** [ ] Pending
- **Goal:** Provide organization-level customization
- **Features:**
  - `organization_settings` JSONB column or separate table
  - Configurable settings:
    - Branding (logo, colors, custom domain)
    - Feature flags (enable/disable modules)
    - Security policies (password requirements, session timeout)
    - Notification preferences
  - `PATCH /organizations/:id/settings` endpoint
  - Settings inheritance for organization members
