# Carbon Skill Pattern Audit → Casbin Adaptation

## Source inspected

- `/home/user/Documents/Rois-Source/carbon/.claude/skills/*/SKILL.md`
- `/home/user/Documents/Rois-Source/carbon/AGENTS.md`
- selected Carbon `llm/cache/*` and `llm/workflows/*`

## Carbon detail pattern observed

Strong Carbon skills usually include:

1. **Announce line** — explicit skill activation sentence.
2. **When to use vs alternatives** — table mapping boundary against other skills.
3. **Prerequisites / read order** — concrete files and durable cache/workflow docs.
4. **Phase workflow** — recon, contract/design, implementation, verification, review.
5. **Project-specific rules** — exact stack/domain constraints and file paths.
6. **Artifacts** — output locations such as `llm/tasks/`, `llm/research/`, docs/specs, evidence.
7. **Gates and stop conditions** — approval, blockers, red flags.
8. **Verification law** — no completion claim without evidence.
9. **Review checklist** — concrete handoff checks.

## Casbin adaptation result

Status: pass.

Detailed Casbin skills now follow Carbon's operational pattern while staying specific to this repo.

## Verified detailed skills

- `.agents/skills/api-endpoint/SKILL.md`
- `.agents/skills/go-service/SKILL.md`
- `.agents/skills/auth-tenant-casbin/SKILL.md`
- `.agents/skills/api-key-scope/SKILL.md`
- `.agents/skills/database-transactions/SKILL.md`
- `.agents/skills/cross-stack-change/SKILL.md`
- `.agents/skills/frontend-surface/SKILL.md`
- `.agents/skills/query-builder-security/SKILL.md`
- `.agents/skills/tus-upload-storage/SKILL.md`
- `.agents/skills/worker-audit-webhook/SKILL.md`
- `.agents/skills/realtime-sse-websocket/SKILL.md`
- `.agents/skills/verification-before-completion/SKILL.md`
- `.agents/skills/feature/SKILL.md`
- `.agents/skills/execute/SKILL.md`
- `.agents/skills/plan/SKILL.md`
- `.agents/skills/brainstorm/SKILL.md`
- `.agents/skills/self-review/SKILL.md`
- `.agents/skills/systematic-debugging/SKILL.md`
- `.agents/skills/test-driven-development/SKILL.md`
- `.agents/skills/forms/SKILL.md`

## Casbin-specific boundaries encoded

- `internal/config/app.go` as composition root
- `internal/router/router.go` as route/access truth
- Redis-backed auth/session checks
- organization tenant boundary and membership/cache behavior
- Casbin policy/enforcer semantics and transactional enforcer warning
- API-key scope checks
- TUS upload metadata/hooks
- Asynq worker/audit/webhook side effects
- `pkg/querybuilder` sensitive field restrictions
- SSE/WebSocket ticket/origin/presence behavior
- active frontend surfaces: `apps/web` and `apps/client`

## Stale stack audit

- No Carbon/Supabase/Kysely/Biome/ERP/MES/Prisma/Drizzle assumptions found in the detailed Casbin skills.
- Existing third-party Vercel frontend skills remain additive and should be used only for React/frontend guidance, not backend stack guidance.
