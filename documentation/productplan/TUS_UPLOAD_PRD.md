# Product Requirements Document (PRD): Global Resumable Upload Service

| Metadata         | Detail                                         |
| :--------------- | :--------------------------------------------- |
| **Project Name** | NexusOS Enterprise Platform                    |
| **Feature Name** | Global Resumable Upload Service (Tus + RustFS) |
| **Version**      | 1.0                                            |
| **Status**       | **Implemented & Verified**                     |
| **Priority**     | High (Infrastructure Foundation)               |

---

## 1. Executive Summary

The current file upload mechanism is "single-shot" (multipart/form-data), which is unreliable for large files (>50MB) or unstable network connections. This feature introduces a **Resumable Upload Service** based on the open standard **Tus Protocol**. It will act as a centralized, module-agnostic service that stores files in **RustFS** (S3-compatible) and notifies specific business modules (User, Project, Document) via an event registry system.

---

## 2. Problem Statement

- **Reliability:** Uploads fail completely if the network drops for a second. Users must restart from 0%.
- **Scalability:** Handling large file streams in memory (standard Gin binding) consumes excessive RAM.
- **Coupling:** Upload logic is currently duplicated in controllers (e.g., `UpdateAvatar`), making it hard to maintain or switch storage providers globally.

---

## 3. Functional Requirements

### FR-01: Resumable Uploads

- **FR-01.1:** The system MUST support the **Tus 1.0.0** protocol.
- **FR-01.2:** Clients MUST be able to pause, resume, and terminate uploads.
- **FR-01.3:** The server MUST accept `PATCH` requests containing binary chunks of the file.

### FR-02: Storage & Persistence

- **FR-02.1:** Files MUST be stored in an S3-compatible object storage (**RustFS**).
- **FR-02.2:** The system MUST NOT store large files on the local disk of the API server (stateless container design).

### FR-03: Dynamic Feature Binding (The Registry)

- **FR-03.1:** The upload service MUST be generic. It should not know about "Avatars" or "Documents".
- **FR-03.2:** Clients MUST provide a `type` metadata field (e.g., `type: avatar`).
- **FR-03.3:** The system MUST route the "Upload Completed" event to the correct business module based on the `type` metadata.

### FR-04: Security

- **FR-04.1:** The upload endpoint MUST be protected by **JWT Authentication**. Anonymous uploads are strictly forbidden.
- **FR-04.2:** The system MUST validate `Content-Type` and `Size` limits before accepting the upload creation.

---

## 4. Technical Specifications

### 4.1 Architecture

- **Protocol:** Tus (via `tusd` v2 library).
- **Storage:** AWS SDK v2 connecting to RustFS.
- **Design Pattern:** Observer / Registry Pattern.

### 4.2 Data Flow

1.  **Client:** `POST /api/v1/upload/files/` (Metadata: `{type: "avatar", user_id: "123"}`).
2.  **Server:** Returns `201 Created` with `Location: .../files/abc-123`.
3.  **Client:** `PATCH .../files/abc-123` (Sends binary data).
4.  **Server (Background):** Detects completion -> Calls `Registry.Get("avatar")` -> Executes `AvatarHook.HandleUpload()`.
5.  **Module:** Updates `users` table with the new S3 URL.

### 4.3 API Endpoints (Tus Standard)

| Method   | Endpoint                   | Purpose               |
| :------- | :------------------------- | :-------------------- |
| `POST`   | `/api/v1/upload/files/`    | Create upload session |
| `HEAD`   | `/api/v1/upload/files/:id` | Check offset          |
| `PATCH`  | `/api/v1/upload/files/:id` | Upload data           |
| `DELETE` | `/api/v1/upload/files/:id` | Cancel upload         |

---

## 5. Non-Functional Requirements

- **Performance:** Minimal memory footprint (streaming directly to S3).
- **Extensibility:** Adding a new upload type (e.g., "Invoice") should require **Zero Changes** to the core upload handler.
- **Compatibility:** Must work with standard Tus clients like **Uppy** and **tus-js-client**.

---

## 6. User Stories

### US-01: User Uploads Avatar

> As a user, I want to upload a high-res profile picture. If my connection drops, it should continue where it left off so I don't waste data.

### US-02: Developer Adds "Document" Feature

> As a backend developer, I want to add a "Project Document" upload feature without modifying the complex TUS server code, so I can ship faster.

---

## 7. Success Metrics

- **100%** of file uploads are processed via Tus.
- **Zero** code modification in `pkg/tus` when adding new features in `internal/modules`.
- **< 50MB** RAM usage during 1GB file upload.
