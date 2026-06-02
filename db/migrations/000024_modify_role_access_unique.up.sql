-- Make role.name and access_rights.name unique per organization
-- Drop global unique constraints on name and add composite unique index (name, organization_id)

-- roles: drop global unique name index, add composite unique index
DROP INDEX `name` ON roles;
ALTER TABLE roles ADD UNIQUE KEY `idx_roles_name_org` (`name`, `organization_id`);

-- access_rights: drop global unique name index, add composite unique index
DROP INDEX `name` ON access_rights;
ALTER TABLE access_rights ADD UNIQUE KEY `idx_access_rights_name_org` (`name`, `organization_id`);
