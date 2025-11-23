# Casbin DB - Go Modular API

This project is a modular, role-based access control (RBAC) API built with Go, Gin, GORM, and Casbin. It serves as a robust starter template for creating secure and scalable backend services.

---

## 🚀 Features

-   **Modular Architecture**: Code is organized by feature modules (`user`, `auth`, `role`, `permission`, `access`), promoting separation of concerns.
-   **Clean Architecture**: Follows a strict layered architecture (Handler, UseCase, Repository) for testability and maintainability.
-   **RBAC with Casbin**: Granular access control managed by Casbin policies stored in the database.
-   **JWT Authentication**: Secure, stateless authentication using JSON Web Tokens.
-   **Database Migrations**: Uses `golang-migrate` for version-controlled database schema management.
-   **Swagger Documentation**: Automatically generated API documentation via `swaggo`.
-   **Configuration Management**: Centralized configuration loaded from `.env` files.
-   **Comprehensive Testing**: Includes unit tests (with mocks) and a full Postman collection for E2E testing.
-   **Live Reloading**: Uses `air` for hot-reloading during development.

---

## 🛠️ Tech Stack

-   **Language**: Go
-   **Web Framework**: Gin
-   **ORM**: GORM
-   **Authorization**: Casbin
-   **Authentication**: JWT (dgrijalva/jwt-go)
-   **Database**: MySQL
-   **Migrations**: golang-migrate
-   **API Documentation**: Swaggo
-   **Live Reload**: Air

---

## 🏗️ Project Structure

The project follows a Clean Architecture-inspired, modular layout:

```
├── cmd/api/            # Main application entrypoint
├── db/                 # Database migrations and seeds
├── docs/               # Generated Swagger files
├── internal/
│   ├── config/         # Application configuration
│   ├── middleware/     # Gin middleware (Auth, CORS, Casbin)
│   ├── modules/        # Business logic modules (user, auth, etc.)
│   │   └── [module]/
│   │       ├── delivery/ # Handlers (HTTP layer)
│   │       ├── usecase/    # Business logic
│   │       ├── repository/ # Data access layer
│   │       ├── model/      # Data Transfer Objects (DTOs)
│   │       └── entity/     # GORM database models
│   ├── router/         # Gin router setup
│   └── utils/          # Shared utilities (response, JWT, etc.)
├── Makefile            # Helper commands
├── docker-compose.yml  # Docker setup for services
└── go.mod              # Go modules
```

---

## ⚙️ Setup & Running the Project

### Prerequisites

-   Go (1.18+)
-   Docker and Docker Compose
-   `make`
-   `golang-migrate` (can be installed with `make migrate-install`)
-   `swag` CLI (for regenerating docs: `go install github.com/swaggo/swag/cmd/swag@latest`)
-   `air` (for live reload: `go install github.com/cosmtrek/air@latest`)

### 1. Environment Configuration

Copy the `.env.example` file to `.env` and fill in the required database and application variables.

```sh
cp .env.example .env
```

### 2. Running with Docker (Recommended)

This is the easiest way to get the entire stack (Go app, MySQL database) running.

```sh
docker-compose up -d
```

### 3. Running Locally (with `make`)

If you have a local MySQL instance running, you can use the `Makefile` commands.

```sh
# 1. Run database migrations
make migrate-up

# 2. Run the application with live reload
air

# OR run without live reload
make run
```

### 4. Running Tests

To run the full suite of unit tests:

```sh
make test
```

### 5. API Documentation

-   **Swagger UI**: Once the server is running, access the interactive API documentation at `http://localhost:8080/api/docs/index.html`.
-   **Postman**: Import `Casbin Project API.postman_collection.json` and `Casbin Project Env.postman_environment.json` into Postman to run end-to-end tests. A detailed guide is available in `POSTMAN_TEST_CASES.md`.

---

## 🌐 API Overview

The API is versioned and prefixed with `/api/v1`.

### Main Endpoints

-   **Authentication** (`/auth`)
    -   `POST /login`: Logs in a user, returns JWT.
    -   `POST /refresh`: Refreshes an access token.
    -   `POST /logout`: Logs out a user.
-   **Users** (`/users`)
    -   `POST /register`: Creates a new user account.
    -   `GET /me`: Retrieves the current user's profile.
    -   `PUT /me`: Updates the current user's profile.
    -   `GET /`: [Admin] Retrieves a list of all users (supports pagination/filtering).
    -   `GET /{id}`: [Admin] Retrieves a specific user by ID.
    -   `DELETE /{id}`: [Admin] Deletes a specific user.
-   **Roles** (`/roles`)
    -   `POST /`: [Admin] Creates a new role.
    -   `GET /`: [Admin] Lists all roles.
-   **Permissions** (`/permissions`)
    -   `POST /grant`: [Admin] Grants a permission to a role.
    -   `DELETE /revoke`: [Admin] Revokes a permission from a role.
    -   `GET /`: [Admin] Lists all permissions.
    -   `GET /{role}`: [Admin] Lists permissions for a specific role.
    -   `POST /assign-role`: [Admin] Assigns a role to a user.

See the Swagger UI or Postman collection for detailed request/response models.
