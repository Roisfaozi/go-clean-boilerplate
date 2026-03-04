-- Add composite indexes for optimization with soft delete
CREATE INDEX idx_user_org_deleted ON users (organization_id, deleted_at);
CREATE INDEX idx_user_status_deleted ON users (status, deleted_at);

CREATE INDEX idx_role_org_deleted ON roles (organization_id, deleted_at);

CREATE INDEX idx_access_org_deleted ON access_rights (organization_id, deleted_at);

CREATE INDEX idx_project_org_deleted ON projects (organization_id, deleted_at);
CREATE INDEX idx_project_user_deleted ON projects (user_id, deleted_at);

CREATE INDEX idx_org_owner_deleted ON organizations (owner_id, deleted_at);
CREATE INDEX idx_org_status_deleted ON organizations (status, deleted_at);

CREATE INDEX idx_audit_org_deleted ON audit_logs (organization_id, deleted_at);
CREATE INDEX idx_audit_user_deleted ON audit_logs (user_id, deleted_at);
