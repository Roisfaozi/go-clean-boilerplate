ALTER TABLE access_rights ADD COLUMN deleted_at BIGINT DEFAULT 0;
CREATE INDEX idx_access_rights_deleted_at ON access_rights(deleted_at);
