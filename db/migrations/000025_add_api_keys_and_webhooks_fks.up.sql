-- Align the charset/collation of api_keys, webhooks, and webhook_logs with the rest of the database (utf8mb4_0900_ai_ci)
ALTER TABLE api_keys CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
ALTER TABLE webhooks CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
ALTER TABLE webhook_logs CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

-- Add new FKs for api_keys and webhooks
ALTER TABLE api_keys
    ADD CONSTRAINT fk_api_keys_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE webhooks
    ADD CONSTRAINT fk_webhooks_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE webhook_logs
    ADD CONSTRAINT fk_webhook_logs_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE;
