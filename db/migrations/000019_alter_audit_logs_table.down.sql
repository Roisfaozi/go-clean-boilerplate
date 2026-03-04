ALTER TABLE audit_logs DROP INDEX idx_audit_logs_deleted_at;
ALTER TABLE audit_logs DROP INDEX idx_audit_org_deleted;
ALTER TABLE audit_logs DROP INDEX idx_audit_user_deleted;

ALTER TABLE audit_logs MODIFY deleted_at TIMESTAMP NULL DEFAULT NULL;

CREATE INDEX idx_audit_logs_deleted_at ON audit_logs(deleted_at);
CREATE INDEX idx_audit_org_deleted ON audit_logs(organization_id, deleted_at);
CREATE INDEX idx_audit_user_deleted ON audit_logs(user_id, deleted_at);
