-- Migration: 000029_fix_invitation_role_id_and_fks.down.sql
-- Purpose: Revert invitation_tokens.role_id column name and drop role foreign keys

ALTER TABLE organization_members DROP FOREIGN KEY fk_organization_members_role;
ALTER TABLE invitation_tokens DROP FOREIGN KEY fk_invitation_tokens_role;

ALTER TABLE invitation_tokens CHANGE COLUMN role_id role VARCHAR(36) NOT NULL;
