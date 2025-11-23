-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    token TEXT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY idx_users_email (email),
    INDEX idx_users_deleted_at (deleted_at)
    ) engine=InnoDB;

-- Create indexes for better query performance
--     uncommet if postgres
-- CREATE INDEX idx_users_email ON users(email);
-- CREATE INDEX idx_users_deleted_at ON users(deleted_at);