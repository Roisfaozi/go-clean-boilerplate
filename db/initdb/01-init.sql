-- Create the database if it doesn't exist
CREATE DATABASE IF NOT EXISTS gin_starter;

-- User creation is handled by MYSQL_USER and MYSQL_PASSWORD env vars in docker-compose.
-- If additional privileges are needed beyond what the default user gets on the specific database,
-- they should be added here, but usually the default user has full access to the created database.
