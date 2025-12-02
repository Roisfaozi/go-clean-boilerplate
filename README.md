# Casbin DB - Go Modular REST API

![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)
![License](https://img.shields.io/badge/License-Apache%202.0-green)
![Architecture](https://img.shields.io/badge/Architecture-Clean%20%26%20Modular-orange)

This project is a production-ready, modular, **Role-Based Access Control (RBAC)** REST API built with Go. It leverages **Gin** for high-performance HTTP handling, **GORM** for database interactions, **Casbin** for robust authorization policy enforcement, and **Redis** for secure session management.

It serves as a solid foundation for building scalable, secure, and maintainable backend services.

---

## 🚀 Key Features

-   **Modular & Clean Architecture**: Codebase is strictly organized by feature modules (`user`, `auth`, `role`, `permission`, `access`) and layers (`Handler` -> `UseCase` -> `Repository`), ensuring scalability and testability.
-   **Advanced RBAC Authorization**: Fine-grained access control using [Casbin](https://casbin.org/). Policies are persisted in the database, allowing dynamic permission updates without restarting the server.
-   **Secure Authentication**:
    -   **JWT (JSON Web Tokens)**: Stateless access tokens carrying user identity and role.
    -   **Session Management**: Stateful refresh tokens stored in **Redis** for secure token rotation and instant revocation (logout/ban).
-   **Real-time Notifications**: Integrated WebSocket support to broadcast events (e.g., user login alerts) to connected clients.
-   **Robust Validation**: Centralized request validation using `go-playground/validator` with user-friendly error messages (HTTP 422).
-   **Standardized Response**: Unified JSON response structure for success (`data`, `paging`) and errors (`message`, `error`), making frontend integration seamless.
-   **Database Migrations**: Version-controlled schema management using `golang-migrate`.
-   **Observability**: Structured logging via `logrus`.
-   **Developer Experience**:
    -   **Swagger/OpenAPI**: Auto-generated interactive API documentation.
    -   **Hot Reload**: Development made easy with `air`.
    -   **Postman Collection**: Ready-to-use collection for end-to-end testing.

---

## 🛠️ Tech Stack

| Category | Technology | Description |
| :--- | :--- | :--- |
| **Language** | [Go (Golang)](https://go.dev/) | Core programming language |
| **Framework** | [Gin](https://github.com/gin-gonic/gin) | High-performance HTTP web framework |
| **Database** | [MySQL](https://www.mysql.com/) | Primary relational database |
| **ORM** | [GORM](https://gorm.io/) | Data access and relationship management |
| **Cache/Session** | [Redis](https://redis.io/) | Session storage and token management |
| **Authorization** | [Casbin](https://casbin.org/) | Authorization library (RBAC model) |
| **Authentication** | [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) | JWT implementation |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) | Database schema migrations |
| **Docs** | [Swaggo](https://github.com/swaggo/swag) | Swagger documentation generator |
| **Test** | [Testify](https://github.com/stretchr/testify) | Assertion and mocking toolkit |

---

## ⚙️ Prerequisites

Ensure you have the following installed on your system:

1.  **Go**: Version 1.21 or higher.
2.  **Docker & Docker Compose**: For running MySQL and Redis services easily.
3.  **Make**: For running automation commands defined in `Makefile`.
4.  **Air** (Optional): For live reloading during development.
    ```bash
    go install github.com/air-verse/air@latest
    ```
5.  **Swag CLI** (Optional): For regenerating API docs.
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```
6.  **Golang Migrate** (Optional): If you want to run migrations manually without the Makefile helper.

---

## 🚀 Getting Started

### 1. Clone the Repository
```bash
git clone https://github.com/yourusername/casbin-db.git
cd casbin-db
```

### 2. Environment Configuration
Copy the example environment file and configure it according to your setup.
```bash
cp .env.example .env
```
*Tip: The default values in `.env.example` usually work out-of-the-box with the provided `docker-compose.yml`.*

### 3. Start Infrastructure (Database & Redis)
Use Docker Compose to spin up MySQL and Redis containers.
```bash
docker-compose up -d
```

### 4. Run Database Migrations
Apply the database schema and seed default data (like `role:admin` and `role:user`).
```bash
make migrate-up
```

### 5. Run the Application
You can run the application in development mode (with hot reload) or standard mode.

**Development Mode (Recommended):**
```bash
air
```
*Or if you don't have `air` installed:*
```bash
go run cmd/api/main.go
```

**Production Build:**
```bash
make build
./bin/api
```

The server will start on **http://localhost:8080** (or the port defined in your `.env`).

---

## 🧪 Testing

### Unit Tests
Run the full suite of unit tests to ensure system integrity.
```bash
make test
```
*Note: This runs `go test ./...` covering all modules.*

### End-to-End Testing (Postman)
We provide a comprehensive Postman collection to test the entire flow:
1.  Import `postman/Casbin_API_Full_Flow.postman_collection.json` into Postman.
2.  Set up your environment variables (`baseURL` = `http://localhost:8080`, `apiVersion` = `v1`).
3.  Run the collection runner to verify Registration -> Login -> RBAC Enforcement -> Cleanup.

For a detailed usage guide, please refer to **[documentation/USAGE.md](documentation/USAGE.md)**.

---

## 📚 API Documentation

### Swagger UI
Interactive API documentation is available at:
> **http://localhost:8080/api/docs/index.html**

### Key Endpoints Overview

| Module | Method | Endpoint | Description | Access |
| :--- | :--- | :--- | :--- | :--- |
| **Auth** | `POST` | `/auth/login` | User login (returns JWT) | Public |
| **Auth** | `POST` | `/auth/refresh` | Refresh access token | Public (Cookie) |
| **User** | `POST` | `/users/register` | Register new user | Public |
| **User** | `GET` | `/users/me` | Get current profile | User |
| **User** | `GET` | `/users` | List all users | Admin |
| **Role** | `POST` | `/roles` | Create new role | Admin |
| **Perm** | `POST` | `/permissions/grant` | Grant permission to role | Admin |
| **Perm** | `POST` | `/permissions/assign-role` | Assign role to user | Admin |

*See [documentation/USAGE.md](documentation/USAGE.md) for detailed workflows.*

---

## 📂 Project Structure

The project follows a standard Go project layout suitable for scalable microservices or monolithic APIs.

```
.
├── .air.toml           # Configuration for Air (live reloading)
├── Makefile            # Automation commands (build, test, migrate, run)
├── README.md           # Main project documentation
├── docker-compose.yml  # Docker services definition (MySQL, Redis)
├── go.mod              # Go dependency definitions
│
├── cmd/
│   └── api/            # Application entry point (main.go)
│
├── db/
│   ├── migrations/     # Database schema migration files (.sql)
│   └── seeds/          # Initial data seeding scripts (e.g. bootstrapping)
│
├── docs/               # Auto-generated Swagger/OpenAPI documentation files
│
├── documentation/      # Project guides and additional documentation
│   ├── USAGE.md        # Detailed guide on how to use the API workflow
│   └── ...
│
├── postman/            # Postman collections for testing
│   ├── Casbin_API_Full_Flow.json  # End-to-end testing collection
│   └── ...
│
└── internal/           # Private application code (not importable by other apps)
    ├── config/         # Configuration loading & app initialization wiring
    ├── middleware/     # HTTP Middlewares (Auth, Casbin Enforcer, CORS)
    ├── router/         # Gin router setup and route registration
    ├── utils/          # Shared utilities (JWT, Response Helper, Validator, WebSocket)
    │
    └── modules/        # Domain-specific modules following Clean Architecture
        ├── auth/       # Authentication logic & JWT handling
        ├── user/       # User management (CRUD)
        ├── role/       # Role management
        ├── permission/ # Permission/Policy management (Casbin)
        └── access/     # Access Right & Endpoint management
            ├── delivery/   # HTTP Handlers (Controller layer)
            ├── usecase/    # Business Logic layer
            ├── repository/ # Data Access layer (DB/Redis)
            ├── model/      # Data structures (DTOs) & Validation structs
            └── entity/     # Database entities (GORM models)
```

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:
1.  Fork the repository.
2.  Create a feature branch (`git checkout -b feature/amazing-feature`).
3.  Commit your changes (`git commit -m 'feat: Add amazing feature'`).
4.  Push to the branch (`git push origin feature/amazing-feature`).
5.  Open a Pull Request.

---

## 📄 License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
