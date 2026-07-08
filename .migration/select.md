# select

2026-07-08, transformation engine + consumer sweep. Verdict: migrated select wrapper to Base UI Select and fixed null-safe consumers.

## Changed

- `packages/ui/src/select.tsx:1-160` — replaced Radix select with `@base-ui/react/select` parts.
- `SelectContent` now uses `Portal > Positioner > Popup > List`; `SelectTrigger`, `SelectItem`, `SelectItemIndicator`, `SelectScrollUpButton`, and `SelectScrollDownButton` all moved to Base UI shapes.
- `apps/web/src/app/[locale]/dashboard/organization/members/_components/member-invite-dialog.tsx` — guarded `onValueChange` null before `setRoleId`.
- `apps/web/src/app/[locale]/dashboard/organization/members/_components/member-list.tsx` — guarded `onValueChange` null before update.
- `apps/client/app/features/organizations/member-role-selector.tsx` — guarded `onValueChange` null before `onChange`.
- Leftover scan on component file clean: no `@radix-ui` import remains in `packages/ui/src/select.tsx`.

## Left alone

- `SelectGroup`, `SelectValue`, `SelectSeparator` — native wrappers, kept simple.

## Behavior changes

- `onValueChange` can yield `null`; consumers now guard before writing to string state.
- `alignItemWithTrigger` is Base UI positioner prop; default preserved.
- `position` prop removed from wrapper API surface.

## Verify by hand

- Open select dropdowns in forms and tables.
- Choose option, ensure value updates.
- Confirm no crash when cleared/null value path hits.
