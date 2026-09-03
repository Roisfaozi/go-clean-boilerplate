ALTER TABLE audit_logs DROP INDEX idx_audit_logs_deleted_at;
ALTER TABLE audit_logs DROP INDEX idx_audit_org_deleted;
ALTER TABLE audit_logs DROP INDEX idx_audit_user_deleted;

-- Down migration for audit logs table alter
ALTER TABLE audit_logs MODIFY deleted_at BIGINT NOT NULL DEFAULT 0;

CREATE INDEX idx_audit_logs_deleted_at ON audit_logs(deleted_at);
CREATE INDEX idx_audit_org_deleted ON audit_logs(organization_id, deleted_at);
CREATE INDEX idx_audit_user_deleted ON audit_logs(user_id, deleted_at);
