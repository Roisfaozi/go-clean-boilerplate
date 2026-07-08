# switch

2026-07-08, transformation engine. Verdict: migrated switch wrapper to Base UI switch primitive.

## Changed

- `packages/ui/src/switch.tsx` — replaced `@radix-ui/react-switch` with `@base-ui/react/switch`. Mapped root and thumb.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Toggle switch, verify thumb moves and background color changes.
