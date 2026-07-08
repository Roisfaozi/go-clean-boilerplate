# navigation-menu

2026-07-08, transformation engine. Verdict: migrated navigation menu wrapper to Base UI Navigation Menu.

## Changed

- `packages/ui/src/navigation-menu.tsx:1-127` — replaced Radix navigation menu with Base UI `NavigationMenu` parts.
- `NavigationMenuTrigger` now uses Base UI `Icon` as chevron and `render` compatibility for trigger composition.
- `NavigationMenuViewport` now composes `Portal > Positioner > Popup > Viewport`.
- `NavigationMenuIndicator` is now an `Icon` shim, matching Base UI pattern.
- Leftover scan on component file clean: no `@radix-ui` import remains in `packages/ui/src/navigation-menu.tsx`.

## Left alone

- `navigationMenuTriggerStyle` — preserved styling helper.
- `NavigationMenuLink` and `NavigationMenuItem` — thin re-exports / aliases.

## Behavior changes

- Radix viewport model replaced by anchored popup model.
- Trigger open state uses Base UI popup markers.

## Verify by hand

- Hover/click top nav items.
- Confirm viewport panel positions under trigger.
- Confirm indicator chevron rotates.
