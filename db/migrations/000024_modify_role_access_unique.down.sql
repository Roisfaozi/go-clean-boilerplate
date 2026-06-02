-- Revert composite unique to global unique on name

-- roles: drop composite index and add global unique
DROP INDEX `idx_roles_name_org` ON roles;
ALTER TABLE roles ADD UNIQUE KEY `name` (`name`);

-- access_rights: drop composite and add global unique
DROP INDEX `idx_access_rights_name_org` ON access_rights;
ALTER TABLE access_rights ADD UNIQUE KEY `name` (`name`);
