# slider

2026-07-08, transformation engine. Verdict: migrated slider wrapper to Base UI slider primitive.

## Changed

- `packages/ui/src/slider.tsx` — replaced `@radix-ui/react-slider` with `@base-ui/react/slider`. Mapped root, track, indicator (range), and thumb parts.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Drag slider thumb, verify it moves smoothly and the track fills correctly.
