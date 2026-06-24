# Worktree Flow

This guide explains recommended branch and worktree strategy for parallel development streams in this repo.

## 1. Branch Model

Use this promotion path:

- `main` = production-ready
- `staging` = release candidate and pre-release integration
- `dev` = daily integration branch

Daily feature work should start from `dev`.

Default rule:

- if current branch is `dev`, `make wt-new feat/name` creates worktree from `dev`
- if current branch is something else, that branch becomes default base
- use second positional arg only when you want to override current branch, for example `make wt-new feat/name staging`

## 2. Why Worktrees

Use git worktrees when:

- frontend and backend move in parallel
- auth, tenant, or Casbin changes need isolated validation
- multiple feature streams run at same time
- one branch must stay open while another needs urgent work

Each worktree should have:

- its own git branch
- its own `.env.local`
- its own docker compose project name
- its own exposed local ports

Default root behavior:

- default worktree root is `.worktrees/` inside repo
- this is safer for restricted or sandboxed environments
- if you prefer sibling folders, pass `WORKTREE_ROOT=../Casbin-worktrees` explicitly

## 3. Recommended Stream Split

Example parallel streams:

- `feat/web-surface`
  - focus: `apps/web`, `apps/client`, `packages/*`
- `feat/auth-hardening`
  - focus: auth middleware, tenant resolution, Casbin rules
- `feat/api-boundary`
  - focus: route protection, API contracts, backend consumers

This split reduces conflicts and keeps review scope smaller.

## 4. Create New Worktree

From main repo checkout on `dev`:

```bash
make wt-new feat/web-surface
make wt-new feat/auth-hardening
```

Then move into each worktree and initialize local env:

```bash
make wt-new feat/web-surface
cd .worktrees/feat-web-surface
make dev-up
```

```bash
make wt-new feat/auth-hardening
cd .worktrees/feat-auth-hardening
make dev-up
```

`make wt-new` now bootstraps `.env.local` automatically.

Override base branch example:

```bash
make wt-new feat/auth-hardening staging
cd .worktrees/feat-auth-hardening
```

For existing worktrees, ensure env and local file state with:

```bash
make wt-enter feat/web-surface
```

To jump into worktree directly from current shell:

```bash
cd "$(make wt-path feat/web-surface)"
```

## 5. Daily Commands

Inspect current state:

```bash
make wt-list
make dev-status
make doctor
```

Keep env template aligned:

```bash
make env-sync
```

Run local migrations:

```bash
make migrate-up-local
```

Run narrow tests:

```bash
make test-local
make test-local TEST_PKG=./internal/modules/auth/...
```

Stop local stack:

```bash
make dev-down
```

## 6. Remove Worktree Safely

When feature work is merged or no longer needed:

```bash
make wt-rm feat/web-surface
```

Current behavior:

- stops local compose stack first when `.env.local` exists
- removes git worktree path
- does not auto-delete remote branch
- refuses to remove the currently active branch checkout

Clean stale git metadata:

```bash
make wt-prune
```

## 7. Merge Strategy

Recommended order:

1. feature branch merges into `dev`
2. integration and regression checks happen on `dev`
3. promote `dev` into `staging`
4. final release checks happen on `staging`
5. promote `staging` into `main`

## 8. Conflict Rules

If parallel streams touch the same API contract:

- settle contract shape first
- merge backend contract branch first if frontend depends on it
- rebase dependent frontend branch after contract merge
- do not let UI-only assumptions redefine backend rules

If streams touch auth, tenant, or Casbin boundaries:

- verify tenant scope in repository and usecase layers
- verify route protection still matches middleware order
- verify frontend consumers still match producer changes

## 9. Verification Expectations

At minimum per stream:

- positive case
- negative case
- edge case
- vulnerability or boundary case

Examples:

- positive: valid authenticated tenant route succeeds
- negative: unauthorized access is rejected
- edge: policy fallback behaves as expected
- vulnerability: cross-tenant access is rejected
