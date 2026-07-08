# aspect-ratio

2026-07-08, engine, replaced Radix aspect ratio primitive with CSS-based wrapper.

## Changed

- `packages/ui/src/aspect-ratio.tsx` — removed `@radix-ui/react-aspect-ratio` and implemented a plain div wrapper using CSS padding-bottom to maintain ratio.
- Leftover sweep on aspect-ratio wrapper is clean: no `radix-ui` / `@radix-ui` imports remain in `packages/ui/src/aspect-ratio.tsx`.

## Left alone

- No app consumer files changed in this batch because there are currently no direct `AspectRatio` consumers in `apps/web` or `apps/client`.

## Behavior changes

- Implementation now uses a plain div + absolutely positioned inner container instead of a Radix primitive. Public `ratio` prop stays available.

## Verify by hand

- Render any future media card/image usage with non-1:1 ratios.
- Confirm child content fills the container and does not overflow unexpectedly.
- Resize viewport and confirm ratio stays stable.
