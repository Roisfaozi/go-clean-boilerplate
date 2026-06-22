# Frontend Redesign Workflow

## Purpose

Primary workflow for upgrading an existing frontend surface while preserving existing product behavior, routing, auth, and backend contract unless change explicitly includes them.

Use this when current UI exists but looks generic, inconsistent, cramped, or structurally weak.

## Use When

- redesigning existing dashboard/page/form/layout
- upgrading visual quality of current route
- improving hierarchy, readability, and polish without rebuilding product flow
- making reviewer-facing visual uplift in already working surface

## Do Not Use When

- task is brand-new surface from scratch; prefer `frontend-design.md`
- task is image-first implementation from screenshot/generated comps; prefer `image-to-frontend.md`
- task is pure image ideation with no code; prefer `image-generation.md`

## Required Read Order

1. `AGENTS.md`
2. `llm/cache/frontend-map.md`
3. `llm/cache/frontend-proxy-system.md`
4. `llm/cache/api-contracts.md` if data shape involved
5. `llm/references/frontend-skill-map.md`
6. target existing route/component tree

## Default Skill Route

Primary stack:

- `frontend-surface`
- `redesign-existing-projects`

Optional support:

- `web-design-guidelines`
- `high-end-visual-design`
- `vercel-react-best-practices`
- `vercel-composition-patterns` if redesign exposes architecture issues
- one style preset only when explicit direction requested

## Workflow Steps

### Step 1 — Audit Current Surface

Capture what already exists:

- route owner
- current layout structure
- current data dependencies
- current interaction states
- what must remain behaviorally identical

### Step 2 — Separate Visual Debt From Product Logic

Classify issues into:

- spacing / hierarchy debt
- typography debt
- component inconsistency
- state feedback weakness
- layout density or clutter
- real product logic issue

Fix visual debt first unless task explicitly includes logic refactor.

### Step 3 — Keep Runtime Boundaries Stable

Default redesign rule:

- keep routing stable
- keep proxy behavior stable
- keep auth flow stable
- keep backend payload mapping stable

If one of those changes, write it as explicit scope expansion, not incidental redesign.

### Step 4 — Redesign Surface by Layer

Preferred order:

- page shell and spacing system
- section hierarchy
- cards/tables/forms/filters
- action emphasis and affordances
- state messaging
- responsive cleanup

### Step 5 — Re-audit UX States

Redesign not complete until these still work:

- loading
- empty
- error
- success
- destructive confirmation or delete state if present
- auth-expired or forbidden state if present

## Common Mistakes

- turning redesign into hidden product rewrite
- adding competing style presets together
- changing data flow because UI felt messy
- forgetting mobile or narrow desktop density
- upgrading visuals while making empty/error states worse

## Verification

### Minimum

- owner app typecheck

### Add When Needed

- owner app build
- browser/manual review of redesigned route
- E2E/manual auth flow if redesign touches protected area

## Stop Conditions

Stop and inspect more if:

- redesign seems to require changing backend contract
- current route owner not clear
- both apps duplicate same surface and redesign scope is unclear
- visual issue is actually caused by upstream data or component API design

## Completion Output

Report:

- what existing surface was redesigned
- what product behavior stayed stable
- what visual debt categories were fixed
- verification run
- any scope intentionally deferred
