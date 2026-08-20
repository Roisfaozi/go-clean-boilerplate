# label

2026-07-08, engine, replaced Radix label primitive with native label.

## Changed

- `packages/ui/src/label.tsx` — removed `@radix-ui/react-label` and now renders a native `<label>` element. Preserved typography and disabled-state classes.
- Leftover sweep on label wrapper is clean: no `radix-ui` / `@radix-ui` imports remain in `packages/ui/src/label.tsx`.

## Left alone

- `packages/ui/src/form.tsx` — still imports Radix label types and will be migrated in a later form-wrapper batch.

## Behavior changes

- No intended behavior change. The wrapper still exposes the same `htmlFor`-style label API, now via native DOM.

## Verify by hand

- Check login, register, forgot-password, and settings forms.
- Click labels and confirm focus moves to the associated input.
- Confirm disabled/readonly form visuals still look correct.
