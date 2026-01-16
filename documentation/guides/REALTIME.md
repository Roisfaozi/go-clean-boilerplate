# Real-time Communication Guide

This project supports two real-time communication protocols: **WebSockets (WS)** for bidirectional communication and **Server-Sent Events (SSE)** for lightweight one-way streaming.

---

## 1. WebSockets (Bidirectional)

The project uses a scalable WebSocket architecture using **Redis Pub/Sub** as a message backplane.

### Architecture
- **Distributed**: Messages sent from Server A will reach clients on Server B via Redis.
- **Channels**: Clients can subscribe to specific channels (e.g., `global_notifications`).

### Connection (Client)
```javascript
const socket = new WebSocket("ws://localhost:8080/ws?token=YOUR_JWT_TOKEN");
socket.onmessage = (event) => console.log(JSON.parse(event.data));
```

### Broadcasting (Server)
```go
notification := map[string]string{"message": "Hello World"}
jsonMsg, _ := json.Marshal(notification)
wsManager.BroadcastToChannel("global_notifications", jsonMsg)
```

---

## 2. Server-Sent Events (One-way)

SSE is ideal for pushing simple notifications or live data updates to web clients without the complexity of WebSockets.

### Connection (Client)
```javascript
const eventSource = new EventSource('http://localhost:8080/api/v1/events');
eventSource.addEventListener('user_login', (e) => {
    console.log("Login Event:", JSON.parse(e.data));
});
```

### Broadcasting (Server)
```go
sseManager.Broadcast("user_login", map[string]string{
    "user_id": "123",
    "message": "User just logged in",
})
```

---

## 3. Comparison Table

| Feature | WebSocket | SSE |
| :--- | :--- | :--- |
| **Direction** | Two-way (Full Duplex) | One-way (Server to Client) |
| **Protocol** | custom (ws://) | HTTP |
| **Scaling** | Needs Redis Pub/Sub | Standard Load Balancing |
| **Use Case** | Chat, Live Dashboards | Notifications, Logs |
