# Image Generation Workflow

## Purpose

Primary workflow for generating premium visual concepts, brand kits, and UI mockups in this repo.

This workflow outputs images only. It does not write code. Treat output as visual specification for later implementation or review.

## Use When

Use this workflow when user wants:

- brand identity direction
- brand board or logo system exploration
- website or landing page mockups as image references
- mobile app screen concepts as image references
- visual direction before code implementation

## Do Not Use When

- task is code implementation; use `frontend-design.md`, `frontend-redesign.md`, or `image-to-frontend.md`
- task is pure frontend refactor without new visual direction
- task is backend-only or docs-only
- task is already a generated-image-to-code implementation request

## Read Order

1. `AGENTS.md`
2. `llm/references/frontend-skill-map.md`
3. `llm/cache/frontend-map.md` if output will feed app work later
4. `llm/cache/project-overview.md` if repo-specific product tone matters

## Skill Selection Matrix

Pick exactly one primary generation skill.

### `brandkit`

Use when target is:

- logo directions
- brand guidelines
- identity boards
- presentation-style visual world boards
- typography/color system exploration

### `imagegen-frontend-web`

Use when target is:

- landing pages
- marketing sites
- web dashboards as design references
- conversion-aware web section concepts

Hard rule:

- one separate horizontal image per section
- do not compress multiple sections into one image
- section count should be explicit if user gives one

### `imagegen-frontend-mobile`

Use when target is:

- iOS screens
- Android screens
- app onboarding or auth flow concepts
- mobile dashboards or app-native flow references

Hard rule:

- favor readable phone-screen composition
- keep device framing subtle and premium
- do not turn mobile concept into desktop page mockup

## Workflow Steps

### Step 1 — Classify Deliverable

State what the output is:

- brand kit
- web section set
- mobile app screen set

If user request mixes more than one deliverable, split into separate runs or ask which one is primary.

### Step 2 — Capture Product Tone

Infer tone from repo and request, then keep it consistent with Casbin product reality.

For this repo, tone may lean toward:

- developer tool
- admin console
- security / authorization product
- enterprise utility
- dashboard or workflow-heavy interface

Do not force consumer-app visuals when repository context is operational or security-heavy.

### Step 3 — Choose Composition Rules

#### For `brandkit`

Use:

- sparse typography
- deliberate grid
- strong negative space
- identity system feel, not collage noise
- logo, palette, type, motion, and usage examples if space allows

#### For `imagegen-frontend-web`

Use:

- one image per section
- varied composition across sections
- avoid repeating left-text/right-image hero by default
- keep hierarchy readable at desktop size
- show CTA, trust, or data cues where relevant

#### For `imagegen-frontend-mobile`

Use:

- single-screen clarity
- app-native component proportions
- comfortable tap target spacing
- legible type on phone mockup
- consistent flow across multiple screens

### Step 4 — Include Section/Screen Map

Before generating, decide:

- how many images are needed
- what each image represents
- what must remain consistent across the set

Suggested set examples:

- web landing page: hero, trust, feature blocks, proof, CTA, footer
- mobile app flow: splash, login, setup, dashboard, detail, confirmation
- brand kit: logo, palette, type, layout rules, mock usage, symbol exploration

### Step 5 — Keep Output Honest

If prompt or repo context constrains design, do not overclaim.

Examples of honest output boundaries:

- missing brand assets
- unclear product positioning
- no confirmed color system
- no confirmed target platform
- no need for multiple variations

### Step 6 — Handoff To Next Workflow

If user wants implementation after images:

- send to `image-to-frontend.md` for image-led frontend implementation
- or to `frontend-design.md` / `frontend-redesign.md` if images are just inspiration

## Common Mistakes

- mixing image-generation skills with code workflow in same run
- generating one all-in-one landing page image for web instead of one per section
- making mobile mockup look like desktop dashboard
- using noisy collage instead of readable product references
- ignoring repo tone and making consumer-gloss visuals when product is security/admin oriented

## Verification

This workflow does not run code verification.

What to check instead:

- image count matches task scope
- composition matches requested platform
- each output is readable and distinct
- brand or UI tone stays consistent across set

## Stop Conditions

Stop and ask for clarification if:

- requested platform is ambiguous
- user asks for both mobile and web in one pass without prioritization
- user wants implementation, not image generation
- section count is not clear for `imagegen-frontend-web` and output would be arbitrary

## Completion Output

Report:

- chosen image skill
- deliverable type
- number of images or screens generated
- major composition rules used
- whether next step should be code implementation or another image pass
