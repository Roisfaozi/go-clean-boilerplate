SET @dbname = DATABASE();
SET @sql = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @dbname AND table_name = 'organizations'
      AND column_name = 'deleted_at'
  ),
  'ALTER TABLE organizations DROP COLUMN deleted_at',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @dbname AND table_name = 'users'
      AND column_name = 'deleted_at'
  ),
  'ALTER TABLE users DROP COLUMN deleted_at',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @dbname AND table_name = 'roles'
      AND column_name = 'deleted_at'
  ),
  'ALTER TABLE roles DROP COLUMN deleted_at',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @dbname AND table_name = 'organization_members'
      AND column_name = 'deleted_at'
  ),
  'ALTER TABLE organization_members DROP COLUMN deleted_at',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
