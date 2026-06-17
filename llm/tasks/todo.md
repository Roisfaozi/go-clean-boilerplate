# Todo

Status:

- Phase 1 complete
- Phase 2 complete
- Phase 3 complete
- Phase 4 complete
- Phase 5 complete
- Phase 6 complete
- Finalization complete

Completed work:

- audited structure and toolchain
- audited architecture, route strata, module boundaries, frontend entrypoints, and database migration surface
- audited domain rules from auth, tenant, permission, API key, TUS, realtime, query-builder, and sentinel learnings
- audited conventions for Go, database, testing, and TypeScript surfaces
- concretized repo workflows for feature, bugfix, API endpoint, Go service, cross-stack change, and database migration
- finalized `AGENTS.md` so it now points agents to the `llm/` starter-pack as the main durable repo context
- completed deep quality pass across cache, convention, workflow, and task files
- expanded remaining thin files into actionable repo-specific guidance

Phase 6 self-audit result:

- core llm cache, convention, workflow, and task files are present
- required commands referenced in workflows map to package scripts, Makefile targets, or CI commands found in repo
- required core paths referenced in cache/workflow files exist in repo
- no unresolved `needs confirmation` markers remain in current phase files
- final AGENTS entrypoint is aligned with the concretized starter-pack
- detailed pass covers frontend/backend boundaries, domain pitfalls, module dependencies, env ownership, testing strategy, and workflow guardrails
