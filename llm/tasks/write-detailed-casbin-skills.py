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

write('improve', '''---
name: improve
description: Use when making reviewer-facing improvements, cleanup, hardening, or maintainability upgrades in the Casbin repo without changing product intent.
---

# Improve: Casbin Hardening and Cleanup

**Announce at start:** "I'm using the improve skill to make a targeted, reviewer-facing improvement."

## Read Order
1. `AGENTS.md`
2. relevant `llm/cache/*`
3. relevant `llm/conventions/*`
4. closest live code precedent

## When To Use
| Use this skill when... | Use another skill when... |
|---|---|
| cleaning up one module or boundary | change is new feature -> `feature` |
| hardening security or maintainability | root cause bug -> `systematic-debugging` |
| improving docs or agent context | API/route change -> `api-endpoint` |

## Workflow
### Phase 1 — Pain Statement
State the exact pain, where it lives, and why it matters.

### Phase 2 — Boundary Check
Identify module, route group, or app surface owning the pain.

### Phase 3 — Improvement
Make smallest useful improvement with minimal blast radius.

### Phase 4 — Verification
Run the narrowest relevant check.

## Output
- what improved
- why it was worth doing
- any follow-up ideas to save into `llm/recommendations/`
''' + COMMON_FOOTER)

write('e2e-test', '''---
name: e2e-test
description: Use when testing a feature flow end-to-end through browser or full request lifecycle in the Casbin repo.
---

# E2E Test: Casbin Flow Verification

**Announce at start:** "I'm using the e2e-test skill to verify the user flow end-to-end."

## Read Order
1. changed diff
2. target workflow cache
3. `llm/cache/frontend-map.md`
4. `llm/cache/api-contracts.md`
5. `llm/cache/frontend-proxy-system.md`
6. `llm/cache/authentication-system.md` if auth is involved

## Workflow
### Phase 1 — Select Flow
Choose the exact user journey and owning app.

### Phase 2 — Setup
Use login/setup pattern if auth is required.

### Phase 3 — Drive Flow
Exercise the actual browser/request lifecycle.

### Phase 4 — Evidence
Capture failure evidence, console/server errors, or screenshots when useful.

### Phase 5 — Save
If stable and reusable, save the steps to `llm/test-playbooks/`.

## Stop Conditions
- stop if backend/app is not running
- stop if the flow needs a data seed or fixture you cannot verify
- stop if auth/tenant boundary is ambiguous
''' + COMMON_FOOTER)

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
- user/role/project/access/stats module work: matching module-domain skill

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

## Output
- file list
- commands run
- verification result
- unresolved blockers
''' + COMMON_FOOTER)

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

## Plan Shape
Every plan must include:
- scope and non-scope
- file list
- route/module owner
- risk boundaries
- stage order
- verification per stage
- stop conditions

## Task Granularity
Each task should be:
- one coherent behavior change
- independently verifiable
- small enough to review

## Writing Rules
- Do not write vague steps like "fix backend".
- Prefer exact file paths and exact commands.
- Add review notes or blocker notes when uncertain.
- Save active state to `llm/tasks/todo.md`.

## Where To Save
- active work: `llm/tasks/todo.md`
- durable staged plan: `llm/plans/`
- future improvement: `llm/recommendations/`
''' + COMMON_FOOTER)

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
''' + COMMON_FOOTER)

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
''' + COMMON_FOOTER)

write('verification-before-completion', '''---
name: verification-before-completion
description: Use when about to claim work is complete, fixed, or passing in the Casbin repo.
---

# Verification Before Completion

**Announce at start:** "I'm using verification-before-completion; no evidence, no success claim."

## Iron Law
Do not claim success without exact verification evidence or exact blocker.

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

## Red Flags
- claiming pass after only reading code
- hiding Docker/Snap/network blockers
- treating unrelated failing tests as fixed
- skipping frontend consumer checks after API contract changes
- skipping integration when tenant/Casbin/API-key behavior changed
''' + COMMON_FOOTER)

write('forms', '''---
name: forms
description: Use when building or changing forms, validation, payload mapping, error rendering, or submit flows across `apps/web`, `apps/client`, and Go backend validation.
---

# Forms: Frontend Validation and Backend Contract

**Announce at start:** "I'm using the forms skill to align UI form state with backend validation and contract behavior."

## Read Order
1. `llm/cache/frontend-map.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/frontend-proxy-system.md`
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

## Project examples
- `apps/web/src/components/auth/login-form.tsx`
- `apps/web/src/app/actions/auth.ts`
- `apps/client/app/features/shared/crud-form-dialog.tsx`
''' + COMMON_FOOTER)

for name, body in {k:v for k,v in locals().items() if k in ['write']}.items():
    pass

print('generator updated')

write('permission-domain', '''---
name: permission-domain
description: Use when changing permission policy CRUD, role/user assignment, access-right expansion, or transactional Casbin behavior in the Casbin repo.
---

# Permission Domain: Policy and Assignment

**Announce at start:** "I'm using the permission-domain skill to preserve policy and assignment semantics."

## Read Order
1. `llm/cache/permission-system.md`
2. `llm/cache/access-right-system.md`
3. `llm/cache/casbin-permission-system.md`
4. `internal/modules/permission/module.go`
5. `internal/modules/permission/usecase/permission_usecase.go`

## Workflow
### Phase 1 — Classify Change
- policy CRUD
- role/user assignment
- batch check
- access-right expansion
- transactional enforcer behavior

### Phase 2 — Patch
- preserve Casbin transaction semantics
- preserve access-right and inheritance behavior
- keep controller thin

### Phase 3 — Verify
- permission usecase tests
- security tests for Casbin errors/failure paths
''')

write('audit-domain', '''---
name: audit-domain
description: Use when changing audit logs, audit outbox, or request side effects that emit audit data in the Casbin repo.
---

# Audit Domain: Audit Logs and Outbox

**Announce at start:** "I'm using the audit-domain skill to preserve audit visibility and outbox behavior."

## Read Order
1. `llm/cache/audit-system.md`
2. `internal/modules/audit/module.go`
3. `internal/modules/audit/usecase/audit_usecase.go`
4. `internal/modules/audit/repository/audit_repository.go`

## Workflow
### Phase 1 — Identify Direction
- log listing/querying
- outbox sync
- write-following side effect

### Phase 2 — Patch
- preserve persistence behavior
- preserve any request/worker coupling

### Phase 3 — Verify
- audit controller/usecase/repository tests
''')

write('webhook-domain', '''---
name: webhook-domain
description: Use when changing webhook CRUD, subscription filters, trigger dispatch, or webhook logs in the Casbin repo.
---

# Webhook Domain: Trigger Dispatch and Logs

**Announce at start:** "I'm using the webhook-domain skill to preserve webhook CRUD, trigger, and async dispatch behavior."

## Read Order
1. `llm/cache/webhook-system.md`
2. `internal/modules/webhook/module.go`
3. `internal/modules/webhook/usecase/webhook_usecase.go`
4. `internal/worker/tasks/webhook.go`

## Workflow
### Phase 1 — Classify Change
- CRUD
- trigger event routing
- log retrieval
- worker payload behavior

### Phase 2 — Patch
- preserve async dispatch intent
- preserve org scope
- keep payload compatible with worker

### Phase 3 — Verify
- webhook usecase/controller tests
- worker webhook task tests if payload changes
''')
