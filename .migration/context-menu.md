# context-menu

2026-07-08, transformation engine. Verdict: migrated wrapper to Base UI Context Menu.

## Changed

- `packages/ui/src/context-menu.tsx:1-200` — replaced Radix context menu with `@base-ui/react/context-menu` parts.
- `ContextMenuContent` now uses `Portal > Positioner > Popup`; `ContextMenuSubContent` composes submenu defaults.
- `ContextMenuTrigger`, `ContextMenuItem`, and `ContextMenuSubTrigger` support `render` path while keeping `asChild` compatibility.
- Leftover scan on component file clean: no `@radix-ui` import remains in `packages/ui/src/context-menu.tsx`.

## Left alone

- `ContextMenuCheckboxItem` / `ContextMenuRadioItem` — kept indicator visuals.
- `ContextMenuShortcut` — unchanged.

## Behavior changes

- Base UI trigger/positioning model used now.
- Item close behavior follows Base UI defaults.

## Verify by hand

- Right-click area using context menu.
- Open submenu.
- Click checkbox/radio entries and confirm selection.
