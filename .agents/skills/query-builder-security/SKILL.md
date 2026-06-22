---
name: query-builder-security
description: Use when changing filtering, sorting, dynamic query fields, repository list endpoints, or `pkg/querybuilder` behavior in this Casbin repo, especially when field allowlists and sensitive-field protections must remain intact.
---

# Query Builder Security

## Overview

Dynamic filtering and sorting is security boundary here, not convenience helper only.

Any change can accidentally expose sensitive fields or unsafe query behavior.

## Read Order

1. `AGENTS.md`
2. `llm/cache/querybuilder-security.md`
3. `llm/conventions/database.md`
4. `pkg/querybuilder/query_builder.go`
5. `pkg/querybuilder/query_builder_test.go`
6. affected repository or list endpoint

## Iron Rules

- field names come from whitelisted struct metadata, never raw user input
- sensitive fields stay denied: password, token, secret, key, salt, and equivalents
- values use placeholders, not unsafe interpolation
- sort and filter convenience must not weaken security

## Workflow

### Step 1 — Identify Input Surface

State whether change affects:

- query params
- filter object
- sort field or direction
- repository list method

### Step 2 — Validate Metadata Semantics

Confirm:

- allowlist and denylist behavior still correct
- struct tags or field metadata are intentional
- sensitive fields remain unqueryable and unsortable

### Step 3 — Patch Narrowly

- keep security logic centralized in querybuilder package or clear repository boundary
- avoid one-off repository exceptions unless explicitly justified

### Step 4 — Test Positive And Negative Paths

Add or update tests for:

- allowed field
- denied sensitive field
- invalid or unknown field
- sort direction edge cases when relevant

## Common Mistakes

- adding repository shortcut that bypasses querybuilder allowlist logic
- exposing field because model tag looked harmless
- validating allowed field but not denied-field regression

## Stop Conditions

- stop if field exposure decision cannot be justified from business need and security model
- stop if repository path starts bypassing central querybuilder rules

## Completion Output

Report:

- input surface changed
- allow/deny implications
- files changed
- verification run and exact result
