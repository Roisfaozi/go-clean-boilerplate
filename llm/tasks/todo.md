# Todo

Status:

- Parity audit complete
- Gap implementation complete
- Skills parity complete
- Final self-audit complete

Current plan:

1. audit template-vs-repo gaps file-by-file
2. patch repo-owned workflow/lifecycle gaps inside writable workspace
3. restore starter-pack parity for skills
4. run final self-audit for paths, commands, workflow sections, and trigger clarity

Completed in this pass:

- audited starter-pack `codex-ready/typescript-go` against current repo structure
- confirmed `llm/cache/*` and `llm/conventions/*` already exceed template depth
- added missing lifecycle folders and README files
- completed workflow section parity for verification/review/stop conditions
- added starter-pack core skills adapted to this repo under `.agents/skills/`
- preserved existing repo-specific skills (`vercel-*`, `web-design-guidelines`, `workflows/bmad`) because they are additive, not conflicting
- completed final parity self-audit across AGENTS, workflows, lifecycle folders, and skills

## Skill quality audit follow-up

Status: complete.

Completed:

- verified project-specific skills now exist under `.agents/skills/`
- verified core skills were tightened for Casbin repo boundaries
- documented final audit result in `llm/tasks/skill-quality-audit.md`
- noted Vercel skill Prisma/Drizzle mentions are generic frontend guidance only, not backend stack guidance

## Carbon-style skill detail follow-up

Status: complete.

Completed:

- analyzed Carbon skill structure and detail pattern
- generated and manually applied detailed Casbin skills
- verified detailed skills contain announce line, read order, phased workflow, stop conditions, and completion output where applicable
- verified no Carbon/Supabase/Kysely/Biome/ERP/MES assumptions were introduced
- documented result in `llm/tasks/carbon-skill-pattern-audit.md`

## Carbon vs Casbin re-analysis

Status: complete.

Completed:

- compared Carbon `.claude/skills` detail level against current Casbin `.agents/skills`
- compared cache granularity: Carbon 42 cache files vs Casbin 8 cache files
- identified remaining maturity gaps: cache split, orchestration depth, debugging depth, forms examples, verification anti-rationalization
- documented result in `llm/tasks/carbon-vs-casbin-skill-analysis.md`

## Remaining maturity gap closure

Status: complete for writable repo files; `.agents/skills` regeneration remains manual due read-only mount for Codex.

Completed:

- created verified granular cache files for auth, tenant/org, Casbin permission, API-key, TUS upload, worker/audit/webhook, querybuilder security, realtime, and frontend proxy/forms
- updated `AGENTS.md` routing to reference new granular cache files
- verified referenced paths exist
- verified no Carbon-specific or wrong-stack assumptions in new cache files
- retained `llm/tasks/write-detailed-casbin-skills.py` as manual regeneration path for Carbon-style detailed skills

## Final gap closure

Status: complete.

Completed:

- added module-level cache files for user, role, project, access-right, and stats systems
- updated `AGENTS.md` routing so module-specific work reads matching cache files first
- extended `llm/tasks/write-detailed-casbin-skills.py` with module-specific skills: `user-domain`, `role-domain`, `project-domain`, `access-domain`, and `stats-domain`
- refreshed `llm/tasks/carbon-vs-casbin-skill-analysis.md` to mark current parity gaps closed
- verified new docs reference existing live paths and contain no wrong-stack assumptions

## Latest closure

Status: complete for writable docs; manual `.agents/skills` regeneration still required.

Added in this pass:

- `llm/cache/permission-system.md`
- `llm/cache/audit-system.md`
- `llm/cache/webhook-system.md`
- module skill entries in `llm/tasks/write-detailed-casbin-skills.py` for `permission-domain`, `audit-domain`, and `webhook-domain`

Manual next step if user wants `.agents/skills` files refreshed:

```bash
python3 llm/tasks/write-detailed-casbin-skills.py
```
