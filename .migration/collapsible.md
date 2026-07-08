# collapsible

2026-07-08, engine, migrated to Base UI.

## Changed

- `packages/ui/src/collapsible.tsx` — replaced `@radix-ui/react-collapsible` with `@base-ui/react/collapsible`. Re-mapped `CollapsibleContent` to `Collapsible.Panel`.
- Leftover sweep on collapsible wrapper is clean.

## Left alone
None

## Behavior changes
None

## Verify by hand

- Click any sidebar group toggle and ensure nested links reveal correctly.
