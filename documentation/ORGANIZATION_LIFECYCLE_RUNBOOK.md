# Organization lifecycle — runbook

This runbook documents operational steps and migration plan for soft-delete, restore, and hard-delete organization lifecycle.

## Principles

- `DELETE /organizations/:id` performs a soft-delete (sets `deleted_at`).
- Soft-deleted organization and its child resources are hidden for non-superadmin callers.
- Superadmin can inspect soft-deleted organizations and may `restore` or `destroy`.
- `destroy` (hard delete) is a privileged/explicit operation and must only be used by platform operators.

## Runbook: soft-delete (already implemented)

1. API: `DELETE /organizations/:id` — marks `organizations.deleted_at` with epoch millis.
2. Post-delete actions (automated):
   - Invalidate membership cache for organization (`InvalidateOrganizationCache(orgID)`).
   - Revoke or disable API keys belonging to organization.
   - Publish audit event `organization.deleted` to outbox.

## Runbook: restore organization

1. Preconditions: only `superadmin` role may call `POST /organizations/:id/restore`.
2. Steps:
   - Validate org exists and `deleted_at != 0`.
   - Clear `deleted_at` field to `0` (or `NULL` depending on migration semantics).
   - Re-enable API keys and reindex membership cache.
   - Publish audit event `organization.restored`.

## Runbook: hard-delete (destroy)

1. Endpoint: `POST /organizations/:id/destroy` (superadmin + operator-only).
2. Preconditions: org must already be soft-deleted and operator must confirm destructive action.
3. Steps:
   - Optional: Backup tenant data (dump tables filtered by organization_id) to object storage.
   - In transaction: delete child records (projects, webhooks, api_keys, roles, access_rights, audit_logs) or rely on FK CASCADE if schema uses CASCADE. Ensure referential integrity.
   - Delete organization row.
   - Publish audit event `organization.hard_deleted`.

## DB Migration Plan (for role/access-right uniqueness)

1. Add migration to convert global `name UNIQUE` into composite `(name, organization_id)` unique index.
   - Add new migration `000014_modify_role_access_unique.up.sql` (already added in this branch).
   - The migration will drop the global `name` index and add `idx_<table>_name_org (name, organization_id)`.
2. Rolling strategy:
   - Run migration in staging and verify no duplicate `(name, organization_id)` combinations exist.
   - For existing duplicates, run a manual dedupe or prefix global roles with `role:global:<name>` before migration.

## Tests to run after deploy

- Integration and E2E suites (tenant and RBAC flows).
- Verify parity: login via JWT vs API key produce identical authorization decisions and DB scoping.

## Safety notes

- Hard-delete is irreversible; require operator confirmation and backups.
- Audit events must be written to outbox before hard-delete to keep trail.
