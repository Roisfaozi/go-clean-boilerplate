# Carbon vs Casbin — Skill and AI Workflow Analysis

## Scope

Analisa ulang konfigurasi AI-native Carbon terhadap Casbin setelah detailed Casbin skills diterapkan.

## Executive summary

Casbin sekarang sudah mengikuti pola struktur skill Carbon pada high-risk engineering skills: announce line, read order, phased workflow, red flags, stop conditions, and completion output.

Namun Carbon masih lebih matang pada dua area:

1. **Cache granularity** — Carbon punya banyak cache domain/module spesifik. Casbin masih punya cache agregat besar.
2. **Orchestration skills** — Carbon `execute`, `plan`, `systematic-debugging`, `forms`, dan `verification-before-completion` jauh lebih panjang, punya examples, templates, and rationalization traps.

Casbin unggul pada repo-specific high-risk boundary coverage: auth/tenant/Casbin, API-key scope, TUS upload, querybuilder security, worker/audit/webhook, realtime.

## Quantitative comparison

### Cache files

- Carbon: 42 files under `llm/cache/`
- Casbin: 8 files under `llm/cache/`

Carbon cache is domain/module-specific:

- `authentication-system.md`
- `audit-log-system.md`
- `database-patterns.md`
- `event-system.md`
- `inventory-system.md`
- `material-tables.md`
- `scheduling-data-structures.md`
- etc.

Casbin cache is architecture-level:

- `project-overview.md`
- `environment.md`
- `architecture.md`
- `backend-map.md`
- `frontend-map.md`
- `api-contracts.md`
- `module-map.md`
- `domain-rules.md`

### Skill length examples

| Skill                            | Carbon lines | Casbin lines | Assessment                                                                                      |
| -------------------------------- | -----------: | -----------: | ----------------------------------------------------------------------------------------------- |
| `feature`                        |          120 |           49 | Casbin has structure, Carbon has richer orchestration examples/artifacts.                       |
| `execute`                        |          201 |           47 | Casbin shorter; Carbon has task tracking, blocker handling, commit steps, parallelization.      |
| `plan`                           |          252 |           40 | Casbin lacks full plan template and task writing examples.                                      |
| `systematic-debugging`           |          296 |           38 | Casbin has core flow; Carbon has deeper anti-rationalization and root-cause technique guidance. |
| `verification-before-completion` |          139 |           46 | Casbin has command matrix; Carbon has stronger failure patterns and rationalization prevention. |
| `forms`                          |          248 |           43 | Casbin covers backend/frontend validation but lacks route/action/component templates.           |
| `database-transactions`          |           77 |           65 | Near parity; Casbin is stack-specific for GORM/Casbin.                                          |

## Structural parity result

Casbin detailed skills now include Carbon-style structure:

- activation sentence
- when-to-use or read order
- phased workflow
- project-specific paths
- stop conditions
- completion output
- high-risk boundary rules

This is enough for operational use. But for Carbon-level maturity, Casbin needs deeper examples and more granular cache files.

## Casbin strengths after patch

### High-risk boundaries are better surfaced than Carbon starter-pack

Casbin now has focused skills for:

- `.agents/skills/auth-tenant-casbin/SKILL.md`
- `.agents/skills/api-key-scope/SKILL.md`
- `.agents/skills/tus-upload-storage/SKILL.md`
- `.agents/skills/worker-audit-webhook/SKILL.md`
- `.agents/skills/query-builder-security/SKILL.md`
- `.agents/skills/realtime-sse-websocket/SKILL.md`
- `.agents/skills/frontend-surface/SKILL.md`

These map directly to real repo risks in `AGENTS.md` and `llm/cache/domain-rules.md`.

### Backend skill is well-targeted

`go-service` now encodes:

- `internal/config/app.go` as composition root
- `internal/modules/*/module.go` constructor ownership
- controller/usecase/repository split
- context propagation
- transactional enforcer warning
- high-risk boundary routing to other skills

### API skill is well-targeted

`api-endpoint` now encodes:

- route strata: public/authenticated/tenantAuthorized/authorized/upload
- API-key scope behavior
- Swagger artifacts
- frontend proxy audit for `apps/web` and `apps/client`
- Redis session and tenant/Casbin warnings

## Remaining maturity gaps vs Carbon

### 1. Cache granularity gap

Casbin cache is accurate but broad. Carbon gains detail by having many topic caches.

Recommended Casbin cache split, only when verified and committed:

- `llm/cache/authentication-system.md`
- `llm/cache/tenant-organization-system.md`
- `llm/cache/casbin-permission-system.md`
- `llm/cache/api-key-system.md`
- `llm/cache/tus-upload-system.md`
- `llm/cache/worker-audit-webhook-system.md`
- `llm/cache/querybuilder-security.md`
- `llm/cache/realtime-system.md`
- `llm/cache/frontend-proxy-system.md`

Do not create these from uncommitted assumptions. They should be filled from live code audits.

### 2. Orchestration skill depth gap

Casbin `feature`, `execute`, and `plan` are structurally correct but not Carbon-deep.

To match Carbon maturity, add:

- task template examples in `plan`
- blocker handling protocol in `execute`
- progress tracker format in `execute`
- approval gate language in `feature` and `plan`
- output report templates
- explicit parallelization/subagent routing only where tool policy allows

### 3. Debugging skill depth gap

Carbon `systematic-debugging` is much stronger. Casbin should add:

- root-cause investigation checklist by layer
- pattern analysis phase
- hypothesis table format
- rationalization warnings
- examples of bad debugging behavior

### 4. Forms skill depth gap

Carbon forms skill has concrete implementation sections. Casbin forms should eventually include examples for:

- backend Gin request struct / validation tags
- `apps/web` form/action pattern
- `apps/client` form/route pattern
- error payload rendering
- auth/session expiry response handling

### 5. Verification skill depth gap

Casbin has command matrix. To match Carbon, add:

- common false-success patterns
- red flag examples
- per-layer gate questions
- “no evidence, no completion” examples

## Serious mismatch check

No serious wrong-stack mismatch found in detailed Casbin skills.

Search audit found no detailed-skill assumptions for:

- Carbon
- Supabase
- Kysely
- Biome
- ERP
- MES
- Prisma
- Drizzle

Existing Vercel skill references to Prisma/Drizzle are generic third-party frontend performance examples, not Casbin backend guidance.

## Recommendation

Current status: **usable and repo-specific**.

If target is **Carbon-equivalent maturity**, next work should not mainly be more skills. Instead:

1. Split Casbin cache into verified domain/module caches.
2. Deepen orchestration skills (`feature`, `execute`, `plan`, `systematic-debugging`, `verification-before-completion`, `forms`) with examples/templates.
3. Keep high-risk project-specific skills as routing adapters to those caches/workflows.

## Suggested next implementation order

1. Create verified domain cache files from live code for auth/tenant/Casbin/API-key/upload/worker/querybuilder/realtime.
2. Expand `plan`, `execute`, and `systematic-debugging` with Carbon-style templates.
3. Expand `forms` only after auditing actual frontend form patterns in both `apps/web` and `apps/client`.
4. Re-run stale-stack scan across `.agents/skills` and `llm/cache`.

## Closure update — remaining gaps

Status: partially closed in this pass.

Completed directly:

- Created granular Casbin cache files from live code evidence:
  - `llm/cache/authentication-system.md`
  - `llm/cache/tenant-organization-system.md`
  - `llm/cache/casbin-permission-system.md`
  - `llm/cache/api-key-system.md`
  - `llm/cache/tus-upload-system.md`
  - `llm/cache/worker-audit-webhook-system.md`
  - `llm/cache/querybuilder-security.md`
  - `llm/cache/realtime-system.md`
  - `llm/cache/frontend-proxy-system.md`
- Updated `AGENTS.md` fast routing so agents read domain-specific cache files before target live code.
- Verified referenced live paths exist.
- Verified new cache files do not contain Carbon/Supabase/Kysely/Biome/ERP/MES/Prisma/Drizzle assumptions.

Still constrained by environment:

- Direct writes to `.agents/skills/*/SKILL.md` still fail from this Codex environment with `Read-only file system`.
- Detailed orchestration skill patch remains maintained in `llm/tasks/write-detailed-casbin-skills.py`; apply it manually when skill files need regeneration.

Updated status:

- Cache granularity gap: closed for high-risk repo boundaries.
- Skill depth gap: closed when `llm/tasks/write-detailed-casbin-skills.py` is applied manually to `.agents/skills`.
