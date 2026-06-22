# Frontend Design Workflow

## Purpose

Primary workflow for new frontend work in this repo when visual quality matters and task is not explicitly image-first.

Use this when agent must design or implement UI in `apps/web`, `apps/client`, or shared frontend packages without turning skill selection into chaos.

## Use When

Use this workflow when:

- building new page, screen, feature surface, or component tree
- improving UI quality while also shipping code
- introducing new frontend route or domain surface
- refining hierarchy, spacing, interaction states, or information architecture

Do not use this as default when task is purely image generation. Use `llm/workflows/image-generation.md` for that.

## Do Not Use When

- task is pure visual ideation with no code output
- task is direct redesign of existing surface with minimal product change; prefer `frontend-redesign.md`
- task begins from supplied screenshot or generated section images; prefer `image-to-frontend.md`
- task is backend-driven contract change first; pair with `cross-stack-change.md` or `api-endpoint.md`

## Required Read Order

1. `AGENTS.md`
2. `llm/cache/project-overview.md`
3. `llm/cache/frontend-map.md`
4. `llm/cache/frontend-proxy-system.md`
5. `llm/cache/api-contracts.md` if backend boundary involved
6. `llm/conventions/typescript.md`
7. `llm/references/frontend-skill-map.md`
8. target owner surface in `apps/web`, `apps/client`, or `packages/*`

## Owner Decision

Prove owning surface before changing code.

### `apps/web`

Prefer when change belongs to:

- Next.js App Router route tree
- locale-aware pages under `src/app/[locale]`
- server actions
- backend proxy route handlers
- cookie/session-heavy server-side auth behavior

### `apps/client`

Prefer when change belongs to:

- React Router route registry and feature screens
- admin/domain pages under `app/features/*`
- Vite-powered client routes
- Playwright-tested client flows

### `packages/*`

Prefer when behavior is truly shared across apps:

- stable API contract typing
- shared UI primitives
- shared hooks or generic utilities

Do not move app-specific logic into shared package just to avoid duplication in one patch.

## Default Skill Route

Primary stack:

- `frontend-surface`
- `design-taste-frontend`

Optional supporting skills by need:

- `vercel-react-best-practices` for data-flow, performance, bundle, or server/client split concerns
- `vercel-composition-patterns` for component API shape or reusable composition problems
- `high-end-visual-design` when visual quality bar needs stronger enforcement
- one style preset only when explicit direction is requested
- `full-output-enforcement` only if exhaustive output matters

## Workflow Steps

### Step 1 — Prove Runtime Owner

- identify route/page/component owner
- identify proxy/API helper owner if data driven
- identify whether both apps consume same contract

### Step 2 — Define UI Job

Before coding, name exact job:

- new screen
- route-level feature
- shared component
- form flow
- dashboard/data surface
- auth-sensitive interaction

This avoids pulling design-heavy skill when actual task is architecture cleanup only.

### Step 3 — Audit State Matrix

For every meaningful UI surface, prove expected handling for:

- loading
- empty
- success
- error
- auth expired / unauthorized
- tenant-sensitive mismatch if relevant
- mutation pending / retry if form or action driven

### Step 4 — Check Contract Boundary

If surface reads or mutates backend data, inspect:

- backend route family in `internal/router/router.go`
- proxy owner file
- app-local API helper
- shared type package if used

If contract changed, do not keep work “frontend-only” in writeup.

### Step 5 — Implement Smallest Coherent Slice

Preferred order:

- shared types or API helper if needed
- route/page shell
- child components
- state and async handling
- polish and accessibility pass

### Step 6 — Reviewer-Facing Check

Before finishing, answer:

- what improved visually?
- what improved structurally?
- what changed in behavior contract?
- what risks remain?

## Common Mistakes

- loading too many style skills together
- guessing app owner from folder name only
- changing backend contract but not proxy/helper code
- polishing happy path only
- treating auth and tenant state as edge cases instead of first-class states
- putting app-specific logic into shared packages too early

## Verification

Start narrow.

### Minimum

- owning app typecheck

### Add When Needed

- owning app build when route/component wiring changed
- browser/manual run when UX/state transitions matter
- E2E when route protection, auth, tenant, or multi-step flow changed

### Repo Notes

- `apps/client` lint runs Biome; run `typecheck` separately for TypeScript validation
- if both apps consume same change, verify both owners or say why not

## Stop Conditions

Stop and inspect more live code if:

- owner app not clear
- proxy and backend route disagree
- route state depends on auth or tenant semantics you did not verify
- shared package change would affect both apps in unclear way

## Completion Output

Report:

- owning app chosen and why
- skill route used
- files changed
- verification run
- unverified states or blockers
