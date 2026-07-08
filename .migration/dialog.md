# dialog

2026-07-08, transformation engine. Verdict: migrated dialog wrapper to Base UI dialog primitive.

## Changed

- `packages/ui/src/dialog.tsx` — replaced `@radix-ui/react-dialog` with `@base-ui/react/dialog`. Mapped Root → Root, Trigger → Trigger, Portal → Portal, Overlay → Backdrop, Content → Popup, Close → Close, Title → Title, Description → Description.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Open dialog, verify backdrop overlay. Press Esc to close. Click outside to close.
