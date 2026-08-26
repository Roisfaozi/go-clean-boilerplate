# popover

2026-07-08, transformation engine. Verdict: migrated popover wrapper to Base UI popover primitive.

## Changed

- `packages/ui/src/popover.tsx` — replaced `@radix-ui/react-popover` with `@base-ui/react/popover`. Mapped Root → Root, Trigger → Trigger, Content → Portal + Positioner + Popup. Preserved positioning props forwarded from Positioner. Used `--transform-origin` CSS var.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Trigger popover, verify it positions correctly near trigger. Click outside to close.
