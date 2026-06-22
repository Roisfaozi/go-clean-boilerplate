# Bugfix Workflow

## Purpose

Primary workflow for fixing incorrect behavior, regressions, security drift, or docs/runtime mismatch without turning the work into feature creep.

## Use When

- fixing incorrect behavior in existing code
- closing regression found by tests, runtime audit, or user report
- tightening broken security path
- correcting docs or cache drift against live code

## Do Not Use When

- request is new capability, not repair
- requested change implies redesign or re-architecture unless bug root cause demands it

## Required Read Order

1. relevant cache files for affected layer
2. failing tests or exact runtime evidence
3. `llm/conventions/testing.md`
4. `AGENTS.md` high-risk rules if auth, tenant, Casbin, API-key, upload, realtime, or worker involved
5. `llm/workflows/benchmarking.md` if fix changes performance-sensitive logic or refactors hot path

## Live Code to Inspect

- exact failing route, module, test, or error path
- `internal/config/app.go` when lifecycle/dependency involved
- `internal/router/router.go` if behavior is HTTP-visible
- target controller/usecase/repository
- frontend proxy/client if bug crosses browser/backend boundary

## Workflow Steps

### Step 1 — Reproduce or Prove Symptom

Collect one of:

- failing test
- exact route behavior
- log/error text
- source-backed docs/runtime contradiction

No bugfix should start from vague hunch alone.

### Step 2 — Find Root Layer

Classify bug location:

- route registration / middleware order
- request validation / parsing
- usecase logic
- repository / query / transaction
- worker async timing
- proxy/frontend consumer mismatch

### Step 3 — Fix Root Cause, Not Only Output Symptom

If bugfix includes refactor or logic rewrite on a measurable hot path, perform benchmark audit before changing code.

Patch smallest root owner that explains failure.

Avoid:

- redundant controller guards when middleware truth is broken
- frontend-only masking for backend auth bug
- test-only patch that leaves runtime wrong

### Step 4 — Contain Scope

Before finalizing, ask:

- did patch accidentally broaden to refactor?
- are unrelated style/cleanup edits creeping in?
- does bug reveal adjacent risk that should only be documented, not fixed now?

## Common Mistakes

- fixing output but not root cause
- changing route stratum because one endpoint failed
- adding duplicate auth/tenant checks in handlers
- masking bug in frontend while backend remains wrong
- “while here” edits unrelated to reported symptom

## Verification

### Minimum

- narrowest failing test or package check

### Add When Needed

- integration check if bug crosses DB/Redis/Casbin/tenant boundary
- E2E/manual route check if browser-visible behavior changed
- docs/cache refresh if fix resolves live-code drift

## Stop Conditions

Stop and mark `needs confirmation` if:

- bug symptom not reproducible and code evidence weak
- fix would require destructive migration or broad redesign not requested
- root cause sits in ambiguous business-rule area without stable source of truth

## Completion Output

Report:

- symptom proved
- root cause layer
- exact fix surface
- verification run
- residual risk if not fully reproducible locally
