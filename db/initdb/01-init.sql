-- Create the database if it doesn't exist
CREATE DATABASE IF NOT EXISTS gin_starter;

-- Create the user with caching_sha2_password authentication
CREATE USER IF NOT EXISTS 'gin_user'@'%' IDENTIFIED WITH caching_sha2_password BY 'ginPassword0!';

-- Grant privileges
GRANT ALL PRIVILEGES ON gin_starter.* TO 'gin_user'@'%';

-- Apply the changes
FLUSH PRIVILEGES;