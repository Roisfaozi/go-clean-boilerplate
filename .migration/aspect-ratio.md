# aspect-ratio

2026-07-08, transformation engine. Verdict: migrated to native CSS aspect-ratio div.

## Changed

- `packages/ui/src/aspect-ratio.tsx` — replaced `@radix-ui/react-aspect-ratio` with plain CSS `aspect-ratio` on a `<div>` wrapper.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Render aspect-ratio container with child content, verify ratio preserved.
