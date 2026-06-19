# Image-To-Frontend Workflow

## Purpose

Primary workflow for frontend implementation that begins from visual references.

This is code-producing flow. It uses images as specification input, then builds frontend to match them as closely as practical within repo architecture.

## Use When

- user provides screenshot or visual reference
- user wants image-generated design converted into code
- task needs high-fidelity UI from image-first direction
- you first generate section/screen images and then implement them

## Do Not Use When

- task is pure image generation only; use `image-generation.md`
- task is normal frontend feature without image-led starting point; use `frontend-design.md`
- task is simple redesign of existing surface without visual reference; use `frontend-redesign.md`

## Required Read Order

1. `AGENTS.md`
2. `llm/cache/frontend-map.md`
3. `llm/cache/frontend-proxy-system.md`
4. `llm/cache/api-contracts.md` if data driven
5. `llm/references/frontend-skill-map.md`
6. target owner in `apps/web` or `apps/client`
7. supplied image reference or generated image set

## Skill Route

Implementation stack:

- `frontend-surface`
- `image-to-code`

Optional support:

- `full-output-enforcement`
- `vercel-react-best-practices`
- one style preset only when explicit art direction needed

Reference-generation inputs can come from:

- `imagegen-frontend-web`
- `imagegen-frontend-mobile`
- `brandkit` for identity direction

Do not mix image-generation-only skills into implementation step unless still actively generating reference images.

## Workflow Steps

### Step 1 — Classify Reference Type

Reference may be:

- user-supplied screenshot
- existing product screenshot
- generated section images from `imagegen-frontend-web`
- generated mobile screens from `imagegen-frontend-mobile`
- brand direction from `brandkit`

This determines what fidelity matters most: layout, palette, typography, motion, or brand tone.

### Step 2 — Pick Owner App

Decide whether implementation belongs to:

- `apps/web`
- `apps/client`
- shared package plus app integration

Do not start coding before proving owner.

### Step 3 — Extract Implementation Spec From Image

List concrete cues from image:

- page sections
- layout grid
- type scale
- spacing rhythm
- card/table/form patterns
- key CTA placement
- color/palette constraints
- responsive assumptions

### Step 4 — Map Visual Spec To Repo Reality

Before coding, decide:

- which existing shared components can be reused
- which parts need new components
- whether backend data contract already exists
- whether image implies fake data or real API data

### Step 5 — Build Closest Honest Match

Implementation goal is closest truthful match.

Do not claim parity with image if:

- asset missing
- motion omitted
- responsive behavior differs materially
- backend data shape forces compromise

### Step 6 — Validate Against Reference

Check:

- section order and hierarchy
- spacing and typography feel
- major visual accents
- state completeness
- responsive usability

## Common Mistakes

- using image workflow for generic frontend task
- implementing image literally without mapping to repo owner/data reality
- mixing many style skills with image-led spec
- ignoring empty/loading/error states because image only showed happy path
- claiming exact match where repo constraints forced simplification

## Verification

### Minimum

- owning app typecheck

### Add When Needed

- owning app build
- browser/manual visual compare against reference
- responsive checks
- auth/tenant flow checks if protected route involved

## Stop Conditions

Stop and inspect more if:

- reference owner app unclear
- image implies backend data/contract not present in live code
- visual design requires asset pipeline or motion system not in scope
- generated image conflicts with existing product behavior user wants preserved

## Completion Output

Report:

- source of visual reference
- owner app chosen
- main implementation compromises, if any
- verification run
- whether next step should be further visual iteration or backend/data integration
