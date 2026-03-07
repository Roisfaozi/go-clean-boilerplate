CREATE TABLE IF NOT EXISTS `user_sso_identities` (
    `id` VARCHAR(36) NOT NULL,
    `user_id` VARCHAR(36) NOT NULL,
    `provider` VARCHAR(50) NOT NULL,
    `provider_id` VARCHAR(255) NOT NULL,
    `created_at` BIGINT NOT NULL,
    `updated_at` BIGINT NOT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_user_sso_identities_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    UNIQUE KEY `idx_sso_provider` (`provider`, `provider_id`),
    INDEX `idx_sso_user` (`user_id`)
) ENGINE=InnoDB;
