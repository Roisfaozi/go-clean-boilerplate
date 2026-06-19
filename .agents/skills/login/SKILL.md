---
name: login
description: Use when browser, E2E, or smoke verification needs authenticated setup.
---

# Login

## Read First

- apps/web or apps/client auth flow

## Workflow

1. Identify which app owns flow.
2. Prefer existing test or manual login pattern.
3. Keep auth flow aligned with backend session and cookie behavior.
4. Record reusable steps in `llm/test-playbooks/` if stable.
