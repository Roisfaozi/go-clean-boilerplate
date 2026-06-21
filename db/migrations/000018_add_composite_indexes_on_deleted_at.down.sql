-- Remove composite indexes
DROP INDEX idx_user_org_deleted ON users;
DROP INDEX idx_user_status_deleted ON users;

DROP INDEX idx_role_org_deleted ON roles;

DROP INDEX idx_access_org_deleted ON access_rights;

DROP INDEX idx_project_org_deleted ON projects;
DROP INDEX idx_project_user_deleted ON projects;

DROP INDEX idx_org_owner_deleted ON organizations;
DROP INDEX idx_org_status_deleted ON organizations;

DROP INDEX idx_audit_org_deleted ON audit_logs;
DROP INDEX idx_audit_user_deleted ON audit_logs;
