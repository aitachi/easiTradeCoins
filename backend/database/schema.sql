-- EasiTradeCoins 数据库完整设计
-- PostgreSQL 14+

-- =================================================================================
-- 1. 用户相关表
-- =================================================================================

-- 用户基础信息表
CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(64) NOT NULL,
    kyc_level INT DEFAULT 0, -- 0未认证 1初级 2高级
    status INT DEFAULT 1, -- 1正常 2冻结 3注销
    vip_level INT DEFAULT 0, -- VIP等级 0-5
    register_ip INET,
    register_time TIMESTAMP DEFAULT NOW(),
    last_login_time TIMESTAMP,
    last_login_ip INET,
    referrer_id BIGINT REFERENCES users(user_id), -- 推荐人
    avatar_url VARCHAR(500),
    nickname VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_referrer ON users(referrer_id);

-- 用户安全设置表
CREATE TABLE user_security (
    user_id BIGINT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    google_2fa_secret VARCHAR(64),
    is_2fa_enabled BOOLEAN DEFAULT FALSE,
    withdrawal_whitelist JSONB, -- 提现地址白名单
    api_key_hash VARCHAR(255),
    api_secret_hash VARCHAR(255),
    api_permissions JSONB, -- API权限配置
    login_password_error_count INT DEFAULT 0,
    login_locked_until TIMESTAMP,
    device_fingerprints JSONB[], -- 设备指纹列表
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- KYC认证表
CREATE TABLE user_kyc (
    kyc_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    real_name VARCHAR(100),
    id_number VARCHAR(50),
    id_type INT, -- 1身份证 2护照 3驾照
    id_front_url VARCHAR(500),
    id_back_url VARCHAR(500),
    face_url VARCHAR(500),
    country_code VARCHAR(10),
    address TEXT,
    date_of_birth DATE,
    submit_time TIMESTAMP DEFAULT NOW(),
    audit_time TIMESTAMP,
    audit_status INT, -- 0待审核 1通过 2拒绝
    audit_remark TEXT,
    auditor_id BIGINT REFERENCES users(user_id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_kyc_user_id ON user_kyc(user_id);
CREATE INDEX idx_kyc_status ON user_kyc(audit_status);

-- 用户登录日志
CREATE TABLE user_login_logs (
    log_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    login_ip INET,
    login_location VARCHAR(100),
    device_type VARCHAR(50),
    device_fingerprint VARCHAR(255),
    user_agent TEXT,
    login_status INT, -- 1成功 2失败
    fail_reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_login_logs_user ON user_login_logs(user_id, created_at DESC);
CREATE INDEX idx_login_logs_ip ON user_login_logs(login_ip, created_at DESC);

-- =================================================================================
-- 2. 资产相关表
-- =================================================================================

-- 用户资产表
CREATE TABLE user_assets (
    asset_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    currency VARCHAR(20) NOT NULL, -- BTC/ETH/USDT
    chain VARCHAR(20), -- ERC20/TRC20/BEP20/SOLANA
    available DECIMAL(36, 18) DEFAULT 0, -- 可用余额
    frozen DECIMAL(36, 18) DEFAULT 0, -- 冻结余额
    total DECIMAL(36, 18) GENERATED ALWAYS AS (available + frozen) STORED,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, currency, chain)
);

CREATE INDEX idx_assets_user ON user_assets(user_id);
CREATE INDEX idx_assets_currency ON user_assets(currency, chain);

-- 充值地址表
CREATE TABLE deposit_addresses (
    address_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    address VARCHAR(200) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, currency, chain, address)
);

CREATE INDEX idx_deposit_addr_user ON deposit_addresses(user_id);
CREATE INDEX idx_deposit_addr ON deposit_addresses(address);

-- 充值记录表
CREATE TABLE deposits (
    deposit_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    amount DECIMAL(36, 18) NOT NULL,
    address VARCHAR(200), -- 充值地址
    from_address VARCHAR(200), -- 来源地址
    txid VARCHAR(200), -- 链上交易哈希
    confirmations INT DEFAULT 0, -- 当前确认数
    required_confirmations INT, -- 需要确认数
    status INT, -- 0待确认 1已到账 2异常
    error_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    confirmed_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_deposits_user ON deposits(user_id, created_at DESC);
CREATE INDEX idx_deposits_txid ON deposits(txid);
CREATE INDEX idx_deposits_status ON deposits(status, created_at DESC);

-- 提现记录表
CREATE TABLE withdrawals (
    withdrawal_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    amount DECIMAL(36, 18) NOT NULL,
    fee DECIMAL(36, 18), -- 提现手续费
    actual_amount DECIMAL(36, 18), -- 实际到账金额
    address VARCHAR(200) NOT NULL, -- 提现地址
    address_tag VARCHAR(100), -- 地址标签(Memo)
    txid VARCHAR(200), -- 链上交易哈希
    status INT, -- 0待审核 1审核通过 2处理中 3已完成 4拒绝 5已取消
    audit_user_id BIGINT REFERENCES users(user_id), -- 审核人
    audit_time TIMESTAMP,
    complete_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    remark TEXT,
    reject_reason TEXT
);

CREATE INDEX idx_withdrawals_user ON withdrawals(user_id, created_at DESC);
CREATE INDEX idx_withdrawals_status ON withdrawals(status, created_at DESC);
CREATE INDEX idx_withdrawals_txid ON withdrawals(txid);

-- 资产流水表
CREATE TABLE asset_transactions (
    tx_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    amount DECIMAL(36, 18) NOT NULL, -- 正数为增加,负数为减少
    balance_before DECIMAL(36, 18) NOT NULL,
    balance_after DECIMAL(36, 18) NOT NULL,
    tx_type INT NOT NULL, -- 1充值 2提现 3交易 4手续费 5转账 6理财 7活动奖励
    ref_id BIGINT, -- 关联ID(订单ID/充值ID/提现ID等)
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_asset_tx_user ON asset_transactions(user_id, created_at DESC);
CREATE INDEX idx_asset_tx_type ON asset_transactions(tx_type, created_at DESC);

-- =================================================================================
-- 3. 交易系统表
-- =================================================================================

-- 交易对配置表
CREATE TABLE trading_pairs (
    pair_id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) UNIQUE NOT NULL, -- BTC_USDT
    base_currency VARCHAR(20) NOT NULL, -- BTC
    quote_currency VARCHAR(20) NOT NULL, -- USDT
    price_precision INT DEFAULT 8, -- 价格小数位
    quantity_precision INT DEFAULT 8, -- 数量小数位
    min_quantity DECIMAL(36, 18), -- 最小下单量
    max_quantity DECIMAL(36, 18), -- 最大下单量
    min_amount DECIMAL(36, 18), -- 最小下单金额
    taker_fee_rate DECIMAL(10, 8) DEFAULT 0.001, -- 吃单手续费率
    maker_fee_rate DECIMAL(10, 8) DEFAULT 0.001, -- 挂单手续费率
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_pairs_symbol ON trading_pairs(symbol);
CREATE INDEX idx_pairs_active ON trading_pairs(is_active);

-- 订单表
CREATE TABLE orders (
    order_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL, -- buy/sell
    type VARCHAR(10) NOT NULL, -- limit/market
    price DECIMAL(36, 18),
    quantity DECIMAL(36, 18) NOT NULL,
    filled_quantity DECIMAL(36, 18) DEFAULT 0,
    filled_amount DECIMAL(36, 18) DEFAULT 0, -- 成交金额
    avg_price DECIMAL(36, 18), -- 平均成交价
    fee DECIMAL(36, 18) DEFAULT 0, -- 手续费
    fee_currency VARCHAR(20), -- 手续费币种
    status INT, -- 0待成交 1部分成交 2完全成交 3已撤销 4已拒绝
    time_in_force VARCHAR(3) DEFAULT 'GTC', -- GTC/IOC/FOK
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_orders_user ON orders(user_id, created_at DESC);
CREATE INDEX idx_orders_symbol ON orders(symbol, created_at DESC);
CREATE INDEX idx_orders_status ON orders(status, created_at DESC);
CREATE INDEX idx_orders_active ON orders(symbol, status) WHERE status IN (0, 1);

-- 成交记录表
CREATE TABLE trades (
    trade_id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    buy_order_id BIGINT NOT NULL REFERENCES orders(order_id),
    sell_order_id BIGINT NOT NULL REFERENCES orders(order_id),
    buyer_id BIGINT NOT NULL REFERENCES users(user_id),
    seller_id BIGINT NOT NULL REFERENCES users(user_id),
    price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    amount DECIMAL(36, 18) NOT NULL,
    buyer_fee DECIMAL(36, 18),
    seller_fee DECIMAL(36, 18),
    is_buyer_maker BOOLEAN, -- 买方是否为Maker
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_trades_symbol ON trades(symbol, created_at DESC);
CREATE INDEX idx_trades_buyer ON trades(buyer_id, created_at DESC);
CREATE INDEX idx_trades_seller ON trades(seller_id, created_at DESC);
CREATE INDEX idx_trades_orders ON trades(buy_order_id, sell_order_id);

-- K线数据表
CREATE TABLE klines (
    kline_id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    interval VARCHAR(10) NOT NULL, -- 1m/5m/15m/1h/4h/1d
    open_time TIMESTAMP NOT NULL,
    open_price DECIMAL(36, 18),
    high_price DECIMAL(36, 18),
    low_price DECIMAL(36, 18),
    close_price DECIMAL(36, 18),
    volume DECIMAL(36, 18),
    close_time TIMESTAMP,
    quote_volume DECIMAL(36, 18), -- 成交额
    trade_count INT,
    UNIQUE(symbol, interval, open_time)
);

CREATE INDEX idx_klines_symbol ON klines(symbol, interval, open_time DESC);

-- =================================================================================
-- 4. 风控与安全表
-- =================================================================================

-- 风控规则配置表
CREATE TABLE risk_rules (
    rule_id SERIAL PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL,
    rule_type VARCHAR(50) NOT NULL, -- DAILY_LIMIT/FREQUENCY/AMOUNT/BLACKLIST
    rule_config JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 风险事件记录表
CREATE TABLE risk_events (
    event_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id),
    event_type VARCHAR(50) NOT NULL,
    risk_level INT, -- 1低 2中 3高 4严重
    description TEXT,
    related_data JSONB,
    status INT, -- 0待处理 1已处理 2已忽略
    handler_id BIGINT REFERENCES users(user_id),
    handle_time TIMESTAMP,
    handle_remark TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_risk_events_user ON risk_events(user_id, created_at DESC);
CREATE INDEX idx_risk_events_level ON risk_events(risk_level, status);

-- 黑名单表
CREATE TABLE blacklist (
    blacklist_id BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(20) NOT NULL, -- USER/IP/ADDRESS/DEVICE
    entity_value VARCHAR(255) NOT NULL,
    reason TEXT,
    added_by BIGINT REFERENCES users(user_id),
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_blacklist_entity ON blacklist(entity_type, entity_value);
CREATE INDEX idx_blacklist_active ON blacklist(is_active, expires_at);

-- =================================================================================
-- 5. 系统配置表
-- =================================================================================

-- 系统参数配置表
CREATE TABLE system_config (
    config_id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    config_type VARCHAR(20), -- STRING/NUMBER/JSON/BOOLEAN
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 币种配置表
CREATE TABLE currencies (
    currency_id SERIAL PRIMARY KEY,
    currency VARCHAR(20) UNIQUE NOT NULL,
    full_name VARCHAR(100),
    symbol VARCHAR(10),
    decimals INT DEFAULT 18,
    is_crypto BOOLEAN DEFAULT TRUE,
    icon_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 链配置表
CREATE TABLE chains (
    chain_id SERIAL PRIMARY KEY,
    chain VARCHAR(20) UNIQUE NOT NULL,
    full_name VARCHAR(100),
    chain_type VARCHAR(20), -- EVM/SOLANA/TRON
    rpc_url VARCHAR(500),
    explorer_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 币种-链映射表
CREATE TABLE currency_chains (
    mapping_id SERIAL PRIMARY KEY,
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20) NOT NULL,
    contract_address VARCHAR(200),
    decimals INT DEFAULT 18,
    deposit_enabled BOOLEAN DEFAULT TRUE,
    withdraw_enabled BOOLEAN DEFAULT TRUE,
    min_deposit DECIMAL(36, 18),
    min_withdraw DECIMAL(36, 18),
    withdraw_fee DECIMAL(36, 18),
    required_confirmations INT DEFAULT 12,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(currency, chain)
);

CREATE INDEX idx_currency_chains ON currency_chains(currency, chain);

-- =================================================================================
-- 6. 营销与活动表
-- =================================================================================

-- 邀请关系表
CREATE TABLE referral_relations (
    relation_id BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES users(user_id),
    referee_id BIGINT NOT NULL REFERENCES users(user_id),
    level INT DEFAULT 1, -- 1一级 2二级
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(referrer_id, referee_id)
);

CREATE INDEX idx_referral_referrer ON referral_relations(referrer_id);
CREATE INDEX idx_referral_referee ON referral_relations(referee_id);

-- 返佣记录表
CREATE TABLE rebate_records (
    rebate_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    from_user_id BIGINT NOT NULL REFERENCES users(user_id),
    trade_id BIGINT REFERENCES trades(trade_id),
    currency VARCHAR(20) NOT NULL,
    amount DECIMAL(36, 18) NOT NULL,
    rate DECIMAL(10, 8),
    level INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_rebate_user ON rebate_records(user_id, created_at DESC);

-- VIP等级配置表
CREATE TABLE vip_levels (
    level INT PRIMARY KEY,
    level_name VARCHAR(50),
    min_volume_30d DECIMAL(36, 18), -- 30日交易量要求
    taker_fee_rate DECIMAL(10, 8),
    maker_fee_rate DECIMAL(10, 8),
    daily_withdraw_limit DECIMAL(36, 18),
    benefits JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- =================================================================================
-- 7. 通知系统表
-- =================================================================================

-- 通知模板表
CREATE TABLE notification_templates (
    template_id SERIAL PRIMARY KEY,
    template_code VARCHAR(50) UNIQUE NOT NULL,
    template_name VARCHAR(100),
    channel VARCHAR(20) NOT NULL, -- EMAIL/SMS/TELEGRAM/PUSH
    subject VARCHAR(200),
    content TEXT,
    variables JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 通知记录表
CREATE TABLE notifications (
    notification_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id),
    template_code VARCHAR(50),
    channel VARCHAR(20),
    recipient VARCHAR(255),
    subject VARCHAR(200),
    content TEXT,
    status INT, -- 0待发送 1已发送 2失败
    sent_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_status ON notifications(status, created_at);

-- =================================================================================
-- 初始化数据
-- =================================================================================

-- 插入默认系统配置
INSERT INTO system_config (config_key, config_value, config_type, description) VALUES
('withdrawal_min_confirmations', '{"BTC": 3, "ETH": 12, "USDT_ERC20": 12, "SOL": 32}', 'JSON', '充值最小确认数'),
('daily_withdrawal_limit', '{"0": 1000, "1": 10000, "2": 100000}', 'JSON', '每日提现限额（USD）'),
('api_rate_limit', '{"public": 100, "private": 50}', 'JSON', 'API请求限流（次/分钟）'),
('maintenance_mode', 'false', 'BOOLEAN', '维护模式'),
('register_enabled', 'true', 'BOOLEAN', '是否允许注册'),
('withdraw_enabled', 'true', 'BOOLEAN', '是否允许提现');

-- 插入默认交易对
INSERT INTO trading_pairs (symbol, base_currency, quote_currency, price_precision, quantity_precision, min_quantity, min_amount, taker_fee_rate, maker_fee_rate) VALUES
('BTC_USDT', 'BTC', 'USDT', 2, 6, 0.0001, 10, 0.001, 0.001),
('ETH_USDT', 'ETH', 'USDT', 2, 4, 0.001, 10, 0.001, 0.001),
('SOL_USDT', 'SOL', 'USDT', 3, 2, 0.1, 5, 0.001, 0.001),
('BNB_USDT', 'BNB', 'USDT', 2, 3, 0.01, 5, 0.001, 0.001);

-- 插入币种配置
INSERT INTO currencies (currency, full_name, symbol, decimals) VALUES
('BTC', 'Bitcoin', 'BTC', 8),
('ETH', 'Ethereum', 'ETH', 18),
('USDT', 'Tether USD', 'USDT', 6),
('SOL', 'Solana', 'SOL', 9),
('BNB', 'Binance Coin', 'BNB', 18);

-- 插入链配置
INSERT INTO chains (chain, full_name, chain_type) VALUES
('BITCOIN', 'Bitcoin', 'BTC'),
('ETHEREUM', 'Ethereum', 'EVM'),
('SOLANA', 'Solana', 'SOLANA'),
('BSC', 'BNB Smart Chain', 'EVM');

-- 插入币种-链映射
INSERT INTO currency_chains (currency, chain, contract_address, decimals, min_deposit, min_withdraw, withdraw_fee, required_confirmations) VALUES
('BTC', 'BITCOIN', NULL, 8, 0.0001, 0.0005, 0.0002, 3),
('ETH', 'ETHEREUM', NULL, 18, 0.01, 0.01, 0.005, 12),
('USDT', 'ETHEREUM', '0xdac17f958d2ee523a2206206994597c13d831ec7', 6, 10, 20, 1, 12),
('USDT', 'BSC', '0x55d398326f99059ff775485246999027b3197955', 18, 10, 20, 0.8, 15),
('SOL', 'SOLANA', NULL, 9, 0.1, 0.1, 0.001, 32),
('BNB', 'BSC', NULL, 18, 0.01, 0.01, 0.0005, 15);

-- 插入VIP等级
INSERT INTO vip_levels (level, level_name, min_volume_30d, taker_fee_rate, maker_fee_rate, daily_withdraw_limit) VALUES
(0, 'VIP 0', 0, 0.001, 0.001, 10000),
(1, 'VIP 1', 100000, 0.0009, 0.0009, 50000),
(2, 'VIP 2', 500000, 0.0008, 0.0008, 100000),
(3, 'VIP 3', 2000000, 0.0007, 0.0006, 500000),
(4, 'VIP 4', 10000000, 0.0005, 0.0004, 2000000),
(5, 'VIP 5', 50000000, 0.0003, 0.0002, 10000000);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为需要的表添加触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_security_updated_at BEFORE UPDATE ON user_security
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_assets_updated_at BEFORE UPDATE ON user_assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
