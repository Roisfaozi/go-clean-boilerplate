# API Access Workflow & Permissions

This guide details the authentication, authorization, and endpoint definitions within the application, implementing secure Role-Based Access Control (RBAC) with Casbin.

---

## 🔐 I. Authentication (Identity)

We use **JWT (JSON Web Tokens)** for stateless authentication, backed by a **Redis-based Session** for instant revocability.

1.  **Login**: User sends credentials to `POST /api/v1/auth/login`.
    - **Success**: Returns `access_token` (Short-lived) and sets `refresh_token` (HTTP-Only Cookie).
    - **Audit**: Logs "LOGIN" activity.
2.  **Access Protected Route**: Client sends `Authorization: Bearer <access_token>`.
    - **Middleware**: Validates signature AND checks if session ID exists in Redis.
3.  **Refresh Token**: Call `POST /api/v1/auth/refresh` with the cookie to rotate sessions.

---

## 🛡️ II. Authorization (Casbin RBAC)

We use **Casbin** with a RESTful model `(Subject, Object, Action)`.

### Policy Structure

- **Subject**: `role:admin`, `role:user`, or specific `user_id`.
- **Object**: API Path (e.g., `/api/v1/users`).
- **Action**: HTTP Method (`GET`, `POST`, `PUT`, `DELETE`).

### Role Privileges Summary

| Role              | capabilities                                                                     |
| :---------------- | :------------------------------------------------------------------------------- |
| `role:superadmin` | Full CRUD access to all resources (Users, Roles, Permissions, Logs).             |
| `role:admin`      | Operational management: View users/roles, update profiles. No deletion/granting. |
| `role:user`       | Self-service: View/Update own profile, use WebSockets, Logout.                   |

---

## 🚀 III. API Endpoints Definition

### 1. Global & Real-time Endpoints

| Method | Path             | Description        | Access        |
| :----- | :--------------- | :----------------- | :------------ |
| `GET`  | `/api/docs/*any` | Swagger/OpenAPI UI | Public        |
| `GET`  | `/api/health`    | Health Check       | Public        |
| `GET`  | `/ws`            | WebSocket Endpoint | Authenticated |
| `GET`  | `/events`        | SSE Stream         | Public        |

### 2. Authentication Module

| Method | Path                               | Description                     | Access          |
| :----- | :--------------------------------- | :------------------------------ | :-------------- |
| `POST` | `/api/v1/auth/login`               | User Login                      | Public          |
| `POST` | `/api/v1/auth/register`            | User Registration               | Public          |
| `POST` | `/api/v1/auth/forgot-password`     | Request Password Reset          | Public          |
| `POST` | `/api/v1/auth/reset-password`      | Reset Password                  | Public          |
| `POST` | `/api/v1/auth/verify-email`        | Verify Email                    | Public          |
| `POST` | `/api/v1/auth/refresh`             | Refresh Token                   | Public (Cookie) |
| `POST` | `/api/v1/auth/logout`              | User Logout                     | Authenticated   |
| `GET`  | `/api/v1/auth/me`                  | Get Current User (Auth Context) | Authenticated   |
| `POST` | `/api/v1/auth/resend-verification` | Resend Verification Email       | Authenticated   |
| `POST` | `/api/v1/auth/ticket`              | Get One-time Ticket             | Authenticated   |

### 3. User Module

| Method   | Path                       | Description         | Required Role     |
| :------- | :------------------------- | :------------------ | :---------------- |
| `POST`   | `/api/v1/users/register`   | Register Account    | Public            |
| `GET`    | `/api/v1/users/me`         | Get Own Profile     | `role:user`       |
| `PUT`    | `/api/v1/users/me`         | Update Own Profile  | `role:user`       |
| `PATCH`  | `/api/v1/users/me/avatar`  | Update Own Avatar   | `role:user`       |
| `GET`    | `/api/v1/users`            | List All Users      | `role:admin`      |
| `POST`   | `/api/v1/users/search`     | Dynamic User Search | `role:admin`      |
| `GET`    | `/api/v1/users/:id`        | Get User Details    | `role:admin`      |
| `PATCH`  | `/api/v1/users/:id/status` | Update User Status  | `role:superadmin` |
| `DELETE` | `/api/v1/users/:id`        | Delete User         | `role:superadmin` |

### 4. Permission & Role Module

| Method   | Path                                    | Description              | Required Role     |
| :------- | :-------------------------------------- | :----------------------- | :---------------- |
| `POST`   | `/api/v1/roles`                         | Create Role              | `role:superadmin` |
| `GET`    | `/api/v1/roles`                         | List Roles               | `role:admin`      |
| `POST`   | `/api/v1/roles/search`                  | Dynamic Role Search      | `role:admin`      |
| `PUT`    | `/api/v1/roles/:id`                     | Update Role              | `role:superadmin` |
| `DELETE` | `/api/v1/roles/:id`                     | Delete Role              | `role:superadmin` |
| `POST`   | `/api/v1/permissions/grant`             | Grant Policy             | `role:superadmin` |
| `DELETE` | `/api/v1/permissions/revoke`            | Revoke Policy            | `role:superadmin` |
| `POST`   | `/api/v1/permissions/assign-role`       | Assign User Role         | `role:superadmin` |
| `DELETE` | `/api/v1/permissions/revoke-role`       | Revoke User Role         | `role:superadmin` |
| `PUT`    | `/api/v1/permissions`                   | Batch Update Permissions | `role:superadmin` |
| `POST`   | `/api/v1/permissions/inheritance`       | Add Inheritance          | `role:superadmin` |
| `DELETE` | `/api/v1/permissions/inheritance`       | Remove Inheritance       | `role:superadmin` |
| `GET`    | `/api/v1/permissions`                   | List All Policies        | `role:superadmin` |
| `GET`    | `/api/v1/permissions/:role`             | Get Role Policies        | `role:superadmin` |
| `GET`    | `/api/v1/permissions/:role/parents`     | Get Parent Roles         | `role:superadmin` |
| `GET`    | `/api/v1/permissions/roles/:role/users` | Get Users in Role        | `role:superadmin` |
| `GET`    | `/api/v1/permissions/resources`         | Get Resource Aggregation | `role:superadmin` |
| `GET`    | `/api/v1/permissions/inheritance-tree`  | Get Inheritance Tree     | `role:superadmin` |
| `POST`   | `/api/v1/permissions/check-batch`       | Batch Permission Check   | Authenticated     |

### 5. Organization Module (Multi-Tenancy)

| Method   | Path                                       | Description           | Access        |
| :------- | :----------------------------------------- | :-------------------- | :------------ |
| `GET`    | `/api/v1/organizations/me`                 | List My Organizations | Authenticated |
| `POST`   | `/api/v1/organizations`                    | Create Organization   | Authenticated |
| `POST`   | `/api/v1/organizations/invitations/accept` | Accept Invitation     | Public        |
| `GET`    | `/api/v1/organizations/:id`                | Get Details           | Member        |
| `GET`    | `/api/v1/organizations/slug/:slug`         | Get Details by Slug   | Member        |
| `PUT`    | `/api/v1/organizations/:id`                | Update Settings       | Owner/Admin   |
| `GET`    | `/api/v1/organizations/:id/members`        | List Members          | Member        |
| `POST`   | `/api/v1/organizations/:id/members/invite` | Invite Member         | Owner/Admin   |
| `PATCH`  | `/api/v1/organizations/:id/members/:uid`   | Update Member Role    | Owner/Admin   |
| `DELETE` | `/api/v1/organizations/:id/members/:uid`   | Remove Member         | Owner/Admin   |
| `DELETE` | `/api/v1/organizations/:id`                | Delete Organization   | Owner         |
| `GET`    | `/api/v1/organizations/:id/presence`       | Get Member Presence   | Member        |

### 6. Project Module (Multi-Tenancy)

| Method   | Path                   | Description    | Access |
| :------- | :--------------------- | :------------- | :----- |
| `GET`    | `/api/v1/projects`     | List Projects  | Member |
| `POST`   | `/api/v1/projects`     | Create Project | Member |
| `GET`    | `/api/v1/projects/:id` | Get Project    | Member |
| `PUT`    | `/api/v1/projects/:id` | Update Project | Member |
| `DELETE` | `/api/v1/projects/:id` | Delete Project | Member |

### 7. Stats Module

| Method | Path                     | Description       | Access        |
| :----- | :----------------------- | :---------------- | :------------ |
| `GET`  | `/api/v1/stats/summary`  | Dashboard Summary | Authenticated |
| `GET`  | `/api/v1/stats/activity` | Activity Stats    | Authenticated |
| `GET`  | `/api/v1/stats/insights` | System Insights   | Authenticated |

### 8. Audit & Access Configuration

| Method   | Path                           | Description                | Required Role     |
| :------- | :----------------------------- | :------------------------- | :---------------- |
| `POST`   | `/api/v1/audit-logs/search`    | Search Audit Logs          | `role:superadmin` |
| `GET`    | `/api/v1/audit-logs/export`    | Export Audit Logs          | `role:superadmin` |
| `GET`    | `/api/v1/access-rights`        | List Access Rights         | `role:superadmin` |
| `POST`   | `/api/v1/access-rights`        | Create Access Right        | `role:superadmin` |
| `POST`   | `/api/v1/access-rights/search` | Search Access Rights       | `role:superadmin` |
| `POST`   | `/api/v1/access-rights/link`   | Link Endpoint to Group     | `role:superadmin` |
| `POST`   | `/api/v1/access-rights/unlink` | Unlink Endpoint from Group | `role:superadmin` |
| `DELETE` | `/api/v1/access-rights/:id`    | Delete Access Right        | `role:superadmin` |
| `POST`   | `/api/v1/endpoints`            | Create API Endpoint        | `role:superadmin` |
| `POST`   | `/api/v1/endpoints/search`     | Search Endpoints           | `role:superadmin` |
| `DELETE` | `/api/v1/endpoints/:id`        | Delete Endpoint            | `role:superadmin` |

---

## 🗝️ IV. Access Rights Grouping (Granular Permissions)

In addition to roles, the system groups specific endpoints into **Access Rights**. This allows for granular permission management. These rights are assigned to roles in the database.

For a complete reference of all Access Rights and their included endpoints, please refer to:
👉 **[Access Rights Reference](./ACCESS_RIGHTS_REFERENCE.md)**

---

## 🚫 V. Common Errors

| Code    | Meaning           | Cause                                                   |
| :------ | :---------------- | :------------------------------------------------------ |
| **401** | Unauthorized      | Token missing, invalid, or expired.                     |
| **403** | Forbidden         | Authenticated, but no Casbin policy allows this action. |
| **422** | Validation Error  | Request body failed validation (e.g., email format).    |
| **429** | Too Many Requests | Rate limit exceeded.                                    |
