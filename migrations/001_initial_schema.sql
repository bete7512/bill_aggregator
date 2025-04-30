-- migrations/001_initial_schema.sql
-- Create users table
CREATE TABLE
    IF NOT EXISTS users (
        id VARCHAR(36) PRIMARY KEY,
        email VARCHAR(255) NOT NULL UNIQUE,
        name VARCHAR(255) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

-- Create providers table
CREATE TABLE
    IF NOT EXISTS providers (
        id VARCHAR(36) PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        type VARCHAR(50) NOT NULL,
        auth_types JSONB NOT NULL,
        required_fields JSONB NOT NULL,
        base_url VARCHAR(255) NOT NULL,
        icon_url VARCHAR(255),
        active BOOLEAN NOT NULL DEFAULT TRUE,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

-- Create linked_accounts table
CREATE TABLE
    IF NOT EXISTS linked_accounts (
        id VARCHAR(36) PRIMARY KEY,
        user_id VARCHAR(36) NOT NULL,
        provider_id VARCHAR(36) NOT NULL,
        account_number VARCHAR(255) NOT NULL,
        credentials JSONB NOT NULL,
        status VARCHAR(50) NOT NULL,
        last_synced TIMESTAMP,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (provider_id) REFERENCES providers (id),
        UNIQUE (user_id, provider_id, account_number)
    );

-- Create bills table
CREATE TABLE
    IF NOT EXISTS bills (
        id VARCHAR(36) PRIMARY KEY,
        user_id VARCHAR(36) NOT NULL,
        linked_account_id VARCHAR(36) NOT NULL,
        provider_id VARCHAR(36) NOT NULL,
        bill_number VARCHAR(255) NOT NULL,
        amount DECIMAL(10, 2) NOT NULL,
        currency VARCHAR(3) NOT NULL,
        status VARCHAR(50) NOT NULL,
        due_date TIMESTAMP NOT NULL,
        issued_date TIMESTAMP NOT NULL,
        paid_date TIMESTAMP,
        period_start TIMESTAMP NOT NULL,
        period_end TIMESTAMP NOT NULL,
        details JSONB,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (linked_account_id) REFERENCES linked_accounts (id) ON DELETE CASCADE,
        FOREIGN KEY (provider_id) REFERENCES providers (id),
        UNIQUE (provider_id, bill_number)
    );

-- Create indexes
CREATE INDEX idx_linked_accounts_user_id ON linked_accounts (user_id);

CREATE INDEX idx_linked_accounts_provider_id ON linked_accounts (provider_id);

CREATE INDEX idx_linked_accounts_status ON linked_accounts (status);

CREATE INDEX idx_bills_user_id ON bills (user_id);

CREATE INDEX idx_bills_linked_account_id ON bills (linked_account_id);

CREATE INDEX idx_bills_provider_id ON bills (provider_id);

CREATE INDEX idx_bills_status ON bills (status);

CREATE INDEX idx_bills_due_date ON bills (due_date);

-- Insert initial providers
INSERT INTO
    providers (
        id,
        name,
        type,
        auth_types,
        required_fields,
        base_url,
        icon_url,
        active
    )
VALUES
    (
        'electricity-provider',
        'PowerCo',
        'electricity',
        '["api_key"]',
        '["account_number", "api_key"]',
        'https://api.powerco.example.com',
        'https://powerco.example.com/icon.png',
        TRUE
    ),
    (
        'water-provider',
        'WaterWorks',
        'water',
        '["api_key"]',
        '["account_number", "api_key"]',
        'https://api.waterworks.example.com',
        'https://waterworks.example.com/icon.png',
        TRUE
    ),
    (
        'internet-provider',
        'NetConnect',
        'internet',
        '["oauth"]',
        '["account_number", "client_id", "client_secret"]',
        'https://api.netconnect.example.com',
        'https://netconnect.example.com/icon.png',
        TRUE
    );

-- Insert test user
INSERT INTO
    users (id, email, name)
VALUES
    ('user123', 'test@example.com', 'Test User');