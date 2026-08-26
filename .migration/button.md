# button

2026-07-08, engine, migrated wrapper to Base UI button and updated remaining `asChild` consumer.

## Changed

- `packages/ui/src/button.tsx` — replaced `@radix-ui/react-slot` usage with `@base-ui/react/button`. Preserved `buttonVariants`, density-aware classes, and custom variants (`magic`, `outline-solid`). The wrapper now exposes Base UI's native `render` prop instead of Radix `asChild`.
- `apps/web/src/components/magicui/bento-grid.tsx` — converted the only `Button asChild` consumer to Base UI composition: `render={<a href={href} />}`.
- Leftover sweep on button wrapper is clean: no `radix-ui` / `@radix-ui` imports remain in `packages/ui/src/button.tsx`.

## Left alone

- `packages/ui/src/nexus-button.tsx` — still uses `@radix-ui/react-slot`. This is a custom brand component, not part of the foundational wrapper batch.
- All non-button `asChild` consumers in dialogs, menus, popovers, and tooltips — intentionally left for their own wrapper migrations.

## Behavior changes

- `Button` call sites must use Base UI `render` instead of Radix `asChild`. Only one direct app consumer existed in this batch and was updated.
- The wrapper no longer accepts `asChild` as a public prop; any future call site must use `render`.

## Verify by hand

- Open landing page / Bento card section and confirm CTA button still renders as link and navigates.
- Click normal primary/outline/ghost buttons in web and client apps.
- Check focus ring still appears on keyboard focus.
- Confirm disabled buttons still show reduced opacity and ignore pointer events.
