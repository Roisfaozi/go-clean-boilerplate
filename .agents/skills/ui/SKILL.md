---
name: ui
description: Use when building, reviewing, or polishing frontend UI in `apps/web`, `apps/client`, or shared packages.
---

# Ui

## Read First

- `llm/cache/frontend-map.md`
- `llm/conventions/typescript.md`
- `.agents/skills/vercel-react-best-practices/SKILL.md`
- `.agents/skills/vercel-composition-patterns/SKILL.md`

## Workflow

1. Confirm owning app before editing: `apps/web` Next.js App Router or `apps/client` React Router.
2. Reuse `packages/ui`, shared hooks, and existing app components before adding new primitives.
3. Check loading, empty, error, success, auth-expired, and tenant-switch states when relevant.
4. Apply Vercel React/performance skills when React component architecture is involved.
5. Verify with typecheck/build or browser flow appropriate to the touched app.

## Watch For

- Do not duplicate API client/proxy logic across apps.
- Do not treat frontend-only visibility as authorization.
