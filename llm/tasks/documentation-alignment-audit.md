# Documentation Alignment Audit

## Goal

Audit whether the concretized AI configuration in `AGENTS.md` and `llm/` follows repository documentation, while treating live code as final authority and treating noisy/redundant docs as secondary evidence.

## Overall judgment

The current AI configuration is mostly aligned with the strongest documentation sources, and in several places it is intentionally more accurate than parts of `documentation/` because it prioritizes live code over stale or product/vision docs.

## Documentation groups by trust level

### High-trust docs for AI configuration

These match live code closely enough to inform agent behavior:

- `documentation/ARCHITECTURE.md`
- `documentation/guides/TESTING.md`
- `documentation/guides/REALTIME.md`
- `documentation/guides/RESUMABLE_UPLOAD.md`
- `documentation/guides/OBSERVABILITY.md`
- `documentation/llm/00-analysis-priority.md`
- `documentation/llm/01-module-map.md`
- `documentation/llm/02-request-flows.md`
- `documentation/llm/04-architecture-security-audit.md`
- `documentation/llm/05-api-key-analysis.md`

### Medium-trust docs

Useful as supporting context, but should still be rechecked against live code:

- `documentation/API_ACCESS_WORKFLOW.md`
- `documentation/API_ANALYSIS_SUMMARY.md`
- `documentation/MODULE_FOLDER_AND_CONTRACT_MAP.md`
- `documentation/MULTI_TENANCY.md`
- `documentation/ORGANIZATION_LIFECYCLE_RUNBOOK.md`
- `documentation/guides/WEBHOOKS.md`
- `documentation/guides/STORAGE.md`
- `documentation/guides/SEARCH.md`

### Low-trust / noisy / drift-prone docs

These should not drive AI configuration without live-code verification:

- `documentation/PROJECT_GUIDE.md`
- `documentation/FRONTEND_STRUCTURE.md`
- `documentation/FRONTEND_PACKAGE_EXTRACTION_PLAN.md`
- `documentation/UI_ROADMAP.md`
- `documentation/productplan/*`
- `documentation/ops/*`
- `documentation/api/AI_STREAMING_CONTRACT.md`

## Confirmed alignments

### Stack and structure

AI configuration aligns with docs and code on:

- Go backend as core runtime
- Gin + GORM + Casbin + Redis + Asynq + OTEL + TUS
- organization/tenant-driven boundaries
- realtime via WebSocket and SSE
- SQL migrations under `db/migrations`
- test layering: unit / integration / E2E

### Module architecture

AI configuration aligns with docs on:

- repository/usecase/controller layering
- composition-root importance of `internal/config/app.go`
- practical route boundary importance of `internal/router/router.go`
- permission layer being more than simple route protection

### Testing

AI configuration aligns with docs on:

- singleton-container integration approach
- Docker requirement for integration/E2E
- broad security/regression coverage around auth, tenant, worker, realtime, and upload paths

### Realtime and upload

AI configuration aligns with docs on:

- SSE and WebSocket being distinct runtime channels
- TUS using registry/hook style upload handling
- upload boundary being separate from normal CRUD routes

## Intentional deviations where AI config is more reliable

### Frontend documentation drift

`documentation/PROJECT_GUIDE.md` says frontend is Next.js 14 and positions the project as Go + Next.js admin.

Current AI configuration does **not** copy that literally because live code shows:

- active `apps/web` using Next.js 16
- active `apps/client` using React Router + Vite
- shared workspace packages in `packages/*`

So the AI configuration is more accurate than that doc.

### `web/` vs `apps/web`

Some older docs still reference `web/` or assume a single frontend. AI configuration correctly treats:

- `apps/web`
- `apps/client`

as active surfaces.

### AI streaming contract doc

`documentation/api/AI_STREAMING_CONTRACT.md` describes `POST /api/v1/ai/chat` for NexusAI.

Current AI configuration intentionally treats this as supporting or planned documentation, not a confirmed live backend contract, because matching route wiring was not found in current backend route registration.

### Tenant header wording

Some docs emphasize `X-Organization-ID` / `X-Organization-Slug` as frontend/request-level concepts. AI configuration keeps that context, but correctly centers runtime truth on middleware/usecase/route behavior instead of assuming every path is header-driven in the same way.

## Documentation noise and redundancy

### Redundant with `llm/`

These docs overlap significantly with the new AI configuration and are now secondary for agent onboarding:

- `documentation/llm/01-module-map.md`
- `documentation/llm/02-request-flows.md`
- `documentation/llm/04-architecture-security-audit.md`
- `documentation/API_ANALYSIS_SUMMARY.md`
- `documentation/MODULE_FOLDER_AND_CONTRACT_MAP.md`

They are still useful as evidence/history, but `llm/cache/*` is now the cleaner agent-facing surface.

### Product/vision contamination

These docs mix product aspirations with technical guidance and should not be treated as agent runtime truth:

- `documentation/productplan/*`
- `documentation/ops/*`
- `documentation/api/AI_STREAMING_CONTRACT.md`
- parts of `documentation/PROJECT_GUIDE.md`

### Potentially stale frontend references

- `documentation/PROJECT_GUIDE.md`
- `documentation/FRONTEND_STRUCTURE.md`
- `documentation/MULTI_TENANCY.md` frontend client examples

These need live-code verification before reuse.

## Practical conclusion for AI configuration

The current `AGENTS.md` + `llm/` setup follows the strongest parts of repository documentation while correctly overriding weak or stale docs with live code.

That is the right behavior for this repo.

## Recommendation

Preferred agent truth order should remain:

1. live code
2. config / wiring / routes
3. tests
4. strong technical docs
5. product/ops/planning docs

`documentation/` should be treated as a mixed-quality knowledge base, not a single authoritative source.
