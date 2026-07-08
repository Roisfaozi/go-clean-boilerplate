# progress

2026-07-08, transformation engine. Verdict: migrated progress wrapper to Base UI progress primitive.

## Changed

- `packages/ui/src/progress.tsx` — replaced `@radix-ui/react-progress` with `@base-ui/react/progress`. Mapped Root → Root, Indicator → Indicator.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Render progress with different values, verify indicator width matches.
