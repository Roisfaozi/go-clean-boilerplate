---
name: self-review
description: Use when Casbin repo changes are complete and need maintainability, correctness, security-boundary, and verification review before handoff.
---

# Self Review: Staff-Level Handoff Check

**Announce at start:** "I'm using self-review before handoff."

## Review Dimensions

### Layering

- controller only parses/validates/responds
- usecase owns business rules
- repository owns persistence details
- app config/router own wiring/protection

### Security Boundaries

- Redis session validation preserved
- tenant org boundary preserved
- Casbin policy/enforcer semantics preserved
- API-key scopes preserved
- querybuilder sensitive fields protected

### Side Effects

- worker/audit/webhook behavior intentional
- upload hooks intentional
- realtime ticket/origin/presence behavior intentional

### Contract

- backend response shape matches frontend consumers
- both `apps/web` and `apps/client` checked when relevant
- Swagger/shared types updated when required

### Verification

- exact commands run
- skipped checks explained
- unrelated failures separated

## Output Format

- Problems found and fixed
- Remaining risks
- Verification summary
