# hover-card

2026-07-08, transformation engine. Verdict: migrated hover-card (PreviewCard) wrapper to Base UI preview-card primitive.

## Changed

- `packages/ui/src/hover-card.tsx` — replaced `@radix-ui/react-hover-card` with `@base-ui/react/preview-card`. Mapped Root → Root, Trigger → Trigger, Content → Portal + Positioner + Popup. Preserved positioning props.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Hover over trigger, verify card appears after delay. Move away, verify it closes.
