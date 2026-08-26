# sidebar

2026-07-08, transformation engine. Verdict: migrated Slot users in sidebar to Base UI render flow, preserved sidebar API shape.

## Changed

- `packages/ui/src/sidebar.tsx:1-763` — replaced `@radix-ui/react-slot` usage in `SidebarGroupLabel`, `SidebarGroupAction`, `SidebarMenuButton`, `SidebarMenuAction`, and `SidebarMenuSubButton` with Base UI `useRender` + `mergeProps`.
- Kept existing button/input/sheet/tooltip layout and state data-attrs.
- Leftover scan on component file clean: no `@radix-ui` import remains in `packages/ui/src/sidebar.tsx`.

## Left alone

- Rest of sidebar structure — not Radix-specific, left intact.
- `SidebarTrigger` uses Base UI `Button` primitive already.

## Behavior changes

- `render` now preferred for polymorphic sidebar subparts.
- `asChild` compatibility remains on affected sidebar parts.

## Verify by hand

- Toggle sidebar collapse/expand.
- Open group labels/actions in expanded mode.
- Confirm menu buttons still navigate and tooltip still shows in collapsed mode.
