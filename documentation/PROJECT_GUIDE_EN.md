# Project Guide: Go Clean Boilerplate API

Welcome to the **Go Clean Boilerplate API** project guide! This document is specifically designed to help you, as a *junior developer*, deeply understand the structure, technologies, and implementation of this project. We will discuss **why** each decision was made, not just **how** it was implemented.

---

## 1. Introduction

### 1.1 What is the Go Clean Boilerplate?

**Go Clean Boilerplate** is a RESTful API project based on Go (Golang), built by applying **Clean Architecture** principles and **modularity**. Its primary goal is to provide a robust, secure, scalable, and maintainable foundation for backend development. The project comes equipped with essential features such as JWT authentication, Role-Based Access Control (RBAC) using Casbin, a dynamic search system, and real-time communication capabilities via WebSocket and Server-Sent Events (SSE).

### 1.2 Project Goals

*   **Learning:** To serve as an example implementation of Clean Architecture and Go best practices for junior developers.
*   **Quick Foundation:** To provide a ready-to-use boilerplate for new backend projects requiring strong foundational features.
*   **Scalability & Maintainability:** To ensure the code is easy to understand, test, and extend in the future.
*   **Security:** To integrate robust authentication (JWT) and authorization (Casbin RBAC) solutions.

### 1.3 Key Features

*   **Clean Architecture:** Modular and organized code structure based on domains and layers.
*   **Advanced RBAC Authorization (Casbin):** Highly granular and dynamic access control.
*   **Secure Authentication (JWT & Redis):** Stateless access tokens, stateful refresh tokens with instant revocation.
*   **Dynamic Search & Filtering:** Flexible, secure, and powerful mechanism for data search and sorting.
*   **Real-time Communication (SSE & WebSocket):** Support for one-way event streaming and bidirectional communication.
*   **Robust Validation:** Centralized request validation with *user-friendly* error messages (HTTP 422).
*   **Standardized Responses:** Consistent JSON response structure for success and errors.
*   **Database Migrations:** Version-controlled database schema management.
*   **Automated API Documentation:** Swagger/OpenAPI integration.

---

## 2. Technology Stack

The technology choices in this project are based on performance, community support, ease of use, and alignment with principles of scalability and maintainability.

### 2.1 Go (Golang)

*   **Why Go?**
    *   **Performance:** Go is a compiled language that offers high performance, similar to C++ or Java, but with a simpler syntax.
    *   **Concurrency:** Go's built-in design with Goroutines and Channels makes it easier to write efficient and scalable concurrent applications (e.g., for WebSocket or SSE).
    *   **Simple Syntax:** Easy to learn and read, minimizing *cognitive load* and promoting consistency among developers.
    *   **Easy Deployment:** Compilation results in a *single binary*, simplifying the deployment process without many dependencies.
    *   **Clear Project Structure:** Go conventions encourage a tidy project structure.

### 2.2 Gin Web Framework

*   **Why Gin?**
    *   **High Performance:** One of the fastest Go web frameworks, ideal for building RESTful APIs.
    *   **Middleware:** Supports middleware, which is very helpful for functionalities like authentication, Casbin authorization, logging, and CORS.
    *   **Routing:** A powerful and easy-to-use routing system.
    *   **API-centric:** Specifically designed for building APIs, not *full-stack* web applications with template rendering.

### 2.3 GORM (Go Object Relational Mapping)

*   **Why GORM?**
    *   **ORM:** Provides an abstraction for interacting with relational databases (e.g., MySQL). Reduces the need to write repetitive raw SQL queries.
    *   **Comprehensive Features:** Supports migrations, relationships (one-to-one, one-to-many, many-to-many), *soft delete*, transactions, etc.
    *   **Easy to Use:** Intuitive API for CRUD operations.
    *   **Flexibility:** Allows execution of raw SQL when ORM abstraction is insufficient.

### 2.4 Casbin

*   **Why Casbin?**
    *   **Robust Authorization:** A very powerful authorization library, supporting various access control models (RBAC, ABAC, ACL, etc.). This project uses RBAC.
    *   **Dynamic Policies:** Authorization policies are stored in the database, allowing policy updates at *runtime* without needing to change code or restart the application.
    *   **Casbin Model:** Defined in a `.conf` file (`internal/config/casbin_model.conf`), highly flexible.
    *   **Adapter:** Uses `gorm-adapter` to store policies in the database.

### 2.5 Redis

*   **Why Redis?**
    *   **In-Memory Data Store:** Very fast due to storing data in memory.
    *   **Sessions & Refresh Tokens:** Used as a store for stateful refresh tokens and user sessions, enabling instant token revocation (e.g., on logout or ban).
    *   **Pub/Sub (Potential):** Can be used for simple *message queuing* systems or *event-driven architectures* in the future.

### 2.6 JWT (JSON Web Tokens)

*   **Why JWT?**
    *   **Stateless Authentication:** The server does not need to store user session status. This is highly scalable for distributed applications.
    *   **Compact & Self-Contained:** Contains all necessary information to identify a user.
    *   **Security:** Cryptographically signed to ensure its integrity.

### 2.7 Others

*   **`go-playground/validator`:** A powerful and feature-rich request validation library.
*   **`logrus`:** A flexible, structured logging library.
*   **`golang-migrate`:** For managing version-controlled database schema migrations.
*   **`air`:** A tool for *live-reloading* code upon file changes, increasing productivity.
*   **`swag`:** A tool for generating Swagger/OpenAPI documentation from Go code annotations.
*   **`stretchr/testify` & `vektra/mockery`:** A complete toolset for unit testing and mocking.

---

## 3. Architecture and Principles

This project strongly adheres to **Clean Architecture** principles, focusing on the separation of *concerns* and independence from frameworks, UI, databases, or other external agents.

### 3.1 Clean Architecture Overview

Clean Architecture organizes code into concentric layers, where dependencies can only flow inwards. Outer layers depend on inner layers, but inner layers should not know the details of outer layers.

*   **Entities (Core):** Contains the most general and high-level business rules. These are the core data objects of the application.
*   **Use Cases (Application Business Rules):** Contains specific business rules for your application. Coordinates the flow of data to and from Entities.
*   **Interface Adapters:** Adapts data from the format most convenient for Use Cases and Entities to the format most convenient for external agents (Database, Web, UI). This includes Controllers, Presenters, Gateways (Repositories).
*   **Frameworks & Drivers (Outermost):** Databases, Web Frameworks, UI, etc. This layer should be the easiest to replace.

### 3.2 Project Folder Structure

The folder structure reflects Clean Architecture principles and standard Go project conventions.

```
.
├── cmd/                # Main application / entry point.
│   └── api/            # API server (main.go).
│
├── db/                 # Database-related scripts.
│   ├── migrations/     # Database schema migration files (.sql).
│   └── seeds/          # Scripts for initial data population (seeding).
│
├── docs/               # Auto-generated API documentation (Swagger).
│
├── documentation/      # Project guides and additional documentation.
│   ├── API_ACCESS_WORKFLOW.md      # Detailed API Access Flow and Role Access Rights.
│   ├── DYNAMIC_SEARCH_EXAMPLES.md  # Examples of Dynamic Search usage.
│   ├── GET_VS_DYNAMIC_SEARCH.md    # Differences between GET and Dynamic Search POST.
│   └── SSE_USAGE.md                # Server-Sent Events usage guide.
│
├── pkg/                # Packages that can be broadly reused within the project.
│   ├── exception/      # Custom error definitions.
│   ├── jwt/            # Utilities for JSON Web Tokens.
│   ├── password/       # Password hashing utilities.
│   ├── querybuilder/   # Dynamic Search Query Builder implementation.
│   ├── response/       # Standardized API response structures.
│   ├── sse/            # Server-Sent Events manager.
│   ├── tx/             # Database transaction manager.
│   ├── validation/     # Custom validation rules and error formatters.
│   └── ws/             # WebSocket manager and client handling.
│
└── internal/           # Internal project code that should not be imported by external Go projects.
    ├── config/         # Application configuration, dependency initialization (DI Container).
    ├── middleware/     # HTTP Middlewares (Auth, Casbin Enforcer, CORS).
    ├── mocking/        # Mock objects for testing.
    ├── router/         # Main routing configuration.
    │
    └── modules/        # Domain-specific modules (Core Business Logic).
        ├── <module_name_1>/ # E.g., auth, user, role, permission, access.
        │   ├── delivery/   # (Interface Adapter) HTTP Handlers (Controller layer).
        │   ├── usecase/    # (Use Case) Module-specific business logic.
        │   ├── repository/ # (Interface Adapter) Data access abstraction.
        │   ├── model/      # (Entities/DTOs) Data structures for requests/responses, database entities.
        │   └── entity/     # (Entities) Domain/database entity representations.
        ├── <module_name_2>/
        └── ...

**Explanation of `pkg/` vs `internal/`:**
*   **`pkg/`**: This directory contains code that can be **broadly reused** within this project, and **could be imported by other external Go projects** if this module were to become a library (though not currently intended). These are generic utilities not tied to a specific business domain.
*   **`internal/`**: According to Go conventions, code within the `internal/` directory **cannot be imported by Go projects outside of this module**. It is used for core project functionalities that are private and not meant to be exposed. Main business modules are placed here to keep internal business domain implementations hidden from the outside world.

### 3.3 SOLID Principles and Dependency Inversion (DIP)

The project heavily emphasizes software development principles, especially:

*   **SOLID Principles:**
    *   **Single Responsibility Principle (SRP):** Every module/class/function has one reason to change.
    *   **Open/Closed Principle (OCP):** Software entities should be open for extension but closed for modification.
    *   **Liskov Substitution Principle (LSP):** Objects in a program should be replaceable with *instances* of their *subtypes* without altering the correctness of that program.
    *   **Interface Segregation Principle (ISP):** Many small *interfaces* are better than one large *interface*.
    *   **Dependency Inversion Principle (DIP):** High-level modules should not depend on low-level modules. Both should depend on abstractions. Abstractions should not depend on details. Details should depend on abstractions.
*   **Dependency Inversion (DIP):** This is key in Clean Architecture. The Use Case layer (high-level) does not directly call the Repository implementation (low-level). Instead, the Use Case depends on a **Repository interface** defined in the Use Case layer. The Repository implementation then satisfies this interface. This allows easy replacement of database implementations without affecting business logic.

```go
// internal/modules/user/usecase/interface.go (Abstraction/Interface)
package usecase

import (
	"context"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	// ... other methods
	FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.User, error)
}

type UserUseCase interface {
	Register(ctx context.Context, request model.RegisterUserRequest) (*model.UserResponse, error)
	Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, string, error)
	// ... other methods
}
```
```go
// internal/modules/user/repository/user_repository.go (Concrete Implementation)
package repository

import (
	"context"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"gorm.io/gorm"
)

type userRepositoryData struct {
	db *gorm.DB
}

// NewUserRepository creates a UserRepository instance.
func NewUserRepository(db *gorm.DB) usecase.UserRepository {
	return &userRepositoryData{
		db: db,
	}
}

func (r *userRepositoryData) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepositoryData) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindAllDynamic implementation
func (r *userRepositoryData) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.User, error) {
	var users []*entity.User
	query := r.db.WithContext(ctx)

	// Logic from pkg/querybuilder applied here
	where, args, _, err := querybuilder.GenerateDynamicQuery[entity.User](filter)
	if err != nil {
		return nil, err
	}
	if where != "" {
		query = query.Where(where, args...)
	}

	sort, err := querybuilder.GenerateDynamicSort[entity.User](filter)
	if err != nil {
		return nil, err
	}
	if sort != "" {
		query = query.Order(sort)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
```

---

## 4. Core Feature Implementation

This section will explain how the main features of this project are implemented and integrated into the existing architecture.

### 4.1 Authentication Flow (Login, Refresh Token, Logout)

Authentication in this project uses a combination of JWT (JSON Web Tokens) for stateless access and stateful Refresh Tokens (stored in Redis) for security and revocation mechanisms.

*   **Login (`POST /auth/login`):**
    1.  The user sends `username` and `password`.
    2.  `AuthUseCase` verifies user credentials from `UserRepository`.
    3.  If valid, `JWTManager` (from `pkg/jwt`) creates a pair of Access Token (short-lived) and Refresh Token (long-lived).
    4.  The Refresh Token is stored in Redis via `TokenRepository` with an appropriate TTL (Time To Live).
    5.  The Access Token is returned in the response body, the Refresh Token is sent as an `HttpOnly Cookie`.
    6.  **Event Broadcast:** `WebSocketManager` is used to broadcast a "UserLoggedIn" event to the `global_notifications` channel (if implemented). This demonstrates the application's ability to notify other systems in real-time.

    ```go
    // Example in internal/modules/auth/usecase/authJWT_usecase.go (excerpt)
    func (uc *authUseCase) Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, string, error) {
    	// ... (User and password verification)

    	// Create tokens
    	accessToken, refreshToken, err := uc.jwtManager.GenerateTokenPair(user.ID, sessionID, userRole, user.Username)
    	if err != nil {
    		return nil, "", fmt.Errorf("failed to generate token pair: %w", err)
    	}

    	// Store refresh token in Redis
    	authData := &model.Auth{
    		ID:           sessionID,
    		UserID:       user.ID,
    		RefreshToken: refreshToken,
    		ExpiresAt:    time.Now().Add(uc.jwtManager.GetRefreshTokenDuration()).Unix(),
    	}
    	if err = uc.tokenRepo.StoreToken(ctx, authData); err != nil {
    		return nil, "", fmt.Errorf("failed to store session: %w", err)
    	}

    	// Broadcast event (if any)
    	uc.wsManager.BroadcastToChannel("global_notifications", []byte(fmt.Sprintf("User %s logged in", user.Username)))

    	// Return response
    	return &model.LoginResponse{
    		AccessToken: accessToken,
    		TokenType:   "Bearer",
    		ExpiresIn:   int(uc.jwtManager.GetAccessTokenDuration().Seconds()),
    		User:        model.UserResponse{ID: user.ID, Username: user.Username, Role: userRole},
    	}, refreshToken, nil
    }
    ```

*   **Refresh Token (`POST /auth/refresh`):**
    1.  The client sends a request without an Access Token, but the Refresh Token must be in the `HttpOnly Cookie`.
    2.  `AuthUseCase` retrieves the Refresh Token from the cookie.
    3.  `JWTManager` validates the Refresh Token.
    4.  `TokenRepository` checks for the Refresh Token's existence in Redis (ensuring it hasn't been revoked).
    5.  If valid, `TokenRepository` deletes the old Refresh Token from Redis.
    6.  `JWTManager` creates a new Access and Refresh Token pair.
    7.  The new Refresh Token is stored in Redis, and the new Access Token is returned.

    ```json
    // Example Response Body from Refresh Token:
    // POST /api/v1/auth/refresh
    // Header: Cookie: refresh_token=<YOUR_REFRESH_TOKEN>
    /*
    {
    	"data": {
    		"access_token": "eyJhbGciOiJIUzI1Ni...",
    		"token_type": "Bearer",
    		"expires_in": 900,
    		"user": {
    			"id": "user-uuid-123",
    			"username": "johndoe",
    			"role": "role:user"
    		}
    	}
    }
    */
    ```

*   **Logout (`POST /auth/logout`):**
    1.  The user sends a request with a valid Access Token.
    2.  `AuthUseCase` obtains the `userID` and `sessionID` from the Access Token claims.
    3.  `TokenRepository` deletes the associated Refresh Token from Redis, effectively revoking the session.
    4.  The Refresh Token cookie is also cleared from the response.

    ```go
    // Example in internal/modules/auth/usecase/authJWT_usecase.go (excerpt)
    func (uc *authUseCase) Logout(ctx context.Context, userID, sessionID string) error {
    	if err := uc.tokenRepo.DeleteToken(ctx, userID, sessionID); err != nil {
    		return fmt.Errorf("failed to delete session token: %w", err)
    	}
    	return nil
    }
    ```

### 4.2 Authorization (Casbin RBAC)

Authorization is at the core of resource access security and is implemented using [Casbin](https://casbin.org/).

*   **Casbin Middleware (`internal/middleware/casbin_middleware.go`):**
    *   Every request requiring authorization passes through `CasbinMiddleware`.
    *   This middleware extracts `sub` (user/role), `obj` (resource being accessed), and `act` (action/HTTP method) from the request.
    *   **Resource (`obj`):** For each request, the middleware attempts to find `Access Rights` associated with the current request `path` and `method`. If an `Access Right` is found, its `name` will be used as the `obj` in Casbin. If no relevant `Access Right` is found, the request `path` will be used directly as the `obj`.
    *   **Action (`act`):** The HTTP method (`GET`, `POST`, `PUT`, `DELETE`).
    *   **Subject (`sub`):** The `role` of the authenticated user (e.g., `role:admin`, `role:user`).
    *   Casbin then performs the authorization check: `enforcer.Enforce(sub, obj, act)`.
    *   If `enforce` fails, the request is rejected with a `403 Forbidden` status.

    ```go
    // Example in internal/middleware/casbin_middleware.go (excerpt)
    func (m *CasbinMiddleware) Authorize(obj string, act string) gin.HandlerFunc {
    	return func(c *gin.Context) {
    		// ... (get user ID and role from JWT context)
    		userRole := c.GetString("userRole") // Assuming role is set in context by previous middleware

    		// Determine the resource object for Casbin
    		// This logic dynamically checks if the current path/method is linked to an AccessRight
    		// If so, it uses the AccessRight name as the object for Casbin.
    		// Otherwise, it falls back to the raw request path.
    		// For a full implementation, refer to the actual Casbin middleware code.

    		ok, err := m.enforcer.Enforce(userRole, obj, act) // 'obj' might be an AccessRight name or raw path
    		if err != nil {
    			response.InternalServerError(c, exception.ErrInternalServer, "Authorization error")
    			c.Abort()
    			return
    		}

    		if !ok {
    			response.Forbidden(c, exception.ErrForbidden, "Forbidden: You don't have permission to access this resource")
    			c.Abort()
    			return
    		}
    		c.Next()
    	}
    }
    ```

*   **Casbin Model (`internal/config/casbin_model.conf`):**
    ```ini
    [request_definition]
    r = sub, obj, act

    [policy_definition]
    p = sub, obj, act

    [role_definition]
    g = _, _

    [policy_effect]
    e = some(where (p.eft == allow))

    [match_definition]
    m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act || r.obj == "/api/v1/users/me" && r.sub == r.obj.ID
    ```
    *   **`r` (request):** Defines the authorization request format (subject, object, action).
    *   **`p` (policy):** Defines the policy format (subject, object, action).
    *   **`g` (role):** Defines role relationships (`g(user, role)` means the user has the role).
    *   **`e` (effect):** Rules on how policies are evaluated.
    *   **`m` (match):** A complex matching function. This includes:
        *   `g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`: Standard RBAC check (if the subject has the appropriate role, the object and action must match).
        *   `r.obj == "/api/v1/users/me" && r.sub == r.obj.ID`: This is a special rule for **self-profile access**. It means if the object is the `/api/v1/users/me` endpoint, then the subject (UserID) must be the same as the UserID extracted from the path (`r.obj.ID` is a placeholder for actual path param parsing), allowing a user to access their own `me` resource without needing a specific role in Casbin, as long as they are authenticated and their ID matches. This is a smart Self-Service feature.
*   **Concepts of Access Rights and Endpoints:**
    *   `Access Rights` and `Endpoints` are database entities managed by the `superadmin` to form the `obj` (resource) in Casbin policies.
    *   An `Endpoint` is a specific combination of `path` and `method` (e.g., `/api/v1/users GET`).
    *   An `Access Right` is a logical grouping of one or more `Endpoints` (e.g., `user_management` could cover `GET /api/v1/users`, `POST /api/v1/users`).
    *   When `CasbinMiddleware` checks permissions, it will look for `Access Rights` associated with the current `path` and `method`. If an `Access Right` is found, its `name` will be used as the `obj` that Casbin checks. If no relevant `Access Right` is found, the raw request `path` will be used directly as the `obj`.
    *   For more details and mapping examples, see [`documentation/API_ACCESS_WORKFLOW.md`](API_ACCESS_WORKFLOW.md).

### 4.3 Dynamic Search (Query Builder)

This feature provides a highly flexible search and sorting mechanism for resources via a `POST /<resource>/search` request.

*   **How it Works (`pkg/querybuilder`):**
    *   The `pkg/querybuilder` library is the core of this functionality. It accepts a `DynamicFilter` JSON structure in the request body.
    *   `DynamicFilter` contains `Filter` (a map from field names to filter objects with an operator `type`, `from` value, `to` optional value) and `Sort` (an array of sort objects with `colId` and `sort` direction).
    *   `querybuilder` uses **reflection** to analyze the Go model struct and map incoming request field names to the corresponding database column names (using `gorm:"column:..."` tags or converting to `snake_case`).
    *   It constructs a secure SQL `WHERE` and `ORDER BY` clause, using *parameterized queries* to prevent SQL Injection.
    *   It automatically adds a *soft delete* clause (`deleted_by IS NULL` or `deleted_at = 0`) if the model has `DeletedBy` or `DeletedAt` fields.

    ```json
    // Example DynamicFilter request structure
    /*
    {
      "filter": {
        "name": { "type": "contains", "from": "John" },
        "age": { "type": "greater_than", "from": 25 }
      },
      "sort": [
        { "colId": "created_at", "sort": "desc" },
        { "colId": "name", "sort": "asc" }
      ]
    }
    */
    ```
    ```go
    // Example in pkg/querybuilder/query_builder.go (excerpt)
    func GenerateDynamicQuery[T any](filter *DynamicFilter) (string, []interface{}, []string, error) {
    	// ... (Reflection logic and WHERE clause construction)
    	switch op {
    	case "contains":
    		queryParts = append(queryParts, fmt.Sprintf("%s LIKE ?", dbCol))
    		args = append(args, "%"+val+"%")
    	case "in_range":
    		queryParts = append(queryParts, fmt.Sprintf("%s >= ? AND %s <= ?", dbCol, dbCol))
    		args = append(args, f.From, f.To)
    	// ... other operators
    	}
    	return strings.Join(queryParts, " AND "), args, warnings, nil
    }

    func GenerateDynamicSort[T any](filter *DynamicFilter) (string, error) {
    	// ... (ORDER BY clause construction logic)
    	for _, s := range *filter.Sort {
    		// ...
    		sortParts = append(sortParts, fmt.Sprintf("%s %s", dbCol, direction))
    	}
    	return strings.Join(sortParts, ", "), nil
    }
    ```

*   **Integration (Repository, Usecase, Controller):**
    *   **Controller:** Receives the `DynamicFilter` JSON body at the `POST /<resource>/search` endpoint. Calls the `Usecase`.
    *   **Usecase:** Receives the `DynamicFilter`, validates it (if necessary), and passes it to the `Repository`.
    *   **Repository:** Uses `querybuilder.GenerateDynamicQuery()` and `querybuilder.GenerateDynamicSort()` to build the GORM query before executing `Find()`.

    ```go
    // Example in internal/modules/user/repository/user_repository.go (excerpt)
    func (r *userRepositoryData) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.User, error) {
    	var users []*entity.User
    	query := r.db.WithContext(ctx)

    	where, args, _, err := querybuilder.GenerateDynamicQuery[entity.User](filter)
    	if err != nil { return nil, err }
    	if where != "" { query = query.Where(where, args...) }

    	sort, err := querybuilder.GenerateDynamicSort[entity.User](filter)
    	if err != nil { return nil, err }
    	if sort != "" { query = query.Order(sort) }

    	if err := query.Find(&users).Error; err != nil { return nil, err }
    	return users, nil
    }
    ```
*   **Usage Examples:** See [`documentation/DYNAMIC_SEARCH_EXAMPLES.md`](DYNAMIC_SEARCH_EXAMPLES.md) for detailed `curl` examples.

### 4.4 Server-Sent Events (SSE)

SSE allows a server to *push* one-way events to web clients via a standard, long-lived HTTP connection.

*   **How it Works (`pkg/sse/manager.go`):**
    *   `sse.Manager` is an event manager that handles client registration/deregistration and event *broadcasting*.
    *   Each client has a private channel (`chan Event`) to receive events.
    *   The `Manager` runs a continuous Goroutine (`run()`) to listen for new clients, disconnected clients, and events to broadcast.
    *   When `Broadcast()` is called, the event is sent to the channels of all connected clients.
    *   Events are formatted according to the SSE standard: `event: <event_name>\ndata: <json_data>\n\n`.

    ```json
    // Example SSE event structure as seen by the client
    // event: user_activity
    // data: {"user_id":"uuid-123","action":"logged_in","timestamp":1678886400}
    ```
    ```go
    // Example of broadcasting an event from application code
    // From any UseCase or service
    sseManager.Broadcast("user_activity", map[string]interface{}{
    	"user_id":   "uuid-user-1",
    	"action":    "logged_in",
    	"timestamp": time.Now().Unix(),
    })
    ```
    ```go
    // Example in pkg/sse/manager.go (excerpt of ServeHTTP handler)
    func (m *Manager) ServeHTTP() gin.HandlerFunc {
    	return func(c *gin.Context) {
    		// ... (SSE header setup)
    		clientChan := make(chan Event)
    		client := &Client{Channel: clientChan}
    		m.register <- client
    		defer func() { m.unregister <- client }() // Unregister on client disconnect

    		c.Stream(func(w io.Writer) bool {
    			select {
    			case <-c.Request.Context().Done(): // Client disconnected
    				return false
    			case event, ok := <-clientChan: // New event to send
    				if !ok { return false } // Channel closed
    				c.Writer.Write([]byte(fmt.Sprintf("event: %s\n", event.Name)))
    				jsonData, err := json.Marshal(event.Data)
    				if err != nil { c.Writer.Write([]byte(fmt.Sprintf("data: %v\n\n", event.Data))) } else { c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", jsonData))) }
    				c.Writer.Flush()
    				return true
    			}
    		})
    	}
    }
    ```
*   **Integration (App Init, Router):**
    *   `sse.Manager` is initialized in `internal/config/app.go` and passed as a dependency.
    *   The `/events` endpoint in `internal/router/router.go` uses `sseManager.ServeHTTP()` as its handler to accept SSE connections.
    *   Events can be broadcast from anywhere in the application (e.g., from `AuthUseCase` after login).

### 4.5 WebSocket

WebSocket provides *bidirectional* and *full-duplex* communication channels over a single TCP connection.

*   **How it Works (`pkg/ws/`):**
    *   **`ws_manager.go`:** `WebSocketManager` is the heart of this functionality. It manages the list of connected clients, channels to which clients can subscribe, and the mechanism for broadcasting messages to specific channels.
    *   **`ws_client.go`:** Each WebSocket client is represented by a `Client` struct that has `ReadPump()` (a Goroutine for reading messages from the client) and `WritePump()` (a Goroutine for sending messages to the client).
    *   **`ws_controller.go`:** `WebSocketController` handles the process of *upgrading* an HTTP connection to a WebSocket connection using `gorilla/websocket`.

    ```json
    // Example message from Client to Server (JSON) for subscription
    /*
    {
    	"type": "subscribe",
    	"channel": "global_notifications"
    }
    */
    ```
    ```json
    // Example message from Server to Client (JSON) for confirmation
    /*
    {
    	"type": "info",
    	"channel": "global_notifications",
    	"data": "Subscribed to channel"
    }
    */
    ```
    ```go
    // Example in pkg/ws/ws_controller.go (excerpt of HandleWebSocket handler)
    func (c *WebSocketController) HandleWebSocket(ctx *gin.Context) {
    	conn, err := c.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
    	if err != nil { /* ... handle error ... */ return }

    	config := &WebSocketConfig{ /* ... */ } // Define WS config (write timeout, pong wait, etc.)
    	client := NewWebsocketClient(conn, c.manager, c.log, config)

    	c.manager.RegisterClient(client) // Register client with the manager

    	go client.WritePump() // Goroutine for sending messages to the client
    	go client.ReadPump()  // Goroutine for reading messages from the client

    	c.log.Infof("New WebSocket connection established: %s", client.ID)
    }
    ```

*   **Integration (App Init, Router):**
    *   `WebSocketManager` is initialized in `internal/config/app.go`.
    *   The `/ws` endpoint in `internal/router/router.go` uses `WebSocketController.HandleWebSocket()` as its handler to accept WebSocket connections.
*   **Differences with SSE:**
    *   **SSE:** One-way (server-to-client), HTTP-based, simpler. Ideal for notifications or data streams that don't require direct client responses.
    *   **WebSocket:** Two-way (server-to-client and client-to-server), full-duplex, separate protocol. Ideal for chat, games, or applications requiring intensive real-time interaction from both sides.

---

## 5. Getting Started & Development

This section will guide you through setting up the development environment, running the application, and understanding the basic workflow for adding or modifying features.

### 5.1 Prerequisites

Before you begin, ensure your system has the following software installed:

1.  **Go:** Version 1.21 or higher. This is the primary programming language of the project.
2.  **Docker & Docker Compose:** Used to run infrastructure services like the database (MySQL) and Redis in isolated containers. This ensures a consistent development environment.
3.  **Make:** The `make` utility is used to run automation commands defined in the `Makefile` (e.g., `make migrate-up`, `make test`).
4.  **Air (Optional):** A tool for *live-reloading* code upon file changes, highly recommended for development.
    ```bash
    go install github.com/air-verse/air@latest
    ```
5.  **Swag CLI (Optional):** Used to regenerate Swagger/OpenAPI documentation from code annotations.
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```
6.  **Golang Migrate (Optional):** If you wish to run migrations manually without the `Makefile`.
    ```bash
    go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    ```
7.  **C/C++ Compiler (GCC/MinGW-w64):** Required if you are running repository tests that use SQLite (due to CGO). Ensure `gcc` is in your system's PATH.

### 5.2 Environment Setup

Follow these steps to set up the project in your local environment.

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/yourusername/go-clean-boilerplate.git
    cd go-clean-boilerplate # Ensure the folder name matches
    ```
    *Note: Make sure your local project folder name precisely matches the module name in `go.mod` (`go-clean-boilerplate`) to avoid Go Modules issues.*

2.  **Environment Configuration:**
    Create a `.env` file from `.env.example`. This file contains crucial configuration for the database, Redis, JWT secrets, and more.
    ```bash
    cp .env.example .env
    ```
    *Tip: The default values in `.env.example` usually work out-of-the-box with the provided `docker-compose.yml`.*

3.  **Start Infrastructure (Database & Redis):**
    Use Docker Compose to spin up MySQL and Redis containers.
    ```bash
    docker-compose up -d
    ```
    This command will create and run containers in the background. You can verify their status with `docker-compose ps`.

### 5.3 Database Migrations & Seeding

After the infrastructure is running, apply the database schema and populate initial data.

1.  **Run Database Migrations:**
    ```bash
    make migrate-up
    ```
    This command will execute all SQL migration files in `db/migrations/` and create the necessary tables.

2.  **Seed Initial Data:**
    ```bash
    make seed-up
    ```
    This command will execute the seeding scripts in `db/seeds/` to populate initial data, such as default roles (`admin`, `user`), or initial users.

### 5.4 Running the Application

You can run the application in development mode (with hot reload) or standard mode.

*   **Development Mode (Recommended):**
    ```bash
    air
    ```
    If you have `air` installed (see Prerequisites), this will start the server and automatically restart the application whenever you save changes to code files. The server will run at `http://localhost:8080` (or the port defined in your `.env`).

*   **Standard Mode:**
    ```bash
    go run cmd/api/main.go
    ```
    This will compile and run the application. Code changes require manual stopping and restarting.

*   **Production Build:**
    ```bash
    make build
    ./bin/api # Or ./bin/api.exe on Windows
    ```
    This will compile the application into a *single executable binary* in the `bin/` directory.

### 5.5 General Development Workflow

To add or modify new features, follow this general workflow:

1.  **Understand Requirements:** Start by understanding what needs to be done.
2.  **Define Entities/Models:** If new data is involved, define it in `modules/<module_name>/entity/` and `modules/<module_name>/model/`.
3.  **Implement Repository:** Add new methods to `modules/<module_name>/repository/interface.go` and their implementation in `modules/<module_name>/repository/<module_name>_repository.go`.
4.  **Implement Use Case:** Add new methods to `modules/<module_name>/usecase/interface.go` and their implementation in `modules/<module_name>/usecase/<module_name>_usecase.go`. This is where the core business logic resides.
5.  **Implement Handler (Controller):** Add new methods to `modules/<module_name>/delivery/http/<module_name>_controller.go` to handle HTTP requests.
6.  **Register Route:** Register new endpoints in `modules/<module_name>/delivery/http/<module_name>_routes.go` and integrate them into `internal/router/router.go`.
7.  **Testing:** Always write tests for your code.
    *   **Unit Tests:** For `Repository` (with mock DB), `Usecase` (with mock Repository), and `Controller` (with mock Usecase).
    *   **Integration Tests:** If necessary, to test interactions between components.
8.  **Database Migrations:** If there are database schema changes, create a new migration (`make migrate create <migration_name>`) and apply it.
9.  **Documentation:** Update relevant documentation (e.g., Swagger, this guide).

*Debugging Tips:*
*   Use `logrus.Debugf()` or `logrus.Errorf()` to trace execution flow and variable values.
*   Use an IDE like VS Code with the Go extension for *breakpoint debugging*.

---

## 6. Testing Strategy

Testing is an integral part of the development process to ensure code functions correctly, prevent regressions, and validate feature implementations. This project adopts several testing strategies.

### 6.1 Unit Tests

*   **Purpose:** To test the smallest units of code (functions, methods) in isolation.
*   **Structure:** Each module (`auth`, `user`, `role`, `permission`, `access`) has a `test/` sub-folder. Within this `test/` folder, there are test files for the `Controller`, `UseCase`, and `Repository` (e.g., `user_controller_test.go`, `user_usecase_test.go`, `user_repository_test.go`).
*   **Philosophy:**
    *   **Repository Tests:** Test interactions with a *real* database (usually using an in-memory database like SQLite for speed, or Docker containers for the actual database). These can be considered low-level *integration tests*.
    *   **Usecase Tests:** Test core business logic. Dependencies like `Repository` or `JWTManager` will be *mocked* or *stubbed*.
    *   **Controller Tests:** Test how HTTP handlers process requests and generate responses. Dependencies like `UseCase` will be *mocked*.
*   **Running Unit Tests:**
    ```bash
    make test
    ```
    This command will run `go test -v ./...` which executes all test files across the entire project.

### 6.2 Mocking: When and How to Use Mockery

*   **What is Mocking?** Mocking is a technique in unit testing where real objects are replaced with simulated (mock) objects that mimic the behavior of the real ones. This allows testing units of code in isolation from their dependencies.
*   **Why Use Mocking?**
    *   **Isolation:** Ensures that the unit under test relies only on its own behavior, not on the correctness or performance of external dependencies (e.g., no need for a real database to test a `UseCase`).
    *   **Control:** Allows you to simulate error scenarios or *edge cases* from dependencies that are difficult to reproduce in a real environment.
    *   **Speed:** Tests run faster because there are no actual network or disk interactions.
*   **Implementation in This Project:**
    *   This project uses the [Mockery](https://vektra.github.io/mockery/) library to automatically generate mock objects from Go interfaces.
    *   The `.mockery.yml` configuration file defines which interfaces should be mocked and where the generated mock files should be stored (e.g., in `internal/modules/<module>/test/mocks/`).
    *   **Example:** `user/usecase/user_usecase_test.go` will import `internal/modules/user/test/mocks/mock_user_repository.go` and use `mocks.MockUserRepository` to simulate the behavior of the `UserRepository` interface.
*   **Regenerating Mocks:**
    If you modify any interface definition (e.g., in `repository/interface.go` or `usecase/interface.go`), you should regenerate the relevant mocks:
    ```bash
    make mocks
    ```

### 6.3 End-to-End Tests (Postman)

*   **Purpose:** To verify the entire application flow, from HTTP request to response, including interactions between all components (router, controller, usecase, repository, database).
*   **Postman Collections:** This project provides several Postman Collections in the `postman/` folder:
    *   `Casbin Project API.postman_collection.json`: For core CRUD, Auth, and RBAC flows.
    *   `Casbin Project API - Dynamic Search.postman_collection.json`: Contains various dynamic search scenarios (positive, negative, edge, security).
    *   `Casbin Project API - Realtime.postman_collection.json`: Examples for WebSocket and SSE connections.
*   **How to Use:**
    1.  **Import:** Import these `.json` files into your Postman application.
    2.  **Environment:** Set up your Postman environment variables (e.g., `baseURL`, `apiVersion`, `authToken`) in "Casbin Project Env.postman_environment.json" to match your local setup.
    3.  **Run Runner:** Use the "Collection Runner" feature in Postman to execute the entire collection. Many requests include test scripts (`pm.test()`) that automatically verify API responses (status codes, data structure, etc.).

### 6.4 Troubleshooting Repository Tests (CGO/SQLite)

*   **Common Issue:** You might encounter an error like `Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub` when running repository tests (especially if you're using `gorm.io/driver/sqlite` for testing).
*   **Cause:** Go defaults to compiling with `CGO_ENABLED=0` on some systems or environments, which prevents the use of C libraries like `sqlite`.
*   **Solution:**
    1.  **Install GCC/MinGW-w64:** Ensure you have a C/C++ compiler installed on your system and accessible from your PATH (e.g., `MinGW-w64` on Windows, `build-essential` on Linux, or `Xcode Command Line Tools` on macOS).
    2.  **Enable CGO:** Run tests with `CGO_ENABLED=1`:
        ```bash
        CGO_ENABLED=1 go test -v ./...
        ```
        Or modify your `Makefile`'s `test` target to include `CGO_ENABLED=1`.

---

## 7. API Documentation

API documentation is crucial for frontend developers and other teams interacting with this API.

### 7.1 Swagger UI

*   **What is it?** Swagger UI provides an interactive web interface to view and test your API endpoints. It is automatically generated from your Go code annotations.
*   **Access:** Once the application is running, open your browser to:
    `http://localhost:8080/api/docs/index.html`
*   **How to Regenerate:**
    If you add or modify endpoints and Swagger annotations (`@Summary`, `@Param`, `@Success`, etc.), you need to regenerate the Swagger files:
    ```bash
    swag init -g cmd/api/main.go
    ```
    *Ensure you are in the project's root directory when running this command.*

### 7.2 Additional Documentation References

In addition to this guide, several other Markdown documents provide specific details:

*   [`documentation/API_ACCESS_WORKFLOW.md`](API_ACCESS_WORKFLOW.md): Full details on all API routes, access workflow, and privileges for each role (superadmin, admin, user).
*   [`documentation/DYNAMIC_SEARCH_EXAMPLES.md`](DYNAMIC_SEARCH_EXAMPLES.md): Detailed `curl` examples for dynamic search endpoints with various filter and sorting types.
*   [`documentation/GET_VS_DYNAMIC_SEARCH.md`](GET_VS_DYNAMIC_SEARCH.md): Explanation of the differences between simple `GET` searches and dynamic `POST /search` searches.
*   [`documentation/SSE_USAGE.md`](SSE_USAGE.md): Guide on how Server-Sent Events (SSE) are implemented and how to use them.

---

## 8. Coding Conventions

Following consistent coding conventions is key to maintaining readability, maintainability, and collaboration in the project.

### 8.1 Naming

*   **Packages:** Use all lowercase (e.g., `user`, `auth`, `pkg`).
*   **Variables, Functions, Methods (Private):** Use `camelCase` and start with a lowercase letter (e.g., `getUserByID`, `jwtManager`).
*   **Variables, Functions, Methods (Public/Exported):** Use `PascalCase` and start with an uppercase letter (e.g., `NewJWTManager`, `GenerateTokenPair`).
*   **Constants:** Use `PascalCase` or `ALL_CAPS` for global constants (e.g., `TestAccessSecret`, `MAX_MESSAGE_SIZE`).
*   **Interfaces:** Start with an `I` or end with `er` (e.g., `UserRepository`, `Writer`).

### 8.2 Formatting

*   Use `go fmt` regularly. Your IDE should be configured to run it automatically on file save.
*   `goimports` is also highly recommended for automatically managing imports.

### 8.3 Error Handling

*   Go uses explicit `error` return values.
*   Always check for errors after calling functions that return an error.
*   Use `errors.Is()` to check for specific error types.
*   Use `fmt.Errorf("...: %w", err)` for error *wrapping*, allowing stack tracing and inspection of the original error.
*   Define custom errors in `pkg/exception/error.go` for frequently occurring or semantically important errors.

---

## 9. Contribution and License

### 9.1 Contribution

We warmly welcome contributions! If you wish to contribute, please follow these general steps:

1.  Fork the project repository.
2.  Create a new feature branch (`git checkout -b feature/your-awesome-feature`).
3.  Implement your changes. Ensure to write corresponding tests, and all tests pass.
4.  Commit your changes (`git commit -m 'feat: Add awesome feature X'`).
5.  Push to your branch (`git push origin feature/your-awesome-feature`).
6.  Open a Pull Request to the main repository.

### 9.2 License

This project is licensed under the Apache 2.0 License - see the `LICENSE` file for more details.

---
