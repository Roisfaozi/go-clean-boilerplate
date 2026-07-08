# tooltip

2026-07-08, transformation engine. Verdict: migrated tooltip wrapper to Base UI tooltip primitive.

## Changed

- `packages/ui/src/tooltip.tsx` — replaced `@radix-ui/react-tooltip` with `@base-ui/react/tooltip`. Mapped Provider → Provider, Root → Root, Trigger → Trigger, Content → Portal + Positioner + Popup. Preserved positioning props.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Hover over trigger, verify tooltip appears. Move away, verify it disappears.
