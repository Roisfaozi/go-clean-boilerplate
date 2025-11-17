-- Drop users table and its indexes
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP TABLE IF EXISTS users;
