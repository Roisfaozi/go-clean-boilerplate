# Image-To-Frontend Workflow

## Purpose

Safe routing workflow for frontend tasks that should begin from image/reference-first design direction.

## Use When

Use when the task is driven by a provided visual reference or when the workflow should generate/analyze images before implementation.

## Read Order

1. `AGENTS.md`
2. `llm/cache/frontend-map.md`
3. `llm/cache/frontend-proxy-system.md`
4. `llm/references/frontend-skill-map.md`
5. target frontend owner in `apps/web` or `apps/client`

## Primary Skill Route

Default stack:

- `frontend-surface`
- `image-to-code`

Optional supporting skills:

- `full-output-enforcement`
- `vercel-react-best-practices`
- one style preset only if explicit art direction is needed

## Rules

- use this only for image-led or visual-reference-led tasks
- do not make this the default frontend workflow
- if implementation changes backend contract or data flow, also load `cross-stack-change`
- if task is redesign of existing UI without image-led requirement, prefer `frontend-redesign.md`

## Verification

- app-specific typecheck
- browser validation of implemented surface when practical
- confirm responsive hierarchy and key states are still usable
