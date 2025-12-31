# Go Clean Boilerplate - Enterprise Modular REST API

![Go Version](https://img.shields.io/badge/Go-1.25.5%2B-blue)
![License](https://img.shields.io/badge/License-Apache%202.0-green)
![Architecture](https://img.shields.io/badge/Architecture-Clean%20%26%20Modular-orange)
![Testing](https://img.shields.io/badge/Testing-Unit%2C%20Integration%2C%20E2E-success)
![Dynamic Search](https://img.shields.io/badge/Dynamic%20Search-Enabled-blueviolet)
![Realtime](https://img.shields.io/badge/Realtime-Distributed%20WS%20%26%20SSE-ff69b4)

Enterprise-ready Go boilerplate implementing Clean Architecture, RBAC with Casbin, Modular Audit Logging, and Distributed WebSocket scaling.

---

## 🚀 Core Features

-   **Clean Architecture**: Strict separation of concerns (Delivery, UseCase, Repository, Entity).
-   **Advanced RBAC with Casbin**: Fine-grained access control using GORM adapter. Policies are stored in the database for dynamic updates.
-   **Distributed WebSockets**: Scalable WebSocket management using **Redis Pub/Sub** backplane, allowing multi-node synchronization.
-   **Modular Audit Logging**: Synchronous activity tracking (LOGIN, REGISTER, UPDATE, DELETE) integrated directly into business UseCases.
-   **Dynamic Search & Filtering**: Secure, reusable query builder supporting complex clauses, range filters, and dynamic sorting.
-   **Secure Authentication**: JWT-based auth with stateful session management in Redis for instant token revocation.
-   **Real-time SSE**: Server-Sent Events manager for live one-way data streaming.
-   **Hardened Security**: 
    -   UseCase-level validation (Regex email, password strength).
    -   Automatic HTTP security headers.
    -   Go 1.25.5 for critical security fixes.
-   **Comprehensive Testing**:
    -   **Unit Tests**: Fast, mock-based verification of logic.
    -   **Integration Tests**: Lightweight testing using **Singleton Testcontainers** pattern.
    -   **E2E Tests**: Full HTTP lifecycle validation.

---

## 🛠️ Technology Stack

| Category | Technology | Description |
| :--- | :--- | :--- |
| **Language** | [Go 1.25.5+](https://go.dev/) | Core programming language |
| **Framework** | [Gin Gonic](https://github.com/gin-gonic/gin) | High-performance HTTP framework |
| **Database** | [MySQL 8.0](https://www.mysql.com/) | Primary relational database |
| **Cache/Session** | [Redis 7](https://redis.io/) | Session storage & WS Pub/Sub backplane |
| **Authorization** | [Casbin](https://casbin.org/) | RBAC model & Policy enforcement |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) | Database schema management |
| **Testing** | [Testcontainers](https://testcontainers.com/) | Real instances for integration tests |
| **Authentication** | [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) | JWT implementation |
| **Realtime** | [Gorilla WebSocket](https://github.com/gorilla/websocket) | WebSocket implementation |
| **Realtime** | Custom SSE Manager | Server-Sent Events implementation |
---

## 🏁 Getting Started

### ⚙️ Prerequisites

Ensure you have the following installed on your system:

1.  **Go**: Version 1.25.5 or higher.
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
7.  **C/C++ Compiler (GCC/MinGW-w64)**: Required for running repository tests that use SQLite (due to CGO). Ensure `gcc` is in your system's PATH.


### Installation
1.  **Clone & Configure**:
    ```bash
    git clone https://github.com/Roisfaozi/go-clean-boilerplate.git
    cd go-clean-boilerplate
    cp .env.example .env
    ```
2.  **Start Infrastructure**:
    ```bash
    docker-compose up -d
    ```
3.  **Run Migrations & Seeding**:
    ```bash
    make migrate-up
    ```
4.  **Run Application**:
    ```bash
    make run
    ```

---



## 📖 API Usage Guides

### Accessing API Documentation (Swagger UI)
Interactive API documentation is available at:
> **http://localhost:8080/api/docs/index.html**

### Postman Collections
Import the Postman collections from the `postman/` directory to explore and test the API:
-   `Casbin Project API.postman_collection.json`: Main collection for core CRUD, Auth, and RBAC flows.
-   `Casbin Project API - Dynamic Search.postman_collection.json`: Dedicated collection for testing all dynamic search endpoints with various filter/sort scenarios.
-   `Casbin Project API - Realtime.postman_collection.json`: Examples for WebSocket and Server-Sent Events (SSE).

### Key Features & Endpoints

| Feature | Method | Endpoint | Description | Access | Documentation |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Auth** | `POST` | `/auth/login` | User login (returns JWT) | Public | |
| **Auth** | `POST` | `/auth/refresh` | Refresh access token | Public (Cookie) | |
| **User** | `POST` | `/users/register` | Register new user | Public | |
| **User** | `GET` | `/users/me` | Get current profile | User | |
| **User** | `GET` | `/users` | List all users (basic filtering) | Admin | [GET vs Dynamic Search](#perbedaan-antara-findall-http-get-dan-dynamic-search-http-post-search) |
| **User** | `POST` | `/users/search` | Dynamic search & filter for users | Admin | [Dynamic Search Examples](#dynamic-search-api-examples-curl) |
| **Role** | `GET` | `/roles` | List all roles | Admin | [GET vs Dynamic Search](#perbedaan-antara-findall-http-get-dan-dynamic-search-http-post-search) |
| **Role** | `POST` | `/roles/search` | Dynamic search & filter for roles | Admin | [Dynamic Search Examples](#dynamic-search-api-examples-curl) |
| **Access** | `POST` | `/endpoints/search` | Dynamic search & filter for endpoints | Admin | [Dynamic Search Examples](#dynamic-search-api-examples-curl) |
| **Access** | `POST` | `/access-rights/search` | Dynamic search & filter for access rights | Admin | [Dynamic Search Examples](#dynamic-search-api-examples-curl) |
| **SSE** | `GET` | `/events` | Server-Sent Events stream | Public | [SSE Usage Guide](#server-sent-events-sse-usage-guide) |
| **WebSocket** | `GET` | `/ws` | WebSocket connection | Public | |

### Detailed Usage Guides
-   **Dynamic Search Examples**: See `documentation/DYNAMIC_SEARCH_EXAMPLES.md` for `curl` examples covering various filter types and scenarios.
-   **SSE Usage Guide**: See `documentation/SSE_USAGE.md` for implementation details and frontend client examples for Server-Sent Events.
-   **GET vs. Dynamic Search**: See `documentation/GET_VS_DYNAMIC_SEARCH.md` for a clear breakdown on when to use each search approach.



---

## 🧪 Testing Strategy

We use a layered testing strategy optimized for both speed and reliability.

| Command | Type | Description |
| :--- | :--- | :--- |
| `make test-unit` | **Unit** | Runs mock-based tests for internal/pkg logic. |
| `make test-integration` | **Integration** | Uses **Singleton Containers** for DB/Redis logic. |
| `make test-e2e` | **E2E** | Validates full HTTP request/response flows. |
| `make test-all` | **Full Suite** | Executes all test layers sequentially. |
| `make test-coverage` | **Coverage** | Generates an interactive HTML coverage report. |

> **Note**: Integration and E2E tests require Docker. We use a **Singleton Container Pattern** to reuse a single database/redis instance across the entire suite, drastically reducing execution time and resource usage.

---

## 📂 Project Structure

The project follows a standard Go project layout suitable for scalable microservices or monolithic APIs.

```
.
├── .air.toml           # Configuration for Air (live reloading)
├── Makefile            # Automation commands (build, test, migrate, run, mocks)
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
│   ├── USAGE.md                        # Detailed guide on how to use the API workflow
│   ├── DYNAMIC_SEARCH_EXAMPLES.md      # Curl examples for dynamic search
│   └── GET_VS_DYNAMIC_SEARCH.md        # Explains GET vs POST search approaches
│   └── SSE_USAGE.md                    # Guide for Server-Sent Events (SSE)
│
├── postman/            # Postman collections for testing
│   ├── Casbin Project API.postman_collection.json         # Main collection
│   ├── Casbin Project API - Dynamic Search.postman_collection.json # Dynamic search tests
│   ├── Casbin Project API - Realtime.postman_collection.json # Realtime features (WS, SSE)
│   └── ...
│
└── internal/           # Private application code (not importable by other apps)
    ├── config/         # Configuration loading & app initialization wiring
    ├── middleware/     # HTTP Middlewares (Auth, Casbin Enforcer, CORS)
    ├── router/         # Gin router setup and route registration
    ├── utils/          # Shared utilities (JWT, Response Helper, Validator, WebSocket, SSE, QueryBuilder)
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

## 📜 Documentation Links

- [Testing Strategy](./documentation/TESTING_STRATEGY.md)
- [Integration Testing Guide](./documentation/INTEGRATION_TESTING_GUIDE.md)
- [Distributed WebSocket Usage](./documentation/WEBSOCKET_USAGE.md)
- [API Access Workflow](./documentation/API_ACCESS_WORKFLOW.md)
- [Technical Debt Status](./documentation/TECHNICAL_DEBT_STATUS.md)

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
