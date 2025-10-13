-- Add first_name and last_name to users table
ALTER TABLE users
    ADD COLUMN first_name VARCHAR(100),
ADD COLUMN last_name VARCHAR(100);

-- Update existing users with data from user_profiles if available
UPDATE users
SET first_name = up.first_name,
    last_name = up.last_name
    FROM user_profiles up
WHERE users.id = up.id AND up.first_name IS NOT NULL;