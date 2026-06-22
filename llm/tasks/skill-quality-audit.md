# Skill Quality Audit — Casbin Repo

## Scope

Audit `.agents/skills/*/SKILL.md` agar sesuai kebutuhan spesifik repo `Casbin`.

## Final result

Status: pass.

Skill set sekarang sudah cukup spesifik untuk kebutuhan repo ini: Go backend, tenant/Casbin boundary, API-key scope, TUS upload, worker/audit/webhook, query-builder security, realtime SSE/WebSocket, dan dua frontend surface aktif.

## Verified project-specific skills

- `.agents/skills/auth-tenant-casbin/SKILL.md`
- `.agents/skills/api-key-scope/SKILL.md`
- `.agents/skills/tus-upload-storage/SKILL.md`
- `.agents/skills/worker-audit-webhook/SKILL.md`
- `.agents/skills/query-builder-security/SKILL.md`
- `.agents/skills/realtime-sse-websocket/SKILL.md`
- `.agents/skills/frontend-surface/SKILL.md`

## Tightened core skills

- `.agents/skills/api-endpoint/SKILL.md` now covers route strata, API-key scope, Swagger, and frontend consumers.
- `.agents/skills/go-service/SKILL.md` now covers `internal/config/app.go`, constructors, usecase/repository split, context, and tenant/Casbin rules.
- `.agents/skills/database-transactions/SKILL.md` now covers GORM, transactional enforcer behavior, tenant semantics, and migration pairing.
- `.agents/skills/cross-stack-change/SKILL.md` now covers both active proxy surfaces and `apps/client` lint caveat.
- `.agents/skills/verification-before-completion/SKILL.md` now covers backend/frontend/DB/worker/upload/realtime mapping and Docker/Snap blocker reporting.
- `.agents/skills/ui/SKILL.md` now covers active frontend ownership and Vercel React skill routing.

## Stale/foreign assumption audit

Search results:

- No Carbon, Supabase, Kysely, Biome, placeholder, or TODO assumptions found in project-specific skills.
- Vercel skill references Prisma/Drizzle in `.agents/skills/vercel-react-best-practices/*` as generic frontend performance examples. This is acceptable only because those files are third-party frontend guidance and must not be used as backend database guidance for this repo.

## Usage routing recommendation

Use high-risk repo skills first when task touches these areas:

- auth/session/tenant/Casbin: `auth-tenant-casbin`
- API-key scopes: `api-key-scope`
- upload/storage: `tus-upload-storage`
- worker/audit/webhook: `worker-audit-webhook`
- dynamic filtering/sorting: `query-builder-security`
- SSE/WebSocket/presence: `realtime-sse-websocket`
- frontend app ownership: `frontend-surface`

Use generic starter-pack skills for task shape only after the repo-specific boundary is selected.
