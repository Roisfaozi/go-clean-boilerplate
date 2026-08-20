# form

2026-07-08, transformation engine. Verdict: removed Radix Slot from form control wrapper, kept field semantics.

## Changed

- `packages/ui/src/form.tsx:1-177` — replaced `@radix-ui/react-slot` with Base UI `useRender` + `mergeProps` for `FormControl`.
- `FormControl` still wires `id`, `aria-describedby`, and `aria-invalid` from `useFormField`.
- Leftover scan on component file clean: no `@radix-ui` import remains in `packages/ui/src/form.tsx`.

## Left alone

- `Form`, `FormField`, `FormItem`, `FormLabel`, `FormDescription`, `FormMessage` — unchanged behavior.
- `Label` component — already native/Base UI aligned.

## Behavior changes

- `FormControl` now expects `render` for polymorphic composition; `asChild` path is no longer the primary model.
- Accessibility wiring unchanged.

## Verify by hand

- Open a form with error state.
- Confirm label, description, and error message still wire through.
- Confirm wrapped input still receives `aria-*` props.
