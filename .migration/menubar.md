# menubar

2026-07-08, transformation engine. Verdict: migrated menubar shell to Base UI Menubar + Menu.

## Changed

- `packages/ui/src/menubar.tsx:1-255` — replaced Radix menubar/menu parts with Base UI `Menubar` + `Menu` parts.
- `MenubarTrigger`, `MenubarItem`, and `MenubarSubTrigger` now support `render`; `asChild` stays compatibility only.
- `MenubarContent` / `MenubarSubContent` now compose Base UI `Portal > Positioner > Popup`.
- Leftover scan on component file clean: no `@radix-ui` import remains in `packages/ui/src/menubar.tsx`.

## Left alone

- `MenubarMenu`, `MenubarGroup`, `MenubarPortal`, `MenubarRadioGroup`, `MenubarSub` — thin re-exports around Base UI menu family.
- `MenubarShortcut` — unchanged helper span.

## Behavior changes

- Menubar now uses Base UI `loopFocus` semantics.
- Menu item close behavior follows Base UI defaults.

## Verify by hand

- Open menubar, move with arrow keys.
- Open submenu.
- Click checkbox/radio items.
