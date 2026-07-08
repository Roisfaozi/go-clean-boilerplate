# separator

2026-07-08, engine, migrated separator to Base UI separator.

## Changed

- `packages/ui/src/separator.tsx` — replaced `@radix-ui/react-separator` with `@base-ui/react/separator`. Preserved orientation handling and shared `bg-border` classes.
- Removed Radix-only `decorative` prop because Base UI separator does not support it.
- Leftover sweep on separator wrapper is clean: no `radix-ui` / `@radix-ui` imports remain in `packages/ui/src/separator.tsx`.

## Left alone

- No related non-wrapper files needed changes in this batch because there were no `decorative` consumers in app code.

## Behavior changes

- `decorative` prop is dropped at wrapper level. Existing consumers in this repo did not use it.

## Verify by hand

- Check auth pages and dashboard header for horizontal/vertical separators.
- Confirm vertical separators still size correctly in header/toolbars.
- Confirm no layout shift around divider lines.
