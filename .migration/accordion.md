# accordion

2026-07-08, engine, migrated to Base UI.

## Changed

- `packages/ui/src/accordion.tsx` — replaced `@radix-ui/react-accordion` with `@base-ui/react/accordion`. Mapped `Content` -> `Panel`, rewrote state selectors (`data-[state=open]` -> `data-open`). Rewrote `animate-accordion-up`/`down` to `data-starting-style` and `data-ending-style` transitions for Base UI. Added `aria-disabled` styles.
- `apps/web/src/app/[locale]/dashboard/access-rights/_components/access-rights-list.tsx` — updated call sites from `type="multiple"` to `multiple` boolean prop.
- `apps/web/src/components/landing/faq.tsx` — updated call site from `type="single" collapsible` to base array-less default state (both props removed).
- Leftover sweep on accordion wrapper is clean: no `radix-ui` / `@radix-ui` imports remain.

## Left alone
None

## Behavior changes

- Base UI accordion takes array values by default (though single strings may pass down). The `collapsible` and `type` props no longer exist on the root. Multiple expansion is driven by the `multiple` boolean prop.

## Verify by hand

- Click FAQ items on landing page. Ensure they expand and collapse.
- Open dashboard access rights. Expand a role, then expand an endpoint group. Ensure multiple groups can be open at once.
