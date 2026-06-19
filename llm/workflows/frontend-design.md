# Frontend Design Workflow

## Purpose

Safe routing workflow for new frontend work that needs strong visual quality without causing skill-selection chaos.

## Use When

Use when building or redesigning a frontend surface where UI quality matters and the task is not explicitly image-led.

## Read Order

1. `AGENTS.md`
2. `llm/cache/frontend-map.md`
3. `llm/cache/frontend-proxy-system.md`
4. `llm/conventions/typescript.md`
5. `llm/references/frontend-skill-map.md`
6. target app surface in `apps/web` or `apps/client`

## Primary Skill Route

Default stack:

- `frontend-surface`
- `design-taste-frontend`

Add supporting skills only if task needs them:

- `vercel-react-best-practices`
- `vercel-composition-patterns`
- `high-end-visual-design`
- one style preset only
- `full-output-enforcement` only when output completeness is critical

## Rules

- prove owning app first
- do not load multiple conflicting style presets by default
- do not use image-first workflow unless task is explicitly visual-reference driven
- if backend contract changes, also load `cross-stack-change` or `api-endpoint`

## Verification

- app-specific typecheck first
- build or browser flow when UI runtime path changed
- review loading, empty, error, success, auth-expired, and tenant-sensitive states when relevant
