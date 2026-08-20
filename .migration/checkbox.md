# checkbox

2026-07-08, transformation engine. Verdict: migrated checkbox wrapper to Base UI checkbox primitive.

## Changed

- `packages/ui/src/checkbox.tsx` — replaced `@radix-ui/react-checkbox` with `@base-ui/react/checkbox`. Mapped root and indicator, preserving all classes.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Click checkbox, verify it toggles and the checkmark icon appears when checked.
