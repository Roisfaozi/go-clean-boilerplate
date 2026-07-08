# scroll-area

2026-07-08, engine, migrated to Base UI.

## Changed

- `packages/ui/src/scroll-area.tsx` — replaced `@radix-ui/react-scroll-area` with `@base-ui/react/scroll-area`. Re-mapped `ScrollAreaScrollbar` -> `Scrollbar` and `ScrollAreaThumb` -> `Thumb`.
- Leftover sweep is clean.

## Left alone
None

## Behavior changes
None

## Verify by hand

- Open notification popover or sidebar and scroll.
- Confirm scrollbar thumb appears and can be dragged.
