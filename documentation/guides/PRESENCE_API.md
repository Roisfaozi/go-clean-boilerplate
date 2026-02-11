# Online Presence API Design

This document specifies the communication protocol for tracking and displaying real-time user presence within organizations.

---

## I. RESTful API (Initial State)

Used by the frontend to fetch the list of online users when the dashboard or a specific organization workspace is loaded.

### 1. Get Online Users
*   **Endpoint**: `GET /api/v1/organizations/:id/presence`
*   **Authentication**: Required (JWT)
*   **Authorization**: Must be a member of the target organization.
*   **Response (200 OK)**:
    ```json
    {
      "data": [
        {
          "user_id": "019c2859-2cc6-7697-9c4e-2058d12845bc",
          "name": "Raven",
          "avatar_url": "https://avatars.githubusercontent.com/u/1",
          "role": "role:admin",
          "status": "online",
          "last_seen": 1770205809736
        }
      ],
      "total": 1
    }
    ```

---

## II. WebSocket Protocol (Live Updates)

The system uses a persistent WebSocket connection for real-time events. Presence events are broadcasted over organization-specific channels.

### 1. Connection & Subscription
*   **Endpoint**: `ws://localhost:8080/ws?token=<JWT>`
*   **Action**: Upon successful connection, the client SHOULD subscribe to the organization channel:
    *   **Channel Name**: `presence:org:{org_id}`

### 2. Inbound Messages (Client → Server)

#### Heartbeat Pong
Used to signal that the client is still active. This refreshes the user's "last seen" timestamp in the presence database.
```json
{
  "type": "presence_heartbeat",
  "data": {}
}
```

### 3. Outbound Messages (Server → Client)

#### Presence Update
Broadcasted to all members of the organization whenever someone's status changes.
```json
{
  "type": "presence_update",
  "channel": "presence:org:019c2863-4a42-73fe-8b7c-ae0ed30bc547",
  "data": {
    "event": "join", 
    "user": {
      "user_id": "019c2859-2cc6-7697-9c4e-2058d12845bc",
      "name": "Raven",
      "avatar_url": "...",
      "role": "role:admin"
    }
  }
}
```
*   **Events**:
    *   `join`: User connected.
    *   `leave`: User disconnected gracefully.
    *   `timeout`: User session expired (no heartbeat).

#### Server Ping
Sent by the server every 30 seconds to verify connection health.
```json
{
  "type": "ping",
  "data": {}
}
```

---

## III. Data Isolation Rules

1.  **Strict Tenancy**: A user MUST ONLY receive `presence_update` events for organizations they are currently a member of.
2.  **Metadata Scrubbing**: Presence payloads MUST NOT include sensitive fields like `email`, `phone`, or `full_permissions`.
3.  **Single Identity**: If a user is connected via multiple tabs, the server consolidates these into a single presence state, but only broadcasts `leave` when the *last* connection is closed.
