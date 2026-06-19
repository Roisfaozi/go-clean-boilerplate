# Image Generation Workflow

## Purpose

Safe routing workflow for generating premium visual concepts, brand kits, and UI mockups. **This workflow does not output code.**

## Use When

Use when the task explicitly requests image generation, visual ideation, brand identity concepts, or UI mockups before any implementation begins.

## Read Order

1. `AGENTS.md`
2. `llm/references/frontend-skill-map.md`

## Primary Skill Route

Pick **one** primary generation skill based on the task:

- `imagegen-frontend-web` (for landing pages, dashboards, websites)
- `imagegen-frontend-mobile` (for iOS/Android native app flows)
- `brandkit` (for logos, typography systems, brand boards)

## Rules

- **Do not write code** when running an image generation task.
- **Do not mix** `imagegen-frontend-web` and `imagegen-frontend-mobile` in the same prompt.
- When using `imagegen-frontend-web`, strictly obey the rule of one image per section.
- Treat the generated images as visual specifications that can later be fed into the `image-to-frontend` workflow.

## Handoff

- Present the generated images clearly to the user.
- Ask if they want to iterate on the design or if they want to move to implementation using the `image-to-code` workflow.
