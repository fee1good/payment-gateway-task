DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gateways') THEN
        CREATE TABLE gateways (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL UNIQUE,
            data_format_supported VARCHAR(50) NOT NULL,  
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'countries') THEN
        CREATE TABLE countries (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL UNIQUE,
            code CHAR(2) NOT NULL UNIQUE,
            currency CHAR(3) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gateway_countries') THEN
        CREATE TABLE gateway_countries (
            gateway_id INT NOT NULL, 
            country_id INT NOT NULL,
            priority INT DEFAULT 0,
            PRIMARY KEY (gateway_id, country_id)
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(255) NOT NULL UNIQUE,
            email VARCHAR(255) NOT NULL UNIQUE,
            password VARCHAR(255) NOT NULL,
            country_id INT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'transactions') THEN
        CREATE TABLE transactions (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL REFERENCES users(id),
            amount DECIMAL(19, 4) NOT NULL,
            currency VARCHAR(3) NOT NULL,
            type VARCHAR(20) NOT NULL, -- 'deposit' or 'withdrawal'
            status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'completed', 'failed'
            gateway_id INT NULL,
            gateway_txn_id VARCHAR(255),
            error_message TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            completed_at TIMESTAMP
        );
    END IF;
END $$;

-- Insert sample data if tables are empty
DO $$
BEGIN
    -- Insert countries if none exist
    IF NOT EXISTS (SELECT 1 FROM countries) THEN
        INSERT INTO countries (name, code, currency) VALUES 
            ('United States', 'US', 'USD'),
            ('United Kingdom', 'GB', 'GBP'),
            ('European Union', 'EU', 'EUR');
    END IF;

    -- Insert gateways if none exist
    IF NOT EXISTS (SELECT 1 FROM gateways) THEN
        INSERT INTO gateways (name, data_format_supported) VALUES 
            ('Stripe', 'JSON'),
            ('PayPal', 'JSON'),
            ('Adyen', 'XML');
    END IF;

    -- Insert users if none exist
    IF NOT EXISTS (SELECT 1 FROM users) THEN
        INSERT INTO users (username, email, password, country_id) VALUES 
            ('testuser1', 'user1@example.com', 'password123', 1),
            ('testuser2', 'user2@example.com', 'password123', 2),
            ('testuser3', 'user3@example.com', 'password123', 3);
    END IF;

    -- Link gateways to countries if none exist
    IF NOT EXISTS (SELECT 1 FROM gateway_countries) THEN
        INSERT INTO gateway_countries (gateway_id, country_id, priority) VALUES 
            (1, 1, 1), -- Stripe for US (highest priority)
            (2, 1, 2), -- PayPal for US (lower priority)
            (1, 2, 1), -- Stripe for UK
            (3, 3, 1); -- Adyen for EU
    END IF;
END $$;
