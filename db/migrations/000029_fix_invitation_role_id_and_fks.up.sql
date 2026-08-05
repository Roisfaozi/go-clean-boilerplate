-- Migration: 000029_fix_invitation_role_id_and_fks.up.sql
-- Purpose: Rename invitation_tokens.role to role_id and add foreign key constraints for roles

ALTER TABLE invitation_tokens CHANGE COLUMN role role_id VARCHAR(36) NOT NULL;

ALTER TABLE invitation_tokens
    ADD CONSTRAINT fk_invitation_tokens_role
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;

ALTER TABLE organization_members
    ADD CONSTRAINT fk_organization_members_role
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;
