# EasiTradeCoins - 项目概述

**项目**: EasiTradeCoins - 专业级加密货币交易平台
**作者**: Aitachi
**联系**: 44158892@qq.com
**日期**: 2025-11-02
**版本**: 1.0

---

## 📋 目录

1. [项目简介](#项目简介)
2. [项目文件结构](#项目文件结构)
3. [技术栈](#技术栈)
4. [系统架构](#系统架构)
5. [数据库设计](#数据库设计)
6. [API文档](#api文档)
7. [部署架构](#部署架构)
8. [开发指南](#开发指南)

---

## 项目简介

### 项目概述

EasiTradeCoins是一个企业级去中心化加密货币交易平台，整合了现货交易、衍生品交易、DeFi生态和社交金融功能。项目采用Go语言构建高性能后端，Solidity开发智能合约，为用户提供安全、高效、专业的数字资产交易服务。

### 核心特性

🚀 **多样化交易**
- 现货交易: 限价单、市价单、止损止盈
- 高级订单: OCO、冰山订单、TWAP
- 自动化策略: 网格交易、DCA定投
- 衍生品: 杠杆交易(1-10x)、期权交易

💎 **DeFi生态**
- DEX聚合器: 多DEX价格聚合，最优路由
- 流动性挖矿: 多池质押，灵活奖励
- 跨链桥接: 多链资产转移(开发中)
- 收益聚合: 自动收益优化(开发中)

👥 **社交金融**
- 跟单交易: 一键跟随专业交易员
- 交易社区: 策略分享，market分析
- 排行榜: 收益、胜率实时排名(开发中)
- NFT徽章: 成就系统(开发中)

🛡️ **企业级安全**
- 完善风控: 订单验证、行为监控、异常检测
- 安全审计: 代码审计、合约审计、依赖扫描
- 多重验证: 2FA、KYC、生物识别
- 资金安全: 冷热钱包分离、多签钱包

⚡ **高性能**
- 撮合引擎: 高频撮合，毫秒级响应
- WebSocket: 实时行情推送
- Redis缓存: 热点数据缓存
- 水平扩展: 微服务架构，可弹性扩容

### 项目统计

| 指标 | 数值 | 说明 |
|------|------|------|
| **总代码行数** | ~26,000 | 包含后端、合约、测试 |
| **Go后端文件** | 35+ | 业务逻辑核心 |
| **智能合约** | 15+ | DeFi功能实现 |
| **测试脚本** | 11 | 完整测试套件 |
| **数据库表** | 34 | PostgreSQL+MySQL |
| **API端点** | 50+ | RESTful API |
| **功能完成度** | 40.3% | 29/72 功能 |

### 项目状态

- **版本**: v0.4.0 (Alpha)
- **状态**: 🔄 持续开发
- **许可**: MIT License
- **代码仓库**: https://github.com/aitachi/easiTradeCoins

---

## 项目文件结构

```
EasiTradeCoins/
│
├── 📁 go-backend/                          # Go后端服务
│   ├── cmd/                                # 命令行工具
│   │   └── server/
│   │       └── main.go                     # 服务入口 (200行)
│   │
│   ├── internal/                           # 内部包
│   │   ├── config/                         # 配置管理
│   │   │   └── config.go                   # 统一配置 (200行)
│   │   │
│   │   ├── services/                       # 业务服务层
│   │   │   ├── auth_service.go             # 用户认证 (300行)
│   │   │   ├── user_service.go             # 用户管理 (250行)
│   │   │   ├── wallet_service.go           # 钱包管理 (350行)
│   │   │   ├── order_service.go            # 订单服务 (400行)
│   │   │   ├── stop_order_service.go       # 止损止盈 (300行)
│   │   │   ├── trailing_stop_service.go    # 跟踪止损 (280行)
│   │   │   ├── conditional_order_service.go# 条件单 (250行)
│   │   │   ├── oco_order_service.go        # OCO订单 (350行)
│   │   │   ├── iceberg_order_service.go    # 冰山订单 (320行)
│   │   │   ├── twap_order_service.go       # TWAP订单 (280行)
│   │   │   ├── grid_trading_service.go     # 网格交易 (518行)
│   │   │   ├── dca_service.go              # DCA定投 (452行)
│   │   │   ├── margin_trading_service.go   # 杠杆交易 (600行)
│   │   │   ├── options_trading_service.go  # 期权交易 (430行)
│   │   │   ├── copy_trading_service.go     # 跟单交易 (550行)
│   │   │   ├── community_service.go        # 交易社区 (400行)
│   │   │   ├── kyc_service.go              # KYC验证 (250行)
│   │   │   ├── audit_service.go            # 审计日志 (200行)
│   │   │   └── services_test.go            # 单元测试 (170行)
│   │   │
│   │   ├── matching/                       # 撮合引擎
│   │   │   ├── engine.go                   # 撮合核心 (400行)
│   │   │   └── orderbook.go                # 订单簿 (300行)
│   │   │
│   │   ├── security/                       # 安全模块
│   │   │   ├── risk_manager.go             # 风控管理 (350行)
│   │   │   ├── rate_limiter.go             # 速率限制 (150行)
│   │   │   └── validator.go                # 输入验证 (200行)
│   │   │
│   │   ├── websocket/                      # WebSocket服务
│   │   │   ├── hub.go                      # 连接管理 (250行)
│   │   │   └── client.go                   # 客户端 (200行)
│   │   │
│   │   └── models/                         # 数据模型
│   │       ├── user.go                     # 用户模型
│   │       ├── order.go                    # 订单模型
│   │       ├── trade.go                    # 交易模型
│   │       └── ... (30+ 模型文件)
│   │
│   ├── go.mod                              # Go模块定义
│   ├── go.sum                              # 依赖校验和
│   └── Dockerfile                          # Docker构建文件
│
├── 📁 contracts/                           # 智能合约
│   ├── src/                                # 合约源码
│   │   ├── DEXAggregator.sol               # DEX聚合器 (280行)
│   │   ├── LiquidityMining.sol             # 流动性挖矿 (280行)
│   │   ├── MockERC20.sol                   # Mock代币 (35行)
│   │   └── ... (12+ 合约文件)
│   │
│   ├── scripts/                            # 部署脚本
│   │   ├── deploy.js                       # 本地部署 (150行)
│   │   └── deploy-sepolia.js               # Sepolia部署 (300行)
│   │
│   ├── test/                               # 合约测试
│   │   └── DEXAggregator.test.js           # 合约测试 (400行)
│   │
│   ├── deployments/                        # 部署记录
│   │   └── sepolia-{timestamp}.json        # Sepolia地址
│   │
│   ├── hardhat.config.js                   # Hardhat配置
│   ├── package.json                        # NPM依赖
│   └── .env                                # 环境变量
│
├── 📁 deployment/                          # 部署配置
│   ├── init_mysql.sql                      # MySQL初始化 (34张表)
│   ├── init_postgres.sql                   # PostgreSQL初始化
│   │
│   ├── nginx/                              # Nginx配置
│   │   └── nginx.conf                      # Nginx配置文件
│   │
│   ├── prometheus/                         # Prometheus配置
│   │   └── prometheus.yml                  # 监控配置
│   │
│   └── grafana/                            # Grafana配置
│       └── dashboards/                     # 仪表板
│
├── 📁 tests/                               # 测试脚本
│   ├── integration_test.sh                 # 集成测试 (200行)
│   ├── performance_test.sh                 # 性能测试 (250行)
│   └── security_audit.sh                   # 安全审计 (300行)
│
├── 📁 test-reports/                        # 测试报告
│   ├── go-unit-tests.log                   # Go测试日志
│   ├── contract-tests.log                  # 合约测试日志
│   ├── integration-tests.log               # 集成测试日志
│   ├── performance-tests.log               # 性能测试日志
│   └── security-audit.log                  # 安全审计日志
│
├── 📁 docs/                                # 文档目录(已废弃)
│
├── 📄 核心文档                             # 项目核心文档
│   ├── README.md                           # 项目主文档 ⭐
│   ├── TESTING.md                          # 测试文档
│   ├── FEATURES.md                         # 功能文档
│   └── PROJECT.md                          # 本文档
│
├── 📄 配置文件                             # 配置文件
│   ├── .env.local                          # 本机环境配置 (120行)
│   ├── docker-compose.yml                  # Docker完整配置 (467行)
│   ├── docker-compose.local.yml            # Docker本机配置 (150行)
│   └── .gitignore                          # Git忽略配置
│
└── 📄 脚本文件                             # 脚本文件
    └── run_all_tests.sh                    # 主测试运行器 (307行)
```

### 文件说明

#### Go后端核心文件

| 文件 | 功能 | 行数 | 重要性 |
|------|------|------|--------|
| `main.go` | 服务启动入口 | 200 | ⭐⭐⭐⭐⭐ |
| `config.go` | 统一配置管理 | 200 | ⭐⭐⭐⭐⭐ |
| `order_service.go` | 订单核心服务 | 400 | ⭐⭐⭐⭐⭐ |
| `matching/engine.go` | 撮合引擎 | 400 | ⭐⭐⭐⭐⭐ |
| `risk_manager.go` | 风控系统 | 350 | ⭐⭐⭐⭐⭐ |
| `margin_trading_service.go` | 杠杆交易 | 600 | ⭐⭐⭐⭐ |
| `options_trading_service.go` | 期权交易 | 430 | ⭐⭐⭐⭐ |
| `grid_trading_service.go` | 网格交易 | 518 | ⭐⭐⭐⭐ |
| `copy_trading_service.go` | 跟单交易 | 550 | ⭐⭐⭐⭐ |
| `websocket/hub.go` | WebSocket管理 | 250 | ⭐⭐⭐⭐ |

#### 智能合约文件

| 文件 | 功能 | 行数 | 网络 |
|------|------|------|------|
| `DEXAggregator.sol` | DEX聚合器 | 280 | Ethereum, BSC |
| `LiquidityMining.sol` | 流动性挖矿 | 280 | Ethereum, BSC |
| `MockERC20.sol` | 测试代币 | 35 | Testnet |

#### 配置文件说明

**`.env.local`** - 本机开发环境配置
```bash
# PostgreSQL配置
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=socialfi
POSTGRES_PASSWORD=socialfi_pg_pass_2024

# MySQL配置
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=root

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka配置
KAFKA_BROKERS=localhost:9092
```

**`docker-compose.yml`** - Docker完整部署配置
- PostgreSQL数据库
- MySQL数据库
- Redis缓存
- Kafka消息队列
- Elasticsearch搜索引擎
- Prometheus监控
- Grafana可视化
- Nginx负载均衡

**`docker-compose.local.yml`** - Docker连接本机服务
- 后端容器
- 通过`host.docker.internal`访问宿主机服务

---

## 技术栈

### 后端技术

| 技术 | 版本 | 用途 |
|------|------|------|
| **Go** | 1.21+ | 后端主语言 |
| **Gin** | 1.9+ | Web框架 |
| **GORM** | 1.25+ | ORM框架 |
| **JWT** | 5.0+ | 身份认证 |
| **WebSocket** | - | 实时通信 |
| **Swagger** | 2.0 | API文档 |

### 数据库

| 数据库 | 版本 | 用途 |
|--------|------|------|
| **PostgreSQL** | 14+ | 主数据库 |
| **MySQL** | 8+ | 辅助存储 |
| **Redis** | 7+ | 缓存&会话 |
| **Kafka** | 3+ | 消息队列 |
| **Elasticsearch** | 8+ | 全文搜索 |

### 智能合约

| 技术 | 版本 | 用途 |
|------|------|------|
| **Solidity** | 0.8.20 | 合约语言 |
| **Hardhat** | 3.0+ | 开发框架 |
| **OpenZeppelin** | 5.4+ | 合约库 |
| **Ethers.js** | 6.0+ | 以太坊库 |

### DevOps

| 工具 | 版本 | 用途 |
|------|------|------|
| **Docker** | 20+ | 容器化 |
| **Docker Compose** | 2.0+ | 容器编排 |
| **Nginx** | 1.24+ | 反向代理 |
| **Prometheus** | 2.45+ | 监控系统 |
| **Grafana** | 10.0+ | 可视化 |

### 测试工具

| 工具 | 用途 |
|------|------|
| **Go Testing** | 单元测试 |
| **Testify** | 断言库 |
| **Hardhat** | 合约测试 |
| **Chai** | 断言库 |
| **Apache Bench** | 性能测试 |
| **Slither** | 合约审计 |

---

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                         用户层                              │
│  Web浏览器  │  移动App  │  API客户端  │  WebSocket客户端  │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                    负载均衡层 (Nginx)                       │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                      应用层 (Go后端)                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ API服务1 │  │ API服务2 │  │ API服务3 │  │  WS服务  │   │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘   │
└────────┼─────────────┼─────────────┼─────────────┼─────────┘
         │             │             │             │
┌────────▼─────────────▼─────────────▼─────────────▼─────────┐
│                       服务层                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │订单服务  │  │撮合引擎  │  │风控系统  │  │钱包服务  │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │杠杆交易  │  │期权交易  │  │网格交易  │  │跟单交易  │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                     数据层                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │PostgreSQL│  │  MySQL   │  │  Redis   │  │  Kafka   │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│  ┌──────────┐                                               │
│  │Elasticsearch│                                            │
│  └──────────┘                                               │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   区块链层 (Ethereum/BSC)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ DEXAggregator│  │LiquidityMining│ │  其他合约    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

### 微服务架构

```
┌────────────────────────────────────────────────────────┐
│                    API Gateway                         │
│              (路由、认证、限流、熔断)                  │
└───────┬────────────────────────────────────────┬───────┘
        │                                        │
┌───────▼──────┐                        ┌───────▼──────┐
│  用户服务    │                        │  交易服务    │
│ ┌──────────┐ │                        │ ┌──────────┐ │
│ │  认证    │ │                        │ │  订单    │ │
│ │  KYC     │ │                        │ │  撮合    │ │
│ │  钱包    │ │                        │ │  成交    │ │
│ └──────────┘ │                        │ └──────────┘ │
└──────────────┘                        └──────────────┘

┌──────────────┐                        ┌──────────────┐
│  衍生品服务  │                        │  社交服务    │
│ ┌──────────┐ │                        │ ┌──────────┐ │
│ │  杠杆    │ │                        │ │  跟单    │ │
│ │  期权    │ │                        │ │  社区    │ │
│ │  网格    │ │                        │ │  排行榜  │ │
│ └──────────┘ │                        │ └──────────┘ │
└──────────────┘                        └──────────────┘

┌──────────────┐                        ┌──────────────┐
│  风控服务    │                        │  数据服务    │
│ ┌──────────┐ │                        │ ┌──────────┐ │
│ │  验证    │ │                        │ │  分析    │ │
│ │  监控    │ │                        │ │  报表    │ │
│ │  审计    │ │                        │ │  统计    │ │
│ └──────────┘ │                        │ └──────────┘ │
└──────────────┘                        └──────────────┘
```

### 数据流架构

```
用户下单 → API Gateway → 订单验证 → 风控检查 → 入库 → 撮合引擎
                                                      │
                                                      ▼
                                                   订单簿
                                                      │
                                                      ▼
                                                   撮合成功
                                                      │
                    ┌─────────────────────────────────┴─────┐
                    │                                       │
                    ▼                                       ▼
                 更新数据库                            WebSocket推送
                    │                                       │
                    │                                       ▼
                    │                                    实时通知
                    ▼
                发送Kafka消息
                    │
         ┌──────────┼──────────┐
         │          │          │
         ▼          ▼          ▼
    风控分析   数据分析   通知服务
```

---

## 数据库设计

### PostgreSQL数据库

**用途**: 核心业务数据存储

#### 核心表结构

##### 1. 用户相关

```sql
-- 用户表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    kyc_level INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 钱包表
CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    currency VARCHAR(10) NOT NULL,
    balance DECIMAL(20,8) DEFAULT 0,
    frozen DECIMAL(20,8) DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, currency)
);
```

##### 2. 交易相关

```sql
-- 订单表
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL,  -- buy/sell
    type VARCHAR(20) NOT NULL, -- limit/market/stop
    price DECIMAL(20,8),
    quantity DECIMAL(20,8) NOT NULL,
    filled_quantity DECIMAL(20,8) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_status (user_id, status),
    INDEX idx_symbol_status (symbol, status)
);

-- 交易表
CREATE TABLE trades (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    buyer_id INT REFERENCES users(id),
    seller_id INT REFERENCES users(id),
    buy_order_id INT REFERENCES orders(id),
    sell_order_id INT REFERENCES orders(id),
    price DECIMAL(20,8) NOT NULL,
    quantity DECIMAL(20,8) NOT NULL,
    amount DECIMAL(20,8) NOT NULL,
    fee DECIMAL(20,8) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_buyer (buyer_id),
    INDEX idx_seller (seller_id),
    INDEX idx_symbol_time (symbol, created_at)
);
```

##### 3. 杠杆交易相关

```sql
-- 杠杆账户表
CREATE TABLE margin_accounts (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) UNIQUE,
    total_assets DECIMAL(20,8) DEFAULT 0,
    total_liabilities DECIMAL(20,8) DEFAULT 0,
    net_assets DECIMAL(20,8) DEFAULT 0,
    margin_level DECIMAL(10,4) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 杠杆持仓表
CREATE TABLE margin_positions (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    account_id INT REFERENCES margin_accounts(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,  -- long/short
    entry_price DECIMAL(20,8) NOT NULL,
    current_price DECIMAL(20,8),
    quantity DECIMAL(20,8) NOT NULL,
    leverage INT NOT NULL,
    margin DECIMAL(20,8) NOT NULL,
    borrowed_amount DECIMAL(20,8) DEFAULT 0,
    unrealized_pnl DECIMAL(20,8) DEFAULT 0,
    liquidation_price DECIMAL(20,8),
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_status (user_id, status)
);
```

##### 4. 期权交易相关

```sql
-- 期权合约表
CREATE TABLE option_contracts (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    type VARCHAR(10) NOT NULL,  -- call/put
    strike_price DECIMAL(20,8) NOT NULL,
    premium DECIMAL(20,8) NOT NULL,
    expiry_date TIMESTAMP NOT NULL,
    style VARCHAR(20) DEFAULT 'european',
    status VARCHAR(20) DEFAULT 'active',
    underlying_price DECIMAL(20,8),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_symbol_expiry (symbol, expiry_date)
);

-- 期权持仓表
CREATE TABLE option_positions (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    contract_id INT REFERENCES option_contracts(id),
    type VARCHAR(10) NOT NULL,  -- buy/sell
    quantity DECIMAL(20,8) NOT NULL,
    entry_price DECIMAL(20,8) NOT NULL,
    current_price DECIMAL(20,8),
    pnl DECIMAL(20,8) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_contract (user_id, contract_id)
);
```

##### 5. 社交金融相关

```sql
-- 交易员表
CREATE TABLE traders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) UNIQUE,
    display_name VARCHAR(100),
    description TEXT,
    win_rate DECIMAL(5,2) DEFAULT 0,
    total_profit DECIMAL(20,8) DEFAULT 0,
    follower_count INT DEFAULT 0,
    max_followers INT DEFAULT 1000,
    commission_rate DECIMAL(5,4) DEFAULT 0.2,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_profit (total_profit DESC)
);

-- 跟单关系表
CREATE TABLE follow_relations (
    id SERIAL PRIMARY KEY,
    follower_id INT REFERENCES users(id),
    trader_id INT REFERENCES traders(id),
    copy_ratio DECIMAL(5,4) DEFAULT 1.0,
    max_copy_amount DECIMAL(20,8),
    stop_loss_ratio DECIMAL(5,4),
    status VARCHAR(20) DEFAULT 'active',
    total_copied INT DEFAULT 0,
    total_profit DECIMAL(20,8) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(follower_id, trader_id),
    INDEX idx_trader (trader_id)
);

-- 交易社区表
CREATE TABLE trading_communities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    creator_id INT REFERENCES users(id),
    member_count INT DEFAULT 0,
    post_count INT DEFAULT 0,
    category VARCHAR(50),
    is_public BOOLEAN DEFAULT true,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 帖子表
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    community_id INT REFERENCES trading_communities(id),
    author_id INT REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(20),  -- 策略/分析/讨论
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    is_top BOOLEAN DEFAULT false,
    tags JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_community_time (community_id, created_at DESC)
);
```

### MySQL数据库

**用途**: 日志、统计、配置等辅助数据

```sql
-- 审计日志表
CREATE TABLE audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT,
    action VARCHAR(100),
    resource VARCHAR(100),
    details JSON,
    ip_address VARCHAR(45),
    status VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_time (user_id, created_at),
    INDEX idx_action (action)
);

-- 系统配置表
CREATE TABLE system_configs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    description VARCHAR(255),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### Redis缓存

**用途**: 热点数据缓存、会话管理

```
缓存结构:

1. 用户会话
   key: session:{token}
   value: {user_id, permissions, ...}
   ttl: 1小时

2. 订单簿缓存
   key: orderbook:{symbol}
   value: {bids: [...], asks: [...]}
   ttl: 10秒

3. 行情数据
   key: ticker:{symbol}
   value: {price, volume, ...}
   ttl: 5秒

4. 速率限制
   key: ratelimit:{ip}:{endpoint}
   value: 请求计数
   ttl: 1分钟
```

---

## API文档

### 认证接口

#### 用户注册
```
POST /api/v1/auth/register
Content-Type: application/json

Request:
{
  "email": "user@example.com",
  "password": "StrongP@ss123",
  "username": "trader001"
}

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 123,
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 用户登录
```
POST /api/v1/auth/login
Content-Type: application/json

Request:
{
  "email": "user@example.com",
  "password": "StrongP@ss123"
}

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 123,
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  }
}
```

### 交易接口

#### 创建订单
```
POST /api/v1/orders
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "symbol": "BTC_USDT",
  "side": "buy",
  "type": "limit",
  "price": "50000.00",
  "quantity": "0.1"
}

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "order_id": 456789,
    "status": "pending",
    "created_at": "2025-11-02T12:00:00Z"
  }
}
```

#### 查询订单
```
GET /api/v1/orders/{order_id}
Authorization: Bearer {token}

Response:
{
  "code": 200,
  "data": {
    "id": 456789,
    "symbol": "BTC_USDT",
    "side": "buy",
    "type": "limit",
    "price": "50000.00",
    "quantity": "0.1",
    "filled_quantity": "0.05",
    "status": "partial_filled",
    "created_at": "2025-11-02T12:00:00Z"
  }
}
```

### 杠杆交易接口

#### 开仓
```
POST /api/v1/margin/open-position
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "symbol": "BTC_USDT",
  "side": "long",
  "entry_price": "50000",
  "quantity": "1.0",
  "leverage": 5
}

Response:
{
  "code": 200,
  "data": {
    "position_id": 123,
    "margin": "10000",
    "borrowed_amount": "40000",
    "liquidation_price": "42000"
  }
}
```

### WebSocket接口

#### 订阅行情
```
ws://localhost:8080/ws

// 连接后发送订阅消息
{
  "action": "subscribe",
  "channel": "ticker",
  "symbol": "BTC_USDT"
}

// 服务器推送
{
  "channel": "ticker",
  "symbol": "BTC_USDT",
  "data": {
    "price": "50000.00",
    "volume_24h": "12345.67",
    "change_24h": "+2.5%",
    "timestamp": 1699123456
  }
}
```

完整API文档请访问: http://localhost:8080/swagger/index.html

---

## 部署架构

### Docker Compose部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  # PostgreSQL数据库
  postgres:
    image: postgres:14
    environment:
      POSTGRES_USER: socialfi
      POSTGRES_PASSWORD: socialfi_pg_pass_2024
      POSTGRES_DB: socialfi
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  # Redis缓存
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  # Kafka消息队列
  kafka:
    image: bitnami/kafka:3.6
    ports:
      - "9092:9092"
    environment:
      - KAFKA_CFG_NODE_ID=0
      - KAFKA_CFG_PROCESS_ROLES=controller,broker
      - KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093
      - KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092

  # Go后端
  backend:
    build: ./go-backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - postgres
      - redis
      - kafka

volumes:
  postgres_data:
  redis_data:
```

### 启动命令

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f backend

# 停止服务
docker-compose down

# 重新构建
docker-compose up -d --build
```

---

## 开发指南

### 本地开发环境搭建

#### 1. 安装依赖

```bash
# Go环境
go version  # 需要 1.21+

# Node.js环境
node --version  # 需要 18+
npm --version

# Docker环境
docker --version  # 需要 20+
docker-compose --version  # 需要 2.0+
```

#### 2. 克隆项目

```bash
git clone https://github.com/aitachi/easiTradeCoins.git
cd easiTradeCoins
```

#### 3. 启动本机服务

```bash
# 启动PostgreSQL
systemctl start postgresql

# 启动MySQL
systemctl start mysql

# 启动Redis
systemctl start redis

# 启动Kafka
systemctl start kafka
```

#### 4. 配置环境变量

```bash
# 复制配置文件
cp .env.local.example .env.local

# 编辑配置
vim .env.local
```

#### 5. 初始化数据库

```bash
# PostgreSQL
psql -U socialfi -d socialfi < deployment/init_postgres.sql

# MySQL
mysql -u root < deployment/init_mysql.sql
```

#### 6. 启动后端服务

```bash
cd go-backend
go mod download
go run cmd/server/main.go
```

#### 7. 访问服务

```bash
# API
curl http://localhost:8080/health

# Swagger文档
open http://localhost:8080/swagger/index.html

# Metrics
curl http://localhost:8081/metrics
```

### 代码规范

#### Go代码规范

```go
// 使用gofmt格式化
gofmt -w .

// 使用golangci-lint检查
golangci-lint run ./...

// 命名规范
// - 包名: 小写单词，如 orderservice
// - 文件名: 小写下划线，如 order_service.go
// - 函数名: 驼峰命名，如 CreateOrder
// - 变量名: 驼峰命名，如 orderID
```

#### Solidity代码规范

```solidity
// 使用solhint检查
npx solhint 'src/**/*.sol'

// 命名规范
// - 合约名: 大写开头驼峰，如 DEXAggregator
// - 函数名: 小写开头驼峰，如 registerDEX
// - 变量名: 小写开头驼峰，如 platformFee
// - 常量: 全大写下划线，如 MAX_UINT256
```

### Git工作流

```bash
# 创建功能分支
git checkout -b feature/new-feature

# 提交代码
git add .
git commit -m "feat: add new feature"

# 推送到远程
git push origin feature/new-feature

# 创建Pull Request
gh pr create --title "Add new feature"
```

### 测试

```bash
# Go单元测试
cd go-backend
go test -v ./...

# 智能合约测试
cd contracts
npx hardhat test

# 集成测试
bash ./tests/integration_test.sh

# 性能测试
bash ./tests/performance_test.sh

# 安全审计
bash ./tests/security_audit.sh
```

---

## 联系方式

**项目**: EasiTradeCoins
**作者**: Aitachi
**邮箱**: 44158892@qq.com
**GitHub**: https://github.com/aitachi/easiTradeCoins
**问题反馈**: https://github.com/aitachi/easiTradeCoins/issues

---

**文档版本**: 1.0
**最后更新**: 2025-11-02
**维护者**: Aitachi
