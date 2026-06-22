---
name: smoke-test
description: Use when running a quick end-to-end confidence check in this Casbin repo after changes, especially to confirm core pages, routes, auth gates, proxies, or high-value flows still work without doing a full E2E suite.
---

# Smoke Test

## Overview

Smoke test is fast confidence pass, not vague clicking.

It should prove core route or page behavior relevant to changed files and report exact surfaces checked.

## When To Use

Use this skill when:

- user wants quick confidence after patch
- route, frontend proxy, page, or auth gate changed
- full E2E is too heavy for current iteration

Do not use this skill as substitute for integration/E2E when deep boundary changed.

## Read Order

1. current diff or changed file list
2. `AGENTS.md`
3. target workflow or domain cache
4. owning app or backend route path
5. `login` skill if auth is required

## Smoke Workflow

### Step 1 — Choose Exact Paths

Pick only high-value paths affected by change, such as:

- changed backend endpoint
- changed frontend page
- changed proxy route
- changed upload or realtime handshake path

### Step 2 — Establish Required State

- unauthenticated baseline if route can be public
- authenticated baseline with `login` if needed
- tenant-specific context if route depends on organization membership

### Step 3 — Execute Quick Check

Check exact visible outcomes:

- page loads or request returns expected class of response
- auth gate behaves correctly
- no obvious server error or frontend crash
- navigation or proxy path still resolves

### Step 4 — Record What Was Not Checked

Be explicit about skipped coverage:

- no broad E2E suite run
- no Docker-backed integration run
- no multi-role matrix if not performed

## Common Targets In This Repo

- routes registered through `internal/router/router.go`
- frontend proxies in `apps/web/src/app/api/v1/[...path]/route.ts`
- frontend proxies in `apps/client/app/routes/api-proxy.ts`
- auth or tenant sensitive pages in active app owner

## Stop Conditions

- stop if target surface cannot be run locally
- stop if auth baseline is missing and route is protected
- stop if exact path under test is still unclear from diff

## Completion Output

Report:

- exact paths checked
- actor used
- observed result
- skipped coverage and blocker
