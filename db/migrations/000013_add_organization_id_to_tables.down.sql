-- Remove organization_id from audit_logs
ALTER TABLE audit_logs DROP FOREIGN KEY fk_audit_logs_organization;
ALTER TABLE audit_logs DROP COLUMN organization_id;

-- Remove organization_id from access_rights
ALTER TABLE access_rights DROP FOREIGN KEY fk_access_rights_organization;
ALTER TABLE access_rights DROP COLUMN organization_id;

-- Remove organization_id from roles
ALTER TABLE roles DROP FOREIGN KEY fk_roles_organization;
ALTER TABLE roles DROP COLUMN organization_id;

-- Remove organization_id from users
ALTER TABLE users DROP FOREIGN KEY fk_users_organization;
ALTER TABLE users DROP COLUMN organization_id;
