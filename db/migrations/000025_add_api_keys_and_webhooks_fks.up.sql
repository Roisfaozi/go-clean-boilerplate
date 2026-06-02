-- Add missing organization FK constraints for api_keys and webhooks
ALTER TABLE api_keys
    ADD CONSTRAINT fk_api_keys_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE webhooks
    ADD CONSTRAINT fk_webhooks_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE webhook_logs
    ADD CONSTRAINT fk_webhook_logs_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE;
