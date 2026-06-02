-- Revert composite unique to global unique on name

-- roles: drop composite index and add global unique
ALTER TABLE roles DROP INDEX IF EXISTS `idx_roles_name_org`;
ALTER TABLE roles ADD UNIQUE KEY `name` (`name`);

-- access_rights: drop composite and add global unique
ALTER TABLE access_rights DROP INDEX IF EXISTS `idx_access_rights_name_org`;
ALTER TABLE access_rights ADD UNIQUE KEY `name` (`name`);
