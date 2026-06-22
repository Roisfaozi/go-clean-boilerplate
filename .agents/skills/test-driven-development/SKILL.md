---
name: test-driven-development
description: Use when implementing a feature or bugfix in this Casbin repo with a meaningful test seam, especially in Go packages, route behavior, middleware logic, or frontend flows where failing-first evidence will reduce risk.
---

# Test Driven Development

## Overview

TDD here means failing-first where repo has real seam.

Do not force fake TDD where harness or architecture makes it noise. Use it when a narrow failing test can meaningfully pin behavior before patching.

## When To Use

Use this skill when:

- backend usecase, repository, or middleware behavior has local test seam
- bugfix can be reproduced by package or route-focused test
- contract or validation behavior can be captured with clear failing assertion

Do not force TDD when:

- only realistic proof is broader manual/E2E flow
- existing harness does not support practical narrow seam

## Read Order

1. `AGENTS.md`
2. target package tests and adjacent patterns
3. relevant `llm/cache/*`
4. relevant workflow file
5. live code owning behavior

## Repo Test Seams

- Go usecase, repository, middleware: package-local tests
- route, auth, tenant, Casbin, API-key: integration or route-focused tests when needed
- querybuilder: `pkg/querybuilder/query_builder_test.go`
- upload and storage: `pkg/tus/*_test.go`, storage/package tests, integration if boundary crosses hooks
- frontend: typecheck plus existing E2E or local component/page test patterns if present

## TDD Workflow

### RED

- write smallest failing test that captures desired behavior
- confirm failure is expected behavior failure, not setup or compile noise

### GREEN

- implement minimal code to pass test
- avoid extra cleanup or speculative enhancements

### REFACTOR

- remove duplication or sharpen names only after green
- keep test suite green during cleanup

## Test Design Rules

- test one behavior per failing case when possible
- use repo-adjacent test style from same package
- prefer boundary-relevant assertions over implementation-detail assertions
- when security boundary changes, include negative-path tests too

## Exceptions

If no practical test seam exists, document:

- why no test was added
- what proof path was used instead
- why alternative seam would be misleading or too costly

## Common Mistakes

- writing too-broad failing test that obscures root cause
- treating compile/setup failure as RED success
- adding implementation before proving test expresses desired behavior
- skipping negative cases for auth, tenant, or scope rules

## Stop Conditions

- stop if failing test cannot isolate intended behavior
- stop if realistic seam is integration/E2E/manual and unit TDD would be fake confidence

## Completion Output

Report:

- failing-first test added or updated
- behavior captured
- files changed
- commands run and exact result
- exception note if no practical seam existed
