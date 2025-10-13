-- Remove first_name and last_name from users table
ALTER TABLE users
DROP COLUMN first_name,
DROP COLUMN last_name;