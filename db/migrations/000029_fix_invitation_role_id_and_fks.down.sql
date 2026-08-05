-- Migration: 000029_fix_invitation_role_id_and_fks.down.sql
-- Purpose: Revert invitation_tokens.role_id column name, restore role names, and drop role foreign keys

ALTER TABLE organization_members DROP FOREIGN KEY fk_organization_members_role;
ALTER TABLE invitation_tokens DROP FOREIGN KEY fk_invitation_tokens_role;

-- Revert role IDs back to role names in invitation_tokens
UPDATE invitation_tokens it
JOIN roles r ON it.role_id = r.id
SET it.role_id = r.name;

-- Revert role IDs back to role names in organization_members
UPDATE organization_members om
JOIN roles r ON om.role_id = r.id
SET om.role_id = r.name;

ALTER TABLE invitation_tokens CHANGE COLUMN role_id role VARCHAR(36) NOT NULL;
