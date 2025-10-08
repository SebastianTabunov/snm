-- Create user_profiles table
CREATE TABLE user_profiles (
                               id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
                               first_name VARCHAR(100),
                               last_name VARCHAR(100),
                               phone VARCHAR(20),
                               address TEXT,
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                               updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for user profiles
CREATE INDEX idx_user_profiles_id ON user_profiles(id);