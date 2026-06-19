# Bugfix Workflow

## Purpose

Workflow ini untuk memperbaiki incorrect behavior, regression, security drift, atau docs/runtime mismatch tanpa berubah menjadi feature creep.

## Use when

- fixing incorrect behavior in existing code
- closing regressions found by tests, runtime audit, or security review
- correcting docs or cache drift against live code
- tightening a specific broken path without redesigning whole feature

## Read first

1. relevant cache files for affected layer
2. failing tests or runtime evidence near affected behavior
3. `llm/conventions/testing.md`
4. `AGENTS.md` high-risk rules if bug touches auth, tenant, Casbin, API key, upload, realtime, or worker paths

## Live code to inspect

- failing test or exact route path
- `internal/config/app.go` when dependency lifecycle is involved
- `internal/router/router.go` if behavior is HTTP-visible
- target module controller/usecase/repository
- frontend proxy or client code if bug crosses browser/backend boundary

## Workflow phases

### Phase 1 — State observed bug

Document:

- observed behavior
- expected behavior
- exact path, actor, or layer affected
- whether bug is reproducible now

### Phase 2 — Reproduce or trace

Prefer:

- failing test
- focused request path
- minimal browser path
- exact worker or async path

Do not patch from memory or assumption.

### Phase 3 — Localize owner layer

Decide whether bug owner is:

- route or middleware
- controller parsing/validation
- usecase business rule
- repository/persistence
- frontend proxy or consumer
- worker or realtime side effect

### Phase 4 — Patch minimal root cause

- fix smallest owner layer that actually caused bug
- add or adjust regression test when meaningful seam exists
- avoid unrelated cleanup

### Phase 5 — Verify fix and adjacent risk

- rerun exact reproduction first
- then run adjacent narrow verification
- widen to integration/E2E only when bug crosses boundary

## Review checklist

- root cause actually addressed
- no broader behavior silently changed
- regression coverage added where practical
- docs/cache updated only if stable truth was clarified

## Stop conditions / needs confirmation

- cannot reproduce or trace enough to isolate likely owner
- multiple root causes remain plausible with no evidence to choose one
- fix would require product decision, not technical correction only
