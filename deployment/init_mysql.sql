-- EasiTradeCoins MySQL Database Initialization Script
-- ======================================================

-- Create database
CREATE DATABASE IF NOT EXISTS easitradecoins
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE easitradecoins;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(64) NOT NULL,
    kyc_level INT DEFAULT 0 COMMENT '0:未认证 1:初级 2:高级',
    status INT DEFAULT 1 COMMENT '1:正常 2:冻结 3:注销',
    register_ip VARCHAR(45),
    register_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login_time DATETIME,
    last_login_ip VARCHAR(45),
    INDEX idx_email (email),
    INDEX idx_phone (phone),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User assets table
CREATE TABLE IF NOT EXISTS user_assets (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    available DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    frozen DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_currency_chain (user_id, currency, chain),
    INDEX idx_user_id (user_id),
    INDEX idx_currency (currency),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Trading pairs table
CREATE TABLE IF NOT EXISTS trading_pairs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    base_currency VARCHAR(20) NOT NULL,
    quote_currency VARCHAR(20) NOT NULL,
    price_precision INT DEFAULT 8,
    quantity_precision INT DEFAULT 8,
    min_quantity DECIMAL(36, 18),
    max_quantity DECIMAL(36, 18),
    min_amount DECIMAL(36, 18),
    taker_fee_rate DECIMAL(10, 8) DEFAULT 0.00100000,
    maker_fee_rate DECIMAL(10, 8) DEFAULT 0.00100000,
    is_active BOOLEAN DEFAULT TRUE,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_symbol (symbol),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL COMMENT 'buy/sell',
    type VARCHAR(10) NOT NULL COMMENT 'limit/market',
    price DECIMAL(36, 18),
    quantity DECIMAL(36, 18) NOT NULL,
    filled_qty DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    filled_amount DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    avg_price DECIMAL(36, 18),
    fee DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    fee_currency VARCHAR(20),
    status VARCHAR(10) COMMENT 'pending/partial/filled/cancelled',
    time_in_force VARCHAR(3) COMMENT 'GTC/IOC/FOK',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time DESC),
    INDEX idx_user_symbol (user_id, symbol),
    INDEX idx_user_status_create_time (user_id, status, create_time DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Trades table
CREATE TABLE IF NOT EXISTS trades (
    id VARCHAR(36) PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    buy_order_id VARCHAR(36) NOT NULL,
    sell_order_id VARCHAR(36) NOT NULL,
    buyer_id BIGINT UNSIGNED NOT NULL,
    seller_id BIGINT UNSIGNED NOT NULL,
    price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    amount DECIMAL(36, 18) NOT NULL,
    buyer_fee DECIMAL(36, 18),
    seller_fee DECIMAL(36, 18),
    trade_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_symbol (symbol),
    INDEX idx_buyer_id (buyer_id),
    INDEX idx_seller_id (seller_id),
    INDEX idx_trade_time (trade_time DESC),
    INDEX idx_buy_order_id (buy_order_id),
    INDEX idx_sell_order_id (sell_order_id),
    INDEX idx_symbol_trade_time (symbol, trade_time DESC),
    FOREIGN KEY (buyer_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (seller_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Deposits table
CREATE TABLE IF NOT EXISTS deposits (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    amount DECIMAL(36, 18) NOT NULL,
    address VARCHAR(200),
    txid VARCHAR(200),
    confirmations INT DEFAULT 0,
    required_confirmations INT,
    status INT COMMENT '0:待确认 1:已到账 2:异常',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    confirm_time DATETIME,
    INDEX idx_user_id (user_id),
    INDEX idx_txid (txid),
    INDEX idx_status_time (status, create_time),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Withdrawals table
CREATE TABLE IF NOT EXISTS withdrawals (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    amount DECIMAL(36, 18) NOT NULL,
    fee DECIMAL(36, 18),
    address VARCHAR(200) NOT NULL,
    txid VARCHAR(200),
    status INT COMMENT '0:待审核 1:审核通过 2:处理中 3:已完成 4:拒绝',
    audit_user_id BIGINT UNSIGNED,
    audit_time DATETIME,
    complete_time DATETIME,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    remark TEXT,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_status_time (status, create_time),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample trading pairs
INSERT INTO trading_pairs (symbol, base_currency, quote_currency, min_quantity, max_quantity, min_amount) VALUES
('BTC_USDT', 'BTC', 'USDT', 0.0001, 1000, 10),
('ETH_USDT', 'ETH', 'USDT', 0.001, 10000, 10),
('BNB_USDT', 'BNB', 'USDT', 0.01, 100000, 10)
ON DUPLICATE KEY UPDATE symbol=symbol;

-- Create audit log table for security
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(20),
    details JSON,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_create_time (create_time DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Risk events table for risk control system
CREATE TABLE IF NOT EXISTS risk_events (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    event_type VARCHAR(50) NOT NULL COMMENT 'order_validation/withdrawal_validation/rate_limit_exceeded/etc',
    severity VARCHAR(20) NOT NULL COMMENT 'low/medium/high/critical',
    description VARCHAR(500) NOT NULL,
    details TEXT,
    action VARCHAR(20) NOT NULL COMMENT 'allowed/blocked/flagged/frozen',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_event_type (event_type),
    INDEX idx_severity (severity),
    INDEX idx_create_time (create_time DESC),
    INDEX idx_user_event_time (user_id, event_type, create_time DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Violations table for tracking user violations
CREATE TABLE IF NOT EXISTS violations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    type VARCHAR(50) NOT NULL COMMENT 'self_trading/wash_trading/suspicious_withdrawal/etc',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT 'active/resolved',
    severity INT NOT NULL COMMENT '1-10 severity score',
    description VARCHAR(500) NOT NULL,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolve_time DATETIME,
    INDEX idx_user_id (user_id),
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time DESC),
    INDEX idx_user_status (user_id, status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Withdrawal whitelists table
CREATE TABLE IF NOT EXISTS withdrawal_whitelists (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    currency VARCHAR(20) NOT NULL,
    address VARCHAR(200) NOT NULL,
    label VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_address (address),
    INDEX idx_user_currency (user_id, currency),
    INDEX idx_user_active (user_id, is_active),
    UNIQUE KEY uk_user_currency_address (user_id, currency, address),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Show tables
SHOW TABLES;

SELECT 'MySQL Database initialized successfully!' as status;
