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
    type VARCHAR(20) NOT NULL COMMENT 'limit/market/stop_loss/take_profit/stop_limit/trailing_stop',
    price DECIMAL(36, 18),
    quantity DECIMAL(36, 18) NOT NULL,
    filled_qty DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    filled_amount DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    avg_price DECIMAL(36, 18),
    fee DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    fee_currency VARCHAR(20),
    status VARCHAR(10) COMMENT 'pending/partial/filled/cancelled',
    time_in_force VARCHAR(3) COMMENT 'GTC/IOC/FOK',

    -- Stop-loss and Take-profit fields
    stop_price DECIMAL(36, 18) COMMENT '止损/止盈触发价格',
    take_profit_price DECIMAL(36, 18) COMMENT '止盈价格',
    trailing_delta DECIMAL(36, 18) COMMENT '跟踪止损价差',
    trigger_condition VARCHAR(10) COMMENT '>=, <=',
    is_triggered BOOLEAN DEFAULT FALSE COMMENT '是否已触发',
    trigger_time DATETIME COMMENT '触发时间',

    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time DESC),
    INDEX idx_user_symbol (user_id, symbol),
    INDEX idx_user_status_create_time (user_id, status, create_time DESC),
    INDEX idx_type_is_triggered (type, is_triggered),
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

-- ======================================================
-- Advanced Trading Features Tables
-- ======================================================

-- OCO (One-Cancels-Other) Orders table
CREATE TABLE IF NOT EXISTS oco_orders (
    id VARCHAR(100) PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL COMMENT 'buy/sell',
    quantity DECIMAL(36, 18) NOT NULL,
    stop_loss_order_id VARCHAR(36) NOT NULL,
    stop_loss_price DECIMAL(36, 18) NOT NULL,
    take_profit_order_id VARCHAR(36) NOT NULL,
    take_profit_price DECIMAL(36, 18) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/filled/cancelled',
    triggered_order_id VARCHAR(36),
    trigger_time DATETIME,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_stop_loss_order (stop_loss_order_id),
    INDEX idx_take_profit_order (take_profit_order_id),
    INDEX idx_create_time (create_time DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Iceberg Orders table
CREATE TABLE IF NOT EXISTS iceberg_orders (
    id VARCHAR(100) PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL COMMENT 'buy/sell',
    type VARCHAR(20) NOT NULL COMMENT 'limit only',
    price DECIMAL(36, 18) NOT NULL,
    total_quantity DECIMAL(36, 18) NOT NULL,
    display_quantity DECIMAL(36, 18) NOT NULL,
    executed_quantity DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    current_child_order_id VARCHAR(36),
    min_display_quantity DECIMAL(36, 18) NOT NULL,
    variance_percent DECIMAL(5, 2) DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/active/filled/cancelled',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- TWAP (Time-Weighted Average Price) Orders table
CREATE TABLE IF NOT EXISTS twap_orders (
    id VARCHAR(100) PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL COMMENT 'buy/sell',
    type VARCHAR(20) NOT NULL COMMENT 'market/limit',
    total_quantity DECIMAL(36, 18) NOT NULL,
    executed_quantity DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    executed_amount DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    average_price DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    duration BIGINT NOT NULL COMMENT 'Duration in seconds',
    intervals INT NOT NULL COMMENT 'Number of intervals',
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    next_slice DATETIME NOT NULL,
    limit_price DECIMAL(36, 18),
    price_tolerance DECIMAL(5, 2) DEFAULT 5.00,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/active/completed/cancelled/failed',
    completed_slices INT DEFAULT 0,
    failed_slices INT DEFAULT 0,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_start_time (start_time),
    INDEX idx_create_time (create_time DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- TWAP Slices table (execution history)
CREATE TABLE IF NOT EXISTS twap_slices (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    twap_order_id VARCHAR(100) NOT NULL,
    order_id VARCHAR(36),
    slice_number INT NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    price DECIMAL(36, 18) NOT NULL,
    status VARCHAR(20) NOT NULL COMMENT 'pending/executing/completed/failed',
    scheduled_at DATETIME NOT NULL,
    executed_at DATETIME,
    error TEXT,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_twap_order (twap_order_id),
    INDEX idx_order_id (order_id),
    INDEX idx_status (status),
    INDEX idx_scheduled_at (scheduled_at),
    FOREIGN KEY (twap_order_id) REFERENCES twap_orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Grid Trading Strategies table
CREATE TABLE IF NOT EXISTS grid_strategies (
    id VARCHAR(100) PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    lower_price DECIMAL(36, 18) NOT NULL,
    upper_price DECIMAL(36, 18) NOT NULL,
    grid_num INT NOT NULL COMMENT 'Number of grids',
    total_investment DECIMAL(36, 18) NOT NULL,
    quantity_per_grid DECIMAL(36, 18) NOT NULL,
    total_profit DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    completed_grids INT DEFAULT 0,
    active_buy_orders INT DEFAULT 0,
    active_sell_orders INT DEFAULT 0,
    auto_restart BOOLEAN DEFAULT TRUE,
    stop_loss DECIMAL(36, 18),
    take_profit DECIMAL(36, 18),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/active/paused/stopped/completed',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    start_time DATETIME,
    stop_time DATETIME,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Grid Levels table
CREATE TABLE IF NOT EXISTS grid_levels (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    strategy_id VARCHAR(100) NOT NULL,
    level INT NOT NULL,
    price DECIMAL(36, 18) NOT NULL,
    buy_order_id VARCHAR(36),
    sell_order_id VARCHAR(36),
    buy_filled BOOLEAN DEFAULT FALSE,
    sell_filled BOOLEAN DEFAULT FALSE,
    profit DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_strategy (strategy_id),
    INDEX idx_buy_order (buy_order_id),
    INDEX idx_sell_order (sell_order_id),
    INDEX idx_level (level),
    FOREIGN KEY (strategy_id) REFERENCES grid_strategies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- DCA (Dollar Cost Averaging) Strategies table
CREATE TABLE IF NOT EXISTS dca_strategies (
    id VARCHAR(100) PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    amount_per_period DECIMAL(36, 18) NOT NULL,
    frequency VARCHAR(20) NOT NULL COMMENT 'daily/weekly/monthly',
    day_of_week INT COMMENT '0-6 for weekly',
    day_of_month INT COMMENT '1-31 for monthly',
    hour_of_day INT DEFAULT 0,
    max_price DECIMAL(36, 18),
    min_price DECIMAL(36, 18),
    stop_loss DECIMAL(36, 18),
    take_profit DECIMAL(36, 18),
    start_date DATETIME NOT NULL,
    end_date DATETIME,
    next_run DATETIME NOT NULL,
    total_invested DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    total_quantity DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    average_cost DECIMAL(36, 18) DEFAULT 0.000000000000000000,
    total_executions INT DEFAULT 0,
    success_executions INT DEFAULT 0,
    failed_executions INT DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/active/paused/stopped/completed',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_next_run (next_run),
    INDEX idx_create_time (create_time DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- DCA Executions table (execution history)
CREATE TABLE IF NOT EXISTS dca_executions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    strategy_id VARCHAR(100) NOT NULL,
    order_id VARCHAR(36),
    amount DECIMAL(36, 18) NOT NULL,
    price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    status VARCHAR(20) NOT NULL COMMENT 'pending/success/failed/skipped',
    reason TEXT,
    scheduled_at DATETIME NOT NULL,
    executed_at DATETIME,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_strategy (strategy_id),
    INDEX idx_order_id (order_id),
    INDEX idx_status (status),
    INDEX idx_scheduled_at (scheduled_at),
    FOREIGN KEY (strategy_id) REFERENCES dca_strategies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ======================================================
-- Phase 3 Features: Margin Trading, Copy Trading, Options, Community
-- ======================================================

-- Margin Trading Tables
CREATE TABLE IF NOT EXISTS margin_accounts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL UNIQUE,
    collateral DECIMAL(36, 18) DEFAULT 0,
    borrowed DECIMAL(36, 18) DEFAULT 0,
    interest DECIMAL(36, 18) DEFAULT 0,
    equity DECIMAL(36, 18) DEFAULT 0,
    margin_level DECIMAL(10, 4) DEFAULT 0,
    leverage INT DEFAULT 1,
    max_leverage INT DEFAULT 10,
    maintenance_rate DECIMAL(10, 4) DEFAULT 0.1,
    status VARCHAR(20) DEFAULT 'active',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS margin_positions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    entry_price DECIMAL(36, 18) NOT NULL,
    current_price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    leverage INT NOT NULL,
    margin DECIMAL(36, 18) NOT NULL,
    unrealized_pnl DECIMAL(36, 18) DEFAULT 0,
    realized_pnl DECIMAL(36, 18) DEFAULT 0,
    liquidation_price DECIMAL(36, 18) NOT NULL,
    stop_loss DECIMAL(36, 18),
    take_profit DECIMAL(36, 18),
    status VARCHAR(20) DEFAULT 'open',
    open_time DATETIME NOT NULL,
    close_time DATETIME,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS margin_loans (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    currency VARCHAR(20) NOT NULL,
    principal DECIMAL(36, 18) NOT NULL,
    interest DECIMAL(36, 18) DEFAULT 0,
    interest_rate DECIMAL(10, 6) NOT NULL,
    repaid DECIMAL(36, 18) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    borrow_time DATETIME NOT NULL,
    repay_time DATETIME,
    last_accrue_time DATETIME NOT NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Copy Trading Tables
CREATE TABLE IF NOT EXISTS traders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL,
    description TEXT,
    roi DECIMAL(10, 4) DEFAULT 0,
    total_pnl DECIMAL(36, 18) DEFAULT 0,
    win_rate DECIMAL(5, 4) DEFAULT 0,
    followers INT DEFAULT 0,
    total_trades INT DEFAULT 0,
    profit_trades INT DEFAULT 0,
    loss_trades INT DEFAULT 0,
    max_drawdown DECIMAL(10, 4) DEFAULT 0,
    sharpe_ratio DECIMAL(10, 4) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    min_follow_amount DECIMAL(36, 18),
    max_followers INT DEFAULT 1000,
    commission_rate DECIMAL(5, 4) DEFAULT 0,
    ranking INT DEFAULT 0,
    verification_level INT DEFAULT 0,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_is_active (is_active),
    INDEX idx_ranking (ranking),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS follow_relations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    follower_id BIGINT UNSIGNED NOT NULL,
    trader_id BIGINT UNSIGNED NOT NULL,
    allocation_ratio DECIMAL(5, 4) NOT NULL,
    max_per_trade DECIMAL(36, 18) NOT NULL,
    stop_loss DECIMAL(10, 4),
    take_profit DECIMAL(10, 4),
    is_active BOOLEAN DEFAULT TRUE,
    total_copied INT DEFAULT 0,
    total_profit DECIMAL(36, 18) DEFAULT 0,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_follower_id (follower_id),
    INDEX idx_trader_id (trader_id),
    INDEX idx_is_active (is_active),
    UNIQUE KEY uk_follower_trader (follower_id, trader_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (trader_id) REFERENCES traders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS copied_orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    follower_id BIGINT UNSIGNED NOT NULL,
    trader_id BIGINT UNSIGNED NOT NULL,
    original_order_id VARCHAR(36) NOT NULL,
    copied_order_id VARCHAR(36) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    price DECIMAL(36, 18) NOT NULL,
    status VARCHAR(20) NOT NULL,
    pnl DECIMAL(36, 18) DEFAULT 0,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_follower_id (follower_id),
    INDEX idx_trader_id (trader_id),
    INDEX idx_original_order_id (original_order_id),
    INDEX idx_copied_order_id (copied_order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS trading_strategies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    trader_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    risk_level VARCHAR(20) NOT NULL,
    min_investment DECIMAL(36, 18) NOT NULL,
    roi DECIMAL(10, 4) DEFAULT 0,
    subscribers INT DEFAULT 0,
    is_public BOOLEAN DEFAULT TRUE,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_trader_id (trader_id),
    INDEX idx_category (category),
    INDEX idx_is_public (is_public),
    FOREIGN KEY (trader_id) REFERENCES traders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Options Trading Tables
CREATE TABLE IF NOT EXISTS option_contracts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    type VARCHAR(10) NOT NULL,
    strike_price DECIMAL(36, 18) NOT NULL,
    premium DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    expiry DATETIME NOT NULL,
    underlying_price DECIMAL(36, 18) NOT NULL,
    implied_volatility DECIMAL(10, 4) DEFAULT 0.3,
    status VARCHAR(20) DEFAULT 'active',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_symbol (symbol),
    INDEX idx_expiry (expiry),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS option_positions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    contract_id BIGINT UNSIGNED NOT NULL,
    position VARCHAR(10) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    entry_premium DECIMAL(36, 18) NOT NULL,
    current_premium DECIMAL(36, 18) NOT NULL,
    unrealized_pnl DECIMAL(36, 18) DEFAULT 0,
    realized_pnl DECIMAL(36, 18) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'open',
    open_time DATETIME NOT NULL,
    close_time DATETIME,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_contract_id (contract_id),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (contract_id) REFERENCES option_contracts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Trading Community Tables
CREATE TABLE IF NOT EXISTS trading_communities (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL UNIQUE,
    description TEXT,
    owner_id BIGINT UNSIGNED NOT NULL,
    member_count INT DEFAULT 0,
    post_count INT DEFAULT 0,
    is_public BOOLEAN DEFAULT TRUE,
    category VARCHAR(50) NOT NULL,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_owner_id (owner_id),
    INDEX idx_category (category),
    INDEX idx_is_public (is_public),
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS community_members (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    community_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    role VARCHAR(20) DEFAULT 'member',
    join_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_community_id (community_id),
    INDEX idx_user_id (user_id),
    UNIQUE KEY uk_community_user (community_id, user_id),
    FOREIGN KEY (community_id) REFERENCES trading_communities(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS posts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    community_id BIGINT UNSIGNED NOT NULL,
    author_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    images TEXT,
    likes INT DEFAULT 0,
    comments INT DEFAULT 0,
    views INT DEFAULT 0,
    is_pinned BOOLEAN DEFAULT FALSE,
    tags VARCHAR(500),
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_community_id (community_id),
    INDEX idx_author_id (author_id),
    INDEX idx_create_time (create_time DESC),
    FOREIGN KEY (community_id) REFERENCES trading_communities(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS comments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    post_id BIGINT UNSIGNED NOT NULL,
    author_id BIGINT UNSIGNED NOT NULL,
    parent_id BIGINT UNSIGNED,
    content TEXT NOT NULL,
    likes INT DEFAULT 0,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_post_id (post_id),
    INDEX idx_author_id (author_id),
    INDEX idx_parent_id (parent_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS likes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    target_type VARCHAR(20) NOT NULL,
    target_id BIGINT UNSIGNED NOT NULL,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_target (target_type, target_id),
    UNIQUE KEY uk_user_target (user_id, target_type, target_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS trading_signals (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    author_id BIGINT UNSIGNED NOT NULL,
    community_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    type VARCHAR(10) NOT NULL,
    entry_price DECIMAL(36, 18) NOT NULL,
    stop_loss DECIMAL(36, 18) NOT NULL,
    take_profit1 DECIMAL(36, 18) NOT NULL,
    take_profit2 DECIMAL(36, 18),
    take_profit3 DECIMAL(36, 18),
    status VARCHAR(20) DEFAULT 'active',
    result DECIMAL(10, 4),
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_author_id (author_id),
    INDEX idx_community_id (community_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (community_id) REFERENCES trading_communities(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Show tables
SHOW TABLES;

SELECT 'MySQL Database initialized successfully with all Phase 2-3 features!' as status;
