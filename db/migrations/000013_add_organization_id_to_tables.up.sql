-- Add organization_id to users
ALTER TABLE users ADD COLUMN organization_id VARCHAR(36) NULL;
ALTER TABLE users ADD CONSTRAINT fk_users_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL;
CREATE INDEX idx_users_organization_id ON users(organization_id);

-- Add organization_id to roles
ALTER TABLE roles ADD COLUMN organization_id VARCHAR(36) NULL;
ALTER TABLE roles ADD CONSTRAINT fk_roles_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
CREATE INDEX idx_roles_organization_id ON roles(organization_id);

-- Add organization_id to access_rights
ALTER TABLE access_rights ADD COLUMN organization_id VARCHAR(36) NULL;
ALTER TABLE access_rights ADD CONSTRAINT fk_access_rights_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
CREATE INDEX idx_access_rights_organization_id ON access_rights(organization_id);

-- Add organization_id to audit_logs
ALTER TABLE audit_logs ADD COLUMN organization_id VARCHAR(36) NULL;
-- Audit logs should not disappear when org is deleted? Usually yes, or set null. 
-- Decision: SET NULL to keep audit trail even if org is deleted, OR CASCADE. 
-- PRD says "Data Segregation... Row-Level Enforcement". If org is deleted, strict isolation implies data (logs) might go too, 
-- but audit usually needs retention. However, cascading is cleaner for "tenant removal". 
-- Let's stick to CASCADE for clean tenant deletion for now, or SET NULL if we want retention. 
-- Given "DeleteOrganization" deletes Members (Cascade), let's use CASCADE for strict tenant cleanup compliance unless specified otherwise.
-- Actually, for audit logs, often we want to keep them. But without an Org ID, who owns them?
-- Let's go with CASCADE for consistent cleanup as per similar FKs above.
ALTER TABLE audit_logs ADD CONSTRAINT fk_audit_logs_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);
