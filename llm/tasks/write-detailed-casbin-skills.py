from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SKILLS = ROOT / '.agents' / 'skills'

def write(name: str, body: str):
    path = SKILLS / name / 'SKILL.md'
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(body.strip() + '\n')

COMMON_FOOTER = '''
## Stop Conditions
- Stop and ask before destructive DB/schema/data operations not explicitly requested.
- Stop if live code contradicts `llm/cache/*`; live code wins, then document drift in `llm/tasks/`.
- Stop if route ownership, tenant boundary, or auth stratum is unclear.

## Completion Output
Report:
- files changed
- commands run and exact result
- verification skipped and exact blocker
- risks or follow-up work
'''

write('api-endpoint', '''---
name: api-endpoint
description: Use when adding or changing backend HTTP endpoints, route protection, API-key scope behavior, Swagger-visible contracts, or frontend-consumed API shape in the Casbin monorepo.
---

# API Endpoint: Route Contract and Protection

**Announce at start:** "I'm using the api-endpoint skill to preserve route, auth, tenant, and contract boundaries."

## When To Use
| Use this skill when... | Use another skill when... |
|---|---|
| adding/changing Gin routes | pure usecase logic only -> `go-service` |
| changing route group/auth/Casbin/API-key scope | persistence-only change -> `database-transactions` |
| changing contract consumed by frontend | frontend-only UI change -> `frontend-surface` / `ui` |

## Read Order
1. `AGENTS.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/backend-map.md`
4. `llm/cache/domain-rules.md`
5. `llm/workflows/api-endpoint.md`
6. `internal/router/router.go`
7. target `internal/modules/*/delivery/http/*routes.go`
8. target controller, usecase, repository, request/response structs

## Route Strata Decision
Choose one intentionally:
- `public`: no auth; safe public auth/invitation style flows only.
- `authenticated`: API-key/JWT/session/status checks, no required tenant/Casbin policy.
- `tenantAuthorized`: auth + tenant org + Casbin policy.
- `authorized`: admin-style scope + optional tenant + Casbin policy.
- `upload`: TUS route with upload-specific handler and auth/status middleware.

## Workflow
### Phase 1 — Recon
- Find existing similar endpoint in same module.
- Trace route registration from `internal/router/router.go` to module routes.
- Identify consumers in `apps/web`, `apps/client`, and `packages/api-types`.

### Phase 2 — Contract
Define:
- method and path
- path/query/body params
- response shape and error shape
- route stratum
- API-key scope behavior (auto or explicit)
- tenant/Casbin domain semantics

### Phase 3 — Implementation
- Add/update request/response structs and validation tags.
- Keep parsing/validation in controller.
- Keep business rules in usecase.
- Keep GORM/query details in repository.
- Update Swagger comments/artifacts when public API contract changes.

### Phase 4 — Consumer Sync
- Audit `apps/web/src/app/api/v1/[...path]/route.ts` if `apps/web` can call it.
- Audit `apps/client/app/routes/api-proxy.ts` if `apps/client` can call it.
- Update shared types/client helpers when payload shape changes.

### Phase 5 — Verification
- Narrow module/controller/usecase tests first.
- Integration/E2E when route stratum, cookie/session, tenant, API-key, or Casbin behavior changed.
- `pnpm go:docs` when Swagger artifacts should change.

## Review Checklist
- [ ] route registered exactly once
- [ ] route group matches product intent
- [ ] API-key scope cannot be bypassed
- [ ] JWT parsing is not treated as enough without Redis session validation
- [ ] tenant context comes from middleware/usecase boundary, not ad hoc query params
- [ ] frontend consumers are updated or confirmed unaffected
''' + COMMON_FOOTER)

write('go-service', '''---
name: go-service
description: Use when changing Go backend module logic, usecases, repositories, module constructors, worker-owned backend behavior, or dependency wiring in the Casbin backend.
---

# Go Service: Backend Module Change

**Announce at start:** "I'm using the go-service skill to preserve Casbin repo backend boundaries."

## When To Use
| Use this skill when... | Use another skill when... |
|---|---|
| changing usecase/repository/module wiring | route protection changes -> `api-endpoint` + `auth-tenant-casbin` |
| changing business rules | schema migrations -> `database-transactions` |
| changing worker-triggered backend behavior | frontend contract changes -> `cross-stack-change` |

## Read Order
1. `AGENTS.md`
2. `llm/cache/backend-map.md`
3. `llm/cache/module-map.md`
4. `llm/cache/domain-rules.md`
5. `llm/conventions/golang.md`
6. `llm/workflows/go-service.md`
7. `internal/config/app.go` when lifecycle/constructor changes
8. target `internal/modules/*/module.go`
9. target controller/usecase/repository/model/tests

## Backend Boundary Map
- controller: bind/validate/request-response only
- usecase: business rules, transaction orchestration, side-effect decisions
- repository: GORM/query details only
- `internal/config/app.go`: dependency graph and constructor wiring
- `internal/router/router.go`: HTTP route/middleware placement

## Workflow
### Phase 1 — Locate Owner
- Identify owning module and exact usecase method.
- Check constructor dependencies in `module.go` and app wiring.
- Identify cross-module calls and side effects: audit, webhook, worker, storage, Casbin, Redis.

### Phase 2 — Risk Classification
Mark touched boundaries:
- auth/session
- tenant/org membership/cache
- Casbin policy/enforcer
- API key
- TUS/storage
- worker/audit/webhook
- querybuilder/list filtering
- realtime/SSE/WebSocket

Load matching project-specific skill if any boundary is touched.

### Phase 3 — Patch Rules
- Do not pass full app config into usecases.
- Preserve context propagation.
- Keep transactional writes together.
- Use transactional enforcer patterns when Casbin policy writes share DB semantics.
- Do not move business rules into handlers.
- Do not make tenant checks optional.

### Phase 4 — Test Strategy
- Use package-local unit tests for pure usecase/repository logic.
- Use integration tests when DB/Redis/Casbin/worker/tenant boundaries are involved.
- Use E2E tests when route lifecycle/cookies/tokens/frontend proxy behavior changes.

## Review Checklist
- [ ] dependency ownership clear
- [ ] no full config object passed into usecase
- [ ] context propagation intact
- [ ] repository hides GORM details
- [ ] side effects intentionally sync/async
- [ ] tests or code trace cover changed behavior
''' + COMMON_FOOTER)

write('auth-tenant-casbin', '''---
name: auth-tenant-casbin
description: Use when changing authentication, Redis-backed sessions, organization tenant resolution, membership cache, Casbin policy/enforcement, protected route middleware, or role/permission behavior.
---

# Auth Tenant Casbin: High-Risk Boundary

**Announce at start:** "I'm using the auth-tenant-casbin skill because this touches the repo's highest-risk access boundary."

## Iron Rule
Do not assume JWT validity is enough. Protected routes need the repo's middleware/session/tenant/Casbin layering as implemented in live code.

## Read Order
1. `AGENTS.md`
2. `llm/cache/domain-rules.md`
3. `llm/cache/backend-map.md`
4. `internal/router/router.go`
5. `internal/middleware/auth_middleware.go`
6. `internal/middleware/tenant_middleware.go`
7. `internal/middleware/casbin_middleware.go`
8. target auth/organization/permission/role usecases

## Boundary Checklist
- bearer/cookie token parsing
- Redis-backed session validation
- user status middleware
- organization tenant context
- membership/cache invalidation
- Casbin subject/domain/object/method enforcement
- API-key scope layering if route is protected

## Workflow
### Phase 1 — Trace Request
Trace exact request path:
`router group -> API-key middleware -> auth/session -> user status -> tenant -> Casbin -> controller -> usecase`.

### Phase 2 — Determine Authority
- Auth/session belongs in middleware/usecase boundaries.
- Tenant org belongs in tenant middleware and organization usecase/cache paths.
- Policy writes belong in permission/Casbin abstractions.
- Route protection belongs in router/middleware, not frontend UI.

### Phase 3 — Patch
- Preserve Redis session checks.
- Preserve tenant context requirements before Casbin on tenant routes.
- Preserve owner/admin/member constraints in organization flows.
- Use transactional enforcer patterns for policy writes tied to DB transactions.

### Phase 4 — Verify
- Narrow middleware/usecase tests first.
- Integration/E2E for route lifecycle, cookies/tokens, tenant isolation, role/permission decisions.

## Red Flags
- "JWT parsed successfully" used as full auth proof.
- org ID accepted from request without membership/tenant boundary.
- Casbin policy changed outside permission abstractions.
- route moved to public/authenticated group for convenience.
''' + COMMON_FOOTER)

write('api-key-scope', '''---
name: api-key-scope
description: Use when changing protected endpoints, API-key authentication, API-key scopes, organization-scoped API keys, or route access behavior involving API keys.
---

# API Key Scope: Protected Route Capability

**Announce at start:** "I'm using the api-key-scope skill to preserve protected endpoint scope behavior."

## Read Order
1. `llm/cache/domain-rules.md`
2. `llm/workflows/api-endpoint.md`
3. `internal/router/router.go`
4. `internal/middleware/api_key_middleware.go`
5. `internal/modules/api_key`
6. affected route/controller tests

## Workflow
### Phase 1 — Route Group
Identify if route is `authenticated`, `tenantAuthorized`, or `authorized`.

### Phase 2 — Scope Decision
Decide:
- no API-key access
- auto scope derived from route
- explicit scope required
- organization-scoped key behavior

### Phase 3 — Patch
- Preserve API-key identity and scope separation.
- Preserve org-scoped behavior and Redis use.
- Keep API-key checks in middleware/router boundary.

### Phase 4 — Verification
- Test allowed key.
- Test missing key/token.
- Test wrong scope.
- Test wrong organization when route is tenant-scoped.

## Red Flags
- accepting identity without scope
- adding route to protected group without scope decision
- checking scope ad hoc in controller while middleware supports it
''' + COMMON_FOOTER)

write('database-transactions', '''---
name: database-transactions
description: Use when changing GORM transactions, all-or-nothing writes, schema migrations, seed data, Casbin policy writes tied to DB state, or tenant-sensitive repository behavior.
---

# Database Transactions: GORM and Casbin Consistency

**Announce at start:** "I'm using the database-transactions skill to preserve DB, tenant, and Casbin consistency."

## When To Use
- multi-table writes
- create/update/delete with audit/webhook side effects
- membership/role/policy changes
- migration or seed changes
- querybuilder/list filter behavior

## Read Order
1. `llm/conventions/database.md`
2. `llm/cache/domain-rules.md`
3. `llm/workflows/database-migration.md`
4. target model/repository/usecase
5. `db/migrations` and `db/seeds/main.go` when schema/seed changes

## Workflow
### Phase 1 — Transaction Boundary
Identify all writes that must commit/rollback together:
- DB rows
- Casbin policies
- audit outbox
- webhook task enqueue
- Redis/cache invalidation

### Phase 2 — Implementation Rules
- Use repo transaction helper/pattern already present.
- Do not split dependent writes into independent commits.
- Use transactional enforcer patterns when policy and DB writes must align.
- Keep tenant constraints inside transaction path when relevant.

### Phase 3 — Migration Rules
- Add paired up/down SQL files only when schema change required.
- No destructive migration without explicit approval.
- Update models/repositories/tests consistently.

### Phase 4 — Verification
- Unit tests for transaction rollback when pattern exists.
- Integration tests for DB/Redis/Casbin interactions.
- Migration up/down commands when environment supports it.

## Review Checklist
- [ ] rollback path understood
- [ ] policy state cannot diverge from DB state
- [ ] tenant/org cache invalidation preserved
- [ ] no sensitive querybuilder field exposure
''' + COMMON_FOOTER)

write('cross-stack-change', '''---
name: cross-stack-change
description: Use when a change spans Go backend plus `apps/web`, `apps/client`, frontend proxies, shared API types, or shared packages.
---

# Cross Stack Change: Backend Contract to Frontend Surfaces

**Announce at start:** "I'm using the cross-stack-change skill to keep backend contracts and both frontend apps aligned."

## Read Order
1. `AGENTS.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/frontend-map.md`
4. `llm/cache/backend-map.md`
5. `llm/workflows/cross-stack-change.md`
6. backend route/controller/usecase
7. `apps/web/src/app/api/v1/[...path]/route.ts`
8. `apps/client/app/routes/api-proxy.ts`
9. `packages/api-types` if shape changes

## Workflow
### Phase 1 — Backend Truth
- Establish contract from live route/controller/usecase.
- Identify auth, cookie, token, tenant header, and error response behavior.

### Phase 2 — Consumer Inventory
Check both active apps:
- `apps/web` Next.js App Router/server actions/proxy
- `apps/client` React Router/proxy/client helpers
- shared `packages/*`

### Phase 3 — Patch Order
1. backend contract
2. shared types/helpers
3. frontend proxy/client
4. feature UI state/error handling

### Phase 4 — Verify
- backend targeted tests
- frontend typecheck/build for touched app
- integration/E2E for auth/cookie/proxy/request lifecycle changes

## Red Flags
- updating only one active frontend app when both consume route
- frontend-only auth/tenant enforcement
- changing payload shape without shared types/proxy update
- treating `apps/client` lint as strong verification
''' + COMMON_FOOTER)

write('frontend-surface', '''---
name: frontend-surface
description: Use when deciding whether a frontend change belongs in `apps/web`, `apps/client`, or shared `packages/*`, or when auditing active frontend ownership.
---

# Frontend Surface: App Ownership Decision

**Announce at start:** "I'm using the frontend-surface skill to choose the correct active frontend surface."

## Read Order
1. `llm/cache/frontend-map.md`
2. `llm/cache/api-contracts.md`
3. `llm/conventions/typescript.md`
4. target app route registry/component tree

## Surface Map
- `apps/web`: Next.js App Router, server components/actions, backend API proxy.
- `apps/client`: React Router, feature folders, route registry, API proxy route.
- `packages/api-types`: shared contract/types.
- `packages/ui`: shared UI primitives.
- `packages/hooks`: shared hooks.
- `packages/utils`: shared utilities.

## Workflow
### Phase 1 — Ownership
- Identify URL/route/component owner.
- Check if both apps expose same domain feature.
- Check shared package candidates before duplicating code.

### Phase 2 — API Boundary
- If backend contract changes, load `cross-stack-change`.
- If auth/tenant route behavior changes, load `auth-tenant-casbin`.

### Phase 3 — UI Patch
- Preserve loading, empty, error, success, auth-expired, and tenant-switch states.
- Prefer existing UI components and variants.
- Avoid new shared abstraction until reused by both apps or clearly stable.

### Phase 4 — Verify
- `apps/web`: lint/typecheck/build as relevant.
- `apps/client`: typecheck/build/E2E as relevant; lint script is placeholder-only.
''' + COMMON_FOOTER)

write('query-builder-security', '''---
name: query-builder-security
description: Use when changing filtering, sorting, dynamic query fields, repository list endpoints, or `pkg/querybuilder` behavior.
---

# Query Builder Security: Dynamic Field Safety

**Announce at start:** "I'm using the query-builder-security skill because filtering/sorting is part of this repo's security model."

## Read Order
1. `llm/cache/domain-rules.md`
2. `llm/conventions/database.md`
3. `pkg/querybuilder`
4. affected repository/list endpoint
5. existing querybuilder tests

## Iron Rules
- Field names must come from whitelisted struct metadata, not raw user input.
- Sensitive fields stay denied: password, token, secret, key, salt, and equivalent secrets.
- Values use GORM placeholders.
- Sorting/filtering convenience must not weaken security.

## Workflow
### Phase 1 — Identify Input Surface
- query params
- filter object
- sort field/order
- repository list method

### Phase 2 — Validate Field Semantics
- Confirm allowlist/denylist behavior.
- Confirm struct tags/field metadata are intended.
- Confirm sensitive fields remain unqueryable.

### Phase 3 — Patch and Test
- Add tests for allowed field.
- Add tests for denied sensitive field.
- Add tests for invalid/unknown field.
- Add affected repository/list endpoint tests when possible.
''' + COMMON_FOOTER)

write('tus-upload-storage', '''---
name: tus-upload-storage
description: Use when changing TUS upload handling, upload metadata, storage provider behavior, hook dispatch, avatar updates, or upload completion flows.
---

# TUS Upload Storage: Upload Lifecycle Boundary

**Announce at start:** "I'm using the tus-upload-storage skill because upload metadata and hooks are high-risk boundaries."

## Read Order
1. `llm/cache/domain-rules.md`
2. `llm/cache/backend-map.md`
3. `internal/config/app.go`
4. `pkg/tus`
5. `pkg/storage`
6. upload route in `internal/router/router.go`
7. upload/storage tests

## Workflow
### Phase 1 — Trace Upload Route
- TUS route is separate from normal JSON CRUD.
- Confirm auth/status middleware and upload handler wiring.
- Confirm storage provider: local/S3-compatible config.

### Phase 2 — Metadata Boundary
- Identify upload type metadata.
- Validate completion hook routing.
- Confirm avatar/user profile update path if relevant.

### Phase 3 — Patch Rules
- Preserve context propagation into storage operations.
- Do not trust upload metadata without existing validation pattern.
- Keep hook side effects intentional and idempotent where possible.

### Phase 4 — Verify
- storage provider tests
- TUS upload tests
- integration checks when route lifecycle or hooks changed
''' + COMMON_FOOTER)

write('worker-audit-webhook', '''---
name: worker-audit-webhook
description: Use when changing Asynq worker tasks, audit outbox/sync behavior, webhook dispatch, email jobs, cleanup jobs, scheduler behavior, or request side effects that enqueue background work.
---

# Worker Audit Webhook: Async Side Effects

**Announce at start:** "I'm using the worker-audit-webhook skill to preserve async side-effect semantics."

## Read Order
1. `llm/cache/backend-map.md`
2. `llm/cache/domain-rules.md`
3. `internal/worker/tasks`
4. `internal/worker/distributor.go`
5. `internal/worker/processor.go`
6. `internal/worker/handlers/*`
7. `internal/modules/audit`
8. `internal/modules/webhook`

## Workflow
### Phase 1 — Trace Task Lifecycle
`usecase -> distributor -> task payload -> processor registration -> handler -> side effect`.

### Phase 2 — Semantics
Decide:
- sync vs async guarantee
- retry behavior
- idempotency
- transaction coupling
- audit/webhook visibility to caller

### Phase 3 — Patch Rules
- Do not silently convert sync behavior to async or async behavior to sync.
- Keep task payload version/shape compatible with handlers.
- Keep audit/webhook consistency with primary request transaction behavior.

### Phase 4 — Verify
- unit tests for task payload/handler logic
- integration tests where request response depends on async side effects
- scheduler tests where timing/cleanup changes
''' + COMMON_FOOTER)

write('realtime-sse-websocket', '''---
name: realtime-sse-websocket
description: Use when changing SSE, WebSocket ticket flow, origin checks, Redis presence, distributed realtime behavior, or frontend realtime consumers.
---

# Realtime SSE WebSocket: Ticket and Presence Boundary

**Announce at start:** "I'm using the realtime-sse-websocket skill because realtime auth and origin checks are security-sensitive."

## Read Order
1. `llm/cache/domain-rules.md`
2. `llm/cache/backend-map.md`
3. `internal/router/router.go`
4. `pkg/sse`
5. `pkg/ws`
6. relevant frontend consumers
7. realtime tests

## Boundary Map
- SSE route requires auth token.
- WebSocket route uses short-lived Redis ticket flow.
- Presence/distributed behavior uses Redis-backed managers.
- Origin validation is security-sensitive.

## Workflow
### Phase 1 — Trace Connection
- ticket issuance
- ticket storage/expiry
- WS upgrade route
- origin check
- presence registration
- message lifecycle

### Phase 2 — Patch Rules
- Do not accept raw access token at WS route unless live code intentionally supports it.
- Preserve ticket one-time/expiry semantics if present.
- Preserve distributed Redis behavior.

### Phase 3 — Verify
- unit tests for ticket validation/origin behavior if present
- integration/realtime tests where available
- frontend smoke flow if UI depends on realtime updates
''' + COMMON_FOOTER)

write('verification-before-completion', '''---
name: verification-before-completion
description: Use before claiming any Casbin repo task is complete, fixed, verified, or safe.
---

# Verification Before Completion

**Announce at start:** "I'm using verification-before-completion; no evidence, no success claim."

## Iron Law
No final success claim without exact verification evidence or exact blocker.

## Gate Function
Before final response, answer:
1. What files changed?
2. Which layers changed: backend, frontend, DB, worker, upload, realtime, docs, skills?
3. What is narrowest meaningful check?
4. Did contract producer and consumer both get checked?
5. What could not run and why?

## Command Matrix
- Go backend: targeted `go test ./path/...` or `pnpm go:test`.
- Integration: `pnpm go:test-integration` for DB/Redis/Casbin/worker/tenant/upload.
- E2E: `pnpm go:test-e2e` for route lifecycle/cookies/tokens/proxies.
- Frontend web: `pnpm --filter casbin-web typecheck`, lint/build as relevant.
- Frontend client: `pnpm --filter casbin-client typecheck`, build/E2E as relevant; lint placeholder is not strong verification.
- Docs/API: `pnpm go:docs` when Swagger changes.

## Red Flags — Stop
- claiming pass after only reading code
- hiding Docker/Snap/network blockers
- treating unrelated failing tests as fixed
- skipping frontend consumer checks after API contract changes
- skipping integration when tenant/Casbin/API-key behavior changed
''' + COMMON_FOOTER)

write('feature', '''---
name: feature
description: End-to-end feature development orchestrator for the Casbin Go plus TypeScript monorepo. Use when building new backend/frontend/cross-stack capability from scratch.
---

# Feature: Casbin End-to-End Development Orchestrator

**Announce at start:** "I'm using the feature skill to orchestrate research, design, plan, implementation, and verification."

## Pipeline
`research -> brainstorm -> plan -> execute -> verification-before-completion -> self-review`

## When To Use
| Use this skill when... | Use individual skills when... |
|---|---|
| feature spans multiple files or layers | change is one small bugfix |
| requirements need design before code | design/plan already exists |
| backend/frontend/API contract may change | pure module internals -> `go-service` |

## Phase 1 — Research
- Read `AGENTS.md` and relevant `llm/cache/*` first.
- Inspect live code before trusting docs.
- Save durable investigation to `llm/research/[feature].md` only if useful.

## Phase 2 — Brainstorm
- Identify module owner and high-risk boundaries.
- Compare 2-3 approaches.
- Choose approach with smallest safe boundary.

## Phase 3 — Plan
- Write staged plan to `llm/tasks/todo.md` for active work.
- Use `llm/plans/` for larger durable roadmap.
- Include verification per stage.

## Phase 4 — Execute
- Backend-first when data/contract changes.
- Use project-specific boundary skill when auth/tenant/Casbin/API-key/upload/worker/query/realtime touched.
- Keep changes small and reviewable.

## Phase 5 — Verify and Review
- Run `verification-before-completion`.
- Run `self-review`.
- Report exact commands and skipped checks.

## Artifacts
- active task state: `llm/tasks/todo.md`
- research: `llm/research/`
- non-urgent recommendations: `llm/recommendations/`
- durable staged plan: `llm/plans/`
''')

write('execute', '''---
name: execute
description: Execute an approved Casbin implementation plan task-by-task with progress tracking and verification. Use after a plan exists.
---

# Execute: Plan Implementation

**Announce at start:** "I'm using the execute skill to implement the approved plan task by task."

## Prerequisites
- Implementation plan exists in `llm/tasks/todo.md` or `llm/plans/*`.
- Route/module ownership is clear.
- High-risk skill selected if needed.

## Workflow
### Step 1 — Load Plan
- Read active plan.
- Check dependencies and blockers.
- Confirm changed layers.

### Step 2 — Select Skills
Use exactly relevant skills:
- backend logic: `go-service`
- route/API: `api-endpoint`
- auth/tenant/Casbin: `auth-tenant-casbin`
- API key: `api-key-scope`
- DB transaction/migration: `database-transactions`
- upload: `tus-upload-storage`
- worker/audit/webhook: `worker-audit-webhook`
- querybuilder: `query-builder-security`
- realtime: `realtime-sse-websocket`
- frontend ownership: `frontend-surface`

### Step 3 — Implement Task
- Do one coherent slice at a time.
- Do not mix unrelated cleanup.
- Update `llm/tasks/todo.md` for multi-step progress.

### Step 4 — Verify Task
- Run narrow verification for changed layer.
- Record exact failures/blockers.
- Stop and re-plan if root assumption is wrong.

### Step 5 — Complete
- Run `self-review`.
- Run `verification-before-completion`.
- Report final file list and verification.
''')

write('plan', '''---
name: plan
description: Use when work needs a staged implementation plan, scoped checklist, or approval gate before editing the Casbin repo.
---

# Plan: Casbin Implementation Planning

**Announce at start:** "I'm using the plan skill to create a staged, verifiable implementation plan."

## Read Order
1. `AGENTS.md`
2. relevant `llm/cache/*`
3. relevant `llm/workflows/*`
4. live code owner paths

## Plan Format
Every plan must include:
- scope and non-scope
- files likely touched
- high-risk boundaries
- task order
- verification per task
- stop conditions

## Task Granularity
Good task:
- 1 coherent behavior change
- clear files
- clear verification
- can be reviewed independently

Bad task:
- "fix backend"
- "update UI"
- "clean up things"

## Where To Save
- active work: `llm/tasks/todo.md`
- durable staged plan: `llm/plans/`
- future improvement: `llm/recommendations/`
''')

write('brainstorm', '''---
name: brainstorm
description: Use when exploring approaches for a Casbin feature, refactor, bug strategy, or architectural decision before planning or implementation.
---

# Brainstorm: Evidence-Based Design Options

**Announce at start:** "I'm using the brainstorm skill to compare options before choosing an implementation path."

## Read First
- relevant `llm/cache/*`
- relevant `llm/conventions/*`
- closest live code precedent

## Workflow
### Phase 1 — Frame Problem
- State goal.
- State constraints.
- State known high-risk boundaries.

### Phase 2 — Options
Produce 2-3 options. For each:
- files touched
- benefits
- risks
- verification needed
- impact on auth/tenant/Casbin/API-key/contracts if any

### Phase 3 — Recommendation
Pick one option based on:
- smallest safe change
- alignment with live wiring
- testability
- low blast radius

## Output
- recommended approach
- rejected alternatives and why
- questions/blockers if any
''')

write('self-review', '''---
name: self-review
description: Use when Casbin repo changes are complete and need maintainability, correctness, security-boundary, and verification review before handoff.
---

# Self Review: Staff-Level Handoff Check

**Announce at start:** "I'm using self-review before handoff."

## Review Dimensions
### Layering
- controller only parses/validates/responds
- usecase owns business rules
- repository owns persistence details
- app config/router own wiring/protection

### Security Boundaries
- Redis session validation preserved
- tenant org boundary preserved
- Casbin policy/enforcer semantics preserved
- API-key scopes preserved
- querybuilder sensitive fields protected

### Side Effects
- worker/audit/webhook behavior intentional
- upload hooks intentional
- realtime ticket/origin/presence behavior intentional

### Contract
- backend response shape matches frontend consumers
- both `apps/web` and `apps/client` checked when relevant
- Swagger/shared types updated when required

### Verification
- exact commands run
- skipped checks explained
- unrelated failures separated

## Output Format
- Problems found and fixed
- Remaining risks
- Verification summary
''')

write('systematic-debugging', '''---
name: systematic-debugging
description: Use when encountering a bug, failing test, suspicious behavior, integration failure, auth/tenant issue, or runtime mismatch in the Casbin repo.
---

# Systematic Debugging: Root Cause Before Fix

**Announce at start:** "I'm using systematic-debugging; root cause before patch."

## Iron Rule
No fix before reproducing or tracing the exact failing path.

## Four Phases
### Phase 1 — Evidence
- capture exact error/test/log/request
- identify changed layer
- read relevant cache and workflow

### Phase 2 — Trace
Trace through live code:
- router/middleware for HTTP behavior
- controller/usecase/repository for module behavior
- worker/distributor/handler for async behavior
- proxy/client/component for frontend behavior

### Phase 3 — Hypothesis
Write 1-3 hypotheses and expected evidence.

### Phase 4 — Patch
- patch minimal root cause
- add regression test if nearby pattern exists
- verify narrow path first

## Stop Signals
- cannot reproduce or trace
- multiple possible root causes remain
- environment blocker hides result
- fix requires product/security decision
''')

write('test-driven-development', '''---
name: test-driven-development
description: Use when implementing a feature or bugfix with a meaningful test seam in the Casbin backend/frontend.
---

# Test Driven Development: Casbin Test-First Flow

**Announce at start:** "I'm using TDD where this repo has a meaningful test seam."

## Repo Test Targets
- Go usecase/repository/middleware: package-local tests
- route/auth/tenant/Casbin/API-key: integration/E2E when needed
- querybuilder: `pkg/querybuilder` tests
- upload/storage: `pkg/tus`, `pkg/storage`, integration tests
- frontend: app typecheck/build and existing UI/E2E patterns

## Red-Green-Refactor
### RED
- Write minimal failing test for desired behavior.
- Confirm failure is expected, not compile/setup error.

### GREEN
- Implement minimal code to pass.
- Do not add extra behavior.

### REFACTOR
- Clean names/duplication after green.
- Keep tests green.

## Exceptions
If no practical test seam exists, document:
- why no test was added
- exact code trace/manual verification used instead
''')

write('forms', '''---
name: forms
description: Use when building or changing forms, validation, payload mapping, error rendering, or submit flows across `apps/web`, `apps/client`, and Go backend validation.
---

# Forms: Frontend Validation and Backend Contract

**Announce at start:** "I'm using the forms skill to align UI form state with backend validation and contract behavior."

## Read Order
1. `llm/cache/frontend-map.md`
2. `llm/cache/api-contracts.md`
3. `llm/conventions/typescript.md`
4. target backend request struct/validation tags
5. closest frontend form precedent

## Workflow
### Phase 1 — Field Ownership
Classify each field:
- UI-only
- API payload
- persisted model
- derived/display-only

### Phase 2 — Validation
- Match frontend validation with backend validation tags and usecase rules.
- Preserve server error display.
- Handle auth/session expiry and tenant errors.

### Phase 3 — States
Cover:
- loading
- empty/default
- dirty/disabled
- validation error
- submit error
- success

### Phase 4 — Verify
- submitted payload shape
- backend response/error shape
- affected app typecheck/build
- browser/E2E flow when user-critical
''')

print('detailed generic Casbin skills appended')
