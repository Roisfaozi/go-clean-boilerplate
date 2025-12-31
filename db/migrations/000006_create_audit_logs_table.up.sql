CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    action VARCHAR(50) NOT NULL,
    entity VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100) NOT NULL,
    old_values JSON,
    new_values JSON,
    ip_address VARCHAR(45),
    user_agent VARCHAR(255),
    created_at BIGINT,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_audit_logs_user_id (user_id),
    INDEX idx_audit_logs_entity_id (entity_id),
    INDEX idx_audit_logs_deleted_at (deleted_at)
);
