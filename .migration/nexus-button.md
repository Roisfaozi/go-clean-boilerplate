# nexus-button

2026-07-08, transformation engine. Verdict: migrated brand button off Radix Slot to Base UI render flow.

## Changed

- `packages/ui/src/nexus-button.tsx:1-75` — replaced `@radix-ui/react-slot` with `@base-ui/react/use-render` + `@base-ui/react/merge-props`.
- Kept custom variants, sizing, loading spinner, and public `asChild` compatibility.
- Leftover scan on component file clean: no `@radix-ui` import remains in `packages/ui/src/nexus-button.tsx`.

## Left alone

- `Button` primitive — separate component, already Base UI.
- Consumers of `NexusButton` — no call-site sweep needed; component stayed backward compatible.

## Behavior changes

- `NexusButton` now prefers `render`; `asChild` stays as compatibility shim.
- Loading state still disables button and shows spinner.

## Verify by hand

- Click ghost/outline/primary Nexus buttons.
- Use `loading` prop and confirm button disables.
- Use `render={<a ... />}` path and confirm link composition works.
