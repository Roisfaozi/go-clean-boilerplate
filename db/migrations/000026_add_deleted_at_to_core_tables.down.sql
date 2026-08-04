ALTER TABLE organizations DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE roles DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE organization_members DROP COLUMN IF EXISTS deleted_at;
