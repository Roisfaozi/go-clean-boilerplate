-- Create casbin_rule table for storing RBAC policies
CREATE TABLE IF NOT EXISTS casbin_rule (
    id BINT AUTO_INCREMENT PRIMARY KEY,
    ptype VARCHAR(30),
    v0 VARCHAR(255),
    v1 VARCHAR(255),
    v2 VARCHAR(255),
    v3 VARCHAR(255),
    v4 VARCHAR(255),
    v5 VARCHAR(255),
    UNIQUE INDEX idx_casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
) ENGINE=InnoDB;

-- Add comments to explain the columns
ALTER TABLE casbin_rule COMMENT 'Casbin rule table for RBAC policies';

-- Example policies (uncomment if you want to add default policies)
-- INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES 
-- ('p', 'admin', '/*', '(GET|POST|PUT|DELETE)'),
-- ('g', 'alice', 'admin', '');