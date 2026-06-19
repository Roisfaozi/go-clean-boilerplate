# Frontend Redesign Workflow

## Purpose

Safe routing workflow for upgrading an existing frontend surface without turning the skill system into overlapping design chaos.

## Use When

Use when a page, dashboard, marketing surface, or existing UI tree already exists and the task is to improve design quality while preserving product behavior.

## Read Order

1. `AGENTS.md`
2. `llm/cache/frontend-map.md`
3. `llm/cache/frontend-proxy-system.md`
4. `llm/references/frontend-skill-map.md`
5. target existing frontend surface

## Primary Skill Route

Default stack:

- `frontend-surface`
- `redesign-existing-projects`

Optional supporting skills:

- `web-design-guidelines`
- `high-end-visual-design`
- `vercel-react-best-practices`
- one style preset only

## Rules

- redesign existing structure before inventing new routing or data flow
- keep product behavior, auth flow, and proxy behavior intact unless task explicitly changes them
- do not combine multiple competing style presets in one pass
- if redesign exposes architecture or composition problems, add `vercel-composition-patterns`

## Verification

- typecheck for owning app
- browser or E2E pass for changed key flows when practical
- check visual states and auth-sensitive surfaces
