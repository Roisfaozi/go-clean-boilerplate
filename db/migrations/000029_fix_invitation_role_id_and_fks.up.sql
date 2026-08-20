-- Migration: 000029_fix_invitation_role_id_and_fks.up.sql
-- Purpose: Rename invitation_tokens.role to role_id, backfill legacy role names to role IDs, and add foreign key constraints

ALTER TABLE invitation_tokens CHANGE COLUMN role role_id VARCHAR(36) NOT NULL;

-- Backfill legacy role names to role IDs for existing rows in organization_members
UPDATE organization_members om
JOIN roles r ON om.role_id = r.name
SET om.role_id = r.id;

-- Backfill legacy role names to role IDs for existing rows in invitation_tokens
UPDATE invitation_tokens it
JOIN roles r ON it.role_id = r.name
SET it.role_id = r.id;

ALTER TABLE invitation_tokens
    ADD CONSTRAINT fk_invitation_tokens_role
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;

ALTER TABLE organization_members
    ADD CONSTRAINT fk_organization_members_role
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;
