CREATE TABLE access_rights (
                               id INT AUTO_INCREMENT PRIMARY KEY,
                               name VARCHAR(191) NOT NULL UNIQUE,
                               description TEXT,
                               created_at BIGINT NOT NULL,
                               updated_at BIGINT NOT NULL,
);

CREATE TABLE endpoints (
                           id INT AUTO_INCREMENT PRIMARY KEY,
                           path VARCHAR(191) NOT NULL,
                           method VARCHAR(10) NOT NULL,
                           deleted_at BIGINT DEFAULT 0
                           UNIQUE KEY idx_path_method (path, method)
);

CREATE TABLE access_right_endpoints (
                                        access_right_id INT NOT NULL,
                                        endpoint_id INT NOT NULL,
                                        PRIMARY KEY (access_right_id, endpoint_id),
                                        FOREIGN KEY (access_right_id) REFERENCES access_rights(id) ON DELETE CASCADE,
                                        FOREIGN KEY (endpoint_id) REFERENCES endpoints(id) ON DELETE CASCADE
);