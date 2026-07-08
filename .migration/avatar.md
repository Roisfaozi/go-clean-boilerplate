# avatar

2026-07-08, engine, migrated to Base UI.

## Changed

- `packages/ui/src/avatar.tsx` — replaced `@radix-ui/react-avatar` with `@base-ui/react/avatar`. Clean 1:1 mapping of `Root`, `Image`, and `Fallback`.
- Leftover sweep is clean.

## Left alone
None

## Behavior changes
None

## Verify by hand

- Check user profile navigation dropdown trigger; ensure initials (Fallback) or image (Image) renders.
