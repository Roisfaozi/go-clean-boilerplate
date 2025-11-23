-- Seed Roles
INSERT INTO roles (id, name, description, created_at, updated_at) VALUES
('role-admin-id', 'role:admin', 'Super Administrator with full access', UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000),
('role-user-id', 'role:user', 'Standard User with basic access', UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000);

-- Seed Casbin Policies (Permissions)
-- role:admin can access everything (*) with any method (*)
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES 
('p', 'role:admin', '*', '*');

-- role:user can only read/update their own profile (handled by logic, but api access is allowed here)
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES 
('p', 'role:user', '/api/v1/users/me', 'GET'),
('p', 'role:user', '/api/v1/users/me', 'PUT');

-- Seed Admin User
-- Password is 'password123' (Bcrypt hash)
INSERT INTO users (id, username, email, password, name, created_at, updated_at, token) VALUES
('admin-user-id', 'admin', 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Super Admin', UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000, '');

-- Assign role:admin to admin user
INSERT INTO casbin_rule (ptype, v0, v1) VALUES 
('g', 'admin-user-id', 'role:admin');
