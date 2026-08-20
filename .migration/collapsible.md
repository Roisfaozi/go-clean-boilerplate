# collapsible

2026-07-08, transformation engine. Verdict: migrated collapsible wrapper to Base UI collapsible primitive.

## Changed

- `packages/ui/src/collapsible.tsx` — replaced `@radix-ui/react-collapsible` with `@base-ui/react/collapsible`. Mapped Root → Root, Trigger → Trigger, Content → Panel.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Click collapsible trigger, verify expand/collapse.
