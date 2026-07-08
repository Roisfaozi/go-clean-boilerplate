# tabs

2026-07-08, engine, migrated to Base UI.

## Changed

- `packages/ui/src/tabs.tsx` — replaced `@radix-ui/react-tabs` with `@base-ui/react/tabs`. Re-mapped `Trigger` -> `Tab` and `Content` -> `Panel`. Replaced `data-[state=active]` with `data-[active]`.
- Leftover sweep is clean.

## Left alone
None

## Behavior changes

- Tabs activation in Base UI defaults to manual (space/enter to activate), replacing Radix's default automatic activation on focus. Flagging this as intended behavior delta.

## Verify by hand

- Open dashboard settings or log details dialog containing tabs.
- Arrow key between tabs. Confirm they receive focus but require Enter/Space to select (manual activation).
