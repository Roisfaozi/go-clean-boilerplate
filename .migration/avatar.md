# avatar

2026-07-08, transformation engine. Verdict: migrated avatar wrapper to Base UI avatar primitive.

## Changed

- `packages/ui/src/avatar.tsx` — replaced `@radix-ui/react-avatar` with `@base-ui/react/avatar`. Mapped Root → Root, Image → Image, Fallback → Fallback.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Render avatar with valid image URL, verify image loads. Without image, verify fallback initials shown.
