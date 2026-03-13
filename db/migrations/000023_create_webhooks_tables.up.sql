CREATE TABLE IF NOT EXISTS webhooks (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    organization_id VARCHAR(36) NOT NULL,
    url TEXT NOT NULL,
    events TEXT NOT NULL, -- JSON array of events
    secret VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT NULL,
    INDEX idx_webhooks_org_id (organization_id),
    INDEX idx_webhooks_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS webhook_logs (
    id VARCHAR(36) PRIMARY KEY,
    webhook_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload LONGTEXT NOT NULL,
    response_status_code INT NULL,
    response_body LONGTEXT NULL,
    execution_time BIGINT NULL, -- in milliseconds
    error_message TEXT NULL,
    retry_count INT DEFAULT 0,
    created_at BIGINT NOT NULL,
    INDEX idx_webhook_logs_webhook_id (webhook_id),
    INDEX idx_webhook_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
