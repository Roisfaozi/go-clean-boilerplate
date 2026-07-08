# tabs

2026-07-08, transformation engine. Verdict: migrated tabs wrapper to Base UI tabs primitive.

## Changed

- `packages/ui/src/tabs.tsx` — replaced `@radix-ui/react-tabs` with `@base-ui/react/tabs`. Mapped Root → Root, List → TabList, Trigger → Tab, Content → TabPanel.

## Left alone

None

## Behavior changes

None

## Verify by hand

- Click tab, verify panel switches. Keyboard navigate with arrow keys.
