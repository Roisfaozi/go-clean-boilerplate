# Casbin Project API - Postman Collection & Test Documentation

This document provides a detailed overview of the Postman collection designed to test the Casbin Project API. The collection covers **end-to-end (E2E) testing**, including positive flows, negative scenarios, security checks, and edge cases.

## 📂 Collection Structure

The collection is organized into logical folders representing different modules of the application.

### 1. Users
Handles user registration and management.

*   **Register New User** (`POST /users/register`)
    *   *Scenario*: Successful user registration.
    *   *Tests*: Checks for 201 Created status, valid JSON response, and presence of `userId`. Saves `userId`, `username`, and `password` to environment variables for subsequent tests.
*   **Register User with Existing Username** (`POST /users/register`)
    *   *Scenario*: Attempt to register a user with a username that already exists.
    *   *Tests*: Checks for 409 Conflict status (or appropriate error code) and validates the error message.
*   **Register User with Bad Payload** (`POST /users/register`)
    *   *Scenario*: Sends various invalid payloads (missing fields, empty body, wrong types).
    *   *Tests*: Checks for 400 Bad Request status.
*   **[SECURITY] POST User with Malformed JSON** (`POST /users/register`)
    *   *Scenario*: Sends syntactically incorrect JSON.
    *   *Tests*: Checks for 400 Bad Request and ensures the server handles it gracefully (no panic/stack trace).
*   **Get Current User** (`GET /users/me`)
    *   *Scenario*: Retrieve profile of the logged-in user.
    *   *Auth*: Bearer Token (`{{authToken}}`).
    *   *Tests*: Checks for 200 OK status and validates that the returned ID matches the logged-in user.
*   **Update Current User** (`PUT /users/me`)
    *   *Scenario*: Update the logged-in user's profile (e.g., fullname).
    *   *Auth*: Bearer Token (`{{authToken}}`).
    *   *Tests*: Checks for 200 OK status and validates the updated data.
*   **[Admin] Get All Users** (`GET /users`)
    *   *Scenario*: Retrieve a list of all registered users.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status and verifies the response is an array. Includes access control tests (admin success vs regular user failure).
*   **[Admin] Get User By ID** (`GET /users/:id`)
    *   *Scenario*: Retrieve details of a specific user by ID.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status and validates the correct user data is returned.
*   **[Admin] Delete User** (`DELETE /users/:id`)
    *   *Scenario*: Delete a specific user.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status.

### 2. Authentication
Handles login, token refresh, and logout.

*   **Login User** (`POST /auth/login`)
    *   *Scenario*: Authenticate a registered user.
    *   *Tests*: Checks for 200 OK status and saves the `accessToken` to the `authToken` environment variable.
*   **Refresh Token** (Sub-folder)
    *   **[Positive] Refresh with Valid Cookie**: Uses the HTTP-only cookie from login to get a new access token.
    *   **[Negative] Refresh with No Cookie**: Attempts refresh without a cookie (Expect 401).
    *   **[Negative] Refresh with Invalid Token**: Attempts refresh with a forged cookie (Expect 401).
*   **Logout** (Sub-folder)
    *   **[Positive] Logout with Valid Token**: Logs out the user, invalidating the refresh token cookie.
    *   **[Negative] Logout with No Token**: Attempts logout without authentication (Expect 401).

### 3. Roles
Manages user roles.

*   **Create a New Role** (`POST /roles`)
    *   *Scenario*: Admin creates a new role definition.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 201 Created status and validates the role name.
*   **List All Roles** (`GET /roles`)
    *   *Scenario*: Admin retrieves all available roles.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status and verifies response is an array.

### 4. Permissions (Casbin Policies)
Manages granular permissions (policies) in Casbin.

*   **Add Policy (Grant Permission)** (`POST /permissions/grant`)
    *   *Scenario*: Grants a specific permission (role + path + method) to a role.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 201 Created status.
*   **View Policies** (`GET /permissions`)
    *   *Scenario*: Lists all active Casbin policies.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status and array response.
*   **Get Permissions for Role** (`GET /permissions/:role`)
    *   *Scenario*: Gets all permissions assigned to a specific role.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status.
*   **Assign Role to User** (`POST /permissions/assign-role`)
    *   *Scenario*: Assigns a role (e.g., 'role_admin') to a specific user ID.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status.
*   **Update Permission** (`PUT /permissions`)
    *   *Scenario*: Modifies an existing policy.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status.
*   **Remove Policy (Revoke Permission)** (`DELETE /permissions/revoke`)
    *   *Scenario*: Removes a specific permission rule.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status.
*   **[SETUP] Grant Admin Access to Get All Users** (`POST /permissions/grant`)
    *   *Description*: A helper request to ensure the admin role has permission to access the `/users` endpoint for testing purposes.

### 5. Access Scenarios (Protected Routes)
Tests the enforcement of RBAC policies.

*   **[SETUP] Prerequisite Roles & Tokens**: A set of requests to register and login users with different personas (`admin`, `editor`, `viewer`) and assign them roles.
*   **[SCENARIO] Article Access**:
    *   **[VIEWER] GET Articles (Should Succeed)**: Tests that a Viewer role *can* read.
    *   **[VIEWER] POST Article (Should be Forbidden)**: Tests that a Viewer role *cannot* write (Expect 403).
    *   **[EDITOR] POST Article (Should Succeed)**: Tests that an Editor role *can* write.

### 6. Endpoints
Manages API endpoint definitions (metadata).

*   **Create an Endpoint** (`POST /endpoints`)
    *   *Scenario*: Registers a new API endpoint in the system database.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 201 Created status.

### 7. Access Rights
Manages abstract access rights and linking them to endpoints.

*   **Create an Access Right** (`POST /access-rights`)
    *   *Scenario*: Creates a high-level access right (e.g., "document:read").
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 201 Created status.
*   **List All Access Rights** (`GET /access-rights`)
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status.
*   **Link Endpoint to Access Right** (`POST /access-rights/link`)
    *   *Scenario*: Associates a concrete API endpoint with an abstract access right.
    *   *Auth*: Bearer Token (`{{adminToken}}`).
    *   *Tests*: Checks for 200 OK status.

### 8. Happy Path Workflow
A sequential folder designed to be run as a complete integration test. It simulates a new user registering, logging in, getting a role assigned by an admin, and then successfully accessing a protected resource.

## 🛠 Environment Variables

The collection relies on the following environment variables, which are often set dynamically by the test scripts:

| Variable | Description |
| :--- | :--- |
| `baseURL` | The root URL of your API (e.g., `http://localhost:8080`). |
| `authToken` | JWT access token for the currently tested regular user. |
| `adminToken` | JWT access token for the admin user. |
| `userId` | ID of the user created during the test run. |
| `roleName` | Name of the role created during the test run. |
| `newRoleName` | Input variable for creating a new role. |
| `endpointId` | ID of a created endpoint metadata. |
| `accessRightId` | ID of a created access right. |

## 🚀 How to Run

1.  Import the collection into Postman.
2.  Create an environment with `baseURL` set to your running server's address.
3.  Run the collection (or specific folders) using the **Collection Runner**.
4.  Ensure your database is clean or handles conflicts gracefully (tests are designed to generate unique data where possible, e.g., using `{{$timestamp}}`).
