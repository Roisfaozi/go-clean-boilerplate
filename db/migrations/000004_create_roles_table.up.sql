CREATE TABLE roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
)Engine=InnoDB;