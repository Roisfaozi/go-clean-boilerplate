# radio-group

2026-07-08, transformation engine. Verdict: migrated radio-group wrapper to Base UI radio primitives.

## Changed

- `packages/ui/src/radio-group.tsx` — replaced `@radix-ui/react-radio-group` with `@base-ui/react/radio-group` and `@base-ui/react/radio`. Mapped root, item, and indicator.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Click radio buttons in a group, verify only one is selected and the indicator circle appears.
