# Architecture Visual And Sequence Guide

Dokumen ini merangkum arsitektur visual project serta sequence diagram untuk flow bisnis inti:

- `register`
- `login`
- `invite member`
- `assign role`

Dokumen ini melengkapi [ARCHITECTURE.md](./ARCHITECTURE.md) dengan fokus pada:

- peta layer runtime
- hubungan antarkomponen
- flow request end-to-end
- titik transaksi, policy, audit, dan background task

## 1. System Context

```mermaid
flowchart LR
    Client[Web App or API Client]
    API[Gin HTTP API]
    WS[WebSocket and SSE]
    Worker[Asynq Worker and Scheduler]
    MySQL[(MySQL)]
    Redis[(Redis)]
    Storage[Local or S3 Storage]
    Casbin[Casbin Policy Store]
    SMTP[SMTP Provider]
    SSO[SSO Providers]
    WebhookTarget[External Webhook Targets]

    Client --> API
    Client --> WS
    API --> MySQL
    API --> Redis
    API --> Storage
    API --> Casbin
    API --> SSO
    API --> Worker
    Worker --> Redis
    Worker --> MySQL
    Worker --> SMTP
    Worker --> WebhookTarget
```

## 2. Runtime Composition

Semua shared dependency dan module wiring dibuat di `internal/config/app.go`.

```mermaid
flowchart TD
    AppConfig[AppConfig]
    Logger[Logger]
    Validator[Validator]
    DB[(MySQL via GORM)]
    Redis[(Redis)]
    JWT[JWT Manager]
    Enforcer[Transactional Casbin Enforcer]
    Storage[Storage Provider]
    WSManager[WebSocket Manager]
    SSEManager[SSE Manager]
    TaskDistributor[Asynq Task Distributor]
    TaskProcessor[Asynq Task Processor]
    Scheduler[Asynq Scheduler]

    AuthModule[Auth Module]
    AuditModule[Audit Module]
    UserModule[User Module]
    OrgModule[Organization Module]
    PermissionModule[Permission Module]
    RoleModule[Role Module]
    AccessModule[Access Module]
    ProjectModule[Project Module]
    ApiKeyModule[API Key Module]
    WebhookModule[Webhook Module]
    StatsModule[Stats Module]

    Router[Router and Middleware]

    AppConfig --> Logger
    AppConfig --> Validator
    AppConfig --> DB
    AppConfig --> Redis
    AppConfig --> JWT
    AppConfig --> Enforcer
    AppConfig --> Storage
    AppConfig --> WSManager
    AppConfig --> SSEManager
    AppConfig --> TaskDistributor
    AppConfig --> TaskProcessor
    AppConfig --> Scheduler

    DB --> AuditModule
    Logger --> AuditModule
    WSManager --> AuditModule
    TaskDistributor --> AuditModule

    DB --> AuthModule
    Redis --> AuthModule
    JWT --> AuthModule
    Enforcer --> AuthModule
    TaskDistributor --> AuthModule
    WSManager --> AuthModule
    SSEManager --> AuthModule

    AuthModule --> UserModule
    AuditModule --> UserModule
    Storage --> UserModule

    UserModule --> ApiKeyModule
    UserModule --> OrgModule

    TaskDistributor --> WebhookModule
    WebhookModule --> UserModule

    Enforcer --> PermissionModule
    AccessModule --> PermissionModule
    UserModule --> PermissionModule
    AuditModule --> PermissionModule

    PermissionModule --> RoleModule

    DB --> StatsModule
    DB --> ProjectModule

    AuthModule --> Router
    UserModule --> Router
    OrgModule --> Router
    PermissionModule --> Router
    RoleModule --> Router
    AccessModule --> Router
    ProjectModule --> Router
    ApiKeyModule --> Router
    WebhookModule --> Router
    StatsModule --> Router
```

## 3. Layered Architecture

```mermaid
flowchart TB
    subgraph Delivery
        Controllers[HTTP Controllers]
        Middleware[Auth Tenant API Key Casbin Middleware]
        Router[Router]
    end

    subgraph UseCase
        AuthUC[Auth UseCase]
        UserUC[User UseCase]
        OrgUC[Organization UseCase]
        MemberUC[Organization Member UseCase]
        PermissionUC[Permission UseCase]
        RoleUC[Role UseCase]
        ProjectUC[Project UseCase]
        ApiKeyUC[API Key UseCase]
        AuditUC[Audit UseCase]
        WebhookUC[Webhook UseCase]
        StatsUC[Stats UseCase]
    end

    subgraph Repository
        UserRepo[User Repository]
        OrgRepo[Organization Repositories]
        TokenRepo[Token Repository]
        RoleRepo[Role Repository]
        AccessRepo[Access Repository]
        ProjectRepo[Project Repository]
        AuditRepo[Audit Repository]
        ApiKeyRepo[API Key Repository]
        WebhookRepo[Webhook Repository]
    end

    subgraph Infra
        GORM[GORM]
        Redis[Redis]
        Asynq[Asynq]
        Casbin[Casbin]
        Storage[Storage Provider]
        WS[WS and SSE]
        TUS[TUS Upload]
    end

    Router --> Middleware
    Middleware --> Controllers
    Controllers --> UseCase
    UseCase --> Repository
    UseCase --> Infra
    Repository --> GORM
    Repository --> Redis
```

## 4. Access Segmentation

Router membagi akses ke empat zona utama:

```mermaid
flowchart LR
    Public[Public]
    Authenticated[Authenticated]
    TenantAuthorized[Tenant Authorized]
    Authorized[Authorized Admin]

    Public -->|login register reset verify invite accept| API1[Public Endpoints]
    Authenticated -->|me logout ticket stats my orgs api keys| API2[User Session Endpoints]
    TenantAuthorized -->|projects webhooks org detail members presence| API3[Tenant Scoped Endpoints]
    Authorized -->|permissions roles access-rights users audit| API4[Admin and RBAC Endpoints]
```

## 5. Request Pipeline

```mermaid
flowchart LR
    Req[HTTP Request]
    MW1[Request ID and Logging]
    MW2[Security and CORS]
    MW3[API Key or JWT Auth]
    MW4[User Status Check]
    MW5[Tenant Middleware]
    MW6[Casbin Middleware]
    Ctrl[Controller]
    UC[UseCase]
    Repo[Repository]
    DB[(DB or Redis)]
    Async[Async Task or WS Broadcast]
    Resp[HTTP Response]

    Req --> MW1 --> MW2 --> MW3 --> MW4 --> MW5 --> MW6 --> Ctrl --> UC --> Repo --> DB
    UC --> Async
    Ctrl --> Resp
```

## 6. Sequence Diagram: Register

Register di sistem ini tidak hanya membuat user, tetapi juga auto-provision workspace default dan auto-login.

```mermaid
sequenceDiagram
    participant C as Client
    participant R as Router or Controller
    participant A as AuthUseCase
    participant U as UserRepo
    participant O as OrgRepo
    participant P as CasbinAdapter
    participant T as TransactionManager
    participant Q as TaskDistributor
    participant S as TokenRepo Redis

    C->>R: POST /api/v1/auth/register
    R->>A: Register(request)
    A->>U: FindByUsername and FindByEmail
    A->>T: Begin transaction
    T->>U: Create user
    T->>P: Assign default global role
    T->>O: Create default organization plus owner member
    T->>Q: Enqueue audit REGISTER
    T-->>A: Commit
    A->>A: Call Login
    A->>S: Store session in Redis
    A-->>R: LoginResponse
    R-->>C: Access token and refresh token
```

### Register business notes

- user baru langsung punya workspace default
- global role default dipasang saat onboarding
- audit `REGISTER` dikirim sebagai async task
- hasil akhirnya bukan sekadar `201 created`, tetapi sesi login aktif

## 7. Sequence Diagram: Login

```mermaid
sequenceDiagram
    participant C as Client
    participant R as Router or Controller
    participant A as AuthUseCase
    participant S as TokenRepo Redis
    participant U as UserRepo
    participant P as CasbinAdapter
    participant O as OrgRepo
    participant Q as TaskDistributor
    participant E as EventPublisher

    C->>R: POST /api/v1/auth/login
    R->>A: Login(username password ip userAgent)
    A->>S: IsAccountLocked
    A->>U: FindByUsername
    A->>A: Check password
    alt password invalid
        A->>S: Increment attempts
        opt threshold reached
            A->>S: Lock account
            A->>Q: Enqueue ACCOUNT_LOCKED audit
        end
        A-->>R: Unauthorized or locked
    else password valid
        A->>S: Reset attempts
        A->>P: Get roles for user
        A->>S: Store token pair and session
        A->>Q: Enqueue LOGIN audit
        A->>O: Find user organizations
        A->>E: Publish user logged in event
        A-->>R: LoginResponse
        R-->>C: Access token and refresh token
    end
```

### Login business notes

- session disimpan di Redis, jadi logout dan revoke bisa benar-benar efektif
- login punya lockout policy
- user status `suspended` atau `banned` akan memblokir login
- login juga jadi event realtime ke organisasi yang relevan

## 8. Sequence Diagram: Invite Member

```mermaid
sequenceDiagram
    participant C as Client Admin or Owner
    participant R as Router or Controller
    participant M as OrganizationMemberUseCase
    participant T as TransactionManager
    participant OR as OrganizationRepo
    participant MR as MemberRepo
    participant UR as UserRepo
    participant IR as InvitationRepo
    participant CE as Casbin Enforcer
    participant Q as TaskDistributor

    C->>R: POST /api/v1/organizations/:id/members/invite
    R->>M: InviteMember(orgID request)
    M->>T: Begin transaction
    T->>OR: Find organization
    T->>MR: Check actor membership and role
    T->>UR: Find user by email
    alt user not found
        T->>UR: Create shadow user with status invited
    end
    T->>MR: Check existing membership
    T->>MR: Create or keep invited member record
    T->>IR: Delete old invitation token
    T->>IR: Create new invitation token
    T->>Q: Enqueue invitation email
    T-->>M: Commit
    M-->>R: MemberResponse
    R-->>C: Invitation queued
```

### Invite member business notes

- sistem mendukung invite user yang belum punya akun
- dibuat `shadow user` agar membership bisa dipersiapkan lebih awal
- invitation bersifat expirable
- email invitation bukan blocking dependency untuk commit transaksi

## 9. Sequence Diagram: Assign Role

```mermaid
sequenceDiagram
    participant C as Admin Client
    participant R as Router or Controller
    participant P as PermissionUseCase
    participant U as UserRepo
    participant RR as RoleRepo
    participant CE as Transactional Enforcer

    C->>R: POST /api/v1/permissions/assign-role
    R->>P: AssignRoleToUser(userID role domain)
    P->>U: Validate user exists
    P->>RR: Validate role exists
    P->>CE: Remove old grouping policy in domain
    P->>CE: Add new grouping policy user -> role -> domain
    P-->>R: Success
    R-->>C: Role assigned
```

### Assign role business notes

- role assignment bersifat domain-aware
- role lama pada domain yang sama dibersihkan dulu
- domain default adalah `global` jika tidak diisi
- policy hidup di Casbin, bukan di kolom role pada tabel user

## 10. Supporting Runtime Flows

### Audit logging

```mermaid
flowchart LR
    UC[UseCase]
    TX{Inside DB transaction}
    Outbox[Audit Outbox]
    AuditLog[Audit Logs]
    Scheduler[Scheduler every 5s]
    Worker[Outbox Handler]
    WS[WebSocket Broadcast]

    UC --> TX
    TX -->|yes| Outbox
    TX -->|no| AuditLog
    AuditLog --> WS
    Scheduler --> Worker --> Outbox --> AuditLog
```

### Tenant context propagation

```mermaid
flowchart LR
    Req[Request with X-Organization-ID]
    TenantMW[Tenant Middleware]
    Cache[CachedOrgReader Redis]
    RepoCheck[Member Repository]
    Ctx[Context with organization_id]
    Scope[GORM OrganizationScope]
    Query[Scoped DB Query]

    Req --> TenantMW
    TenantMW --> Cache
    Cache -->|miss| RepoCheck
    TenantMW --> Ctx --> Scope --> Query
```

## 11. Summary

Secara visual, sistem ini lebih tepat dibaca sebagai:

- platform identity dan session management
- multi-tenant workspace and membership engine
- RBAC orchestration layer berbasis Casbin
- operational control plane dengan audit, worker, webhook, dan realtime

Modul seperti `project`, `stats`, `api key`, `webhook`, dan `upload` berdiri di atas fondasi tersebut.
