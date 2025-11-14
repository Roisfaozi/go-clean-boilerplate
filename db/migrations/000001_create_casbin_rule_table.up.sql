-- +migrate Up
CREATE TABLE casbin_rule (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ptype VARCHAR(255) NOT NULL,
    v0 VARCHAR(255),
    v1 VARCHAR(255),
    v2 VARCHAR(255),
    v3 VARCHAR(255),
    v4 VARCHAR(255),
    v5 VARCHAR(255),
    INDEX idx_casbin_rule (ptype, v0, v1)
);

-- +migrate Down
DROP TABLE IF EXISTS casbin_rule;
