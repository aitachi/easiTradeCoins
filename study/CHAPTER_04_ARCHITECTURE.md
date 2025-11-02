# 第四章: 系统架构与数据库设计

**作者**: Aitachi
**联系**: 44158892@qq.com
**项目**: EasiTradeCoins - Professional Decentralized Trading Platform
**日期**: 2025-11-02

---

## 目录

1. [系统整体架构](#1-系统整体架构)
2. [数据库设计深度分析](#2-数据库设计深度分析)
3. [基础设施组件](#3-基础设施组件)
4. [Go 语言技术栈应用](#4-go-语言技术栈应用)

---

## 1. 系统整体架构

### 1.1 微服务架构图

```
┌────────────────────── 前端层 ──────────────────────┐
│                                                     │
│  Web Frontend (React/Vue - 未实现)                 │
│  Mobile App (React Native - 未实现)                │
│                                                     │
└──────────────────────┬──────────────────────────────┘
                       │ HTTPS/WSS
┌──────────────────────▼──────────────────────────────┐
│                  Nginx (反向代理)                    │
│  - 负载均衡                                          │
│  - SSL 终止                                          │
│  - 静态资源服务                                      │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│              Go Backend API 服务器                   │
│  ┌─────────────────────────────────────────┐       │
│  │  Gin Web Framework                      │       │
│  │  - RESTful API (/api/v1/*)              │       │
│  │  - WebSocket (/ws)                      │       │
│  │  - Swagger文档 (/swagger/*)             │       │
│  └─────────────────────────────────────────┘       │
│  ┌─────────────────────────────────────────┐       │
│  │  Middleware 中间件                       │       │
│  │  - JWT 认证                              │       │
│  │  - CORS 跨域                             │       │
│  │  - Rate Limiting 限流                    │       │
│  │  - Logging 日志                          │       │
│  └─────────────────────────────────────────┘       │
│  ┌─────────────────────────────────────────┐       │
│  │  Business Services 业务服务              │       │
│  │  - OrderService 订单服务                 │       │
│  │  - UserService 用户服务                  │       │
│  │  - MarginTradingService 杠杆服务         │       │
│  │  - GridTradingService 网格服务           │       │
│  │  - RiskManager 风控服务                  │       │
│  │  - 其他 23+ 服务...                      │       │
│  └─────────────────────────────────────────┘       │
│  ┌─────────────────────────────────────────┐       │
│  │  Matching Engine 撮合引擎                │       │
│  │  - 内存订单簿                            │       │
│  │  - 高性能撮合                            │       │
│  │  - 异步成交通知                          │       │
│  └─────────────────────────────────────────┘       │
└────────────┬────────────┬─────────────┬────────────┘
             │            │             │
             ▼            ▼             ▼
┌────────────────┐ ┌────────────┐ ┌──────────────┐
│  PostgreSQL    │ │   Redis    │ │    Kafka     │
│  (TimescaleDB) │ │  (缓存层)   │ │  (消息队列)   │
│                │ │            │ │              │
│  - 主数据库    │ │  - 会话    │ │  - 成交事件  │
│  - 34个表      │ │  - 订单簿  │ │  - 风控通知  │
│  - 时序数据    │ │  - 限流    │ │  - 用户通知  │
└────────────────┘ └────────────┘ └──────────────┘

┌──────────────────── 辅助服务 ───────────────────────┐
│                                                     │
│  Elasticsearch  │  Prometheus   │   Grafana        │
│  (日志/搜索)     │  (监控指标)    │  (可视化面板)     │
│                                                     │
└─────────────────────────────────────────────────────┘

┌──────────────────── 区块链层 ───────────────────────┐
│                                                     │
│  Smart Contracts (Ethereum/BSC/Polygon)            │
│  - DEXAggregator 去中心化交易聚合                   │
│  - LiquidityMining 流动性挖矿                       │
│  - EasiToken 平台代币                               │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### 1.2 数据流向

```
用户下单:
  用户 → API → 风控验证 → 撮合引擎 → 生成成交 → Kafka → 持久化到数据库

实时推送:
  撮合引擎 → WebSocket Hub → WebSocket Client → 前端

缓存策略:
  API → Redis查询 → 命中返回
              ↓ 未命中
            数据库查询 → 写入Redis → 返回

异步任务:
  Kafka → 消费者服务 → 执行业务逻辑 → 更新数据库
```

---

## 2. 数据库设计深度分析

### 2.1 数据库架构

**文件**: `deployment/init_mysql.sql` (752 行, 34 张表)

#### 核心表分类

| 分类 | 表数量 | 主要表 |
|------|--------|--------|
| **用户与资产** | 3 | users, user_assets, kyc_verifications |
| **交易核心** | 5 | orders, trades, trading_pairs, stop_orders, oco_orders |
| **高级交易** | 8 | iceberg_orders, twap_orders, grid_strategies, dca_strategies, margin_positions, option_contracts |
| **社交金融** | 6 | traders, follow_relations, copy_trades, trading_communities, posts, comments |
| **资金管理** | 2 | deposits, withdrawals |
| **风控安全** | 5 | audit_logs, risk_events, violations, withdrawal_whitelists, api_keys |
| **系统配置** | 3 | system_configs, notifications, market_data, rate_limits |

### 2.2 核心表设计详解

#### 2.2.1 用户表 (users)

```sql
CREATE TABLE users (
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
```

**设计亮点**:
1. ✅ **双重唯一性**: email 和 phone 都是唯一索引
2. ✅ **密码安全**: 使用 salt 加密
3. ✅ **KYC 分级**: 不同等级不同权限
4. ✅ **IP 追踪**: 注册和登录 IP 记录
5. ✅ **状态管理**: 正常/冻结/注销

#### 2.2.2 用户资产表 (user_assets)

```sql
CREATE TABLE user_assets (
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
) ENGINE=InnoDB;
```

**关键特性**:
1. ✅ **高精度**: DECIMAL(36, 18) 支持极高精度
2. ✅ **冻结机制**: available + frozen 分离
3. ✅ **多链支持**: 同一币种可在不同链上
4. ✅ **复合唯一键**: (user_id, currency, chain)
5. ✅ **级联删除**: 用户删除时自动删除资产

**精度选择**:
```
DECIMAL(36, 18):
- 整数部分: 18 位
- 小数部分: 18 位
- 最大值: 999,999,999,999,999,999.999999999999999999
- 适用于: BTC (8位小数), ETH (18位小数), ERC20代币
```

#### 2.2.3 订单表 (orders)

```sql
CREATE TABLE orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL COMMENT 'buy/sell',
    type VARCHAR(20) NOT NULL COMMENT 'limit/market/stop_loss/take_profit',
    price DECIMAL(36, 18),
    quantity DECIMAL(36, 18) NOT NULL,
    filled_qty DECIMAL(36, 18) DEFAULT 0,
    filled_amount DECIMAL(36, 18) DEFAULT 0,
    avg_price DECIMAL(36, 18),
    fee DECIMAL(36, 18) DEFAULT 0,
    status VARCHAR(10) COMMENT 'pending/partial/filled/cancelled',
    time_in_force VARCHAR(3) COMMENT 'GTC/IOC/FOK',

    -- 高级订单字段
    stop_price DECIMAL(36, 18),
    take_profit_price DECIMAL(36, 18),
    trailing_delta DECIMAL(36, 18),
    trigger_condition VARCHAR(10),
    is_triggered BOOLEAN DEFAULT FALSE,
    trigger_time DATETIME,

    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_user_status_create_time (user_id, status, create_time DESC),
    INDEX idx_type_is_triggered (type, is_triggered),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;
```

**索引策略**:
1. ✅ **单列索引**: 常用查询字段
2. ✅ **复合索引**: 组合查询优化
3. ✅ **覆盖索引**: 包含查询所需所有字段
4. ✅ **排序优化**: create_time DESC

**查询优化示例**:
```sql
-- 使用复合索引
SELECT * FROM orders
WHERE user_id = 1000
  AND status = 'pending'
ORDER BY create_time DESC
LIMIT 10;
-- 使用索引: idx_user_status_create_time

-- 使用单列索引
SELECT * FROM orders
WHERE symbol = 'BTC_USDT'
  AND create_time > '2025-01-01';
-- 使用索引: idx_symbol
```

#### 2.2.4 成交表 (trades)

```sql
CREATE TABLE trades (
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
    INDEX idx_symbol_trade_time (symbol, trade_time DESC),
    FOREIGN KEY (buyer_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (seller_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;
```

**时序优化**:
- trade_time 索引支持快速时间范围查询
- 复合索引 (symbol, trade_time) 支持按交易对查询历史

#### 2.2.5 杠杆持仓表 (margin_positions)

```sql
CREATE TABLE margin_positions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(5) NOT NULL COMMENT 'long/short',
    entry_price DECIMAL(36, 18) NOT NULL,
    current_price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    leverage INT NOT NULL,
    margin DECIMAL(36, 18) NOT NULL COMMENT '保证金',
    unrealized_pnl DECIMAL(36, 18) DEFAULT 0 COMMENT '未实现盈亏',
    realized_pnl DECIMAL(36, 18) DEFAULT 0 COMMENT '已实现盈亏',
    liquidation_price DECIMAL(36, 18) NOT NULL COMMENT '强平价',
    stop_loss DECIMAL(36, 18),
    take_profit DECIMAL(36, 18),
    status VARCHAR(10) DEFAULT 'open' COMMENT 'open/closed/liquidated',
    open_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    close_time DATETIME,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;
```

**风控字段**:
- liquidation_price: 强平价计算
- unrealized_pnl: 实时盈亏跟踪
- stop_loss/take_profit: 自动止损止盈

### 2.3 数据库性能优化

#### 2.3.1 分区策略 (建议)

```sql
-- 对于大量历史数据的表,使用分区
ALTER TABLE trades
PARTITION BY RANGE (YEAR(trade_time)) (
    PARTITION p2023 VALUES LESS THAN (2024),
    PARTITION p2024 VALUES LESS THAN (2025),
    PARTITION p2025 VALUES LESS THAN (2026),
    PARTITION pmax VALUES LESS THAN MAXVALUE
);
```

#### 2.3.2 TimescaleDB 时序优化

```sql
-- 将 trades 表转换为时序表 (PostgreSQL + TimescaleDB)
SELECT create_hypertable('trades', 'trade_time');

-- 自动压缩历史数据
ALTER TABLE trades SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'symbol'
);

-- 压缩 30 天前的数据
SELECT add_compression_policy('trades', INTERVAL '30 days');
```

---

## 3. 基础设施组件

### 3.1 Redis 缓存策略

**用途**:
1. **会话管理**: JWT token 黑名单
2. **订单簿缓存**: 热门交易对深度数据
3. **行情缓存**: Ticker, K线数据
4. **限流计数**: 用户 API 调用频率

**Key 设计**:
```
session:{token}           → 用户会话
orderbook:{symbol}        → 订单簿深度
ticker:{symbol}           → 最新价格
kline:{symbol}:{period}   → K线数据
ratelimit:{user_id}       → 限流计数
```

### 3.2 Kafka 消息队列

**Topics 设计**:
```
trades.executed           → 成交事件
orders.created            → 订单创建
orders.cancelled          → 订单取消
risk.alerts               → 风控告警
user.notifications        → 用户通知
```

**消费者组**:
```
persistence-group         → 持久化服务
notification-group        → 通知服务
analytics-group           → 数据分析
```

### 3.3 Elasticsearch

**索引设计**:
```
audit_logs-2025-01        → 审计日志
trades-2025-01            → 成交记录
user_activities-2025-01   → 用户行为
```

**用途**:
- 全文搜索
- 日志分析
- 审计追踪

### 3.4 监控体系

**Prometheus 指标**:
```
# 业务指标
orders_total              → 订单总数
trades_total              → 成交总数
active_users              → 活跃用户数
trading_volume            → 交易量

# 性能指标
api_request_duration      → API 响应时间
matching_latency          → 撮合延迟
orderbook_depth           → 订单簿深度

# 系统指标
go_goroutines             → Goroutine 数量
go_memstats_alloc         → 内存使用
```

**Grafana 面板**:
- 实时交易大盘
- 系统性能监控
- 用户行为分析
- 风控告警面板

---

## 4. Go 语言技术栈应用

### 4.1 核心库使用

| 库 | 版本 | 用途 |
|---|------|------|
| **gin-gonic/gin** | v1.9+ | Web 框架 |
| **gorm.io/gorm** | v1.25+ | ORM |
| **shopspring/decimal** | latest | 精确计算 |
| **golang-jwt/jwt** | v5+ | JWT 认证 |
| **gorilla/websocket** | latest | WebSocket |
| **redis/go-redis** | v9+ | Redis 客户端 |
| **segmentio/kafka-go** | latest | Kafka 客户端 |
| **google/uuid** | latest | UUID 生成 |
| **prometheus/client_golang** | latest | 监控指标 |

### 4.2 并发模型

```go
// Goroutine 池模式
type WorkerPool struct {
	tasks    chan Task
	workers  int
	wg       sync.WaitGroup
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for task := range p.tasks {
		task.Execute()
	}
}
```

### 4.3 错误处理

```go
// 自定义错误类型
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("code:%d, message:%s", e.Code, e.Message)
}

// 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			if apiErr, ok := err.(*APIError); ok {
				c.JSON(apiErr.Code, apiErr)
			} else {
				c.JSON(500, gin.H{"error": err.Error()})
			}
		}
	}
}
```

### 4.4 优雅关闭

```go
func main() {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 5秒超时关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
```

---

## 总结

### 架构优势

1. ✅ **微服务架构**: 模块解耦,易于扩展
2. ✅ **多层缓存**: Redis + 内存,性能优异
3. ✅ **异步处理**: Kafka 削峰填谷
4. ✅ **完善监控**: Prometheus + Grafana
5. ✅ **数据库优化**: 索引 + 分区 + 时序

### 可扩展性

- **水平扩展**: 无状态 API 服务器
- **垂直扩展**: 数据库读写分离
- **分库分表**: 支持用户数和交易量增长
- **多区域部署**: 跨地域容灾

### 性能指标 (设计目标)

| 指标 | 目标 |
|------|------|
| **API 响应时间** | <100ms (P99) |
| **撮合延迟** | <10ms |
| **吞吐量** | >10,000 TPS |
| **并发用户** | >100,000 |
| **可用性** | 99.9% |

---

**下一章预告**: [第五章: 安全、风控与项目分析](./CHAPTER_05_SECURITY_ANALYSIS.md)

---

**文档版本**: v1.0
**最后更新**: 2025-11-02
**作者**: Aitachi (44158892@qq.com)
