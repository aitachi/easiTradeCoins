# EasiTradeCoins 部署指南

## 📋 目录

1. [系统要求](#系统要求)
2. [快速开始](#快速开始)
3. [开发环境部署](#开发环境部署)
4. [生产环境部署](#生产环境部署)
5. [配置说明](#配置说明)
6. [监控和日志](#监控和日志)
7. [故障排查](#故障排查)
8. [性能优化](#性能优化)

---

## 系统要求

### 硬件要求

**开发环境**:
- CPU: 2核心+
- 内存: 4GB+
- 硬盘: 20GB+

**生产环境**:
- CPU: 8核心+
- 内存: 16GB+
- 硬盘: 100GB+ (SSD推荐)

### 软件要求

- Docker: 20.10+
- Docker Compose: 2.0+
- Git: 2.30+

---

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/your-org/EasiTradeCoins.git
cd EasiTradeCoins
```

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑.env文件,设置必要的配置
# 重点配置:
# - POSTGRES_PASSWORD
# - REDIS_PASSWORD
# - JWT_SECRET
# - ETHEREUM_RPC_URL
# - PRIVATE_KEY (用于智能合约交互)
```

### 3. 启动服务

```bash
# 启动所有核心服务
docker-compose up -d

# 查看日志
docker-compose logs -f backend

# 检查服务状态
docker-compose ps
```

### 4. 初始化数据库

数据库会自动使用 `deployment/init_mysql.sql` 初始化。如需手动初始化:

```bash
docker-compose exec postgres psql -U postgres -d easitradecoins -f /docker-entrypoint-initdb.d/init.sql
```

### 5. 访问服务

- **API文档**: http://localhost:8080/swagger/index.html
- **健康检查**: http://localhost:8080/health
- **Prometheus指标**: http://localhost:8081/metrics

---

## 开发环境部署

### 启动开发环境

```bash
# 启动核心服务 + 开发工具
docker-compose --profile dev up -d

# 可用的开发工具:
# - PgAdmin: http://localhost:5050 (数据库管理)
# - Redis Commander: http://localhost:8082 (Redis管理)
```

### 开发工具访问

#### PgAdmin (PostgreSQL管理)
- URL: http://localhost:5050
- Email: admin@easitrade.com (可在.env中配置)
- Password: 在.env中设置

连接数据库:
- Host: postgres
- Port: 5432
- Database: easitradecoins
- Username: postgres
- Password: 在.env中设置

#### Redis Commander
- URL: http://localhost:8082
- 自动连接到Redis

### 本地开发

如果需要在本地运行Go后端(不使用Docker):

```bash
cd go-backend

# 安装依赖
go mod download

# 运行数据库迁移
# (确保Docker的postgres和redis在运行)

# 启动服务
go run cmd/server/main.go
```

---

## 生产环境部署

### 1. 使用生产配置

```bash
# 启动生产环境 (包含Redis集群)
docker-compose --profile production up -d
```

### 2. 启用监控栈

```bash
# 启动Prometheus + Grafana + Alertmanager
docker-compose --profile monitoring up -d

# 访问监控面板:
# - Grafana: http://localhost:3000
# - Prometheus: http://localhost:9090
# - Alertmanager: http://localhost:9093
```

### 3. 启用日志栈

```bash
# 启动ELK栈 (Elasticsearch + Logstash + Kibana)
docker-compose --profile logging up -d

# 访问Kibana: http://localhost:5601
```

### 4. 完整生产环境

```bash
# 一次性启动所有生产服务
docker-compose --profile production --profile monitoring --profile logging up -d
```

### 5. 配置Redis集群

Redis集群需要手动初始化:

```bash
# 创建集群
docker exec -it easitrade-redis-1 redis-cli --cluster create \
  172.25.0.11:6379 172.25.0.12:6379 172.25.0.13:6379 \
  --cluster-replicas 0 \
  -a your-redis-password
```

---

## 配置说明

### 环境变量分类

#### 必需配置 (生产环境必须修改)

```bash
# 数据库密码
POSTGRES_PASSWORD=生产环境强密码

# Redis密码
REDIS_PASSWORD=生产环境强密码

# JWT密钥 (至少64字符)
JWT_SECRET=生产环境随机字符串-至少64字符

# 以太坊RPC
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR-PROJECT-ID

# 私钥 (用于智能合约交互,绝对保密!)
PRIVATE_KEY=0x...
```

#### 可选配置

```bash
# 风控参数
ORDER_RATE_LIMIT=10
MAX_PRICE_DEVIATION=0.20
WITHDRAWAL_DAILY_LIMIT=100000

# 功能开关
ENABLE_STOP_ORDER_MONITOR=true
ENABLE_RISK_MANAGER=true
ENABLE_SWAGGER=true
```

### 智能合约地址配置

部署智能合约后,更新以下地址:

```bash
CONTRACT_ADDRESS_STAKING=0x...
CONTRACT_ADDRESS_AIRDROP=0x...
CONTRACT_ADDRESS_TOKEN_FACTORY=0x...
CONTRACT_ADDRESS_MULTISIG=0x...
```

---

## 监控和日志

### Prometheus监控

**指标收集**:
- 应用指标: http://localhost:8081/metrics
- 系统指标: Node Exporter
- 数据库指标: Postgres Exporter
- 缓存指标: Redis Exporter

**关键指标**:
- `http_requests_total`: HTTP请求总数
- `http_request_duration_seconds`: 请求延迟
- `order_processing_duration_seconds`: 订单处理时间
- `active_orders_count`: 活跃订单数
- `trades_executed_total`: 成交笔数

### Grafana仪表盘

访问 http://localhost:3000

默认登录:
- Username: admin
- Password: 在.env中配置

推荐仪表盘:
1. **系统概览**: CPU, 内存, 磁盘使用率
2. **应用性能**: 请求QPS, 延迟分布
3. **交易监控**: 订单量, 成交量, 撮合延迟
4. **风控监控**: 风控事件, 违规行为统计

### ELK日志分析

访问 Kibana: http://localhost:5601

**日志收集流程**:
1. 应用输出JSON格式日志
2. Logstash收集和解析
3. Elasticsearch存储
4. Kibana可视化

**日志查询示例**:
```
# 查询风控事件
log_level: "warn" AND event_type: "risk_event"

# 查询订单创建
service: "order" AND action: "create"

# 查询错误日志
log_level: "error"
```

---

## 故障排查

### 常见问题

#### 1. 数据库连接失败

```bash
# 检查PostgreSQL状态
docker-compose ps postgres

# 查看日志
docker-compose logs postgres

# 手动连接测试
docker-compose exec postgres psql -U postgres -d easitradecoins
```

#### 2. Redis连接失败

```bash
# 检查Redis状态
docker-compose ps redis

# 测试连接
docker-compose exec redis redis-cli -a your-password ping
```

#### 3. 后端服务无法启动

```bash
# 查看详细日志
docker-compose logs --tail=100 backend

# 检查健康状态
curl http://localhost:8080/health

# 进入容器调试
docker-compose exec backend sh
```

#### 4. Kafka无法连接

```bash
# 检查Zookeeper
docker-compose logs zookeeper

# 检查Kafka
docker-compose logs kafka

# 测试Kafka
docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
```

### 健康检查

```bash
# 检查所有服务健康状态
docker-compose ps

# 检查API健康
curl http://localhost:8080/health

# 检查数据库
docker-compose exec postgres pg_isready -U postgres

# 检查Redis
docker-compose exec redis redis-cli -a your-password ping
```

### 日志查看

```bash
# 实时查看所有日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend

# 查看最近100行日志
docker-compose logs --tail=100 backend

# 查看带时间戳的日志
docker-compose logs -t backend
```

---

## 性能优化

### 数据库优化

```sql
-- 分析查询性能
EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 1;

-- 更新统计信息
ANALYZE orders;

-- 重建索引
REINDEX TABLE orders;
```

### Redis优化

```bash
# 查看内存使用
docker-compose exec redis redis-cli -a your-password INFO memory

# 查看慢查询
docker-compose exec redis redis-cli -a your-password SLOWLOG GET 10

# 监控实时命令
docker-compose exec redis redis-cli -a your-password MONITOR
```

### 应用优化

**连接池配置**:
```bash
# .env中调整
DB_MAX_CONNECTIONS=100
DB_MAX_IDLE_CONNECTIONS=10
```

**缓存TTL调整**:
```bash
CACHE_TTL_PRICES=5s
CACHE_TTL_ORDER_BOOK=1s
CACHE_TTL_USER_PROFILE=5m
```

### 负载均衡

启用多个后端实例:

```bash
# 调整.env
BACKEND_REPLICAS=3

# 重启服务
docker-compose up -d --scale backend=3
```

---

## 备份和恢复

### 数据库备份

```bash
# 手动备份
docker-compose exec postgres pg_dump -U postgres easitradecoins > backup_$(date +%Y%m%d).sql

# 自动备份 (cron任务)
0 2 * * * docker-compose exec postgres pg_dump -U postgres easitradecoins > /backups/easitrade_$(date +\%Y\%m\%d).sql
```

### 数据库恢复

```bash
# 恢复备份
cat backup_20250102.sql | docker-compose exec -T postgres psql -U postgres easitradecoins
```

### Redis备份

```bash
# Redis会自动持久化到 redis_data volume
# 手动触发保存
docker-compose exec redis redis-cli -a your-password BGSAVE
```

---

## 安全检查清单

- [ ] 修改所有默认密码
- [ ] 生成强随机JWT_SECRET (64+字符)
- [ ] 启用HTTPS (生产环境)
- [ ] 配置防火墙规则
- [ ] 限制数据库远程访问
- [ ] 定期更新Docker镜像
- [ ] 配置备份策略
- [ ] 启用监控告警
- [ ] 审计日志定期检查
- [ ] 私钥绝对保密,不提交到版本控制

---

## 升级指南

### 滚动更新

```bash
# 拉取最新镜像
docker-compose pull

# 滚动重启服务
docker-compose up -d --no-deps --build backend

# 验证新版本
curl http://localhost:8080/health
```

### 数据库迁移

```bash
# 1. 备份数据库
docker-compose exec postgres pg_dump -U postgres easitradecoins > backup_before_migration.sql

# 2. 运行迁移脚本
docker-compose exec postgres psql -U postgres -d easitradecoins -f /path/to/migration.sql

# 3. 验证迁移
docker-compose exec postgres psql -U postgres -d easitradecoins -c "\dt"
```

---

## 支持和反馈

- 问题反馈: https://github.com/your-org/EasiTradeCoins/issues
- 技术文档: https://docs.easitrade.com
- 社区讨论: https://community.easitrade.com

---

生成时间: 2025-01-XX
版本: 1.0.0

🤖 Generated with [Aitachi Development](https://github.com/aitachi/claude-code)
