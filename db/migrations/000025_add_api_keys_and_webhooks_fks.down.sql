ALTER TABLE webhook_logs
    DROP FOREIGN KEY fk_webhook_logs_webhook;

ALTER TABLE webhooks
    DROP FOREIGN KEY fk_webhooks_organization;

ALTER TABLE api_keys
    DROP FOREIGN KEY fk_api_keys_organization;
