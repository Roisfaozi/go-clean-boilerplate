---
name: login
description: Use when browser, E2E, smoke, or manual verification in this Casbin repo needs authenticated setup, especially when app ownership, session cookies, tenant selection, or role-specific access must be established before testing behavior.
---

# Login

## Overview

Authenticated verification in this repo is not only JWT presence.

Login setup must respect owning app, proxy path, backend session behavior, and tenant-sensitive access.

## When To Use

Use this skill when:

- browser or E2E flow requires signed-in user
- route behavior differs by role, membership, or tenant
- smoke test needs authenticated baseline

## Read Order

1. `AGENTS.md`
2. `llm/cache/authentication-system.md`
3. `llm/cache/tenant-organization-system.md` when tenant matters
4. `llm/cache/user-system.md` when role/profile matters
5. target app ownership paths in `apps/web` or `apps/client`
6. relevant backend auth routes or middleware when needed

## Login Workflow

### Step 1 — Identify Owning Surface

Decide whether flow belongs to:

- `apps/web`
- `apps/client`
- backend-only request verification

Check matching proxy path before assuming browser route behavior.

### Step 2 — Define Actor

State exact actor needed:

- unauthenticated public user
- authenticated user without tenant-sensitive permission
- tenant member
- admin or superadmin
- API-key actor instead of browser user

### Step 3 — Use Existing Stable Path

Prefer existing test, fixture, seeded account, or documented manual flow.

Do not invent alternate auth shortcuts that bypass real session or tenant behavior.

### Step 4 — Validate Session Context

Confirm at least relevant subset:

- login success path works
- cookies or headers are present where expected
- backend session rules are respected
- active tenant or organization context is correct when required

### Step 5 — Hand Off To Verification Skill

After login baseline is established:

- use `smoke-test` for broad quick checks
- use `e2e-test` for scripted flow verification

## Common Mistakes

- assuming JWT alone proves authenticated behavior
- using wrong frontend surface for route under test
- forgetting tenant switch or membership-dependent state
- testing admin path with normal user actor

## Stop Conditions

- stop if owning app is unclear
- stop if account/seed path is unknown and cannot be proven locally
- stop if auth failure may actually be proxy or session-layer issue

## Completion Output

Report:

- owning app or surface
- actor used
- login method used
- validated session or tenant context
- blockers for downstream verification
