# Module, Folder, And Contract Map

Dokumen ini membedah struktur folder dan kontrak utama per modul pada project.

Fokus dokumen:

- folder apa berisi apa
- entity inti
- use case utama
- repository utama
- route yang diekspos
- dependency penting
- catatan peran bisnis modul

## 1. Struktur Folder Utama

```text
cmd/api/
  entrypoint server, lifecycle, graceful shutdown

internal/config/
  wiring dependency, bootstrap DB, Redis, Casbin, storage, worker

internal/router/
  setup router, route grouping, middleware placement

internal/middleware/
  auth, tenant, API key, Casbin, rate limit, logging, metrics

internal/modules/
  business modules by bounded area

internal/worker/
  async task distributor, processor, scheduler, handlers

pkg/
  reusable infra helpers: jwt, ws, sse, storage, tus, telemetry, tx, querybuilder

db/migrations/
  schema evolution

db/seeds/
  role, policy, and bootstrap seeding

tests/
  unit, integration, e2e, realtime, business scenarios
```

## 2. Standard Module Shape

Setiap modul umumnya mengikuti pola ini:

```text
internal/modules/<module>/
  module.go
  entity/
  model/
  repository/
  usecase/
  delivery/http/
  test/
```

Kontrak per layer:

- `entity/`: representasi domain atau persistence object utama
- `model/`: request, response, DTO, converter
- `repository/`: kontrak akses data dan implementasi persistence
- `usecase/`: aturan bisnis dan orkestrasi
- `delivery/http/`: controller dan route binding
- `module.go`: composition mini untuk modul itu sendiri

## 3. Router Contract Map

Route group utama didefinisikan di `internal/router/router.go`.

### Public routes

- `/auth/login`
- `/auth/register`
- `/auth/refresh`
- `/auth/forgot-password`
- `/auth/reset-password`
- `/auth/verify-email`
- `/auth/sso/:provider`
- `/auth/sso/:provider/callback`
- `/users/register`
- `/organizations/invitations/accept`

### Authenticated routes

- `/auth/logout`
- `/auth/ticket`
- `/auth/resend-verification`
- `/auth/me`
- `/stats/summary`
- `/stats/activity`
- `/stats/insights`
- `/users/me`
- `/organizations`
- `/organizations/me`
- `/permissions/check-batch`
- `/api-keys`

### Tenant authorized routes

- `/organizations/:id`
- `/organizations/slug/:slug`
- `/organizations/:id/members/*`
- `/organizations/:id/presence`
- `/projects`
- `/webhooks`

### Authorized admin routes

- `/permissions/*`
- `/access-rights/*`
- `/endpoints/*`
- `/roles/*`
- `/users/*`
- `/audit-logs/*`

## 4. Auth Module

### Folder

`internal/modules/auth/`

### Core entities and models

- auth session model
- password reset token
- email verification token
- SSO identity link support

### Use case contract

Use case inti:

- `Register`
- `Login`
- `RefreshToken`
- `ValidateAccessToken`
- `ValidateRefreshToken`
- `Verify`
- `RevokeToken`
- `RevokeAllSessions`
- `ForgotPassword`
- `ResetPassword`
- `RequestVerification`
- `VerifyEmail`
- `GetTicket`
- `GetSSORedirectURL`
- `HandleSSOCallback`

### Repository contract

Repository auth utama:

- `TokenRepository`
  - simpan session Redis
  - manage reset token dan verification token di DB
  - track login attempts dan account lock di Redis
- `AuthzManager`
  - assign default role
  - baca role user dari Casbin

### Routes

- public auth routes
- authenticated auth routes
- websocket ticket route

### Dependencies

- JWT manager
- Redis
- User repository
- Organization repository
- transaction manager
- event publisher ke WS dan SSE
- Casbin adapter
- task distributor
- ticket manager
- SSO providers

### Business role

Modul ini adalah pintu masuk identity, session, SSO, dan onboarding default workspace.

## 5. User Module

### Folder

`internal/modules/user/`

### Core entities and models

- `User`
- request dan response untuk register, update, delete, status change
- avatar update hook untuk TUS

### Use case contract

- `Create`
- `GetUserByID`
- `GetAllUsers`
- `GetAllUsersDynamic`
- `Current`
- `Update`
- `UpdateStatus`
- `UpdateAvatar`
- `SetAvatarURL`
- `GetAvatarUrl`
- `DeleteUser`
- `HardDeleteSoftDeletedUsers`

### Repository contract

- CRUD user
- dynamic search user
- status update
- hard delete soft-deleted user

### Routes

- public: `/users/register`
- authenticated: `/users/me`
- admin: list, search, detail, status update, delete

### Dependencies

- transaction manager
- Casbin enforcer
- audit use case
- auth use case
- webhook use case
- storage provider

### Business role

Modul ini mengelola lifecycle user dan menjadi penghubung ke audit, policy, session revocation, dan file storage.

## 6. Organization Module

### Folder

`internal/modules/organization/`

### Core entities and models

- `Organization`
- `OrganizationMember`
- `InvitationToken`
- request response create org, invite member, accept invite, update member

### Use case contract

Organization use case:

- `CreateOrganization`
- `GetOrganization`
- `GetOrganizationBySlug`
- `UpdateOrganization`
- `GetUserOrganizations`
- `DeleteOrganization`

Member use case:

- `InviteMember`
- `GetMembers`
- `UpdateMember`
- `RemoveMember`
- `AcceptInvitation`
- `GetPresence`

Reader contract:

- `ValidateMembership`
- `GetMemberRole`
- `InvalidateMembershipCache`
- `InvalidateOrganizationCache`

### Repository contract

- organization repository
- organization member repository
- invitation repository

### Routes

- authenticated: create org, get my organizations
- public: accept invitation
- tenant: org detail, org update, delete, member management, presence

### Dependencies

- Redis cached reader
- user repository
- transaction manager
- task distributor
- Casbin enforcer
- presence reader
- frontend base URL untuk invitation link

### Business role

Ini adalah inti multi-tenancy: workspace creation, membership, invitation, role-in-tenant, dan presence.

## 7. Permission Module

### Folder

`internal/modules/permission/`

### Core entities and models

Modul ini lebih dominan di policy/model daripada entity GORM sendiri.

Konsep utamanya:

- user-role assignment
- role-permission assignment
- role inheritance
- resource aggregation matrix
- access-right bundle assignment

### Use case contract

- `AssignRoleToUser`
- `RevokeRoleFromUser`
- `GrantPermissionToRole`
- `RevokePermissionFromRole`
- `GetAllPermissions`
- `GetPermissionsForRole`
- `GetUsersForRole`
- `UpdatePermission`
- `AddParentRole`
- `RemoveParentRole`
- `GetParentRoles`
- `BatchCheckPermission`
- `GetResourceAggregation`
- `GetInheritanceTree`
- `GetRoleAccessRights`
- `AssignAccessRight`
- `RevokeAccessRight`
- `DeleteRole`

### Repository contract

Tidak punya repository sendiri untuk policy storage karena policy utama dikelola lewat Casbin `IEnforcer`, tetapi modul ini memakai:

- `RoleRepository`
- `UserRepository`
- `AccessRepository`
- `AuditUseCase`

### Routes

- `/permissions/assign-role`
- `/permissions/revoke-role`
- `/permissions/grant`
- `/permissions/revoke`
- `/permissions/inheritance`
- `/permissions/resources`
- `/permissions/inheritance-tree`
- `/permissions/check-batch`
- `/permissions/assign-access-right`

### Dependencies

- transactional Casbin enforcer
- role repo
- user repo
- access repo
- audit use case

### Business role

Ini modul orkestrasi authorization. Ia menjadi pusat RBAC domain-aware, bukan hanya endpoint admin biasa.

## 8. Access Module

### Folder

`internal/modules/access/`

### Core entities and models

- `AccessRight`
- `Endpoint`
- link antara access right dan endpoint

### Use case contract

- `CreateAccessRight`
- `GetAllAccessRights`
- `CreateEndpoint`
- `LinkEndpointToAccessRight`
- `UnlinkEndpointFromAccessRight`
- `DeleteAccessRight`
- `DeleteEndpoint`
- `GetEndpointsDynamic`
- `GetAccessRightsDynamic`

### Repository contract

- CRUD access rights
- CRUD endpoints
- linking and unlinking
- dynamic filtering

### Routes

- `/access-rights`
- `/access-rights/link`
- `/access-rights/unlink`
- `/endpoints`

### Dependencies

- DB repository
- logger
- validator

### Business role

Modul ini menyediakan abstraction layer agar permission tidak harus dikelola endpoint-per-endpoint.

## 9. Role Module

### Folder

`internal/modules/role/`

### Core entities and models

- `Role`

### Use case contract

- `Create`
- `Update`
- `GetAll`
- `Delete`
- `GetAllRolesDynamic`

### Repository contract

- role CRUD
- role lookup by id and name
- dynamic role search

### Routes

- `/roles`
- `/roles/search`

### Dependencies

- transaction manager
- role repository
- permission use case

### Business role

Role adalah master data authorization. Delete role akan memicu cleanup policy di Casbin.

## 10. Project Module

### Folder

`internal/modules/project/`

### Core entities and models

- `Project`

### Use case contract

- `CreateProject`
- `GetProjects`
- `GetProjectByID`
- `UpdateProject`
- `DeleteProject`

### Repository contract

- CRUD project per organization

### Routes

- `/projects`

### Dependencies

- project repository
- tenant context dari middleware

### Business role

Ini resource bisnis tenant-scoped yang saat ini masih tipis dan bergantung pada fondasi auth, tenant, dan RBAC.

## 11. Audit Module

### Folder

`internal/modules/audit/`

### Core entities and models

- `AuditLog`
- `AuditOutbox`
- export payload and response models

### Use case contract

- `LogActivity`
- `GetLogsDynamic`
- `ExportLogs`
- `ExportLogsAsync`

### Repository contract

- create audit log
- create outbox
- find logs dynamically
- batch export
- find pending outbox
- update outbox status
- delete outbox

### Routes

- `/audit-logs/search`
- `/audit-logs/export`
- `/audit-logs/export-async`

### Dependencies

- WS manager
- task distributor
- DB repository

### Business role

Ini modul jejak operasional. Ia penting untuk integrity, observability, dan compliance behavior.

## 12. API Key Module

### Folder

`internal/modules/api_key/`

### Core entities and models

- `ApiKey`
- API key identity
- create and list request response

### Use case contract

- `Create`
- `List`
- `Revoke`
- `Authenticate`

### Repository contract

- create key
- list by organization
- find by id
- find by hash
- update metadata seperti `last_used_at`
- delete key

### Routes

- `/api-keys`

### Dependencies

- user repository
- Redis cache
- logger

### Business role

Memberi machine access yang tetap user-bound dan scope-bound, bukan bypass dari auth model utama.

## 13. Webhook Module

### Folder

`internal/modules/webhook/`

### Core entities and models

- `Webhook`
- webhook delivery logs
- trigger payload

### Use case contract

- `Create`
- `Update`
- `Delete`
- `FindByID`
- `FindByOrganizationID`
- `Trigger`
- `FindLogs`

### Repository contract

- CRUD webhook
- find by event
- read webhook logs

### Routes

- `/webhooks`
- `/webhooks/:id/logs`

### Dependencies

- task distributor
- validator
- logger

### Business role

Webhook adalah outbound integration surface untuk mengirim event domain keluar sistem secara async.

## 14. Stats Module

### Folder

`internal/modules/stats/`

### Core entities and models

- dashboard summary
- dashboard activity
- system insights

### Use case contract

- `GetDashboardSummary`
- `GetDashboardActivity`
- `GetSystemInsights`

### Repository contract

Modul ini tidak memakai repository terpisah; use case langsung membaca DB.

### Routes

- `/stats/summary`
- `/stats/activity`
- `/stats/insights`

### Dependencies

- GORM DB
- organization scope dari context

### Business role

Menyediakan agregasi monitoring dan dashboard ringkas, baik global maupun tenant-scoped.

## 15. Worker Layer Contract

### Folder

`internal/worker/`

### Komponen

- `distributor.go`
  - enqueue audit, email, webhook, cleanup tasks
- `processor.go`
  - register handler dan proses queue
- `scheduler.go`
  - periodic cleanup dan audit outbox sync
- `handlers/`
  - email
  - webhook
  - audit
  - cleanup
  - outbox

### Business role

Worker layer menjalankan side effect yang tidak cocok di request path sinkron:

- email invitation
- webhook delivery
- audit export
- cleanup maintenance
- audit outbox flush

## 16. Middleware Contract

### `AuthMiddleware`

- validasi access token
- cek sesi Redis
- inject `user_id`, `session_id`, `user_role`, `username`

### `APIKeyMiddleware`

- auth dengan `X-API-Key`
- inject identity user dan organization
- enforce scope otomatis atau eksplisit

### `TenantMiddleware`

- resolve org id atau slug
- validasi membership
- inject `organization_id` ke context dan gin context

### `CasbinMiddleware`

- final authorization gate untuk route yang membutuhkan RBAC enforcement

## 17. Data And Policy Backbone

Backbone skema yang paling penting:

- `users`
- `roles`
- `casbin_rule`
- `organizations`
- `organization_members`
- `invitation_tokens`
- `audit_logs`
- `audit_outbox`
- `projects`
- `api_keys`
- `webhooks`

Interpretasi desain:

- `users` menyimpan identity
- `organization_members` menyimpan hubungan user ke tenant
- `roles` menyimpan katalog role
- `casbin_rule` menyimpan policy efektif
- `audit_logs` dan `audit_outbox` menyimpan jejak perubahan

## 18. Module Dependency Summary

```text
Audit
  standalone infra-aware module

Auth
  depends on UserRepo, OrgRepo, JWT, Redis TokenRepo, AuthzManager, WS, SSE, Worker

User
  depends on Repo, AuditUC, AuthUC, WebhookUC, Enforcer, Storage

Organization
  depends on OrgRepo, MemberRepo, InvitationRepo, UserRepo, Worker, Enforcer, Presence

Permission
  depends on Enforcer, RoleRepo, UserRepo, AccessRepo, AuditUC

Role
  depends on RoleRepo, PermissionUC

Access
  depends on AccessRepo

Project
  depends on ProjectRepo

ApiKey
  depends on ApiKeyRepo, UserRepo, Redis

Webhook
  depends on WebhookRepo, Worker

Stats
  depends on DB and tenant context
```

## 19. Reading Order Recommendation

Untuk memahami codebase ini dengan cepat, urutan terbaik:

1. `cmd/api/main.go`
2. `internal/config/app.go`
3. `internal/router/router.go`
4. `internal/middleware/auth_middleware.go`
5. `internal/middleware/tenant_middleware.go`
6. `internal/modules/auth/usecase/auth_usecase.go`
7. `internal/modules/organization/usecase/*.go`
8. `internal/modules/permission/usecase/*.go`
9. `internal/modules/user/usecase/user_usecase.go`
10. `internal/modules/audit/usecase/audit_usecase.go`

## 20. Summary

Struktur folder project ini rapi dan cukup konsisten terhadap pola Clean Architecture, tetapi secara bisnis pusat gravitasi sistemnya ada pada:

- `auth`
- `organization`
- `permission`
- `user`
- `audit`

Modul lain berfungsi sebagai capability layer yang dibangun di atas lima area inti tersebut.
