# Plan: Online Presence Implementation

## 1. Backend: Presence Logic (Go)

### 1.1 Presence Manager (`pkg/ws/presence_manager.go`)
- [ ] Define `PresenceManager` interface.
- [ ] Implement `RedisPresenceManager`.
- [ ] Use `ZSET` for `presence:org:{orgID}` with timestamp as score to allow efficient pruning.
- [ ] Use `HASH` (or JSON String) for `presence:user:{userID}` metadata.
- [ ] Implement `SetUserOnline`, `SetUserOffline`, `GetOnlineUsers`.
- [ ] Implement `PruneStaleUsers` to remove users from `ZSET` who haven't heartbeat-ed.

### 1.2 WebSocket Integration (`pkg/ws/ws_manager.go` & `pkg/ws/client.go`)
- [ ] Add `PresenceManager` to `WebSocketManager` struct.
- [ ] Update `Client` struct to include `UserID`, `OrgID`, `UserData`.
- [ ] **Heartbeat Loop:**
    - Server sends `ping` every 30s.
    - Client sends `pong`.
    - On `pong`, call `PresenceManager.RefreshUserHeartbeat`.
- [ ] **Connection Events:**
    - On `Register`: Call `SetUserOnline` and broadcast `presence_update:join`.
    - On `Unregister`: Call `SetUserOffline` and broadcast `presence_update:leave`.

### 1.3 HTTP Endpoint (`internal/modules/organization/...`)
- [ ] Add `GetPresence` method to `OrganizationController`.
- [ ] Add route `GET /api/v1/organizations/:id/presence`.
- [ ] Call `PresenceManager.GetOnlineUsers` via `WebSocketManager` (or inject PresenceManager directly).

### 1.4 Pruning Worker (`internal/worker/...`)
- [ ] Create a lightweight background ticker (every 30s) in `ws.Manager` (or separate worker) to call `PruneStaleUsers`.
- [ ] Ensure pruned users broadcast a `leave` event.

---

## 2. Frontend: Real-time UI (Next.js)

### 2.1 Presence Store (`web/src/stores/use-presence-store.ts`)
- [ ] Create Zustand store to hold `onlineUsers` array.
- [ ] Actions: `setUsers`, `addUser`, `removeUser`.

### 2.2 WebSocket Hook (`web/src/hooks/use-presence.ts`)
- [ ] Listen for `presence_update` events.
- [ ] Dispatch actions to store.
- [ ] Handle `ping` (browser usually handles this at low level, but if application-level ping: reply `pong`).

### 2.3 UI Components
- [ ] `AvatarStack` component in `Navbar`.
- [ ] Tooltip showing list of names.
- [ ] Badge `+N` if more than 5 users.

---

## 3. Testing Strategy
- [ ] **Unit Test:** `PresenceManager` (Redis mocks).
- [ ] **Integration Test:** Connect 2 WS clients, verify `join` event received.
- [ ] **Manual:** Open 2 tabs, check Avatar Stack updates.
