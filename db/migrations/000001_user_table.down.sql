-- Drop users table and its indexes
ALTER TABLE users
    DROP INDEX idx_users_email,
    DROP INDEX idx_users_deleted_at;
DROP TABLE users;
