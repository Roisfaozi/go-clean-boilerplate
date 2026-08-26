# toast

2026-07-08, transformation engine. Verdict: migrated toast wrapper and legacy toast hook from Radix Toast to Base UI Toast manager.

## Changed

- `packages/ui/src/toast.tsx` — replaced `@radix-ui/react-toast` with `@base-ui/react/toast` parts. Preserved custom `toastVariants`, viewport placement, title, description, action, and close wrapper exports.
- `packages/ui/src/toaster.tsx` — moved rendering to Base UI `Toast.Provider` with a shared `toastManager`, and renders toast objects from `Toast.useToastManager()`.
- `packages/ui/src/hooks/use-toast.ts` — replaced local reducer/timeout queue with Base UI `createToastManager()` while preserving exported `toast`, `toast.success`, `toast.error`, `toast.warning`, `useToast()`, and `dismiss()` APIs.
- `packages/ui/package.json` / `pnpm-lock.yaml` — removed `@radix-ui/react-toast`.
- Leftover scan on component files clean: no `@radix-ui/react-toast` import remains.

## Left alone

- `packages/ui/src/sonner.tsx` — third-party Sonner wrapper, intentionally untouched.
- `@radix-ui/react-dialog` dependency remains because `vaul` drawer types leak Radix Dialog declarations.

## Behavior changes

- Toast lifecycle now follows Base UI Toast manager behavior instead of the old local reducer queue.
- Toast action elements are converted to Base UI action props. Existing call shapes compile, but manual click-through should be checked.

## Verify by hand

- Trigger `toast({ title, description })` and confirm it appears.
- Trigger `toast.success`, `toast.error`, and `toast.warning`.
- Click close button and confirm dismissal.
- Trigger toast with action button and confirm action click works.
