SET @dbname = DATABASE();
SET @sql = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @dbname AND table_name = 'organizations'
      AND column_name = 'deleted_at'
  ),
  'SELECT 1',
  'ALTER TABLE organizations ADD COLUMN deleted_at BIGINT NOT NULL DEFAULT 0'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @dbname AND table_name = 'users'
      AND column_name = 'deleted_at'
  ),
  'SELECT 1',
  'ALTER TABLE users ADD COLUMN deleted_at BIGINT NOT NULL DEFAULT 0'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @dbname AND table_name = 'roles'
      AND column_name = 'deleted_at'
  ),
  'SELECT 1',
  'ALTER TABLE roles ADD COLUMN deleted_at BIGINT NOT NULL DEFAULT 0'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @dbname AND table_name = 'organization_members'
      AND column_name = 'deleted_at'
  ),
  'SELECT 1',
  'ALTER TABLE organization_members ADD COLUMN deleted_at BIGINT NOT NULL DEFAULT 0'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
