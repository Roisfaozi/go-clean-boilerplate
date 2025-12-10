# Server-Sent Events (SSE) Usage Guide

This document explains how to integrate and use the generic SSE Manager located at `internal/utils/sse`.

## 1. Overview

The SSE Manager provides a thread-safe way to broadcast server-side events to multiple connected clients via HTTP streaming. It handles client registration, disconnection cleanup, and event broadcasting automatically.

**Core Features:**
- **Generic Payload:** Can broadcast any struct, map, or primitive type (auto-marshaled to JSON).
- **Named Events:** Supports custom event names (e.g., `user_created`, `notification`).
- **Thread-Safe:** Uses channels and mutexes for concurrency safety.

## 2. Setup & Initialization

### A. Initialize Manager
Ideally, initialize the SSE Manager once in your application setup (e.g., `internal/config/app.go` or `cmd/api/main.go`) and pass it as a dependency.

```go
// cmd/api/main.go or internal/config/app.go

import "github.com/Roisfaozi/casbin-db/internal/utils/sse"

func main() {
    // 1. Initialize SSE Manager
    sseManager := sse.NewManager()

    // 2. Inject into Handlers/Controllers
    // (See section 3 for details)
}
```

### B. Register Route
Expose an endpoint for clients to connect to.

```go
// internal/router/router.go

func SetupRouter(sseManager *sse.Manager) *gin.Engine {
    r := gin.Default()

    // SSE Endpoint
    // Clients connect to: http://localhost:8080/events
    r.GET("/events", sseManager.ServeHTTP())

    return r
}
```

## 3. Dependency Injection

To send events from your business logic (Controllers/UseCases), you need to inject the `*sse.Manager`.

### Example: Injecting into UserHandler

**Update `UserHandler` struct:**

```go
// internal/modules/user/delivery/http/user_controller.go

type UserHandler struct {
    UserUseCase usecase.UserUseCase
    Log         *logrus.Logger
    validate    *validator.Validate
    SSE         *sse.Manager // Add this field
}

func NewUserHandler(uc usecase.UserUseCase, log *logrus.Logger, v *validator.Validate, sse *sse.Manager) *UserHandler {
    return &UserHandler{
        UserUseCase: uc,
        Log:         log,
        validate:    v,
        SSE:         sse, // Inject here
    }
}
```

**Update Wiring in `main.go`:**

```go
userHandler := http.NewUserHandler(userUseCase, log, validate, sseManager)
```

## 4. Broadcasting Events

You can now broadcast events from anywhere `sseManager` is available.

### Example: Broadcast on User Creation

```go
// internal/modules/user/delivery/http/user_controller.go

func (h *UserHandler) RegisterUser(c *gin.Context) {
    // ... existing registration logic ...
    
    user, err := h.UserUseCase.Create(ctx, &req)
    if err != nil {
        // ... handle error ...
        return
    }

    // BROADCAST EVENT
    // Event Name: "user_registered"
    // Data: The new user object (or any custom struct)
    h.SSE.Broadcast("user_registered", map[string]interface{}{
        "message": "A new user just joined!",
        "username": user.Username,
        "id": user.ID,
    })

    response.Created(c, user)
}
```

## 5. Client-Side Implementation (Frontend)

To receive these events in a browser (JavaScript/React/Vue):

```javascript
// Connect to the SSE endpoint
const eventSource = new EventSource('http://localhost:8080/events');

// Listen for connection open
eventSource.onopen = function() {
    console.log("Connection to server opened.");
};

// Listen for specific named event 'user_registered'
eventSource.addEventListener('user_registered', function(event) {
    const data = JSON.parse(event.data);
    console.log("New User Event Received:", data);
    alert(`New user registered: ${data.username}`);
});

// Listen for generic messages (if any)
eventSource.onmessage = function(event) {
    console.log("Generic message:", event.data);
};

// Handle errors
eventSource.onerror = function(err) {
    console.error("EventSource failed:", err);
};
```

## 6. Testing with Curl

You can also test the connection using `curl`:

```bash
curl -N http://localhost:8080/events
```

You will see output appear in real-time as events are broadcasted:

```text
event: user_registered
data: {"id":"...", "message":"A new user just joined!", "username":"testuser"}

```
