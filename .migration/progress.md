# progress

2026-07-08, engine, migrated to Base UI.

## Changed

- `packages/ui/src/progress.tsx` — replaced `@radix-ui/react-progress` with `@base-ui/react/progress`. Rebuilt anatomy: `Root > Track > Indicator > Value`. Removed manual `-translateX` style since Base UI handles fill computing natively.
- Leftover sweep is clean.

## Left alone
None

## Behavior changes
None

## Verify by hand

- View a loading progress bar if any exists, confirm fill advances appropriately.
