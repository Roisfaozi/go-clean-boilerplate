INSERT IGNORE INTO roles (id, name, description, created_at, updated_at) VALUES
    (UUID(), 'role:user', 'Default role for all new users', NOW(), NOW()),
    (UUID(), 'role:admin', 'Administrator role with full access', NOW(), NOW());
