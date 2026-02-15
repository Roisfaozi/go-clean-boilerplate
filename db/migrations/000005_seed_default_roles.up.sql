-- Insert Basic Roles
INSERT IGNORE INTO roles (id, name, description, created_at, updated_at) VALUES
    (UUID(), 'role:user', 'Standard User with basic access', UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000),
    (UUID(), 'role:admin', 'Administrator role with full access',  UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000),
    (UUID(), 'role:superadmin', 'Administrator role with root access access',  UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000);

-- Insert Basic Policies for Admin
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'role:admin', '/api/v1/users/*', 'GET'),
    ('p', 'role:admin', '/api/v1/users/*', 'PUT'),
    ('p', 'role:admin', '/api/v1/users/*', 'PATCH');

-- Insert Basic Policies for User
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'role:user', '/api/v1/users/me', 'GET'),
    ('p', 'role:user', '/api/v1/users/me', 'PUT');

-- Insert Basic Policies for Admin
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'role:superadmin', '*', '*');