CREATE TABLE projects (
    id VARCHAR(36) PRIMARY KEY,
    organization_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(191) NOT NULL,
    domain VARCHAR(191) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0,
    CONSTRAINT fk_projects_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_projects_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_projects_organization_id (organization_id),
    INDEX idx_projects_user_id (user_id),
    INDEX idx_projects_deleted_at (deleted_at)
) ENGINE=InnoDB;
