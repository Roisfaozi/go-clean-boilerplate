# project

2026-07-08, full whole-project Base UI migration. Verdict: all Radix-eligible wrappers migrated to `@base-ui/react`, all consumer apps updated, all Radix direct dependencies removed.

## Changed

- Migrated all 33 Radix UI wrappers in `packages/ui/src` to `@base-ui/react`:
  - **Foundational**: button, label, separator, aspect-ratio
  - **Disclosure/display**: accordion, collapsible, avatar, progress, tabs, scroll-area
  - **Form controls**: checkbox, radio-group, switch, slider, toggle, toggle-group
  - **Overlays**: dialog, alert-dialog, sheet, popover, hover-card, tooltip
  - **Menus/nav**: dropdown-menu, context-menu, menubar, select, navigation-menu
  - **Internal**: breadcrumb, nexus-button, form, sidebar
  - **Toast**: toast, toaster, use-toast
- Migrated `asChild` → `render` prop and consumer call-site props across `apps/web` and `apps/client`.
- Updated positioning, state attributes, and CSS class names per class-mapping guides.
- Removed 27 `@radix-ui/react-*` direct dependencies from `apps/web/package.json`.
- Removed `@radix-ui/react-dialog` last direct dep from `packages/ui/package.json`.
- Fixed drawer type annotations to suppress transitive vaul → Radix type leakage.

## Left alone

- Third-party wrappers intentionally untouched: `command` (cmdk), `drawer` (vaul), `sonner`, `input-otp`, `calendar` (react-day-picker), `chart` (recharts).
- Transitive Radix packages remain in `pnpm-lock.yaml` only via vaul dependency.

## Behavior changes

- Radix `data-[state=open]` / `data-[state=closed]` → Base UI `data-open` / `data-closed`.
- Select callbacks now allow `null` values; consumers wrapped with null guards.
- DropdownMenuCheckboxItem/DropdownMenuRadioItem close behavior follows Base UI defaults.
- Toast lifecycle now uses Base UI Toast manager instead of custom reducer.
- Drawer types explicitly cast to suppress transitive vaul→Radix type emissions.

## Verify by hand

- Open all overlays (dialog, sheet, popover, tooltip, hover-card, context-menu, dropdown-menu).
- Verify form controls (checkbox, radio, switch, slider, toggle) in both web and client apps.
- Test keyboard navigation on menus, select, tabs, accordion.
- Trigger toast notifications and verify close/success/error/destructive variants.
