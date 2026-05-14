# API Access Workflow & Permissions

This guide details the authentication, authorization, and endpoint definitions within the application, implementing secure Role-Based Access Control (RBAC) with Casbin.

---

## 🔐 I. Authentication (Identity)

We use **JWT (JSON Web Tokens)** for stateless authentication, backed by a **Redis-based Session** for instant revocability.

1.  **Login**: User sends credentials to `POST /api/v1/auth/login`.
    *   **Success**: Returns `access_token` (Short-lived) and sets `refresh_token` (HTTP-Only Cookie).
    *   **Audit**: Logs "LOGIN" activity.
2.  **Access Protected Route**: Client sends `Authorization: Bearer <access_token>`.
    *   **Middleware**: Validates signature AND checks if session ID exists in Redis.
3.  **Refresh Token**: Call `POST /api/v1/auth/refresh` with the cookie to rotate sessions.

---

## 🛡️ II. Authorization (Casbin RBAC)

We use **Casbin** with a RESTful model `(Subject, Object, Action)`.

### Policy Structure
-   **Subject**: `role:admin`, `role:user`, or specific `user_id`.
-   **Object**: API Path (e.g., `/api/v1/users`).
-   **Action**: HTTP Method (`GET`, `POST`, `PUT`, `DELETE`).

### 🌳 Advanced Role Hierarchy (Multiple Inheritance)

The system supports a complex role hierarchy where a single role can inherit permissions from **multiple parent roles**.

*   **Mechanism**: Handled via `g` grouping policies in Casbin.
*   **Tree Visualization**: The `GET /permissions/inheritance-tree` endpoint generates a graph of all roles, identifying all parents and children.
*   **Cycle Detection**: To prevent infinite loops in deep hierarchies, the system implements a recursive traversal with a "visited" set. If a circular dependency is detected (e.g., Role A -> Role B -> Role A), the recursion stops safely to prevent system crashes.

### 🛡️ Transactional Integrity

Authorization policies are kept in sync with the database using a **Transactional Enforcer**.

*   **Context Propagation**: All permission operations must use `.WithContext(ctx)` to ensure they use the correct GORM transaction handle.
*   **Atomicity**: If a business operation fails (e.g., user creation), the associated Casbin policy changes (e.g., role assignment) are automatically rolled back along with the database transaction.

---

## 🚀 III. API Endpoints Definition

### 1. Global & Real-time Endpoints
| Method | Path | Description | Access |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/docs/*any` | Swagger/OpenAPI UI | Public |
| `GET` | `/api/health` | Health Check | Public |
| `GET` | `/ws` | WebSocket Endpoint | Authenticated |
| `GET` | `/events` | SSE Stream | Public |

### 2. Authentication Module
| Method | Path | Description | Access |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/login` | User Login | Public |
| `POST` | `/api/v1/auth/refresh` | Refresh Token | Public (Cookie) |
| `POST` | `/api/v1/auth/logout` | User Logout | Authenticated |

### 3. User Module
| Method | Path | Description | Required Role |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/users/register`| Register Account | Public |
| `GET` | `/api/v1/users/me` | Get Own Profile | `role:user` |
| `PUT` | `/api/v1/users/me` | Update Own Profile | `role:user` |
| `PATCH`| `/api/v1/users/me/avatar`| Update Own Avatar | `role:user` |
| `GET` | `/api/v1/users` | List All Users | `role:admin` |
| `POST` | `/api/v1/users/search` | Dynamic User Search | `role:admin` |
| `DELETE`| `/api/v1/users/:id` | Delete User | `role:superadmin` |

### 4. Permission & Role Module
| Method | Path | Description | Required Role |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/roles` | Create Role | `role:superadmin` |
| `GET` | `/api/v1/roles` | List Roles | `role:admin` |
| `POST` | `/api/v1/permissions/grant` | Grant Policy | `role:superadmin` |
| `POST` | `/api/v1/permissions/assign-role`| Assign User Role | `role:superadmin` |
| `GET` | `/api/v1/permissions/inheritance-tree` | Get Full RBAC Graph | `role:admin` |

### 5. Audit & Access Configuration
| Method | Path | Description | Required Role |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/audit-logs/search`| Search Audit Logs | `role:superadmin` |
| `POST` | `/api/v1/access-rights` | Create Access Right| `role:superadmin` |
| `POST` | `/api/v1/endpoints` | Create API Endpoint | `role:superadmin` |

---

## 🚫 IV. Common Errors

| Code | Meaning | Cause |
| :--- | :--- | :--- |
| **401** | Unauthorized | Token missing, invalid, or expired. |
| **403** | Forbidden | Authenticated, but no Casbin policy allows this action. |
| **422** | Validation Error | Request body failed validation (e.g., email format). |
| **429** | Too Many Requests | Rate limit exceeded. |