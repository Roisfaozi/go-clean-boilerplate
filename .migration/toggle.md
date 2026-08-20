# toggle

2026-07-08, transformation engine. Verdict: migrated toggle wrapper to Base UI toggle primitive.

## Changed

- `packages/ui/src/toggle.tsx` — replaced `@radix-ui/react-toggle` with `@base-ui/react/toggle`. 
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Click toggle button, verify styling changes between pressed/unpressed states.
