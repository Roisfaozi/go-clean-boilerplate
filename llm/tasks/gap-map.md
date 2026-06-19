# Gap Map — Codex-ready Parity Audit

## Scope

Audit penuh repo `Casbin` terhadap starter-pack `codex-ready/typescript-go`.

## Final result

Parity penuh tercapai dengan penyesuaian repo-aware.

## File-by-file result

### `AGENTS.md`

- Status: fixed.
- Result: lifecycle folder rules, memory/update rules, dan repo routing sudah eksplisit.

### `llm/cache/*`

- Status: already stronger than template.
- Result: tidak diubah dalam pass ini karena aturan user melarang update cache dengan info uncommitted.

### `llm/conventions/*`

- Status: parity met.
- Result: tetap valid dan repo-aware.

### `llm/workflows/*`

- Status: fixed.
- Result: semua workflow inti kini punya section eksplisit `Verification commands`, `Review checklist`, dan `Stop conditions / needs confirmation`.

### `llm/research/README.md`

- Status: fixed.

### `llm/recommendations/README.md`

- Status: fixed.

### `llm/test-playbooks/README.md`

- Status: fixed.

### `llm/plans/README.md`

- Status: fixed.

### `llm/plans/improve/README.md`

- Status: fixed.

### `llm/plans/roadmap/README.md`

- Status: fixed.

### `llm/tasks/todo.md`

- Status: fixed.
- Result: task state sekarang mencerminkan parity pass ini.

### `.agents/skills/*`

- Status: fixed.
- Result: starter-pack core skills sekarang ada dan sudah diadaptasi ke repo ini.
- Retained: `vercel-*`, `web-design-guidelines`, dan `workflows/bmad` dipertahankan karena tidak konflik dengan skill inti starter-pack.

## Self-audit summary

- required lifecycle folders now exist with README guidance
- required starter-pack workflows exist and contain operational sections
- required starter-pack core skills now exist under `.agents/skills/`
- no Carbon-specific stack assumptions introduced into new skills
- live-code-first rules remain centered in `AGENTS.md`
