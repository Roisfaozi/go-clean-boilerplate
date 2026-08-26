INSERT INTO roles (id, name, description, created_at, updated_at)
SELECT 'role:org-owner', 'role:org-owner', 'Organization owner role', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'role:org-owner');
