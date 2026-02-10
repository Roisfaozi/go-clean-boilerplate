# Access Rights Reference

This document provides a detailed reference for the **Access Rights** (grouped permissions) available in the system. These rights are assigned to roles to grant access to specific sets of API endpoints.

> **Note:** This configuration is based on the system seeder (`db/seeds/main.go`). Custom deployments may modify these mappings.

## 1. Access Right Definitions

| Access Right Key        | Description                             | Included Endpoints (Methods & Paths)                                                                                                                                                                                                                                             |
| :---------------------- | :-------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`dashboard:view`**    | View dashboard statistics               | `GET /api/v1/stats/summary`<br>`GET /api/v1/stats/activity`<br>`GET /api/v1/stats/insights`                                                                                                                                                                                      |
| **`user:view`**         | View user list and details              | `GET /api/v1/users/`<br>`POST /api/v1/users/search`<br>`GET /api/v1/users/:id`                                                                                                                                                                                                   |
| **`user:manage`**       | Manage user status/deletion             | `PATCH /api/v1/users/:id/status`<br>`DELETE /api/v1/users/:id`                                                                                                                                                                                                                   |
| **`org:view`**          | View organizations                      | `GET /api/v1/organizations/me`<br>`GET /api/v1/organizations/:id`<br>`GET /api/v1/organizations/slug/:slug`                                                                                                                                                                      |
| **`org:manage`**        | Create/Update/Delete organizations      | `POST /api/v1/organizations`<br>`PUT /api/v1/organizations/:id`<br>`DELETE /api/v1/organizations/:id`                                                                                                                                                                            |
| **`member:manage`**     | Manage organization members             | `POST /api/v1/organizations/:id/members/invite`<br>`GET /api/v1/organizations/:id/members`<br>`PATCH /api/v1/organizations/:id/members/:userId`<br>`DELETE /api/v1/organizations/:id/members/:userId`                                                                            |
| **`presence:view`**     | View member presence                    | `GET /api/v1/organizations/:id/presence`                                                                                                                                                                                                                                         |
| **`project:view`**      | View projects                           | `GET /api/v1/projects`<br>`GET /api/v1/projects/:id`                                                                                                                                                                                                                             |
| **`project:manage`**    | Create/Update/Delete projects           | `POST /api/v1/projects`<br>`PUT /api/v1/projects/:id`<br>`DELETE /api/v1/projects/:id`                                                                                                                                                                                           |
| **`role:view`**         | View roles                              | `GET /api/v1/roles`<br>`POST /api/v1/roles/search`                                                                                                                                                                                                                               |
| **`role:manage`**       | Create/Update/Delete roles              | `POST /api/v1/roles`<br>`PUT /api/v1/roles/:id`<br>`DELETE /api/v1/roles/:id`                                                                                                                                                                                                    |
| **`permission:view`**   | View permissions and inheritance        | `GET /api/v1/permissions`<br>`GET /api/v1/permissions/:role`<br>`GET /api/v1/permissions/roles/:role/users`<br>`GET /api/v1/permissions/:role/parents`<br>`GET /api/v1/permissions/resources`<br>`GET /api/v1/permissions/inheritance-tree`                                      |
| **`permission:manage`** | Manage permissions and role assignments | `POST /api/v1/permissions/assign-role`<br>`DELETE /api/v1/permissions/revoke-role`<br>`POST /api/v1/permissions/grant`<br>`PUT /api/v1/permissions`<br>`DELETE /api/v1/permissions/revoke`<br>`POST /api/v1/permissions/inheritance`<br>`DELETE /api/v1/permissions/inheritance` |
| **`access:view`**       | View access rights definitions          | `GET /api/v1/access-rights`<br>`POST /api/v1/access-rights/search`<br>`POST /api/v1/endpoints/search`                                                                                                                                                                            |
| **`access:manage`**     | Manage access rights definitions        | `POST /api/v1/access-rights`<br>`DELETE /api/v1/access-rights/:id`<br>`POST /api/v1/access-rights/link`<br>`POST /api/v1/access-rights/unlink`<br>`POST /api/v1/endpoints`<br>`DELETE /api/v1/endpoints/:id`                                                                     |
| **`audit:view`**        | View and export audit logs              | `POST /api/v1/audit-logs/search`<br>`GET /api/v1/audit-logs/export`                                                                                                                                                                                                              |

## 2. Default Role Assignments

These are the default Access Rights assigned to the standard roles.

| Role                  | Assigned Access Rights                                                                                                                                                              |
| :-------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`role:superadmin`** | _All Permissions_ (via wildcard policy `*`, `*`)                                                                                                                                    |
| **`role:admin`**      | `dashboard:view`<br>`user:view`<br>`role:view`, `role:manage`<br>`project:view`, `project:manage`<br>`org:view`, `org:manage`<br>`member:manage`<br>`presence:view`<br>`audit:view` |
| **`role:user`**       | `dashboard:view`<br>`project:view`<br>`org:view`<br>`presence:view`                                                                                                                 |

## 3. Excluded Endpoints

The following endpoints are **NOT** managed via Access Rights because they are either Public or strictly Self-Service (Authenticated User context only).

### Public Endpoints

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/verify-email`
- `POST /api/v1/auth/refresh`
- `GET /api/health`
- `GET /api/docs/*any`

### Self-Service (Authenticated Context)

- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`
- `PUT /api/v1/users/me`
- `PATCH /api/v1/users/me/avatar`
- `POST /api/v1/organizations` (Creation is allowed for all authenticated users)
- `GET /api/v1/organizations/me`
