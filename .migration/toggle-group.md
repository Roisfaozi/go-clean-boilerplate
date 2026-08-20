# toggle-group

2026-07-08, transformation engine. Verdict: migrated toggle-group wrapper to Base UI toggle-group primitives.

## Changed

- `packages/ui/src/toggle-group.tsx` — replaced `@radix-ui/react-toggle-group` with `@base-ui/react/toggle-group` and base toggle.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Click items in a toggle group (single/multiple), verify correct selection state.
