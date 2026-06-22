# Frontend Skill Map

## Purpose

This file prevents frontend skill routing chaos.

It classifies high-signal frontend and React skills by role so agents do not treat style packs, rulebooks, and orchestration skills as if they were the same thing.

## Rule of Use

Do not load all frontend skills together by default.

Pick:

1. one primary workflow skill
2. zero or more supporting reference skills
3. zero or one style preset
4. `full-output-enforcement` only when exhaustive output is required

## Primary Workflow Skills

These can drive a task from start to finish.

### `design-taste-frontend`

Use as the default premium frontend design workflow when the task is to design or implement a new frontend surface and visual quality matters.

Best for:

- new landing pages
- premium UI direction
- frontend redesigns where no image-first workflow is required

### `redesign-existing-projects`

Use when an existing frontend already exists and the task is to upgrade its design quality without rebuilding routing or product behavior from scratch.

Best for:

- reviewer-facing redesigns
- visual uplift of existing pages
- removing generic AI-looking UI from current surface

### `image-to-code`

Use when the task should begin from a visual reference or generated image and then implement the UI to match it.

Best for:

- image-led implementation
- visual reverse engineering
- highly art-directed frontend delivery

## Supporting Reference Skills

These should inform implementation or review, but should not be treated as the main workflow by default.

### `vercel-composition-patterns`

Role: React component architecture reference.

Use when:

- component APIs are getting boolean-prop heavy
- compound components or provider patterns are needed
- shared component composition is being refactored

### `vercel-react-best-practices`

Role: React and Next.js performance reference.

Use when:

- reviewing React or Next performance
- implementing client/server data flow
- reducing waterfalls, rerenders, or bundle waste

### `web-design-guidelines`

Role: UI audit and review rubric.

Use when:

- reviewing design quality
- checking usability or accessibility
- auditing interface quality against broad web standards

### `high-end-visual-design`

Role: visual-quality reference.

Use when:

- premium spacing, typography, and hierarchy need to be enforced
- a workflow skill needs stronger visual guardrails

## Image Generation Skills

These skills are purely for generating high-end reference images. They DO NOT write code. Use them when exploring visual direction before coding.

### `imagegen-frontend-web`

Use for generating premium, conversion-aware website and landing page mockups. Output constraint: one horizontal image per section.

### `imagegen-frontend-mobile`

Use for generating elite, app-native mobile screens and flows (iOS/Android). Prioritizes UI legibility and realistic device mockups.

### `brandkit`

Use for high-end brand guidelines, logo systems, and visual-world presentations.

## Style Presets

These are optional style directions, not default workflows.

### `minimalist-ui`

Use when clean editorial simplicity is the intended look.

### `industrial-brutalist-ui`

Use when raw, mechanical, high-contrast brutalist direction is the intended look.

### `gpt-taste`

Use when bold, highly art-directed, motion-heavy, conversion-aware visual direction is explicitly desired.

## Execution Modifier

### `full-output-enforcement`

Role: output completeness modifier.

Use when:

- generated frontend output must be exhaustive
- partial code, placeholders, or truncation would break handoff quality

Do not treat this as a design workflow.

## Legacy / Compatibility

### `design-taste-frontend-v1`

Keep only as fallback when a task explicitly depends on older behavior.

Do not use as default if `design-taste-frontend` is acceptable.

## Recommended Routing Patterns

### New premium frontend surface

- `frontend-surface`
- `design-taste-frontend`
- optional `vercel-react-best-practices`
- optional `vercel-composition-patterns`
- optional one style preset

### Existing frontend redesign

- `frontend-surface`
- `redesign-existing-projects`
- optional `web-design-guidelines`
- optional `high-end-visual-design`
- optional one style preset

### Visual Ideation & Image Generation

- `brandkit` (for brand identity) OR
- `imagegen-frontend-web` (for web design exploration) OR
- `imagegen-frontend-mobile` (for mobile app exploration)
- (Do not mix with coding workflow skills in the same pass if outputting images only)

### Image-led frontend implementation

- `frontend-surface`
- `image-to-code`
- optional `full-output-enforcement`
- optional `vercel-react-best-practices`

### React architecture cleanup

- `frontend-surface`
- `ui`
- `vercel-composition-patterns`
- optional `vercel-react-best-practices`

### React or Next performance cleanup

- `frontend-surface`
- `ui`
- `vercel-react-best-practices`

## What To Avoid

- loading all taste and style skills together by default
- treating style presets as architecture workflows
- treating reference rulebooks as implementation orchestrators
- routing every frontend task through image-first flow
- using `design-taste-frontend-v1` as normal default
