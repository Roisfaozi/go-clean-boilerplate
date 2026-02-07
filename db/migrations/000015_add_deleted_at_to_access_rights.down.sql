DROP INDEX idx_access_rights_deleted_at ON access_rights;
ALTER TABLE access_rights DROP COLUMN deleted_at;
