-- Make role.name and access_rights.name unique per organization
-- Drop global unique constraints on name and add composite unique index (name, organization_id)

-- roles: drop unique on name if exists, add composite unique
ALTER TABLE roles DROP INDEX IF EXISTS `name`;
ALTER TABLE roles ADD UNIQUE KEY `idx_roles_name_org` (`name`, `organization_id`);

-- access_rights: drop unique on name and add composite unique
ALTER TABLE access_rights DROP INDEX IF EXISTS `name`;
ALTER TABLE access_rights ADD UNIQUE KEY `idx_access_rights_name_org` (`name`, `organization_id`);
