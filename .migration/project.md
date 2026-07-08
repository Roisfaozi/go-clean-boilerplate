# project

2026-07-08, partial whole-project Base UI migration. Verdict: core wrapper batch migrated and consumer sweep completed for changed triggers.

## Changed

- Migrated `breadcrumb`, `nexus-button`, `form`, `sidebar`, `dropdown-menu`, `context-menu`, `menubar`, `select`, and `navigation-menu` off Radix primitives.
- Updated consumer call sites in client/web apps from `asChild` to Base UI `render` where needed.
- Fixed Base UI `Select` nullability at consumer sites.
- Verified `pnpm --filter @casbin/ui typecheck`, `pnpm --filter casbin-web typecheck`, and `pnpm --filter casbin-client typecheck` all pass.

## Left alone

- `packages/ui/src/toast.tsx` — still Radix toast wrapper; not migrated yet.
- Third-party wrappers intentionally untouched: `command`, `drawer`, `sonner`, `input-otp`, `calendar`, `chart`.

## Behavior changes

- `DropdownMenuCheckboxItem` / `DropdownMenuRadioItem` close behavior now follows Base UI defaults.
- `Select` callbacks now allow `null`.
- `render` is preferred on Base UI-bridged wrappers; `asChild` only compatibility.

## Verify by hand

- Open menus, selects, nav menu, and breadcrumb links in both apps.
- Confirm sidebar buttons, forms, and Nexus buttons still render.
- Click logout/avatar/theme/locale controls.
