-- EasiTradeCoins 虚拟测试数据
-- 请在执行schema.sql后运行此脚本

-- =================================================================================
-- 1. 创建测试用户
-- =================================================================================

-- 创建10个测试用户 (密码都是: Test123456!)
-- 盐: test_salt_xxxx
-- 密码哈希是用 bcrypt(password + salt, 10) 生成的

INSERT INTO users (email, phone, password_hash, salt, kyc_level, status, vip_level, register_ip, nickname) VALUES
('admin@easitradecoins.com', '+8613800000001', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'admin_salt_001', 2, 1, 5, '192.168.1.1', 'Admin User'),
('trader1@test.com', '+8613800000002', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'test_salt_001', 1, 1, 2, '192.168.1.2', 'Trader One'),
('trader2@test.com', '+8613800000003', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'test_salt_002', 1, 1, 1, '192.168.1.3', 'Trader Two'),
('trader3@test.com', '+8613800000004', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'test_salt_003', 0, 1, 0, '192.168.1.4', 'Trader Three'),
('whale1@test.com', '+8613800000005', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'test_salt_004', 2, 1, 4, '192.168.1.5', 'Whale Trader'),
('maker1@test.com', '+8613800000006', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'test_salt_005', 1, 1, 3, '192.168.1.6', 'Market Maker 1'),
('maker2@test.com', '+8613800000007', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'test_salt_006', 1, 1, 3, '192.168.1.7', 'Market Maker 2'),
('newbie1@test.com', '+8613800000008', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'test_salt_007', 0, 1, 0, '192.168.1.8', 'New Trader 1'),
('newbie2@test.com', '+8613800000009', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'test_salt_008', 0, 1, 0, '192.168.1.9', 'New Trader 2'),
('frozen@test.com', '+8613800000010', '$2b$10$X1Y2Z3A4B5C6D7E8F9G0H1I2J3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y', 'test_salt_009', 0, 2, 0, '192.168.1.10', 'Frozen User');

-- =================================================================================
-- 2. 创建用户安全配置
-- =================================================================================

INSERT INTO user_security (user_id, is_2fa_enabled, login_password_error_count)
SELECT user_id, false, 0 FROM users;

-- 为admin用户启用2FA
UPDATE user_security SET
  google_2fa_secret = 'JBSWY3DPEHPK3PXP',
  is_2fa_enabled = true
WHERE user_id = (SELECT user_id FROM users WHERE email = 'admin@easitradecoins.com');

-- =================================================================================
-- 3. 初始化用户资产
-- =================================================================================

-- 为所有活跃用户创建主要币种资产
DO $$
DECLARE
    user_rec RECORD;
BEGIN
    FOR user_rec IN SELECT user_id, email FROM users WHERE status = 1 LOOP
        -- BTC资产
        INSERT INTO user_assets (user_id, currency, chain, available, frozen) VALUES
        (user_rec.user_id, 'BTC', 'BITCOIN',
         CASE
            WHEN user_rec.email = 'admin@easitradecoins.com' THEN 100
            WHEN user_rec.email LIKE 'whale%' THEN 10
            WHEN user_rec.email LIKE 'maker%' THEN 5
            WHEN user_rec.email LIKE 'trader%' THEN 1
            ELSE 0.1
         END, 0);

        -- ETH资产
        INSERT INTO user_assets (user_id, currency, chain, available, frozen) VALUES
        (user_rec.user_id, 'ETH', 'ETHEREUM',
         CASE
            WHEN user_rec.email = 'admin@easitradecoins.com' THEN 1000
            WHEN user_rec.email LIKE 'whale%' THEN 100
            WHEN user_rec.email LIKE 'maker%' THEN 50
            WHEN user_rec.email LIKE 'trader%' THEN 10
            ELSE 1
         END, 0);

        -- USDT资产 (ERC20)
        INSERT INTO user_assets (user_id, currency, chain, available, frozen) VALUES
        (user_rec.user_id, 'USDT', 'ETHEREUM',
         CASE
            WHEN user_rec.email = 'admin@easitradecoins.com' THEN 1000000
            WHEN user_rec.email LIKE 'whale%' THEN 500000
            WHEN user_rec.email LIKE 'maker%' THEN 100000
            WHEN user_rec.email LIKE 'trader%' THEN 10000
            ELSE 1000
         END, 0);

        -- USDT资产 (BSC)
        INSERT INTO user_assets (user_id, currency, chain, available, frozen) VALUES
        (user_rec.user_id, 'USDT', 'BSC',
         CASE
            WHEN user_rec.email = 'admin@easitradecoins.com' THEN 500000
            WHEN user_rec.email LIKE 'whale%' THEN 250000
            WHEN user_rec.email LIKE 'maker%' THEN 50000
            WHEN user_rec.email LIKE 'trader%' THEN 5000
            ELSE 500
         END, 0);

        -- SOL资产
        INSERT INTO user_assets (user_id, currency, chain, available, frozen) VALUES
        (user_rec.user_id, 'SOL', 'SOLANA',
         CASE
            WHEN user_rec.email = 'admin@easitradecoins.com' THEN 10000
            WHEN user_rec.email LIKE 'whale%' THEN 5000
            WHEN user_rec.email LIKE 'maker%' THEN 1000
            WHEN user_rec.email LIKE 'trader%' THEN 100
            ELSE 10
         END, 0);

        -- BNB资产
        INSERT INTO user_assets (user_id, currency, chain, available, frozen) VALUES
        (user_rec.user_id, 'BNB', 'BSC',
         CASE
            WHEN user_rec.email = 'admin@easitradecoins.com' THEN 1000
            WHEN user_rec.email LIKE 'whale%' THEN 500
            WHEN user_rec.email LIKE 'maker%' THEN 100
            WHEN user_rec.email LIKE 'trader%' THEN 10
            ELSE 1
         END, 0);
    END LOOP;
END $$;

-- =================================================================================
-- 4. 创建历史订单数据
-- =================================================================================

-- 创建一些已完成的历史订单
DO $$
DECLARE
    trader1_id INT;
    trader2_id INT;
    whale_id INT;
    order_count INT := 0;
BEGIN
    SELECT user_id INTO trader1_id FROM users WHERE email = 'trader1@test.com';
    SELECT user_id INTO trader2_id FROM users WHERE email = 'trader2@test.com';
    SELECT user_id INTO whale_id FROM users WHERE email = 'whale1@test.com';

    -- 创建50个历史订单
    FOR i IN 1..50 LOOP
        -- 买单
        INSERT INTO orders (
            user_id, symbol, side, type, price, quantity,
            filled_quantity, filled_amount, avg_price, fee, fee_currency,
            status, time_in_force, created_at, completed_at
        ) VALUES (
            CASE WHEN i % 3 = 0 THEN trader1_id WHEN i % 3 = 1 THEN trader2_id ELSE whale_id END,
            CASE WHEN i % 4 = 0 THEN 'BTC_USDT' WHEN i % 4 = 1 THEN 'ETH_USDT' WHEN i % 4 = 2 THEN 'SOL_USDT' ELSE 'BNB_USDT' END,
            'buy',
            'limit',
            CASE
                WHEN i % 4 = 0 THEN 34000 + (i * 10)
                WHEN i % 4 = 1 THEN 1800 + (i * 5)
                WHEN i % 4 = 2 THEN 20 + (i * 0.1)
                ELSE 250 + (i * 1)
            END,
            CASE
                WHEN i % 4 = 0 THEN 0.1
                WHEN i % 4 = 1 THEN 1
                WHEN i % 4 = 2 THEN 10
                ELSE 5
            END,
            CASE
                WHEN i % 4 = 0 THEN 0.1
                WHEN i % 4 = 1 THEN 1
                WHEN i % 4 = 2 THEN 10
                ELSE 5
            END,
            CASE
                WHEN i % 4 = 0 THEN (34000 + (i * 10)) * 0.1
                WHEN i % 4 = 1 THEN (1800 + (i * 5)) * 1
                WHEN i % 4 = 2 THEN (20 + (i * 0.1)) * 10
                ELSE (250 + (i * 1)) * 5
            END,
            CASE
                WHEN i % 4 = 0 THEN 34000 + (i * 10)
                WHEN i % 4 = 1 THEN 1800 + (i * 5)
                WHEN i % 4 = 2 THEN 20 + (i * 0.1)
                ELSE 250 + (i * 1)
            END,
            CASE
                WHEN i % 4 = 0 THEN (34000 + (i * 10)) * 0.1 * 0.001
                WHEN i % 4 = 1 THEN (1800 + (i * 5)) * 1 * 0.001
                WHEN i % 4 = 2 THEN (20 + (i * 0.1)) * 10 * 0.001
                ELSE (250 + (i * 1)) * 5 * 0.001
            END,
            'USDT',
            2,
            'GTC',
            NOW() - INTERVAL '1 day' * i,
            NOW() - INTERVAL '1 day' * i + INTERVAL '5 minutes'
        );
    END LOOP;
END $$;

-- =================================================================================
-- 5. 创建充值记录
-- =================================================================================

INSERT INTO deposits (
    user_id, currency, chain, amount, address, from_address,
    txid, confirmations, required_confirmations, status, confirmed_at
)
SELECT
    u.user_id,
    'USDT',
    'ETHEREUM',
    10000,
    '0x' || md5(random()::text),
    '0x' || md5(random()::text),
    '0x' || md5(random()::text || random()::text),
    12,
    12,
    1,
    NOW() - INTERVAL '1 day'
FROM users u
WHERE u.email LIKE 'trader%'
LIMIT 5;

-- =================================================================================
-- 6. 创建提现记录
-- =================================================================================

INSERT INTO withdrawals (
    user_id, currency, chain, amount, fee, actual_amount,
    address, txid, status, complete_time
)
SELECT
    u.user_id,
    'USDT',
    'ETHEREUM',
    1000,
    1,
    999,
    '0x' || md5(random()::text),
    '0x' || md5(random()::text || random()::text),
    3,
    NOW() - INTERVAL '12 hours'
FROM users u
WHERE u.email IN ('trader1@test.com', 'trader2@test.com')
LIMIT 2;

-- =================================================================================
-- 7. 创建K线数据 (BTC_USDT 1小时线,最近24小时)
-- =================================================================================

DO $$
DECLARE
    base_price DECIMAL := 34000;
    base_time TIMESTAMP := DATE_TRUNC('hour', NOW() - INTERVAL '24 hours');
BEGIN
    FOR i IN 0..23 LOOP
        INSERT INTO klines (
            symbol, interval, open_time, close_time,
            open_price, high_price, low_price, close_price,
            volume, quote_volume, trade_count
        ) VALUES (
            'BTC_USDT',
            '1h',
            base_time + (i || ' hours')::INTERVAL,
            base_time + ((i + 1) || ' hours')::INTERVAL,
            base_price + (random() * 200 - 100),
            base_price + (random() * 300),
            base_price - (random() * 300),
            base_price + (random() * 200 - 100),
            random() * 100 + 50,
            (base_price + (random() * 200 - 100)) * (random() * 100 + 50),
            floor(random() * 1000 + 500)::INT
        );
        base_price := base_price + (random() * 200 - 100);
    END LOOP;
END $$;

-- =================================================================================
-- 8. 数据验证查询
-- =================================================================================

-- 查看用户统计
SELECT
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE status = 1) as active_users,
    COUNT(*) FILTER (WHERE kyc_level > 0) as kyc_users,
    COUNT(*) FILTER (WHERE vip_level > 0) as vip_users
FROM users;

-- 查看资产统计
SELECT
    currency,
    chain,
    COUNT(DISTINCT user_id) as holders,
    SUM(available::DECIMAL) as total_available,
    SUM(frozen::DECIMAL) as total_frozen
FROM user_assets
GROUP BY currency, chain
ORDER BY currency, chain;

-- 查看订单统计
SELECT
    symbol,
    COUNT(*) as total_orders,
    COUNT(*) FILTER (WHERE status = 2) as completed_orders,
    SUM(filled_amount::DECIMAL) FILTER (WHERE status = 2) as total_volume
FROM orders
GROUP BY symbol;

-- 显示测试用户信息
SELECT
    user_id,
    email,
    nickname,
    kyc_level,
    vip_level,
    status
FROM users
ORDER BY user_id;

-- 显示测试用户余额
SELECT
    u.email,
    u.nickname,
    ua.currency,
    ua.chain,
    ua.available,
    ua.frozen
FROM users u
JOIN user_assets ua ON u.user_id = ua.user_id
WHERE u.email LIKE '%test.com'
ORDER BY u.user_id, ua.currency, ua.chain;
