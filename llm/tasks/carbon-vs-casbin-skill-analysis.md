# Carbon vs Casbin — Skill and AI Workflow Analysis

## Scope

Analisa konfigurasi AI-native Carbon terhadap Casbin setelah detailed Casbin skills, granular cache, dan module-level cache diterapkan.

## Final status

Status: complete for repo documentation and generator scripts.

Casbin now follows the Carbon operational pattern while remaining project-specific to this Go/Gin/GORM/Casbin monorepo.

## Carbon pattern matched

Casbin docs/skills now cover the Carbon-style structure:

- skill activation line (`Announce at start`)
- concrete read order
- skill boundary routing
- phased workflows
- project-specific paths and modules
- stop conditions / red flags
- completion output expectations
- verification command routing
- durable cache and task artifact locations

## Cache maturity result

Casbin now has architecture-level and module/domain-level cache files.

### Core cache

- `llm/cache/project-overview.md`
- `llm/cache/environment.md`
- `llm/cache/architecture.md`
- `llm/cache/backend-map.md`
- `llm/cache/frontend-map.md`
- `llm/cache/api-contracts.md`
- `llm/cache/module-map.md`
- `llm/cache/domain-rules.md`

### High-risk domain cache

- `llm/cache/authentication-system.md`
- `llm/cache/tenant-organization-system.md`
- `llm/cache/casbin-permission-system.md`
- `llm/cache/api-key-system.md`
- `llm/cache/tus-upload-system.md`
- `llm/cache/worker-audit-webhook-system.md`
- `llm/cache/querybuilder-security.md`
- `llm/cache/realtime-system.md`
- `llm/cache/frontend-proxy-system.md`

### Module cache

- `llm/cache/user-system.md`
- `llm/cache/role-system.md`
- `llm/cache/project-system.md`
- `llm/cache/access-right-system.md`
- `llm/cache/stats-system.md`

## Skill maturity result

Detailed skill generator now includes Carbon-style orchestration and Casbin-specific module skills:

- core orchestration: `feature`, `execute`, `plan`, `brainstorm`, `systematic-debugging`, `verification-before-completion`, `self-review`, `test-driven-development`, `forms`
- high-risk boundaries: `auth-tenant-casbin`, `api-key-scope`, `tus-upload-storage`, `worker-audit-webhook`, `query-builder-security`, `realtime-sse-websocket`, `frontend-surface`
- backend/API foundations: `api-endpoint`, `go-service`, `database-transactions`, `cross-stack-change`
- module-specific skills: `user-domain`, `role-domain`, `project-domain`, `access-domain`, `stats-domain`

The manual regeneration script is:

- `llm/tasks/write-detailed-casbin-skills.py`

## Project-specific evidence encoded

Docs are anchored to real paths:

- `internal/config/app.go`
- `internal/router/router.go`
- `internal/middleware/auth_middleware.go`
- `internal/middleware/tenant_middleware.go`
- `internal/middleware/casbin_middleware.go`
- `internal/middleware/api_key_middleware.go`
- `internal/modules/user`
- `internal/modules/role`
- `internal/modules/project`
- `internal/modules/access`
- `internal/modules/stats`
- `internal/modules/permission`
- `internal/modules/api_key`
- `internal/modules/audit`
- `internal/modules/webhook`
- `internal/worker`
- `pkg/tus`
- `pkg/querybuilder`
- `pkg/ws`
- `pkg/sse`
- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`

## Closed gaps

- Cache granularity gap: closed for core, high-risk, and key module domains.
- Skill specificity gap: closed through detailed generator and applied skill docs.
- Forms gap: closed at repo-specific pattern level with web React Hook Form/server action and client CRUD dialog/zod references.
- Verification gap: closed at command-matrix and stop-condition level.
- Wrong-stack risk: checked; no Carbon/Supabase/Kysely/Biome/ERP/MES/Prisma/Drizzle assumptions in new Casbin-specific docs.

## Remaining optional improvement

Only optional future work remains:

- Add more module caches if new modules become high-change areas.
- Add screenshots/manual browser playbooks under `llm/test-playbooks/` after real UI/E2E runs.
- Add more concrete examples to individual skill files if recurring agent mistakes appear.

No blocking gap remains for current AI-native workflow parity except manual `.agents/skills` regeneration outside this read-only sandbox.

## Latest module-level closure

Added verified module caches:

- `llm/cache/permission-system.md`
- `llm/cache/audit-system.md`
- `llm/cache/webhook-system.md`

Added generator-backed module skills:

- `permission-domain`
- `audit-domain`
- `webhook-domain`

These are Casbin-specific and anchored to live module paths.
