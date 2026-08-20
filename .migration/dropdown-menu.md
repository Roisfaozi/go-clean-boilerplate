# dropdown-menu

2026-07-08, transformation engine + consumer sweep. Verdict: migrated wrapper to Base UI Menu and moved trigger consumers to `render`.

## Changed

- `packages/ui/src/dropdown-menu.tsx:1-200` — replaced Radix dropdown menu with `@base-ui/react/menu` parts.
- `DropdownMenuContent` now composes `Portal > Positioner > Popup`; `DropdownMenuSubContent` composes submenu defaults from the wrapper.
- `DropdownMenuTrigger`, `DropdownMenuItem`, and `DropdownMenuSubTrigger` support `render` path; `asChild` remains compatibility only.
- App consumers updated:
  - `apps/client/app/components/layout/app-navbar.tsx`
  - `apps/client/app/features/organizations/members-table.tsx`
  - `apps/client/app/features/shared/crud-table.tsx`
  - `apps/web/src/app/[locale]/dashboard/users/_components/users-toolbar.tsx`
  - `apps/web/src/components/dashboard/user-nav.tsx`
  - `apps/web/src/components/dashboard/users/user-table.tsx`
  - `apps/web/src/components/shared/locale-toggler.tsx`
  - `apps/web/src/components/shared/theme-toggle.tsx`
- Leftover scan on component file clean: no `@radix-ui` import remains in `packages/ui/src/dropdown-menu.tsx`.

## Left alone

- `DropdownMenuCheckboxItem` / `DropdownMenuRadioItem` indicator layout — kept same visuals.
- `DropdownMenuShortcut` — native helper span, unchanged.

## Behavior changes

- Menu item close-on-click behavior now follows Base UI defaults.
- `forceMount` consumer usage removed from app code; no wrapper prop kept.
- `asChild` call sites now use `render`.

## Verify by hand

- Open dropdown menus in web/client.
- Check trigger composition on icon buttons and avatar button.
- Click item, checkbox item, radio item, and submenu.
