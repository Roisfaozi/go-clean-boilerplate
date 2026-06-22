from pathlib import Path

root = Path(__file__).resolve().parents[2]
skills_dir = root / '.agents' / 'skills'

def write_skill(name: str, description: str, read_first: list[str], workflow: list[str], watch: list[str] | None = None):
    path = skills_dir / name / 'SKILL.md'
    path.parent.mkdir(parents=True, exist_ok=True)
    title = ' '.join(part.capitalize() for part in name.replace('-', ' ').split())
    read = ''.join(f'- `{item}`\n' if item.endswith('.md') or '/' in item else f'- {item}\n' for item in read_first)
    steps = ''.join(f'{idx}. {item}\n' for idx, item in enumerate(workflow, 1))
    watch_block = ''
    if watch:
        watch_block = '\n## Watch For\n' + ''.join(f'- {item}\n' for item in watch)
    path.write_text(f'''---\nname: {name}\ndescription: {description}\n---\n\n# {title}\n\n## Read First\n{read}\n## Workflow\n{steps}{watch_block}''')

# Tightened starter-pack skills
write_skill(
    'api-endpoint',
    'Use when adding or changing backend HTTP endpoints, route protection, API-key scope behavior, or frontend-consumed contracts in the Casbin monorepo.',
    ['AGENTS.md', 'llm/cache/api-contracts.md', 'llm/cache/backend-map.md', 'llm/cache/domain-rules.md', 'llm/workflows/api-endpoint.md', 'internal/router/router.go'],
    [
        'Choose route stratum intentionally: public, authenticated, tenantAuthorized, authorized, or upload.',
        'Confirm API-key middleware and explicit/auto scope behavior before registering protected routes.',
        'Implement controller -> usecase -> repository in owning `internal/modules/*` module.',
        'Update Swagger artifacts when public API contract changes.',
        'Audit `apps/web` and `apps/client` proxies/clients when frontend can consume the route.',
        'Run narrow backend verification first; use integration/E2E when auth, tenant, cookie, or Casbin behavior changes.',
    ],
    ['Do not move route protection into frontend-only checks.', 'Do not bypass Redis-backed session validation.', 'Do not weaken tenant/Casbin/API-key layering.'],
)

write_skill(
    'go-service',
    'Use when changing Go backend module logic, usecases, repositories, module constructors, or worker-owned backend behavior.',
    ['AGENTS.md', 'llm/cache/backend-map.md', 'llm/cache/module-map.md', 'llm/cache/domain-rules.md', 'llm/conventions/golang.md', 'llm/workflows/go-service.md', 'internal/config/app.go'],
    [
        'Start at `internal/config/app.go` for dependency wiring when lifecycle or constructor changes.',
        'Keep business rules in usecase layer and persistence in repository layer.',
        'Pass only required values/dependencies into usecases; do not pass full app config.',
        'Preserve context propagation, especially storage, request-scoped, worker, and transaction flows.',
        'Add/adjust tests near the owning package when patterns exist.',
    ],
    ['Casbin writes that share DB transaction semantics need transactional enforcer patterns.', 'Tenant reads must not become ad hoc query-param checks.'],
)

write_skill(
    'database-transactions',
    'Use when changing all-or-nothing GORM persistence, multi-table writes, Casbin policy writes, or tenant-sensitive repository behavior.',
    ['llm/conventions/database.md', 'llm/cache/domain-rules.md', 'llm/workflows/database-migration.md', 'internal/config/app.go'],
    [
        'Confirm transaction owner and failure semantics before editing.',
        'Keep all dependent DB writes and policy changes inside intended transaction boundary.',
        'Use transactional enforcer patterns for Casbin policy changes when DB semantics must match.',
        'Preserve tenant organization checks and membership cache invalidation behavior.',
        'Add paired up/down migrations when schema changes are required.',
    ],
    ['No destructive migration without explicit user request.', 'Do not weaken query-builder field restrictions.'],
)

write_skill(
    'cross-stack-change',
    'Use when a change spans Go backend plus `apps/web`, `apps/client`, or shared packages in this monorepo.',
    ['AGENTS.md', 'llm/cache/api-contracts.md', 'llm/cache/frontend-map.md', 'llm/cache/backend-map.md', 'llm/workflows/cross-stack-change.md'],
    [
        'Establish backend contract from route/controller/usecase first.',
        'Check both active frontend proxy surfaces: `apps/web/src/app/api/v1/[...path]/route.ts` and `apps/client/app/routes/api-proxy.ts`.',
        'Update shared type/client usage under `packages/*` if contract shape changes.',
        'Verify cookies, authorization headers, tenant headers, and error payload handling.',
        'Run backend and affected frontend checks; do not treat `apps/client` lint as strong verification.',
    ],
    ['Both frontend apps are active surfaces.', 'Frontend checks must not replace backend route/middleware authorization.'],
)

write_skill(
    'verification-before-completion',
    'Use before claiming a Casbin repo change is complete, fixed, or passing.',
    ['AGENTS.md', 'llm/conventions/testing.md', 'llm/tasks/todo.md'],
    [
        'Map changed files to backend, frontend, DB, worker, upload, realtime, or docs layers.',
        'Run the narrowest relevant check first.',
        'Use `pnpm go:test` or targeted Go tests for backend behavior; use integration/E2E for DB/Redis/Casbin/worker/tenant/upload boundaries.',
        'Use app-specific frontend typecheck/build for affected app; remember frontend lint runs Biome and typecheck remains a separate TypeScript gate.',
        'Report exact commands, pass/fail status, and skipped checks with blockers such as Docker or Snap Go.',
    ],
    ['No evidence, no success claim.', 'Do not hide unrelated failures.'],
)

write_skill(
    'ui',
    'Use when building, reviewing, or polishing frontend UI in `apps/web`, `apps/client`, or shared packages.',
    ['llm/cache/frontend-map.md', 'llm/conventions/typescript.md', '.agents/skills/vercel-react-best-practices/SKILL.md', '.agents/skills/vercel-composition-patterns/SKILL.md'],
    [
        'Confirm owning app before editing: `apps/web` Next.js App Router or `apps/client` React Router.',
        'Reuse `packages/ui`, shared hooks, and existing app components before adding new primitives.',
        'Check loading, empty, error, success, auth-expired, and tenant-switch states when relevant.',
        'Apply Vercel React/performance skills when React component architecture is involved.',
        'Verify with typecheck/build or browser flow appropriate to the touched app.',
    ],
    ['Do not duplicate API client/proxy logic across apps.', 'Do not treat frontend-only visibility as authorization.'],
)

# Project-specific high-risk skills
write_skill(
    'auth-tenant-casbin',
    'Use when changing auth, Redis-backed sessions, organization tenant resolution, Casbin policy/enforcement, or protected route middleware.',
    ['AGENTS.md', 'llm/cache/domain-rules.md', 'llm/cache/backend-map.md', 'internal/router/router.go', 'internal/middleware/'],
    [
        'Trace request from router group through API-key/JWT/session/status/tenant/Casbin middleware.',
        'Verify JWT parsing is not treated as enough without Redis session validation.',
        'Confirm organization context and membership/cache behavior before usecase changes.',
        'Use permission/Casbin abstractions for policy writes; preserve transactional semantics.',
        'Run tenant/auth/Casbin focused tests or integration checks when behavior changes.',
    ],
    ['This is high-risk. Stop if route stratum or ownership is unclear.', 'Never bypass tenant boundary or API-key scope checks.'],
)

write_skill(
    'api-key-scope',
    'Use when adding or changing protected endpoints, API-key authentication, API-key scopes, or organization-scoped API-key behavior.',
    ['llm/cache/domain-rules.md', 'llm/workflows/api-endpoint.md', 'internal/middleware/api_key_middleware.go', 'internal/modules/api_key'],
    [
        'Identify if route uses authenticated, tenantAuthorized, or authorized group.',
        'Confirm auto scope vs explicit scope behavior before route registration.',
        'Preserve organization scope checks and Redis-backed API-key behavior.',
        'Update tests around API key lifecycle or route access if behavior changes.',
    ],
    ['Identity alone is not scope authorization.', 'Do not add protected endpoints without scope decision.'],
)

write_skill(
    'tus-upload-storage',
    'Use when changing TUS upload handling, upload metadata, storage providers, hook dispatch, or avatar/upload completion behavior.',
    ['llm/cache/domain-rules.md', 'llm/cache/backend-map.md', 'pkg/tus', 'pkg/storage', 'internal/config/app.go'],
    [
        'Trace TUS route separately from normal CRUD route groups.',
        'Preserve auth/status middleware and upload metadata validation.',
        'Check hook dispatch path for upload completion type.',
        'Keep storage context propagation and local/S3 provider behavior intact.',
        'Run upload/storage tests or integration checks when upload lifecycle changes.',
    ],
    ['Do not treat upload route as normal JSON CRUD.', 'Metadata trust boundary is security-sensitive.'],
)

write_skill(
    'worker-audit-webhook',
    'Use when changing Asynq worker tasks, audit outbox/sync behavior, webhook dispatch, email jobs, cleanup jobs, or scheduler behavior.',
    ['llm/cache/backend-map.md', 'llm/cache/domain-rules.md', 'internal/worker', 'internal/modules/audit', 'internal/modules/webhook'],
    [
        'Trace distributor method, task payload, processor registration, and handler side effects.',
        'Check retry/idempotency assumptions before changing handler writes.',
        'Preserve audit/webhook consistency with primary request transaction behavior.',
        'Use integration tests when request response depends on async side effects.',
    ],
    ['Worker side effects can affect audit, webhook, email, and cleanup behavior.', 'Do not silently make async behavior synchronous or vice versa.'],
)

write_skill(
    'query-builder-security',
    'Use when changing filtering, sorting, dynamic query fields, repository list endpoints, or `pkg/querybuilder` behavior.',
    ['llm/cache/domain-rules.md', 'pkg/querybuilder', 'llm/conventions/database.md'],
    [
        'Confirm fields come from whitelisted struct metadata, not raw user-controlled names.',
        'Keep sensitive field deny rules for password/token/secret/key/salt style fields.',
        'Use GORM placeholders for values.',
        'Run querybuilder tests and affected repository/list endpoint tests.',
    ],
    ['Field filtering/sorting is part of security model.', 'Do not weaken allow/deny behavior for convenience.'],
)

write_skill(
    'realtime-sse-websocket',
    'Use when changing SSE, WebSocket ticket flow, origin checks, Redis presence, or distributed realtime behavior.',
    ['llm/cache/domain-rules.md', 'llm/cache/backend-map.md', 'pkg/ws', 'pkg/sse', 'internal/router/router.go'],
    [
        'Distinguish SSE auth token path from WebSocket short-lived Redis ticket path.',
        'Check origin validation and ticket consumption behavior.',
        'Preserve presence manager and distributed Redis behavior.',
        'Run realtime tests or focused integration checks when flow changes.',
    ],
    ['WebSocket origin/ticket/presence is high-risk.', 'Do not accept raw access token at WS route unless live code already supports it intentionally.'],
)

write_skill(
    'frontend-surface',
    'Use when deciding whether a frontend change belongs in `apps/web`, `apps/client`, or shared `packages/*`.',
    ['llm/cache/frontend-map.md', 'llm/cache/api-contracts.md', 'llm/conventions/typescript.md'],
    [
        'Identify owning app route/component before editing.',
        'Check if shared code belongs under `packages/ui`, `packages/hooks`, `packages/utils`, or `packages/api-types`.',
        'Audit both app proxies if backend contract changes.',
        'Use app-specific verification; do not rely on `apps/client` lint placeholder.',
    ],
    ['Both apps are active.', 'Avoid duplicating API helpers or UI primitives without reason.'],
)

print('project-specific skill quality patch complete')
