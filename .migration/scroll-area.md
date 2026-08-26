# scroll-area

2026-07-08, transformation engine. Verdict: migrated scroll-area wrapper to Base UI scroll-area primitive.

## Changed

- `packages/ui/src/scroll-area.tsx` — replaced `@radix-ui/react-scroll-area` with `@base-ui/react/scroll-area`. Mapped parts accordingly.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Render content longer than container, verify scrollbar appears and works.
