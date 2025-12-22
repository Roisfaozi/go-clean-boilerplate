# API Overview & Access Control Workflow

This document provides a comprehensive overview of all API routes and endpoints, along with the defined access control workflow and privileges for `superadmin`, `admin`, and `user` roles using Casbin RBAC.

---

## I. API Routes and Endpoints Definition

This section lists all available API routes and their corresponding HTTP methods, descriptions, and access types.

### Global Endpoints

These endpoints are generally public or handle foundational services.

| Method | Path               | Description                       | Access Type |
| :----- | :----------------- | :-------------------------------- | :---------- |
| `GET`  | `/api/docs/*any`   | Swagger/OpenAPI UI                | Public      |
| `GET`  | `/api/health`      | Application Health Check          | Public      |
| `GET`  | `/ws`              | WebSocket Connection Endpoint     | Public      |
| `GET`  | `/events`          | Server-Sent Events (SSE) Stream   | Public      |

### Authentication Module

These endpoints handle user authentication and token management.

**Public Routes (`/api/v1` prefix):**

| Method | Path             | Description            |
| :----- | :--------------- | :--------------------- |
| `POST` | `/auth/login`    | User Login (returns JWT token pair) |
| `POST` | `/auth/refresh`  | Refresh Access Token using Refresh Token |

**Authenticated Routes (`/api/v1` prefix):**

Requires a valid JWT Access Token. These routes are handled by `authMiddleware.ValidateToken()` but may not have specific Casbin policies unless they involve resource access.

| Method | Path             | Description            |
| :----- | :--------------- | :--------------------- |
| `POST` | `/auth/logout`   | User Logout (invalidates refresh token) |

### User Module

These endpoints manage user accounts and profiles.

**Public Routes (`/api/v1` prefix):**

| Method | Path             | Description            |
| :----- | :--------------- | :--------------------- |
| `POST` | `/users/register`| Register a New User Account |

**Authorized Routes (`/api/v1` prefix):**

Requires a valid JWT Access Token and additional Casbin RBAC authorization.

| Method | Path             | Description                                |
| :----- | :--------------- | :----------------------------------------- |
| `GET`  | `/users/me`      | Get Profile of the Currently Authenticated User |
| `PUT`  | `/users/me`      | Update Profile of the Currently Authenticated User |
| `GET`  | `/users`         | List All Users (supports basic query params like `page`, `limit`, `username`, `email`) |
| `POST` | `/users/search`  | Dynamic Search and Filter for Users (supports complex JSON filters in body) |
| `GET`  | `/users/:id`     | Get Details of a Specific User by ID       |
| `DELETE`|`/users/:id`     | Delete a User by ID                        |

### Role Module

These endpoints manage user roles.

**Authorized Routes (`/api/v1` prefix):**

Requires a valid JWT Access Token and Casbin RBAC authorization.

| Method | Path             | Description                                |
| :----- | :--------------- | :----------------------------------------- |
| `POST` | `/roles`         | Create a New Role                          |
| `GET`  | `/roles`         | List All Roles                             |
| `POST` | `/roles/search`  | Dynamic Search and Filter for Roles        |
| `DELETE`|`/roles/:id`     | Delete a Role by ID                        |

### Permission Module

These endpoints directly manage Casbin authorization policies and user-role assignments.

**Authorized Routes (`/api/v1` prefix):**

Requires a valid JWT Access Token and Casbin RBAC authorization.

| Method | Path                       | Description                                    |
| :----- | :------------------------- | :--------------------------------------------- |
| `POST` | `/permissions/assign-role` | Assign a Role to a User                        |
| `POST` | `/permissions/grant`       | Grant a Permission Policy to a Role (`p` rule) |
| `PUT`  | `/permissions`             | Update an Existing Permission Policy           |
| `DELETE`|`/permissions/revoke`      | Revoke a Permission Policy from a Role         |
| `GET`  | `/permissions`             | View All Active Casbin Policies                |
| `GET`  | `/permissions/:role`       | Get Permissions for a Specific Role            |

### Access Module (Endpoints & Access Rights)

These endpoints manage granular access rights and API endpoints.

**Authorized Routes (`/api/v1` prefix):**

Requires a valid JWT Access Token and Casbin RBAC authorization.

| Method | Path                       | Description                                    |
| :----- | :------------------------- | :--------------------------------------------- |
| `POST` | `/access-rights`           | Create a New Access Right (e.g., `user_read`)  |
| `GET`  | `/access-rights`           | List All Access Rights                         |
| `POST` | `/access-rights/search`    | Dynamic Search and Filter for Access Rights    |
| `DELETE`|`/access-rights/:id`       | Delete an Access Right by ID                   |
| POST | `/access-rights/link`      | Link an API Endpoint to an Access Right        |
| POST | `/endpoints`               | Create a New API Endpoint (e.g., `/users GET`) |
| POST | `/endpoints/search`        | Dynamic Search and Filter for Endpoints        |
| DELETE|`/endpoints/:id`           | Delete an Endpoint by ID                       |

### Audit Module

Requires a valid JWT Access Token and Casbin RBAC authorization (usually restricted to `superadmin`).

| Method | Path                  | Description                               |
| :----- | :-------------------- | :---------------------------------------- |
| `POST` | `/audit-logs/search`  | Dynamic Search and Filter for Audit Logs  |

---

## II. Access Control Workflow & Role Privileges (Casbin RBAC)

This section details the privileges associated with each predefined role within the application, enforced by Casbin's Role-Based Access Control (RBAC) model. Privileges are defined as `(role, resource_path, action_method)`.

### Defined Roles:

*   `role:superadmin`
*   `role:admin`
*   `role:user` (Default role for newly registered users)

### Privilege Definitions:

#### 1. `role:superadmin`

**Description:** The highest level of access. `superadmin` has full control over all administrative functions, including managing users, roles, permissions, endpoints, and access rights.

| Access Right Name                 | Action (Method)        | Privileges & Notes                                      |
| :-------------------------------- | :--------------------- | :------------------------------------------------------ |
| `user_management`                 | `all` (`GET`, `POST`, `PUT`, `DELETE`) | Full CRUD access for all users.                         |
| `role_management`                 | `all` (`GET`, `POST`, `PUT`, `DELETE`) | Full CRUD access for all roles.                         |
| `permission_management`           | `all` (`GET`, `POST`, `PUT`, `DELETE`) | Full control over Casbin policies (grant/revoke/assign).|
| `endpoint_configuration`          | `all` (`GET`, `POST`, `PUT`, `DELETE`) | Full control over API endpoint definitions.             |
| `access_right_configuration`      | `all` (`GET`, `POST`, `PUT`, `DELETE`) | Full control over access right definitions and linking. |
| `self_profile_management`         | `GET`, `PUT`           | Manage own user profile.                                |
| *(Direct Path for Logout)*        | `/auth/logout` `POST`  | Logout from current session.                            |
| *(Direct Path for Admin APIs)*    | `/permissions` `GET`   | View all Casbin policies.                               |
| *(Direct Path for Admin APIs)*    | `/permissions/:role` `GET`| View permissions of a specific role.                   |


#### 2. `role:admin`

**Description:** `admin` possesses strong administrative capabilities, primarily focused on viewing and managing operational aspects like users and roles. `admin` typically does not manage core Casbin policies directly or critical infrastructure components.

| Access Right Name                 | Action (Method)        | Privileges & Notes                                      |
| :-------------------------------- | :--------------------- | :------------------------------------------------------ |
| `user_management`                 | `read`, `update` (`GET`, `PUT`, `POST /search`) | View all users, perform dynamic search, update user profiles. (No user creation/deletion by default). |
| `role_management`                 | `read` (`GET`, `POST /search`) | View all roles, perform dynamic search. (No role creation/deletion by default). |
| `endpoint_configuration`          | `read` (`GET`, `POST /search`) | View all registered API endpoints.                      |
| `access_right_configuration`      | `read` (`GET`, `POST /search`) | View all defined access rights.                         |
| `self_profile_management`         | `GET`, `PUT`           | Manage own user profile.                                |
| *(Direct Path for Logout)*        | `/auth/logout` `POST`  | Logout from current session.                            |


#### 3. `role:user`

**Description:** A standard application user. `user` has limited privileges, primarily focused on managing their own profile and accessing resources explicitly granted to them.

| Access Right Name                 | Action (Method)        | Privileges & Notes                                      |
| :-------------------------------- | :--------------------- | :------------------------------------------------------ |
| `self_profile_management`         | `GET`, `PUT`           | View and update own user profile.                       |
| *(Direct Path for Logout)*        | `/auth/logout` `POST`  | Logout from current session.                            |

### 4. Pertimbangan untuk Role Masa Depan (Contoh: `role:supervisor`)

Ketika peran baru seperti `role:supervisor` diperkenalkan, fleksibilitas model `Access Rights` menjadi sangat jelas.

**`role:supervisor`:**

**Deskripsi:** Peran ini mungkin membutuhkan kemampuan untuk melihat data yang lebih luas daripada `user` standar, tetapi tidak memiliki akses administratif penuh seperti `admin`. Mereka mungkin bisa menyetujui, meninjau, atau mengelola entitas spesifik yang relevan dengan pekerjaan mereka.

| Access Right Name                 | Action (Method)        | Privileges & Notes                                      |
| :-------------------------------- | :--------------------- | :------------------------------------------------------ |
| `self_profile_management`         | `GET`, `PUT`           | Manage own user profile.                                |
| `user_management`                 | `read` (`GET`, `POST /search`) | Hanya melihat daftar dan detail user (misal untuk laporan atau audit). |
| `report_viewer`                   | `read` (`GET`)         | Melihat laporan terkait. (Asumsi ada `Access Right` baru untuk laporan) |
| *(Direct Path for Logout)*        | `/auth/logout` `POST`  | Logout dari current session.                            |

---

## III. Konsep Access Rights dan Endpoints dalam RBAC

... (Konten Bagian III sebelumnya) ...

### 5. Contoh Konkret Pemetaan Access Rights ke Endpoints

Berikut adalah contoh pemetaan beberapa `Access Rights` yang umum digunakan ke `Endpoints` yang sesuai. Ini adalah bagaimana `superadmin` akan mengkonfigurasi sistem untuk mengelompokkan izin.

#### A. `Access Right: user_management`
**Deskripsi:** Hak akses untuk mengelola pengguna (CRUD).

| Endpoint Path           | Method |
| :---------------------- | :----- |
| `/api/v1/users`         | `GET`  |
| `/api/v1/users/search`  | `POST` |
| `/api/v1/users`         | `POST` |
| `/api/v1/users/:id`     | `GET`  |
| `/api/v1/users/:id`     | `PUT`  |
| `/api/v1/users/:id`     | `DELETE` |

#### B. `Access Right: role_management`
**Deskripsi:** Hak akses untuk mengelola peran (CRUD).

| Endpoint Path           | Method |
| :---------------------- | :----- |
| `/api/v1/roles`         | `GET`  |
| `/api/v1/roles/search`  | `POST` |
| `/api/v1/roles`         | `POST` |
| `/api/v1/roles/:id`     | `GET`  |
| `/api/v1/roles/:id`     | `PUT`  |
| `/api/v1/roles/:id`     | `DELETE` |

#### C. `Access Right: permission_management`
**Deskripsi:** Hak akses untuk mengelola kebijakan izin Casbin (grant/revoke).

| Endpoint Path                 | Method |
| :---------------------------- | :----- |
| `/api/v1/permissions`         | `GET`  |
| `/api/v1/permissions/:role`   | `GET`  |
| `/api/v1/permissions/grant`   | `POST` |
| `/api/v1/permissions/revoke`  | `DELETE` |
| `/api/v1/permissions/assign-role`| `POST` |
| `/api/v1/permissions`         | `PUT`  |

#### D. `Access Right: endpoint_configuration`
**Deskripsi:** Hak akses untuk mengelola definisi Endpoint API.

| Endpoint Path           | Method |
| :---------------------- | :----- |
| `/api/v1/endpoints`     | `GET`  |
| `/api/v1/endpoints/search`| `POST` |
| `/api/v1/endpoints`     | `POST` |
| `/api/v1/endpoints/:id` | `GET`  |
| `/api/v1/endpoints/:id` | `DELETE` |

#### E. `Access Right: access_right_configuration`
**Deskripsi:** Hak akses untuk mengelola definisi Access Right dan Linking-nya.

| Endpoint Path                 | Method |
| :---------------------------- | :----- |
| `/api/v1/access-rights`       | `GET`  |
| `/api/v1/access-rights/search`| `POST` |
| `/api/v1/access-rights`       | `POST` |
| `/api/v1/access-rights/:id`   | `GET`  |
| `/api/v1/access-rights/:id`   | `DELETE` |
| `/api/v1/access-rights/link`  | `POST` |

#### F. `Access Right: self_profile_management`
**Deskripsi:** Hak akses untuk mengelola profil pengguna sendiri.

| Endpoint Path           | Method |
| :---------------------- | :----- |
| `/users/me`             | `GET`  |
| `/users/me`             | `PUT`  |

---