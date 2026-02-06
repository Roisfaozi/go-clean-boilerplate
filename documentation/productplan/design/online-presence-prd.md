# Product Requirements Document (PRD): Online Presence

| Metadata | Detail |
| :--- | :--- |
| **Feature Name** | Online Presence (Real-time Collaboration) |
| **Module** | Real-time / Organization |
| **Version** | 1.0 |
| **Status** | **Draft** |
| **Priority** | High (Collaboration Enabler) |

---

## 1. Executive Summary
The **Online Presence** feature provides real-time visibility into team activity within the NexusOS dashboard. By displaying active users in real-time, we aim to enhance collaboration, prevent data conflicts (e.g., two users editing the same role), and create a livelier "virtual office" atmosphere for remote teams.

---

## 2. User Stories

### 2.1 As a Team Member
*   **I want** to see who else is currently online in my organization.
*   **So that** I know who is available for quick collaboration or chat.

### 2.2 As an Admin
*   **I want** to see active user sessions.
*   **So that** I can monitor system usage and security in real-time.

---

## 3. Functional Requirements

### FR-01: Real-time Connection Tracking
*   **FR-01.1:** The system MUST track active WebSocket connections per Organization.
*   **FR-01.2:** Users opening multiple tabs MUST be counted as a single active presence session.
*   **FR-01.3:** The system MUST detect disconnection events (tab close, network loss) within 60 seconds (Heartbeat timeout).

### FR-02: Initial State Synchronization
*   **FR-02.1:** Upon loading the dashboard, the client MUST fetch the current list of online users via API.
*   **FR-02.2:** The list MUST include: User ID, Name, Avatar URL, and Role.

### FR-03: Live Updates (Broadcast)
*   **FR-03.1:** When a user joins (connects), a `user_joined` event MUST be broadcast to all other members of the *same* organization.
*   **FR-03.2:** When a user leaves (disconnects), a `user_left` event MUST be broadcast.
*   **FR-03.3:** Broadcasts MUST be strictly scoped to the Organization ID. Cross-tenant leakage is a critical security failure.

### FR-04: UI Representation (Avatar Stack)
*   **FR-04.1:** Display a horizontal stack of user avatars in the top Navbar.
*   **FR-04.2:** Limit display to 5 users. Show a `+N` badge for excess users.
*   **FR-04.3:** Hovering over an avatar MUST show a tooltip with the user's Name.
*   **FR-04.4:** The current user's own avatar SHOULD NOT be included in the presence list (to save space and reduce noise).

---

## 4. Technical Specifications

### 4.1 Data Architecture (Redis)
*   **Storage Strategy:** Redis `SET` for active IDs and `HASH` for user metadata.
    *   `presence:org:{org_id}` (Set of UserIDs)
    *   `presence:user:{user_id}` (Hash: name, avatar)
*   **TTL:** Keys expire after 5 minutes to prevent "zombie" presence if server crashes. Keys are refreshed on every Heartbeat.

### 4.2 WebSocket Protocol
*   **Ping/Pong:** Server sends `ping` every 30s. Client responds `pong`.
*   **Payloads:**
    ```json
    // Event: User Joined
    {
      "type": "presence_update",
      "event": "join",
      "data": { "user_id": "u1", "name": "Alice", "avatar": "..." }
    }
    ```

### 4.3 API Endpoint
*   `GET /api/v1/organizations/:id/presence`
    *   **Auth:** Bearer Token + Tenant Membership Check.
    *   **Response:** Array of User objects.

---

## 5. Security & Performance

### 5.1 Security
*   **Tenant Isolation:** Before broadcasting or fetching presence, the system MUST validate that the requester belongs to the target Organization.
*   **Data Minimization:** Only public profile data (Name, Avatar) is shared. Email/Phone are excluded from presence payloads.

### 5.2 Performance
*   **Redis Operations:** All presence operations are O(1) or O(N) where N is the number of online users (typically small per org).
*   **Debounce:** "Join" events for the same user (e.g., opening 2nd tab) should typically be suppressed or handled gracefully by the frontend (idempotent list).

---

## 6. Success Metrics
*   **Accuracy:** User status updates in UI within < 2 seconds of actual connection/disconnection.
*   **Reliability:** Zero "ghost users" (users shown as online but actually offline) persisting for > 2 minutes.
