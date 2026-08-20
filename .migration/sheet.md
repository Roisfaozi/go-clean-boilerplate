# sheet

2026-07-08, transformation engine. Verdict: migrated sheet wrapper to Base UI dialog primitive (used as sheet).

## Changed

- `packages/ui/src/sheet.tsx` — replaced `@radix-ui/react-dialog` with `@base-ui/react/dialog` (SheetPrimitive is an alias). Mapped Root → Root, Trigger → Trigger, Portal → Portal, Overlay → Backdrop, Content → Popup, Close → Close, Title → Title, Description → Description. Preserved custom `sheetVariants` CVA for side animations.
- Verified leftover scan is clean.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Open sheet from each side, verify slide-in animation. Press Esc to close.
