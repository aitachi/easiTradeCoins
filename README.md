# EasiTradeCoins

> **专业级去中心化加密货币交易平台**  
> 集成现货交易、衍生品交易、DeFi生态和社交金融的企业级交易系统

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Solidity](https://img.shields.io/badge/Solidity-0.8.20-orange.svg)](https://soliditylang.org/)
[![GitHub](https://img.shields.io/github/stars/aitachi/easiTradeCoins?style=social)](https://github.com/aitachi/easiTradeCoins)

**作者**: Aitachi | **联系**: 44158892@qq.com | **仓库**: https://github.com/aitachi/easiTradeCoins

---

## 📋 目录

- [项目简介](#项目简介)
- [核心特性](#核心特性)
- [项目结构](#项目结构)
- [快速开始](#快速开始)
- [文档导航](#文档导航)
- [开发路线](#开发路线)

---

## 项目简介

EasiTradeCoins 是一个**企业级去中心化加密货币交易平台**，采用 Go + Solidity 构建，为用户提供安全、高效、专业的数字资产交易服务。

### 项目统计

```
📊 项目规模
├── 总代码行数: ~26,000+
├── Go后端文件: 35+
├── 智能合约: 15+
├── 测试脚本: 11
├── 数据库表: 34
└── API端点: 50+

📈 完成度
├── 交易功能: 100% (12/12)
├── DeFi生态: 25% (2/8)
├── 社交金融: 20% (2/10)
├── 风控系统: 100% (8/8)
└── 总体: 40.3% (29/72)
```

---

## 核心特性

### 🚀 多样化交易
- **现货交易**: 限价单、市价单、止损止盈、跟踪止损
- **高级订单**: OCO、冰山订单、TWAP时间加权
- **自动化策略**: 网格交易、DCA定投
- **衍生品**: 杠杆交易(1-10x)、期权交易(Call/Put)

### 💎 DeFi生态
- **DEX聚合器**: 多DEX价格聚合，最优路由
- **流动性挖矿**: 质押LP代币赚取奖励
- **跨链桥接**: 多链资产转移(开发中)
- **收益聚合**: 自动收益优化(开发中)

### 👥 社交金融
- **跟单交易**: 一键跟随专业交易员
- **交易社区**: 策略分享、市场分析
- **排行榜**: 收益胜率排名(开发中)
- **NFT徽章**: 成就系统(开发中)

### 🛡️ 企业级安全
- **完善风控**: 订单验证、行为监控、异常检测
- **安全审计**: 代码审计、合约审计、0 critical issues
- **多重验证**: JWT、2FA、KYC分级
- **资金安全**: 冷热钱包分离、多签钱包

### ⚡ 高性能
- **撮合引擎**: < 10ms 撮合延迟
- **WebSocket**: 实时行情推送
- **Redis缓存**: 热点数据缓存
- **水平扩展**: 微服务架构

---

## 项目结构

```
EasiTradeCoins/
├── go-backend/           # Go后端 (~15,000行)
│   ├── cmd/server/       # 服务入口
│   ├── internal/
│   │   ├── services/     # 业务服务 (25+ 文件)
│   │   ├── matching/     # 撮合引擎
│   │   ├── security/     # 安全模块
│   │   └── websocket/    # WebSocket
│   └── ...
├── contracts/            # 智能合约 (~3,000行)
│   ├── src/             # Solidity源码
│   ├── scripts/         # 部署脚本
│   └── test/            # 合约测试
├── tests/               # 测试脚本 (~750行)
│   ├── integration_test.sh
│   ├── performance_test.sh
│   └── security_audit.sh
├── 📄 核心文档
│   ├── README.md        # 项目主文档 (本文件)
│   ├── TESTING.md       # 测试文档 (1283行)
│   ├── FEATURES.md      # 功能文档 (1686行)
│   └── PROJECT.md       # 项目概述 (1113行)
└── ...
```

---

## 快速开始

### 环境要求

- **Go** 1.21+
- **Node.js** 18+
- **PostgreSQL** 14+
- **Redis** 7+
- **Docker** 20+ (可选)

### Docker Compose 部署 (推荐)

```bash
# 1. 克隆项目
git clone https://github.com/aitachi/easiTradeCoins.git
cd easiTradeCoins

# 2. 启动服务
docker-compose up -d

# 3. 访问服务
curl http://localhost:8080/health  # API健康检查
open http://localhost:8080/swagger/index.html  # API文档
```

### 本机开发

```bash
# 1. 安装依赖服务
sudo apt-get install postgresql redis-server

# 2. 初始化数据库
psql -U postgres < deployment/init_postgres.sql

# 3. 启动后端
cd go-backend
go run cmd/server/main.go

# 4. 运行测试
bash ./run_all_tests.sh
```

更多部署方式请查看 [PROJECT.md - 开发指南](PROJECT.md#开发指南)

---

## 文档导航

### 📚 核心文档 (4082行)

| 文档 | 行数 | 说明 |
|------|------|------|
| **[TESTING.md](TESTING.md)** | 1283 | **测试文档** - 全面的测试指南、测试基础设施、测试执行方法 |
| **[FEATURES.md](FEATURES.md)** | 1686 | **功能文档** - 详细的后端功能说明、智能合约功能列表 |
| **[PROJECT.md](PROJECT.md)** | 1113 | **项目概述** - 项目架构、数据库设计、API文档、部署指南 |

### 📖 文档内容速查

#### TESTING.md - 测试文档
- ✅ 测试概述和策略
- ✅ 单元测试 (Go + Solidity)
- ✅ 集成测试 (数据库/API/消息队列)
- ✅ 性能测试 (响应时间/吞吐量/负载)
- ✅ 安全审计 (代码安全/合约审计/依赖扫描)
- ✅ Sepolia测试网部署
- ✅ 测试执行指南和故障排除

#### FEATURES.md - 功能文档
- ✅ 后端功能列表 (交易/衍生品/社交/风控)
- ✅ 智能合约功能 (DEX聚合器/流动性挖矿)
- ✅ 功能完成度统计
- ✅ 技术架构图
- ✅ 详细业务流程说明

#### PROJECT.md - 项目概述
- ✅ 项目文件结构详解
- ✅ 技术栈说明
- ✅ 系统架构图
- ✅ 数据库设计 (ER图/表结构)
- ✅ API文档和示例
- ✅ 部署架构和开发指南

### 📊 测试报告

```
test-reports/
├── go-unit-tests.log          # Go单元测试
├── contract-tests.log         # 合约测试
├── integration-tests.log      # 集成测试
├── performance-tests.log      # 性能测试
└── security-audit.log         # 安全审计
```

### 🔗 API文档

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:8081/metrics

---

## 开发路线

### Q1 2025 ✅ 已完成
- ✅ 基础交易功能 (限价/市价/止损)
- ✅ 高级订单 (OCO/冰山/TWAP)
- ✅ 风控系统 (完整的8大模块)
- ✅ 撮合引擎 (高频撮合)
- ✅ DEX聚合器
- ✅ 流动性挖矿

### Q2 2025 🔄 进行中
- 🔄 DeFi生态完善 (跨链桥/收益聚合器)
- 🔄 社交功能增强 (排行榜/NFT徽章)
- ⏳ 移动App开发 (React Native)
- ⏳ 数据分析平台

### Q3-Q4 2025 📅 计划中
- 📅 DAO治理
- 📅 NFT市场
- 📅 机器学习策略
- 📅 主网部署

---

## 技术栈

| 类别 | 技术 | 版本 |
|------|------|------|
| **后端** | Go + Gin + GORM | 1.21+ |
| **合约** | Solidity + Hardhat | 0.8.20 |
| **数据库** | PostgreSQL + MySQL + Redis | 14+ / 8+ / 7+ |
| **消息队列** | Kafka | 3+ |
| **搜索** | Elasticsearch | 8+ |
| **DevOps** | Docker + Nginx + Prometheus | 20+ |

---

## 贡献

我们欢迎各种形式的贡献！

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

**贡献要求**:
- ✅ 通过所有测试
- ✅ 代码覆盖率 > 80%
- ✅ 遵循代码规范
- ✅ 添加相应文档

---

## 许可证

本项目采用 [MIT License](LICENSE) 开源协议。

---

## 联系方式

- **GitHub**: https://github.com/aitachi/easiTradeCoins
- **Issues**: https://github.com/aitachi/easiTradeCoins/issues
- **Email**: 44158892@qq.com

---

## 致谢

感谢以下开源项目:
- [Go](https://golang.org/) - 后端语言
- [Gin](https://gin-gonic.com/) - Web框架
- [GORM](https://gorm.io/) - ORM框架
- [Solidity](https://soliditylang.org/) - 合约语言
- [Hardhat](https://hardhat.org/) - 合约框架
- [OpenZeppelin](https://openzeppelin.com/) - 安全合约库

---

<p align="center">
  <b>EasiTradeCoins - 专业级加密货币交易平台</b><br>
  <i>企业级 · 安全 · 高性能 · 可扩展</i><br><br>
  Made with ❤️ by <a href="https://github.com/aitachi">Aitachi</a>
</p>

<p align="center">
  <a href="#项目简介">项目简介</a> •
  <a href="#核心特性">核心特性</a> •
  <a href="#项目结构">项目结构</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#文档导航">文档导航</a> •
  <a href="#开发路线">开发路线</a>
</p>

---

**文档版本**: 1.0 | **最后更新**: 2025-11-02 | **维护者**: Aitachi
