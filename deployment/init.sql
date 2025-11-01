-- Create database
CREATE DATABASE easitradecoins;

\c easitradecoins;

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "timescaledb";

-- Users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(64) NOT NULL,
    kyc_level INT DEFAULT 0,
    status INT DEFAULT 1,
    register_ip INET,
    register_time TIMESTAMP DEFAULT NOW(),
    last_login_time TIMESTAMP,
    last_login_ip INET
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);

-- User assets table
CREATE TABLE user_assets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    available DECIMAL(36, 18) DEFAULT 0,
    frozen DECIMAL(36, 18) DEFAULT 0,
    update_time TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, currency, chain)
);

CREATE INDEX idx_user_assets_user_id ON user_assets(user_id);

-- Trading pairs table
CREATE TABLE trading_pairs (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    base_currency VARCHAR(20) NOT NULL,
    quote_currency VARCHAR(20) NOT NULL,
    price_precision INT DEFAULT 8,
    quantity_precision INT DEFAULT 8,
    min_quantity DECIMAL(36, 18),
    max_quantity DECIMAL(36, 18),
    min_amount DECIMAL(36, 18),
    taker_fee_rate DECIMAL(10, 8) DEFAULT 0.001,
    maker_fee_rate DECIMAL(10, 8) DEFAULT 0.001,
    is_active BOOLEAN DEFAULT TRUE,
    create_time TIMESTAMP DEFAULT NOW()
);

-- Orders table (partitioned by time)
CREATE TABLE orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL,
    type VARCHAR(10) NOT NULL,
    price DECIMAL(36, 18),
    quantity DECIMAL(36, 18) NOT NULL,
    filled_qty DECIMAL(36, 18) DEFAULT 0,
    filled_amount DECIMAL(36, 18) DEFAULT 0,
    avg_price DECIMAL(36, 18),
    fee DECIMAL(36, 18) DEFAULT 0,
    fee_currency VARCHAR(20),
    status VARCHAR(10),
    time_in_force VARCHAR(3),
    create_time TIMESTAMP DEFAULT NOW(),
    update_time TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_symbol ON orders(symbol);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_create_time ON orders(create_time DESC);

-- Trades table (partitioned by time)
CREATE TABLE trades (
    id VARCHAR(36) PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    buy_order_id VARCHAR(36) NOT NULL,
    sell_order_id VARCHAR(36) NOT NULL,
    buyer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    amount DECIMAL(36, 18) NOT NULL,
    buyer_fee DECIMAL(36, 18),
    seller_fee DECIMAL(36, 18),
    trade_time TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_trades_symbol ON trades(symbol);
CREATE INDEX idx_trades_buyer_id ON trades(buyer_id);
CREATE INDEX idx_trades_seller_id ON trades(seller_id);
CREATE INDEX idx_trades_time ON trades(trade_time DESC);

-- Deposits table
CREATE TABLE deposits (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    amount DECIMAL(36, 18) NOT NULL,
    address VARCHAR(200),
    txid VARCHAR(200),
    confirmations INT DEFAULT 0,
    required_confirmations INT,
    status INT,
    create_time TIMESTAMP DEFAULT NOW(),
    confirm_time TIMESTAMP
);

CREATE INDEX idx_deposits_user_id ON deposits(user_id);
CREATE INDEX idx_deposits_txid ON deposits(txid);
CREATE INDEX idx_deposits_status_time ON deposits(status, create_time);

-- Withdrawals table
CREATE TABLE withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    amount DECIMAL(36, 18) NOT NULL,
    fee DECIMAL(36, 18),
    address VARCHAR(200) NOT NULL,
    txid VARCHAR(200),
    status INT,
    audit_user_id BIGINT,
    audit_time TIMESTAMP,
    complete_time TIMESTAMP,
    create_time TIMESTAMP DEFAULT NOW(),
    remark TEXT
);

CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);
CREATE INDEX idx_withdrawals_status_time ON withdrawals(status, create_time);

-- K-lines table (TimescaleDB hypertable)
CREATE TABLE klines (
    symbol VARCHAR(20) NOT NULL,
    interval VARCHAR(10) NOT NULL,
    open_time TIMESTAMP NOT NULL,
    open_price DECIMAL(36, 18),
    high_price DECIMAL(36, 18),
    low_price DECIMAL(36, 18),
    close_price DECIMAL(36, 18),
    volume DECIMAL(36, 18),
    close_time TIMESTAMP,
    quote_volume DECIMAL(36, 18),
    trade_count INT,
    PRIMARY KEY (symbol, interval, open_time)
);

-- Convert to hypertable
SELECT create_hypertable('klines', 'open_time');

-- Insert sample trading pairs
INSERT INTO trading_pairs (symbol, base_currency, quote_currency, min_quantity, max_quantity, min_amount) VALUES
('BTC_USDT', 'BTC', 'USDT', 0.0001, 1000, 10),
('ETH_USDT', 'ETH', 'USDT', 0.001, 10000, 10),
('BNB_USDT', 'BNB', 'USDT', 0.01, 100000, 10);
