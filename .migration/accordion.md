# accordion

2026-07-08, transformation engine. Verdict: migrated accordion wrapper to Base UI accordion primitive.

## Changed

- `packages/ui/src/accordion.tsx` — replaced radix accordion parts with `@base-ui/react/accordion`. Mapped Root → Root, Item → Item, Trigger → Header + Trigger, Content → Backdrop + Panel.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Click accordion item, verify expand/collapse animation.
