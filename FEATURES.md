# EasiTradeCoins - 功能文档

**项目**: EasiTradeCoins - 专业级加密货币交易平台
**作者**: Aitachi
**联系**: 44158892@qq.com
**日期**: 2025-11-02
**版本**: 1.0

---

## 📋 目录

1. [功能概述](#功能概述)
2. [后端功能列表](#后端功能列表)
3. [智能合约功能列表](#智能合约功能列表)
4. [功能完成度](#功能完成度)
5. [技术架构](#技术架构)

---

## 功能概述

EasiTradeCoins是一个企业级去中心化加密货币交易平台，集成了现货交易、衍生品交易、DeFi生态和社交金融功能。

### 核心特性

- ✅ **多样化交易**: 现货、杠杆、期权、网格、DCA等多种交易方式
- ✅ **DeFi集成**: DEX聚合器、流动性挖矿、跨链桥接
- ✅ **社交金融**: 跟单交易、交易社区、排行榜系统
- ✅ **企业级安全**: 完善的风控系统、安全审计、多重验证
- ✅ **高性能**: 高频撮合引擎、实时WebSocket推送
- ✅ **可扩展**: 微服务架构、容器化部署、水平扩展

### 项目完成度

| 模块 | 完成度 | 功能数 | 状态 |
|------|--------|--------|------|
| 交易功能 | 100% | 12/12 | ✅ 完成 |
| DeFi生态 | 25% | 2/8 | 🔄 进行中 |
| 社交金融 | 20% | 2/10 | 🔄 进行中 |
| 风控系统 | 100% | 8/8 | ✅ 完成 |
| 基础设施 | 100% | 6/6 | ✅ 完成 |
| **总计** | **40.3%** | **29/72** | 🔄 **持续开发** |

---

## 后端功能列表

### 1. 交易功能 (100% 完成)

#### 1.1 基础交易

##### 限价单 (Limit Order)
**文件**: `order_service.go`
**状态**: ✅ 完成
**代码行数**: ~200

**功能描述**:
限价单是指定价格买入或卖出数字资产的订单类型。订单将在指定价格或更优价格执行。

**核心特性**:
- 价格精确控制
- 部分成交支持
- 订单簿实时更新
- 手续费自动计算

**API端点**:
```
POST /api/v1/orders
Body: {
  "symbol": "BTC_USDT",
  "side": "buy",
  "type": "limit",
  "price": "50000.00",
  "quantity": "0.1"
}
```

**业务流程**:
```
1. 用户提交限价单
2. 风控验证（余额、限额、频率）
3. 订单入库
4. 发送到撮合引擎
5. 订单簿更新
6. WebSocket推送订单状态
```

##### 市价单 (Market Order)
**文件**: `order_service.go`
**状态**: ✅ 完成
**代码行数**: ~150

**功能描述**:
市价单以当前市场最优价格立即执行买入或卖出操作。

**核心特性**:
- 立即执行
- 最优价格匹配
- 滑点保护
- 深度检查

**API端点**:
```
POST /api/v1/orders
Body: {
  "symbol": "BTC_USDT",
  "side": "buy",
  "type": "market",
  "quantity": "0.1"
}
```

**业务流程**:
```
1. 用户提交市价单
2. 检查订单簿深度
3. 计算滑点
4. 立即撮合
5. 成交通知
6. WebSocket推送成交信息
```

##### 止损止盈 (Stop-Loss/Take-Profit)
**文件**: `stop_order_service.go`
**状态**: ✅ 完成
**代码行数**: ~300

**功能描述**:
当市场价格达到设定的触发价格时，自动执行买入或卖出操作，用于风险管理和利润保护。

**核心特性**:
- 价格监控
- 自动触发
- 止损保护
- 止盈锁定

**API端点**:
```
POST /api/v1/orders/stop
Body: {
  "symbol": "BTC_USDT",
  "side": "sell",
  "stop_price": "48000.00",
  "quantity": "0.1",
  "type": "stop_loss"
}
```

**业务流程**:
```
1. 用户设置止损/止盈价格
2. 订单进入监控队列
3. 实时监控市场价格
4. 价格触发时转为市价单
5. 执行成交
6. 通知用户
```

##### 跟踪止损 (Trailing Stop)
**文件**: `trailing_stop_service.go`
**状态**: ✅ 完成
**代码行数**: ~280

**功能描述**:
根据市场价格动态调整止损价格，在价格上涨时保护利润，同时允许继续获利。

**核心特性**:
- 动态止损价格
- 利润保护
- 回撤控制
- 百分比或固定金额跟踪

**API端点**:
```
POST /api/v1/orders/trailing-stop
Body: {
  "symbol": "BTC_USDT",
  "side": "sell",
  "trailing_delta": "1000.00",
  "quantity": "0.1"
}
```

**业务流程**:
```
1. 用户设置跟踪距离
2. 监控最高价/最低价
3. 动态调整止损价
4. 价格回撤触发时执行
5. 成交通知
```

##### 条件单 (Conditional Order)
**文件**: `conditional_order_service.go`
**状态**: ✅ 完成
**代码行数**: ~250

**功能描述**:
基于多个条件触发的订单，支持复杂的交易策略。

**核心特性**:
- 多条件组合
- 逻辑运算符支持
- 技术指标触发
- 时间条件

**API端点**:
```
POST /api/v1/orders/conditional
Body: {
  "symbol": "BTC_USDT",
  "conditions": [
    {"type": "price", "operator": ">", "value": "51000"},
    {"type": "volume", "operator": ">", "value": "1000000"}
  ],
  "action": {
    "side": "buy",
    "quantity": "0.1"
  }
}
```

#### 1.2 高级订单

##### OCO订单 (One-Cancels-Other)
**文件**: `oco_order_service.go`
**状态**: ✅ 完成
**代码行数**: ~350

**功能描述**:
同时创建两个订单，当其中一个订单成交时，另一个订单自动取消。常用于同时设置止盈和止损。

**核心特性**:
- 双订单关联
- 自动取消机制
- 风险收益平衡
- 止盈止损组合

**数据模型**:
```go
type OCOOrder struct {
    ID              uint
    UserID          uint
    Symbol          string
    Side            string
    StopPrice       decimal.Decimal  // 止损价
    StopLimitPrice  decimal.Decimal  // 止损限价
    LimitPrice      decimal.Decimal  // 止盈价
    Quantity        decimal.Decimal
    Status          string
    StopOrderID     *uint            // 止损订单ID
    LimitOrderID    *uint            // 止盈订单ID
}
```

**API端点**:
```
POST /api/v1/orders/oco
Body: {
  "symbol": "BTC_USDT",
  "side": "sell",
  "quantity": "0.1",
  "stop_price": "48000.00",
  "stop_limit_price": "47800.00",
  "limit_price": "52000.00"
}
```

**业务流程**:
```
1. 创建OCO订单主记录
2. 创建止损订单（stop_price）
3. 创建止盈订单（limit_price）
4. 关联三个订单
5. 监控任一订单状态
6. 一个成交时取消另一个
7. 更新OCO订单状态
```

##### 冰山订单 (Iceberg Order)
**文件**: `iceberg_order_service.go`
**状态**: ✅ 完成
**代码行数**: ~320

**功能描述**:
将大额订单分割成多个小订单，只显示部分数量在订单簿中，避免对市场造成冲击。

**核心特性**:
- 大单隐藏
- 分批执行
- 市场影响最小化
- 动态数量显示

**数据模型**:
```go
type IcebergOrder struct {
    ID              uint
    UserID          uint
    Symbol          string
    Side            string
    Price           decimal.Decimal
    TotalQuantity   decimal.Decimal  // 总数量
    VisibleQuantity decimal.Decimal  // 可见数量
    ExecutedQuantity decimal.Decimal // 已执行数量
    Status          string
    ChildOrders     []Order          // 子订单列表
}
```

**API端点**:
```
POST /api/v1/orders/iceberg
Body: {
  "symbol": "BTC_USDT",
  "side": "buy",
  "price": "50000.00",
  "total_quantity": "10.0",
  "visible_quantity": "0.5"
}
```

**业务流程**:
```
1. 接收大额订单
2. 分割为多个子订单
3. 首个子订单进入订单簿
4. 子订单成交后
5. 下一个子订单自动进入
6. 循环直到全部成交
7. 隐藏总量信息
```

**优势**:
- 防止大单被跟单
- 避免价格滑点
- 保护交易策略
- 减少市场冲击

##### TWAP订单 (Time-Weighted Average Price)
**文件**: `twap_order_service.go`
**状态**: ✅ 完成
**代码行数**: ~280

**功能描述**:
在指定时间段内，将大额订单均匀分割并按时间间隔执行，以获得时间加权平均价格。

**核心特性**:
- 时间分散执行
- 价格平滑
- 市场冲击最小
- 自动执行

**数据模型**:
```go
type TWAPOrder struct {
    ID              uint
    UserID          uint
    Symbol          string
    Side            string
    TotalQuantity   decimal.Decimal
    Duration        int              // 执行时长(分钟)
    Interval        int              // 间隔时间(秒)
    StartTime       time.Time
    EndTime         time.Time
    ExecutedQuantity decimal.Decimal
    Status          string
    Executions      []TWAPExecution  // 执行记录
}
```

**API端点**:
```
POST /api/v1/orders/twap
Body: {
  "symbol": "BTC_USDT",
  "side": "buy",
  "total_quantity": "5.0",
  "duration": 60,        // 60分钟
  "interval": 300        // 每5分钟执行一次
}
```

**业务流程**:
```
1. 接收TWAP订单
2. 计算执行次数和单次数量
3. 设置定时任务
4. 按间隔执行市价单
5. 记录每次执行价格
6. 计算平均成交价
7. 完成后返回TWAP价格
```

**计算示例**:
```
总量: 5 BTC
时长: 60分钟
间隔: 5分钟
执行次数: 12次
单次数量: 0.4167 BTC

执行记录:
时间    数量      价格      金额
00:00  0.4167   50000   20835
00:05  0.4167   50100   20876.67
...
00:55  0.4167   50500   21043.35

TWAP价格 = 总金额 / 总数量
```

#### 1.3 自动化策略

##### 网格交易 (Grid Trading)
**文件**: `grid_trading_service.go`
**状态**: ✅ 完成
**代码行数**: 518

**功能描述**:
在指定价格区间内设置多个买卖网格，自动高抛低吸，适合震荡市场。

**核心特性**:
- 自动化交易
- 高抛低吸
- 震荡市获利
- 风险可控

**数据模型**:
```go
type GridStrategy struct {
    ID              uint
    UserID          uint
    Symbol          string
    LowerPrice      decimal.Decimal  // 下限价格
    UpperPrice      decimal.Decimal  // 上限价格
    GridNumber      int              // 网格数量
    InvestmentAmount decimal.Decimal // 投资金额
    Status          string
    TotalProfit     decimal.Decimal  // 总收益
    GridOrders      []GridOrder      // 网格订单
}

type GridOrder struct {
    ID          uint
    StrategyID  uint
    Level       int              // 网格层级
    BuyPrice    decimal.Decimal  // 买入价
    SellPrice   decimal.Decimal  // 卖出价
    Quantity    decimal.Decimal
    Status      string
    Profit      decimal.Decimal
}
```

**API端点**:
```
POST /api/v1/grid-trading
Body: {
  "symbol": "BTC_USDT",
  "lower_price": "45000",
  "upper_price": "55000",
  "grid_number": 10,
  "investment_amount": "10000"
}
```

**策略逻辑**:
```
价格区间: 45000 - 55000
网格数: 10
网格间距: (55000-45000)/10 = 1000

网格层级:
Level 1: Buy 45000, Sell 46000
Level 2: Buy 46000, Sell 47000
Level 3: Buy 47000, Sell 48000
...
Level 10: Buy 54000, Sell 55000

执行规则:
1. 价格下跌到买入价 -> 买入
2. 价格上涨到卖出价 -> 卖出
3. 每次交易赚取网格利润
4. 循环往复
```

**收益计算**:
```
单次网格利润 = 卖出价 - 买入价
            = 1000 USDT (每格)

假设来回交易5次:
总收益 = 5 × 1000 = 5000 USDT
收益率 = 5000 / 10000 = 50%
```

##### DCA定投 (Dollar Cost Averaging)
**文件**: `dca_service.go`
**状态**: ✅ 完成
**代码行数**: 452

**功能描述**:
按固定时间间隔、固定金额自动买入数字资产，平均成本，降低择时风险。

**核心特性**:
- 定时定额
- 成本平均
- 风险分散
- 长期投资

**数据模型**:
```go
type DCAStrategy struct {
    ID              uint
    UserID          uint
    Symbol          string
    Amount          decimal.Decimal  // 每次投资金额
    Frequency       string           // 频率: daily, weekly, monthly
    StartDate       time.Time
    EndDate         *time.Time       // 可选
    MaxInvestments  *int             // 最大次数
    Status          string
    TotalInvested   decimal.Decimal  // 总投入
    TotalQuantity   decimal.Decimal  // 总持仓
    AverageCost     decimal.Decimal  // 平均成本
    Purchases       []DCAPurchase    // 购买记录
}

type DCAPurchase struct {
    ID          uint
    StrategyID  uint
    ExecuteTime time.Time
    Price       decimal.Decimal
    Quantity    decimal.Decimal
    Amount      decimal.Decimal
    Status      string
}
```

**API端点**:
```
POST /api/v1/dca
Body: {
  "symbol": "BTC_USDT",
  "amount": "1000",
  "frequency": "weekly",
  "start_date": "2025-01-01",
  "max_investments": 52    // 投资一年
}
```

**执行示例**:
```
定投计划:
币种: BTC
金额: 1000 USDT/周
频率: 每周一
期限: 52周

执行记录:
周次  日期      价格    数量      总投入  平均成本
1    01-01   50000   0.0200   1000    50000
2    01-08   48000   0.0208   2000    49000
3    01-15   52000   0.0192   3000    50000
...
52   12-30   60000   0.0167   52000   52000

总投入: 52,000 USDT
总持仓: 1.04 BTC
平均成本: 50,000 USDT/BTC
当前价值: 62,400 USDT
收益: 10,400 USDT (20%)
```

**优势**:
- 避免择时压力
- 平滑价格波动
- 纪律性投资
- 长期复利效应

#### 1.4 衍生品交易

##### 杠杆交易 (Margin Trading)
**文件**: `margin_trading_service.go`
**状态**: ✅ 完成
**代码行数**: 600

**功能描述**:
使用借贷资金放大交易规模，提高资金利用效率，支持1-10倍杠杆。

**核心特性**:
- 多倍杠杆(1-10x)
- 借贷机制
- 强平保护
- 利息计算
- 风险控制

**数据模型**:
```go
type MarginAccount struct {
    ID              uint
    UserID          uint
    TotalAssets     decimal.Decimal  // 总资产
    TotalLiabilities decimal.Decimal  // 总负债
    NetAssets       decimal.Decimal  // 净资产
    MarginLevel     decimal.Decimal  // 保证金水平
    Status          string
}

type MarginPosition struct {
    ID              uint
    UserID          uint
    AccountID       uint
    Symbol          string
    Side            string           // long/short
    EntryPrice      decimal.Decimal
    CurrentPrice    decimal.Decimal
    Quantity        decimal.Decimal
    Leverage        int
    Margin          decimal.Decimal  // 保证金
    BorrowedAmount  decimal.Decimal  // 借入金额
    UnrealizedPnL   decimal.Decimal  // 未实现盈亏
    LiquidationPrice decimal.Decimal // 强平价格
    Status          string
}
```

**API端点**:
```
# 开仓
POST /api/v1/margin/open-position
Body: {
  "symbol": "BTC_USDT",
  "side": "long",
  "entry_price": "50000",
  "quantity": "1.0",
  "leverage": 5
}

# 平仓
POST /api/v1/margin/close-position
Body: {
  "position_id": 123,
  "quantity": "1.0"
}
```

**杠杆计算**:
```
假设:
本金: 10,000 USDT
杠杆: 5x
交易规模: 50,000 USDT
开仓价: 50,000 USDT/BTC
数量: 1 BTC

保证金 = 交易规模 / 杠杆 = 10,000 USDT
借入金额 = 交易规模 - 保证金 = 40,000 USDT

盈亏计算:
价格涨到 55,000:
收益 = (55,000 - 50,000) × 1 = 5,000 USDT
收益率 = 5,000 / 10,000 = 50%

价格跌到 45,000:
亏损 = (50,000 - 45,000) × 1 = 5,000 USDT
亏损率 = 5,000 / 10,000 = 50%
```

**强平机制**:
```
强平价格计算 (做多):
强平价 = 开仓价 × (1 - 1/杠杆 × 0.8)
      = 50,000 × (1 - 1/5 × 0.8)
      = 50,000 × 0.84
      = 42,000 USDT

保证金水平 = 净资产 / 持仓价值
当保证金水平 < 120%时，发出预警
当保证金水平 < 110%时，强制平仓
```

**利息计算**:
```
借贷年利率: 10%
日利率: 10% / 365 = 0.0274%
借入金额: 40,000 USDT
持仓1天利息: 40,000 × 0.0274% = 10.96 USDT
```

##### 期权交易 (Options Trading)
**文件**: `options_trading_service.go`
**状态**: ✅ 完成
**代码行数**: 430

**功能描述**:
提供看涨期权(Call)和看跌期权(Put)交易，支持期权定价、买卖和行权。

**核心特性**:
- 看涨/看跌期权
- Black-Scholes定价
- 美式/欧式期权
- 自动行权
- 风险对冲

**数据模型**:
```go
type OptionContract struct {
    ID              uint
    Symbol          string           // 标的资产
    Type            string           // call/put
    StrikePrice     decimal.Decimal  // 行权价
    Premium         decimal.Decimal  // 期权费
    ExpiryDate      time.Time        // 到期日
    Style           string           // american/european
    Status          string
    UnderlyingPrice decimal.Decimal  // 标的价格
}

type OptionPosition struct {
    ID          uint
    UserID      uint
    ContractID  uint
    Type        string           // buy/sell
    Quantity    decimal.Decimal
    EntryPrice  decimal.Decimal  // 买入价
    CurrentPrice decimal.Decimal
    PnL         decimal.Decimal
    Status      string
}
```

**API端点**:
```
# 创建期权
POST /api/v1/options/create
Body: {
  "symbol": "BTC_USDT",
  "type": "call",
  "strike_price": "55000",
  "expiry_date": "2025-12-31",
  "style": "european"
}

# 购买期权
POST /api/v1/options/buy
Body: {
  "contract_id": 123,
  "quantity": "1.0"
}

# 行权
POST /api/v1/options/exercise
Body: {
  "position_id": 456
}
```

**Black-Scholes定价**:
```
期权定价公式:
C = S×N(d1) - K×e^(-r×T)×N(d2)  (看涨期权)
P = K×e^(-r×T)×N(-d2) - S×N(-d1) (看跌期权)

其中:
S = 标的资产当前价格
K = 行权价
r = 无风险利率
T = 到期时间(年)
σ = 波动率
N() = 标准正态分布累积函数

d1 = [ln(S/K) + (r + σ²/2)×T] / (σ×√T)
d2 = d1 - σ×√T
```

**交易示例**:
```
看涨期权 (Call Option):
标的: BTC
当前价: 50,000 USDT
行权价: 55,000 USDT
到期日: 30天后
期权费: 500 USDT

情景1 - 价格上涨:
到期价格: 60,000 USDT
行权收益: 60,000 - 55,000 = 5,000 USDT
净收益: 5,000 - 500 = 4,500 USDT

情景2 - 价格下跌:
到期价格: 48,000 USDT
不行权,损失期权费: -500 USDT
```

### 2. DeFi生态 (25% 完成)

#### 2.1 已完成功能

##### DEX聚合器 (DEX Aggregator)
**文件**: `contracts/src/DEXAggregator.sol`
**状态**: ✅ 完成
**代码行数**: 280

**功能描述**:
聚合多个去中心化交易所(DEX)的流动性，为用户提供最优交易价格和路径。

**核心特性**:
- 多DEX集成
- 价格聚合
- 最优路由
- Gas优化
- 平台费收取

**智能合约结构**:
```solidity
contract DEXAggregator is Ownable, ReentrancyGuard {
    // DEX信息
    struct DEXInfo {
        address dexAddress;
        string name;
        bool isActive;
    }

    mapping(address => DEXInfo) public registeredDEXes;
    address[] public dexList;
    uint256 public platformFeeRate;  // 平台费率

    // 注册DEX
    function registerDEX(address dex, string memory name) external onlyOwner;

    // 获取最优价格
    function getBestPrice(
        address tokenIn,
        address tokenOut,
        uint256 amountIn
    ) public view returns (uint256 bestPrice, address bestDEX);

    // 执行交换
    function swap(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 minAmountOut,
        address dex
    ) external nonReentrant returns (uint256 amountOut);
}
```

**使用示例**:
```javascript
// 查询最优价格
const [bestPrice, bestDEX] = await dexAggregator.getBestPrice(
    USDT.address,
    BTC.address,
    ethers.utils.parseUnits("1000", 6)  // 1000 USDT
);

// 执行交换
await dexAggregator.swap(
    USDT.address,
    BTC.address,
    ethers.utils.parseUnits("1000", 6),
    bestPrice.mul(99).div(100),  // 1% 滑点保护
    bestDEX
);
```

**支持的DEX**:
- Uniswap V2/V3
- SushiSwap
- PancakeSwap
- Balancer
- Curve

##### 流动性挖矿 (Liquidity Mining)
**文件**: `contracts/src/LiquidityMining.sol`
**状态**: ✅ 完成
**代码行数**: 280

**功能描述**:
用户质押LP代币赚取奖励，支持多池、权重系统和灵活的奖励分配。

**核心特性**:
- 多池管理
- 质押奖励
- 权重系统
- 灵活提取
- 紧急提取

**智能合约结构**:
```solidity
contract LiquidityMining is Ownable, ReentrancyGuard, Pausable {
    // 池子信息
    struct PoolInfo {
        IERC20 lpToken;              // LP代币
        uint256 allocPoint;          // 分配权重
        uint256 lastRewardBlock;     // 上次奖励区块
        uint256 accRewardPerShare;   // 累计每股奖励
        uint256 totalStaked;         // 总质押量
    }

    // 用户信息
    struct UserInfo {
        uint256 amount;              // 质押数量
        uint256 rewardDebt;          // 奖励债务
        uint256 pendingReward;       // 待领取奖励
    }

    PoolInfo[] public poolInfo;
    mapping(uint256 => mapping(address => UserInfo)) public userInfo;

    IERC20 public rewardToken;       // 奖励代币
    uint256 public rewardPerBlock;   // 每块奖励

    // 质押
    function stake(uint256 pid, uint256 amount) external nonReentrant;

    // 解除质押
    function unstake(uint256 pid, uint256 amount) external nonReentrant;

    // 领取奖励
    function claimReward(uint256 pid) external nonReentrant;

    // 查询待领取奖励
    function pendingReward(uint256 pid, address user)
        external view returns (uint256);
}
```

**收益计算**:
```
假设:
池子总权重: 1000
Pool A权重: 400 (40%)
Pool B权重: 600 (60%)
每块奖励: 10代币

Pool A每块奖励 = 10 × 400/1000 = 4代币

用户质押:
Pool A总质押: 10,000 LP
用户质押: 1,000 LP (10%)
质押100个区块

用户奖励 = 4 × 100 × 1,000/10,000 = 40代币
```

**使用示例**:
```javascript
// 质押LP代币
await lpToken.approve(liquidityMining.address, stakeAmount);
await liquidityMining.stake(0, stakeAmount);

// 查询待领取奖励
const pending = await liquidityMining.pendingReward(0, userAddress);

// 领取奖励
await liquidityMining.claimReward(0);

// 解除质押
await liquidityMining.unstake(0, unstakeAmount);
```

#### 2.2 开发中功能

##### 跨链桥 (Cross-Chain Bridge)
**状态**: ⏳ 开发中
**完成度**: 0%

**计划功能**:
- 多链资产转移
- 原子交换
- 流动性池
- 安全验证
- 手续费优化

##### 收益聚合器 (Yield Aggregator)
**状态**: ⏳ 开发中
**完成度**: 0%

**计划功能**:
- 自动收益优化
- 多协议聚合
- 复投策略
- Gas优化
- 风险评级

##### 质押服务 (Staking Service)
**状态**: ⏳ 计划中
**完成度**: 0%

**计划功能**:
- 代币质押
- 流动性质押
- 锁仓奖励
- 提前解锁
- 治理权益

### 3. 社交金融 (20% 完成)

#### 3.1 已完成功能

##### 跟单交易 (Copy Trading)
**文件**: `copy_trading_service.go`
**状态**: ✅ 完成
**代码行数**: 550

**功能描述**:
普通用户可以跟随专业交易员的交易策略，自动复制其交易操作。

**核心特性**:
- 交易员认证
- 跟单设置
- 自动复制
- 收益分成
- 风险控制

**数据模型**:
```go
type Trader struct {
    ID              uint
    UserID          uint
    DisplayName     string
    Description     string
    WinRate         decimal.Decimal  // 胜率
    TotalProfit     decimal.Decimal  // 总收益
    FollowerCount   int              // 跟随者数量
    MaxFollowers    int              // 最大跟随数
    CommissionRate  decimal.Decimal  // 佣金比例
    Status          string
    PerformanceStats TraderStats     // 业绩统计
}

type FollowRelation struct {
    ID              uint
    FollowerID      uint             // 跟随者ID
    TraderID        uint             // 交易员ID
    CopyRatio       decimal.Decimal  // 跟单比例 (0.1-1.0)
    MaxCopyAmount   decimal.Decimal  // 单次最大跟单金额
    StopLossRatio   decimal.Decimal  // 止损比例
    Status          string
    TotalCopied     int              // 跟单次数
    TotalProfit     decimal.Decimal  // 总收益
}
```

**API端点**:
```
# 注册交易员
POST /api/v1/copy-trading/register-trader
Body: {
  "display_name": "CryptoMaster",
  "description": "10年交易经验",
  "max_followers": 100,
  "commission_rate": "0.2"
}

# 跟随交易员
POST /api/v1/copy-trading/follow
Body: {
  "trader_id": 123,
  "copy_ratio": "0.5",
  "max_copy_amount": "1000",
  "stop_loss_ratio": "0.1"
}

# 查询交易员列表
GET /api/v1/copy-trading/traders
Query: ?sort=profit&order=desc&page=1&limit=20
```

**跟单逻辑**:
```
交易员下单:
币种: BTC_USDT
方向: 买入
价格: 50,000 USDT
数量: 1 BTC
金额: 50,000 USDT

跟随者A (跟单比例50%):
数量: 0.5 BTC
金额: 25,000 USDT

跟随者B (跟单比例30%, 最大金额10,000):
理论金额: 50,000 × 0.3 = 15,000 USDT
实际金额: 10,000 USDT (受限于最大金额)
数量: 0.2 BTC

跟随者C (止损10%):
如果亏损达到 10% → 自动停止跟单
```

**收益分成**:
```
交易员佣金: 20%

跟随者盈利 1000 USDT:
佣金 = 1000 × 20% = 200 USDT
跟随者净收益 = 800 USDT
交易员收入 = 200 USDT

跟随者亏损:
不收取佣金
```

##### 交易社区 (Trading Community)
**文件**: `community_service.go`
**状态**: ✅ 完成
**代码行数**: 400

**功能描述**:
用户可以创建或加入交易社区，分享交易策略、市场分析和投资心得。

**核心特性**:
- 社区创建
- 帖子发布
- 评论互动
- 点赞收藏
- 内容审核

**数据模型**:
```go
type TradingCommunity struct {
    ID          uint
    Name        string
    Description string
    CreatorID   uint
    MemberCount int
    PostCount   int
    Category    string           // 分类: 策略/分析/新手
    IsPublic    bool
    Status      string
}

type Post struct {
    ID          uint
    CommunityID uint
    AuthorID    uint
    Title       string
    Content     string           // 支持Markdown
    Type        string           // 策略/分析/讨论
    ViewCount   int
    LikeCount   int
    CommentCount int
    IsTop       bool             // 是否置顶
    Tags        []string
    CreatedAt   time.Time
}

type Comment struct {
    ID          uint
    PostID      uint
    UserID      uint
    Content     string
    LikeCount   int
    ParentID    *uint            // 回复评论
    CreatedAt   time.Time
}
```

**API端点**:
```
# 创建社区
POST /api/v1/community/create
Body: {
  "name": "BTC策略交流",
  "description": "专注比特币交易策略",
  "category": "策略",
  "is_public": true
}

# 发布帖子
POST /api/v1/community/:id/posts
Body: {
  "title": "BTC突破关键阻力位分析",
  "content": "## 技术分析\n...",
  "type": "分析",
  "tags": ["BTC", "技术分析", "突破"]
}

# 评论帖子
POST /api/v1/community/posts/:id/comment
Body: {
  "content": "分析很到位！",
  "parent_id": null
}

# 点赞帖子
POST /api/v1/community/posts/:id/like
```

**社区功能**:
```
1. 内容分类:
   - 交易策略
   - 市场分析
   - 新手教程
   - 行情讨论
   - 工具分享

2. 互动功能:
   - 发帖/回复
   - 点赞/收藏
   - 关注/订阅
   - 分享/转发

3. 内容管理:
   - 置顶/加精
   - 审核/删除
   - 举报/封禁
   - 标签/搜索

4. 积分系统:
   - 发帖奖励
   - 精华奖励
   - 等级升级
   - 权限解锁
```

#### 3.2 开发中功能

##### 排行榜系统 (Leaderboard)
**状态**: ⏳ 开发中
**完成度**: 0%

**计划功能**:
- 收益排行
- 胜率排行
- 交易量排行
- 实时更新
- 历史记录

##### 社交挖矿 (Social Mining)
**状态**: ⏳ 计划中
**完成度**: 0%

**计划功能**:
- 内容激励
- 互动奖励
- 推荐奖励
- 等级系统

##### NFT徽章 (NFT Badges)
**状态**: ⏳ 计划中
**完成度**: 0%

**计划功能**:
- 成就徽章
- 等级勋章
- 交易记录NFT
- 可交易/展示

### 4. 风控系统 (100% 完成)

#### 4.1 订单验证
**文件**: `risk_manager.go`
**功能**: ✅ 完成

**验证项**:
- 余额充足性
- 订单合法性
- 价格合理性
- 数量限制
- 重复订单检测

**代码示例**:
```go
func (rm *RiskManager) ValidateOrder(order *Order) error {
    // 余额检查
    if !rm.checkBalance(order.UserID, order.TotalAmount) {
        return errors.New("insufficient balance")
    }

    // 价格偏离检查
    marketPrice := rm.getMarketPrice(order.Symbol)
    deviation := order.Price.Sub(marketPrice).Div(marketPrice).Abs()
    if deviation.GreaterThan(decimal.NewFromFloat(0.1)) {  // 10%
        return errors.New("price deviation too large")
    }

    // 单笔限额检查
    if order.TotalAmount.GreaterThan(rm.config.MaxOrderAmount) {
        return errors.New("order amount exceeds limit")
    }

    return nil
}
```

#### 4.2 行为监控
**文件**: `risk_manager.go`
**功能**: ✅ 完成

**监控项**:
- 频繁交易
- 异常下单
- 对敲交易
- 账户异常

**检测逻辑**:
```go
// 频繁交易检测
func (rm *RiskManager) DetectFrequentTrading(userID uint) bool {
    count := rm.getRecentOrderCount(userID, 1*time.Minute)
    return count > rm.config.MaxOrdersPerMinute
}

// 对敲交易检测
func (rm *RiskManager) DetectSelfTrading(buyerID, sellerID uint) bool {
    // 同一用户
    if buyerID == sellerID {
        return true
    }

    // 关联账户
    if rm.isRelatedAccount(buyerID, sellerID) {
        return true
    }

    return false
}
```

#### 4.3 提现风控
**文件**: `risk_manager.go`
**功能**: ✅ 完成

**风控措施**:
- 提现限额
- 白名单地址
- 人工审核
- 延迟提现
- 多重签名

#### 4.4 速率限制
**文件**: `rate_limiter.go`
**功能**: ✅ 完成
**代码行数**: 150

**限制类型**:
- IP限制
- 用户限制
- API限制
- WebSocket限制

**实现**:
```go
type RateLimiter struct {
    redis *redis.Client
}

func (rl *RateLimiter) Allow(key string, limit int, window time.Duration) bool {
    count, _ := rl.redis.Incr(context.Background(), key).Result()
    if count == 1 {
        rl.redis.Expire(context.Background(), key, window)
    }
    return count <= int64(limit)
}

// 使用
allowed := rateLimiter.Allow(
    fmt.Sprintf("api:%s", userID),
    100,                    // 限制
    1 * time.Minute,        // 时间窗口
)
```

#### 4.5 风险评分
**文件**: `risk_manager.go`
**功能**: ✅ 完成

**评分因素**:
```
1. 账户年龄 (10分)
2. KYC状态 (20分)
3. 交易历史 (20分)
4. 账户余额 (15分)
5. 违规记录 (35分)

总分: 100分
高风险: < 40分
中风险: 40-70分
低风险: > 70分
```

#### 4.6 异常检测
**文件**: `risk_manager.go`
**功能**: ✅ 完成

**检测算法**:
- 统计分析
- 机器学习
- 规则引擎
- 实时监控

#### 4.7 账户冻结
**文件**: `user_service.go`
**功能**: ✅ 完成

**冻结触发**:
- 风险评分过低
- 异常行为
- 用户举报
- 监管要求

#### 4.8 审计日志
**文件**: `audit_service.go`
**功能**: ✅ 完成
**代码行数**: 200

**日志内容**:
- 操作时间
- 操作用户
- 操作类型
- 操作详情
- IP地址
- 结果状态

### 5. 核心基础设施 (100% 完成)

#### 5.1 撮合引擎
**文件**: `matching/engine.go`
**状态**: ✅ 完成
**代码行数**: 400

**功能特性**:
- 价格优先
- 时间优先
- 高频撮合
- 实时处理

#### 5.2 订单簿
**文件**: `matching/orderbook.go`
**状态**: ✅ 完成
**代码行数**: 300

**数据结构**:
- 买单队列
- 卖单队列
- 价格索引
- 深度统计

#### 5.3 WebSocket
**文件**: `websocket/hub.go`
**状态**: ✅ 完成
**代码行数**: 450

**推送内容**:
- 实时行情
- 订单更新
- 成交通知
- 账户变动

#### 5.4 用户认证
**文件**: `auth_service.go`
**状态**: ✅ 完成
**代码行数**: 300

**认证方式**:
- JWT Token
- OAuth2.0
- 2FA验证
- 生物识别

#### 5.5 KYC验证
**文件**: `kyc_service.go`
**状态**: ✅ 完成
**代码行数**: 250

**验证级别**:
- Level 1: 邮箱验证
- Level 2: 身份证验证
- Level 3: 人脸识别

#### 5.6 钱包管理
**文件**: `wallet_service.go`
**状态**: ✅ 完成
**代码行数**: 350

**钱包功能**:
- 充值
- 提现
- 转账
- 余额查询

---

## 智能合约功能列表

### 1. DEX聚合器合约

**合约**: `DEXAggregator.sol`
**网络**: Ethereum, BSC
**审计**: ✅ 已完成

**主要函数**:
```solidity
// DEX管理
function registerDEX(address dex, string memory name) external onlyOwner
function removeDEX(address dex) external onlyOwner
function getRegisteredDEXes() external view returns (address[] memory)

// 价格查询
function getBestPrice(
    address tokenIn,
    address tokenOut,
    uint256 amountIn
) public view returns (uint256 bestPrice, address bestDEX)

// 交换执行
function swap(
    address tokenIn,
    address tokenOut,
    uint256 amountIn,
    uint256 minAmountOut,
    address dex
) external nonReentrant returns (uint256 amountOut)

// 费用管理
function setPlatformFee(uint256 newFee) external onlyOwner
function withdrawFees() external onlyOwner
```

### 2. 流动性挖矿合约

**合约**: `LiquidityMining.sol`
**网络**: Ethereum, BSC
**审计**: ✅ 已完成

**主要函数**:
```solidity
// Pool管理
function createPool(
    address lpToken,
    uint256 allocPoint,
    uint256 startBlock,
    uint256 endBlock,
    uint256 weight
) external onlyOwner

function updatePool(uint256 pid) public
function massUpdatePools() external

// 质押操作
function stake(uint256 pid, uint256 amount) external nonReentrant
function unstake(uint256 pid, uint256 amount) external nonReentrant
function emergencyWithdraw(uint256 pid) external

// 奖励查询
function pendingReward(uint256 pid, address user)
    external view returns (uint256)
function claimReward(uint256 pid) external nonReentrant
```

### 3. Mock ERC20合约

**合约**: `MockERC20.sol`
**用途**: 测试
**网络**: Sepolia, Testnet

**功能**:
```solidity
function mint(address to, uint256 amount) external
function burn(address from, uint256 amount) external
```

---

## 功能完成度

### 总体完成情况

**总功能数**: 72
**已完成**: 29
**完成度**: 40.3%

### 分模块完成度

| 模块 | 已完成 | 总数 | 完成度 | 优先级 |
|------|--------|------|--------|--------|
| 基础交易 | 5/5 | 100% | ✅ | P0 |
| 高级订单 | 3/3 | 100% | ✅ | P0 |
| 自动化策略 | 2/2 | 100% | ✅ | P1 |
| 衍生品交易 | 2/2 | 100% | ✅ | P1 |
| DeFi生态 | 2/8 | 25% | 🔄 | P1 |
| 社交金融 | 2/10 | 20% | 🔄 | P2 |
| 风控系统 | 8/8 | 100% | ✅ | P0 |
| 基础设施 | 6/6 | 100% | ✅ | P0 |
| 智能合约 | 3/8 | 37.5% | 🔄 | P1 |
| 移动应用 | 0/10 | 0% | ⏳ | P2 |
| 数据分析 | 0/10 | 0% | ⏳ | P3 |

### 开发路线图

#### Q1 2025 (已完成)
- ✅ 基础交易功能
- ✅ 高级订单类型
- ✅ 风控系统
- ✅ 撮合引擎

#### Q2 2025 (进行中)
- 🔄 DeFi生态完善
- 🔄 社交功能增强
- ⏳ 移动App开发
- ⏳ 数据分析平台

#### Q3 2025 (计划中)
- 📅 跨链桥接
- 📅 NFT市场
- 📅 DAO治理
- 📅 机器学习策略

#### Q4 2025 (计划中)
- 📅 主网部署
- 📅 全球化扩展
- 📅 合规认证
- 📅 生态整合

---

## 技术架构

### 后端架构

```
                    ┌─────────────────┐
                    │   Nginx/LB      │
                    └────────┬────────┘
                             │
                ┌────────────┼────────────┐
                │            │            │
         ┌──────▼─────┐ ┌───▼────┐ ┌────▼─────┐
         │  API服务1  │ │API服务2│ │API服务N  │
         └──────┬─────┘ └───┬────┘ └────┬─────┘
                │            │            │
                └────────────┼────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
      ┌───────▼────┐  ┌─────▼─────┐  ┌────▼──────┐
      │ PostgreSQL │  │   Redis    │  │   Kafka   │
      └────────────┘  └────────────┘  └───────────┘
```

### 智能合约架构

```
                 ┌─────────────────┐
                 │  DEXAggregator  │
                 └────────┬────────┘
                          │
         ┌────────────────┼────────────────┐
         │                │                │
   ┌─────▼─────┐   ┌─────▼─────┐   ┌─────▼─────┐
   │ Uniswap V3│   │ SushiSwap │   │  Balancer │
   └───────────┘   └───────────┘   └───────────┘


                ┌──────────────────┐
                │ LiquidityMining  │
                └────────┬─────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
   ┌─────▼─────┐   ┌────▼────┐   ┌─────▼─────┐
   │  Pool A   │   │ Pool B  │   │  Pool C   │
   └───────────┘   └─────────┘   └───────────┘
```

---

## 联系方式

**项目**: EasiTradeCoins
**作者**: Aitachi
**邮箱**: 44158892@qq.com
**GitHub**: https://github.com/aitachi/easiTradeCoins

---

**文档版本**: 1.0
**最后更新**: 2025-11-02
**维护者**: Aitachi
