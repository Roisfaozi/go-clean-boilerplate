# alert-dialog

2026-07-08, transformation engine. Verdict: migrated alert-dialog wrapper to Base UI alert-dialog primitive.

## Changed

- `packages/ui/src/alert-dialog.tsx` — replaced `@radix-ui/react-alert-dialog` with `@base-ui/react/alert-dialog`. Mapped Root → Root, Trigger → Trigger, Portal → Portal, Overlay → Backdrop, Content → Popup, Action → custom button render wrapper, Cancel → Close. Title and Description mapped to corresponding primitives.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

- `AlertDialogAction` now accepts a `render` prop instead of `asChild` for injecting custom button elements. Default returns a plain `<button>` with `buttonVariants()`.

## Verify by hand

- Open alert dialog, verify "Cancel" and "Action" buttons. Press Esc should close.
