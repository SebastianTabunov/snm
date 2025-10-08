-- Таблица пользователей
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица токенов аутентификации
CREATE TABLE auth_tokens (
                             id SERIAL PRIMARY KEY,
                             user_id INTEGER REFERENCES users(id),
                             token VARCHAR(255) NOT NULL,
                             expires_at TIMESTAMP NOT NULL,
                             created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица профилей пользователей
CREATE TABLE user_profiles (
                               id INTEGER PRIMARY KEY REFERENCES users(id),
                               first_name VARCHAR(100),
                               last_name VARCHAR(100),
                               phone VARCHAR(20),
                               address TEXT,
                               created_at TIMESTAMP DEFAULT NOW(),
                               updated_at TIMESTAMP DEFAULT NOW()
);

-- Таблица заказов
CREATE TABLE orders (
                        id SERIAL PRIMARY KEY,
                        user_id INTEGER REFERENCES users(id),
                        title VARCHAR(255) NOT NULL,
                        description TEXT,
                        price DECIMAL(10,2) NOT NULL,
                        status VARCHAR(50) DEFAULT 'pending',
                        created_at TIMESTAMP DEFAULT NOW(),
                        updated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для улучшения производительности
CREATE INDEX idx_auth_tokens_token ON auth_tokens(token);
CREATE INDEX idx_auth_tokens_expires_at ON auth_tokens(expires_at);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_users_email ON users(email);