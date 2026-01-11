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