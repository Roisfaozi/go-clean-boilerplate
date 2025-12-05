-- Delete Basic Policies (Permissions)
DELETE FROM casbin_rule WHERE v0 IN ('role:admin', 'role:user');

-- Delete Basic Roles
DELETE FROM roles WHERE name IN ('role:admin', 'role:user');