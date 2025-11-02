# EasiTradeCoins 专业炒币平台功能设计文档（v2.0）

## 一、文档修订说明

**版本**：v2.0
 **修订重点**：

1. 补充中等难度常见功能
2. 强化交易稳定性与安全性设计
3. 新增订单撮合与交易引擎模块
4. 优化用户体验设计

------

## 二、新增中等难度常见功能

### 2.1 代币与合约功能补充

#### 2.1.1 代币销毁机制增强

- 自动销毁触发器（Auto-Burn Trigger）
  - 按交易量百分比销毁（如每笔交易销毁 0.1%）
  - 按时间周期销毁（每周/月销毁固定数量）
  - 按价格触发销毁（价格低于阈值时启动）
- 销毁证明与透明度
  - 销毁记录上链存储
  - 生成销毁证明 NFT
  - 实时销毁数据仪表盘

#### 2.1.2 代币发行辅助工具

- 空投管理系统（Airdrop Manager）
  - 批量地址导入（CSV/Excel）
  - 反女巫攻击检测（地址聚类分析）
  - 分批发放（避免 Gas 峰值）
  - 领取凭证（Merkle Tree 验证）
- 预售与公募管理（Presale/ICO Manager）
  - 白名单管理（KYC 集成）
  - 分层定价（早鸟/普通/延迟）
  - 软顶/硬顶设置
  - 退款机制（未达软顶自动退款）
  - 线性解锁（TGE 释放 + Vesting）

#### 2.1.3 代币实用功能

- 持币生息（Staking as a Service）
  - 灵活期限（7/30/90/365 天）
  - 复利选项（自动重投）
  - 提前赎回罚金机制
- 代币兑换池（Token Swap Pool）
  - 恒定乘积做市商（CPMM）
  - 集中流动性（Uniswap V3 风格）
  - 手续费分级（VIP 用户降费）

### 2.2 交易辅助功能

#### 2.2.1 价格监控与预警

- 智能价格提醒（Smart Price Alert）
  - 绝对价格提醒（达到 $X）
  - 百分比变动提醒（涨跌 ±X%）
  - 成交量异常提醒（放量 >3 倍）
  - 巨鲸动向提醒（单笔 >$100k）
- K 线与技术指标
  - 实时 K 线图（1m/5m/15m/1h/4h/1d）
  - 技术指标：MA/EMA/MACD/RSI/BOLL/KDJ
  - 自定义指标组合
  - 图表模板保存

#### 2.2.2 订单类型丰富化

- 限价单（Limit Order）
  - 普通限价单
  - 冰山订单（隐藏总量）
  - 时间加权订单（TWAP）
- 条件单（Conditional Order）
  - 止盈止损单（Take Profit/Stop Loss）
  - 追踪止损（Trailing Stop）
  - OCO 订单（One Cancels Other）
  - 时间触发单（定时执行）

#### 2.2.3 交易分析工具

- 交易日志与复盘（Trading Journal）
  - 自动记录每笔交易
  - 盈亏统计（日/周/月）
  - 成功率分析
  - 交易热力图（最佳时段）
- 持仓分析（Portfolio Analytics）
  - 资产分布饼图
  - 收益曲线图
  - 风险敞口分析
  - 相关性矩阵

### 2.3 用户体验增强

#### 2.3.1 界面与交互

- 交易模板（Trading Templates）
  - 保存常用交易参数
  - 一键下单
  - 批量修改订单
- 快捷键系统（Keyboard Shortcuts）
  - F1 买入 / F2 卖出
  - Ctrl+L 限价单 / Ctrl+M 市价单
  - Esc 取消所有订单
- 暗色/亮色主题（Dark/Light Mode）
  - 护眼模式
  - 自定义配色方案

#### 2.3.2 通知与提醒

- 多渠道通知（Multi-Channel Notification）
  - 站内信
  - 邮件
  - Telegram Bot
  - 微信公众号
  - Discord Webhook
- 消息优先级分类
  - 紧急：账户安全、大额异动
  - 重要：订单成交、价格到达
  - 普通：系统公告、活动通知

#### 2.3.3 新手引导

- 交互式教程（Interactive Tutorial）
  - 步骤式引导（5 步完成首次交易）
  - 模拟交易账户（虚拟资金练习）
  - 视频教程库
- 风险测评（Risk Assessment）
  - 问卷评估风险承受能力
  - 推荐适合策略等级
  - 定期重新评估

### 2.4 数据与工具

#### 2.4.1 链上数据分析

- 地址分析器（Address Analyzer）
  - 地址标签（交易所/巨鲸/合约）
  - 资金流向可视化
  - 持仓变动追踪
- 代币安全检测（Token Safety Scanner）
  - 合约代码审计
  - 流动性检查
  - 持币分布分析
  - 社交媒体情绪分析

#### 2.4.2 API 与自动化

- RESTful API
  - 行情数据 API
  - 交易 API（需 API Key）
  - WebSocket 实时推送
- Webhook 通知
  - 订单状态变化
  - 价格触发事件
  - 账户余额变动

------

## 三、交易稳定性与安全性保障

### 3.1 交易引擎架构

#### 3.1.1 核心组件设计



```clojure
┌─────────────────────────────────────────┐
│         负载均衡层（Nginx/HAProxy）        │
└─────────────────┬───────────────────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
┌───▼────┐   ┌───▼────┐   ┌───▼────┐
│ API 网关│   │ API 网关│   │ API 网关│
│ (Node) │   │ (Node) │   │ (Node) │
└───┬────┘   └───┬────┘   └───┬────┘
    │            │            │
    └────────────┼────────────┘
                 │
         ┌───────▼───────┐
         │  消息队列层    │
         │ (Kafka/Redis) │
         └───────┬───────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
┌───▼────┐  ┌───▼────┐  ┌───▼────┐
│撮合引擎1│  │撮合引擎2│  │撮合引擎3│
│  (Go)  │  │  (Go)  │  │  (Go)  │
└───┬────┘  └───┬────┘  └───┬────┘
    │           │           │
    └───────────┼───────────┘
                │
         ┌──────▼──────┐
         │   数据层     │
         │ PostgreSQL  │
         │   + Redis   │
         └─────────────┘
```

#### 3.1.2 撮合引擎设计

- **内存撮合（In-Memory Matching）**

  - 订单簿存储在内存（Red-Black Tree）
  - 价格时间优先算法（Price-Time Priority）
  - 单线程顺序处理（无锁设计）
  - 性能：100,000+ TPS

- **订单状态机**

  

  ```awk
  新建订单 → 验证 → 入队列 → 撮合 → 部分成交/完全成交/取消
                                ↓
                            持久化存储
  ```

- **撮合优先级**

  1. 市价单优先于限价单
  2. 价格优先（买单价高/卖单价低优先）
  3. 时间优先（先下单先成交）
  4. VIP 用户优先（可选）

### 3.2 交易安全机制

#### 3.2.1 订单验证层

- 预检查（Pre-Check）
  - 余额充足性验证
  - 价格合理性检查（偏离市价 <10%）
  - 最小/最大订单量限制
  - 频率限制（单用户 10 单/秒）
- 风险控制规则
  - 单笔订单金额上限
  - 单日交易次数限制
  - 持仓集中度限制（单币种 <30% 总资产）
  - 净仓位限制（杠杆交易）

#### 3.2.2 资金安全

- 冷热钱包分离
  - 热钱包：<10% 总资产，处理日常提现
  - 温钱包：30% 总资产，定时批量结算
  - 冷钱包：60% 总资产，多签离线存储
- 提现审核机制
  - 小额提现：自动审核（<$1,000）
  - 中额提现：人工复核（$1,000–$10,000）
  - 大额提现：多人审批 + 延迟到账（>$10,000）
- 异常检测
  - 首次提现地址需邮件确认
  - 提现地址白名单
  - IP 地址异常登录提醒
  - 设备指纹识别

#### 3.2.3 交易反作弊

- 自成交检测（Self-Trading Detection）
  - 禁止同一用户/关联账户对敲
  - IP 关联检测
  - 设备指纹关联
- 洗盘交易检测（Wash Trading Detection）
  - 短时间内高频对倒
  - 成交量异常放大
  - 价格异常波动
- 刷量防护（Volume Manipulation Prevention）
  - 手续费消耗验证
  - 交易模式识别（机器人特征）
  - 惩罚措施：限制交易/冻结账户

### 3.3 系统稳定性保障

#### 3.3.1 高可用架构

- 多活数据中心（Multi-Active DC）
  - 主数据中心（北京/东京）
  - 备份数据中心（新加坡/法兰克福）
  - 自动故障转移（<5 秒）
- 数据库架构
  - 主从复制（1 主 + 2 从）
  - 读写分离
  - 定时备份（每 6 小时）
  - 增量备份（每 1 小时）

#### 3.3.2 性能优化

- 缓存策略
  - L1 缓存：内存（订单簿/K 线）
  - L2 缓存：Redis（用户信息/配置）
  - L3 缓存：CDN（静态资源）
- 数据库优化
  - 分库分表（按交易对/时间分片）
  - 索引优化（组合索引）
  - 慢查询监控（>100ms 告警）
- 异步处理
  - 订单通知异步发送
  - 数据统计后台计算
  - 日志写入批量提交

#### 3.3.3 监控与告警

- 实时监控指标
  - 系统指标：CPU/内存/磁盘/网络
  - 业务指标：TPS/延迟/成交量/在线用户
  - 安全指标：登录失败率/异常提现/API 调用频率
- 告警机制
  - Level 1：邮件通知（CPU >70%）
  - Level 2：短信 + 电话（CPU >85%）
  - Level 3：自动扩容 + 紧急响应（CPU >95%）
- 日志体系
  - 访问日志（Nginx）
  - 应用日志（ELK Stack）
  - 审计日志（区块链存证）
  - 错误日志（Sentry）

------

## 四、用户体验优化设计

### 4.1 交易体验优化

#### 4.1.1 订单执行优化

- 智能路由（Smart Order Routing）
  - 自动拆分大单（避免滑点）
  - 多池聚合（DEX Aggregator）
  - 最优价格执行
- 订单预估
  - 实时滑点计算
  - Gas 费用估算
  - 预计成交价格
  - 最坏情况模拟

#### 4.1.2 界面响应速度

- 前端性能
  - 虚拟滚动（长列表）
  - 懒加载（图片/数据）
  - Service Worker（离线缓存）
  - WebAssembly（复杂计算）
- 数据更新
  - WebSocket 推送（<100ms 延迟）
  - 增量更新（仅传输变化数据）
  - 心跳检测（断线自动重连）

#### 4.1.3 错误处理

- 友好错误提示
  - 余额不足 → 显示差额 + 充值入口
  - 网络错误 → 自动重试 + 离线模式
  - 滑点过大 → 建议分批交易
- 操作可撤销
  - 订单下达后 3 秒内可撤销
  - 误操作恢复（30 天内）

### 4.2 个性化功能

#### 4.2.1 用户偏好设置

- 交易偏好
  - 默认交易对
  - 默认订单类型
  - 默认滑点容忍度
  - 确认弹窗开关
- 显示偏好
  - 价格小数位数
  - 时区设置
  - 货币单位（USD/CNY/EUR）
  - 图表颜色方案

#### 4.2.2 智能推荐

- 交易建议
  - 基于历史行为推荐交易对
  - 热门交易提醒
  - 相似用户策略推荐
- 资讯推送
  - 持仓币种相关新闻
  - 监控币种公告
  - 市场热点解读

### 4.3 移动端优化

#### 4.3.1 移动 APP 功能

- 生物识别登录
  - 指纹识别
  - 面容识别
  - 手势密码
- 简化交易流程
  - 快速买卖（3 步完成）
  - 语音下单
  - 扫码转账
- 离线功能
  - 缓存历史数据
  - 离线查看持仓
  - 网络恢复自动同步

#### 4.3.2 响应式设计

- 自适应布局
  - 手机（<768px）：单列布局
  - 平板（768px–1024px）：双列布局
  - 桌面（>1024px）：三列布局
- 触控优化
  - 按钮最小尺寸 44×44px
  - 滑动操作（左滑取消订单）
  - 长按显示详情

------

## 五、订单撮合系统详细设计

### 5.1 订单簿数据结构

#### 5.1.1 价格层级（Price Level）

go



```go
type PriceLevel struct {
    Price      decimal.Decimal
    Volume     decimal.Decimal
    OrderCount int
    Orders     *list.List // 同价订单队列
}

type OrderBook struct {
    BuyLevels  *rbtree.Tree // 买单红黑树（降序）
    SellLevels *rbtree.Tree // 卖单红黑树（升序）
    OrderMap   map[string]*Order // 订单 ID 索引
    Symbol     string
    UpdateTime time.Time
}
```

#### 5.1.2 订单对象

go



```go
type Order struct {
    OrderID       string
    UserID        string
    Symbol        string
    Side          string // "buy" / "sell"
    Type          string // "limit" / "market"
    Price         decimal.Decimal
    Quantity      decimal.Decimal
    FilledQty     decimal.Decimal
    Status        string // "pending" / "partial" / "filled" / "cancelled"
    TimeInForce   string // "GTC" / "IOC" / "FOK"
    CreateTime    time.Time
    UpdateTime    time.Time
}
```

### 5.2 撮合算法流程

#### 5.2.1 限价单撮合

armasm



```armasm
1. 接收限价买单（Price = P1, Quantity = Q1）
2. 遍历卖单价格层级（从低到高）
   WHILE 存在卖单 AND 卖单价格 <= P1 AND Q1 > 0:
       a. 取出最早卖单（Price = P2, Quantity = Q2）
       b. 成交量 = min(Q1, Q2)
       c. 成交价 = P2（Maker 价格优先）
       d. 更新订单状态
       e. Q1 -= 成交量
       f. 若 Q2 完全成交则移除，否则更新
   END WHILE
3. 若 Q1 > 0 则将剩余订单加入买单簿
4. 广播成交记录与订单簿更新
```

#### 5.2.2 市价单撮合



```armasm
1. 接收市价买单（Quantity = Q1）
2. 遍历卖单价格层级（从低到高）
   WHILE 存在卖单 AND Q1 > 0:
       a. 按限价单撮合逻辑执行
   END WHILE
3. 若 Q1 > 0 则拒绝订单（流动性不足）
4. 广播成交记录
```

#### 5.2.3 特殊订单处理

- IOC（Immediate or Cancel）
  - 立即成交可成交部分
  - 未成交部分自动撤销
- FOK（Fill or Kill）
  - 必须完全成交
  - 无法完全成交则拒绝订单
- Post-Only
  - 仅作为 Maker（挂单方）
  - 若会立即成交则拒绝

### 5.3 撮合性能优化

#### 5.3.1 数据结构选择

- 红黑树 vs 跳表
  - 红黑树：插入/删除 O(log n)，适合价格层级少
  - 跳表：插入/删除 O(log n)，并发性能更好
- 价格精度优化
  - 价格量化（如精确到 0.0001）
  - 减少价格层级数量

#### 5.3.2 并发处理

- 订单分区（Order Sharding）
  - 按交易对分区（BTC/USDT 独立引擎）
  - 单引擎单线程（无锁）
- 消息队列解耦
  - 接单 → Kafka → 撮合引擎
  - 撮合结果 → Kafka → 通知服务

#### 5.3.3 内存管理

- 对象池（Object Pool）
  - 预分配订单对象
  - 减少 GC 压力
- 内存限制
  - 单交易对订单簿 <100MB
  - 超限时拒绝新订单

### 5.4 订单簿快照与恢复

#### 5.4.1 快照机制

- 增量快照
  - 每 1000 笔成交生成快照
  - 快照 + 增量日志
- 快照内容
  - 所有价格层级
  - 所有未成交订单
  - 最后成交价/时间

#### 5.4.2 灾难恢复



```markdown
1. 加载最新快照
2. 重放增量日志（Redo Log）
3. 验证订单簿一致性（Checksum）
4. 恢复服务（<30 秒）
```

------

## 六、链上交易集成设计

### 6.1 DEX 聚合交易

#### 6.1.1 支持协议

- EVM 链
  - Uniswap V2/V3
  - SushiSwap
  - PancakeSwap
  - Curve
  - Balancer
- Solana
  - Raydium
  - Orca
  - Jupiter Aggregator
- 跨链桥
  - Wormhole
  - LayerZero
  - Stargate

#### 6.1.2 智能路由算法



```markdown
输入：代币 A, 代币 B, 数量 Q
输出：最优交易路径

1. 扫描所有 DEX 的 A/B 直接池
2. 扫描所有 A/C + C/B 两跳路径（C 为中间币）
3. 计算每条路径的：
   - 成交价格
   - Gas 费用
   - 滑点
   - 综合成本 = 价格差 + Gas + 滑点
4. 选择综合成本最低的路径
5. 支持路径拆分（如 50% 路径 1 + 50% 路径 2）
```

#### 6.1.3 交易执行

- 批量交易（Batch Transaction）
  - 多笔交易打包为一笔
  - 原子性保证（全部成功或全部失败）
- MEV 保护
  - Flashbots Protect RPC
  - 私有内存池提交
  - 抗三明治攻击

### 6.2 跨链交易

#### 6.2.1 跨链桥集成

- Lock-Mint 模式
  - 源链锁定资产
  - 目标链铸造映射资产
- Burn-Mint 模式
  - 源链销毁资产
  - 目标链铸造等量资产
- 流动性池模式
  - 双链流动性池
  - 快速跨链（无需等待确认）

#### 6.2.2 跨链安全

- 多签验证
  - 中继节点多签确认
  - 门槛 >2/3
- 欺诈证明
  - 挑战期机制
  - 恶意中继惩罚

------

## 七、完整技术架构图



```gcode
┌───────────────────────────────────────────────────────────┐
│                      用户层（Multi-Platform）                │
│   Web App (React) │ Mobile App (React Native) │ API       │
└─────────────────────────┬─────────────────────────────────┘
                          │
┌─────────────────────────▼─────────────────────────────────┐
│                     API 网关层（Kong/Traefik）               │
│   鉴权 │ 限流 │ 路由 │ 协议转换 │ 监控                       │
└─────────────────────────┬─────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
┌───────▼────────┐ ┌─────▼──────┐ ┌───────▼────────┐
│   用户服务      │ │  交易服务   │ │   资产服务      │
│  - 注册/登录   │ │ - 下单     │ │  - 充值/提现   │
│  - KYC        │ │ - 撤单     │ │  - 划转       │
│  - 2FA        │ │ - 查询     │ │  - 对账       │
└───────┬────────┘ └─────┬──────┘ └───────┬────────┘
        │                 │                 │
        └─────────────────┼─────────────────┘
                          │
                  ┌───────▼───────┐
                  │   消息队列层    │
                  │  Kafka/RabbitMQ│
                  └───────┬───────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
┌───────▼────────┐ ┌─────▼──────┐ ┌───────▼────────┐
│   撮合引擎      │ │  风控引擎   │ │   通知引擎      │
│  - 订单簿      │ │ - 反洗钱    │ │  - 邮件       │
│  - 撮合算法    │ │ - 限额     │ │  - 短信       │
│  - 成交推送    │ │ - 异常检测  │ │  - Telegram   │
└───────┬────────┘ └─────┬──────┘ └───────┬────────┘
        │                 │                 │
        └─────────────────┼─────────────────┘
                          │
                  ┌───────▼───────┐
                  │    数据层      │
                  │ PostgreSQL (主)│
                  │ + TimescaleDB  │
                  │ + Redis (缓存) │
                  │ + MongoDB (日志)│
                  └───────┬───────┘
                          │
                  ┌───────▼───────┐
                  │   区块链层     │
                  │ - EVM RPC     │
                  │ - Solana RPC  │
                  │ - 智能合约     │
                  └───────────────┘
```

------

## 八、核心数据库设计

### 8.1 用户相关表

sql



```sql
-- 用户基础信息表
CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(64) NOT NULL,
    kyc_level INT DEFAULT 0, -- 0未认证 1初级 2高级
    status INT DEFAULT 1, -- 1正常 2冻结 3注销
    register_ip INET,
    register_time TIMESTAMP DEFAULT NOW(),
    last_login_time TIMESTAMP,
    last_login_ip INET,
    INDEX idx_email (email),
    INDEX idx_phone (phone)
);

-- 用户安全设置表
CREATE TABLE user_security (
    user_id BIGINT PRIMARY KEY REFERENCES users(user_id),
    google_2fa_secret VARCHAR(64),
    is_2fa_enabled BOOLEAN DEFAULT FALSE,
    withdrawal_whitelist JSONB, -- 提现地址白名单
    api_key_hash VARCHAR(255),
    api_secret_hash VARCHAR(255),
    api_permissions JSONB, -- API 权限配置
    update_time TIMESTAMP DEFAULT NOW()
);

-- KYC 认证表
CREATE TABLE user_kyc (
    kyc_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id),
    real_name VARCHAR(100),
    id_number VARCHAR(50),
    id_type INT, -- 1身份证 2护照 3驾照
    id_front_url VARCHAR(500),
    id_back_url VARCHAR(500),
    face_url VARCHAR(500),
    country_code VARCHAR(10),
    submit_time TIMESTAMP DEFAULT NOW(),
    audit_time TIMESTAMP,
    audit_status INT, -- 0待审核 1通过 2拒绝
    audit_remark TEXT,
    INDEX idx_user_id (user_id)
);
```

### 8.2 资产相关表

sql



```sql
-- 用户资产表
CREATE TABLE user_assets (
    asset_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id),
    currency VARCHAR(20) NOT NULL, -- BTC/ETH/USDT
    chain VARCHAR(20), -- ERC20/TRC20/BEP20
    available DECIMAL(36, 18) DEFAULT 0, -- 可用余额
    frozen DECIMAL(36, 18) DEFAULT 0, -- 冻结余额
    update_time TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, currency, chain),
    INDEX idx_user_id (user_id)
);

-- 充值记录表
CREATE TABLE deposits (
    deposit_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id),
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    amount DECIMAL(36, 18) NOT NULL,
    address VARCHAR(200), -- 充值地址
    txid VARCHAR(200), -- 链上交易哈希
    confirmations INT DEFAULT 0, -- 确认数
    required_confirmations INT, -- 需要确认数
    status INT, -- 0待确认 1已到账 2异常
    create_time TIMESTAMP DEFAULT NOW(),
    confirm_time TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_txid (txid),
    INDEX idx_status_time (status, create_time)
);

-- 提现记录表
CREATE TABLE withdrawals (
    withdrawal_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id),
    currency VARCHAR(20) NOT NULL,
    chain VARCHAR(20),
    amount DECIMAL(36, 18) NOT NULL,
    fee DECIMAL(36, 18), -- 提现手续费
    address VARCHAR(200) NOT NULL, -- 提现地址
    txid VARCHAR(200), -- 链上交易哈希
    status INT, -- 0待审核 1审核通过 2处理中 3已完成 4拒绝
    audit_user_id BIGINT, -- 审核人
    audit_time TIMESTAMP,
    complete_time TIMESTAMP,
    create_time TIMESTAMP DEFAULT NOW(),
    remark TEXT,
    INDEX idx_user_id (user_id),
    INDEX idx_status_time (status, create_time)
);
```

### 8.3 交易相关表

sql



```sql
-- 订单表（分表：按交易对 + 时间分片）
CREATE TABLE orders_btc_usdt_202511 (
    order_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    symbol VARCHAR(20) NOT NULL, -- BTC_USDT
    side CHAR(4) NOT NULL, -- buy/sell
    type CHAR(10) NOT NULL, -- limit/market
    price DECIMAL(36, 18),
    quantity DECIMAL(36, 18) NOT NULL,
    filled_quantity DECIMAL(36, 18) DEFAULT 0,
    filled_amount DECIMAL(36, 18) DEFAULT 0, -- 成交金额
    avg_price DECIMAL(36, 18), -- 平均成交价
    fee DECIMAL(36, 18) DEFAULT 0, -- 手续费
    fee_currency VARCHAR(20), -- 手续费币种
    status INT, -- 0待成交 1部分成交 2完全成交 3已撤销
    time_in_force CHAR(3), -- GTC/IOC/FOK
    create_time TIMESTAMP DEFAULT NOW(),
    update_time TIMESTAMP DEFAULT NOW(),
    INDEX idx_user_time (user_id, create_time DESC),
    INDEX idx_status_time (status, create_time DESC)
) PARTITION BY RANGE (create_time);

-- 成交记录表（分表）
CREATE TABLE trades_btc_usdt_202511 (
    trade_id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    buy_order_id BIGINT NOT NULL,
    sell_order_id BIGINT NOT NULL,
    buyer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    price DECIMAL(36, 18) NOT NULL,
    quantity DECIMAL(36, 18) NOT NULL,
    amount DECIMAL(36, 18) NOT NULL,
    buyer_fee DECIMAL(36, 18),
    seller_fee DECIMAL(36, 18),
    trade_time TIMESTAMP DEFAULT NOW(),
    INDEX idx_buy_order (buy_order_id),
    INDEX idx_sell_order (sell_order_id),
    INDEX idx_time (trade_time DESC)
) PARTITION BY RANGE (trade_time);

-- K 线数据表（TimescaleDB 超表）
CREATE TABLE klines (
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
    PRIMARY KEY (symbol, interval, open_time)
);

SELECT create_hypertable('klines', 'open_time');
```

### 8.4 系统配置表

sql



```sql
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
    create_time TIMESTAMP DEFAULT NOW()
);

-- 系统参数配置表
CREATE TABLE system_config (
    config_key VARCHAR(100) PRIMARY KEY,
    config_value TEXT,
    description TEXT,
    update_time TIMESTAMP DEFAULT NOW()
);

-- 插入示例配置
INSERT INTO system_config VALUES
('withdrawal_min_confirmations', '{"BTC": 3, "ETH": 12, "USDT_ERC20": 12}', '提现最小确认数'),
('daily_withdrawal_limit', '{"level0": 1000, "level1": 10000, "level2": 100000}', '每日提现限额（USD）'),
('api_rate_limit', '{"public": 100, "private": 50}', 'API 请求限流（次/分钟）');
```

------

## 九、API 接口设计

### 9.1 RESTful API

#### 9.1.1 认证接口

awk



```awk
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh-token
POST /api/v1/auth/enable-2fa
POST /api/v1/auth/verify-2fa
```

#### 9.1.2 市场行情接口（公开）

awk



```awk
GET /api/v1/market/ticker/{symbol}          # 24h 行情
GET /api/v1/market/depth/{symbol}           # 订单簿深度
GET /api/v1/market/trades/{symbol}          # 最新成交
GET /api/v1/market/klines/{symbol}          # K 线数据
GET /api/v1/market/symbols                  # 交易对列表
```

**示例响应（Depth）：**

json



```json
{
  "symbol": "BTC_USDT",
  "timestamp": 1698825600000,
  "bids": [
    ["34250.50", "1.5"],
    ["34250.00", "2.3"],
    ["34249.50", "0.8"]
  ],
  "asks": [
    ["34251.00", "1.2"],
    ["34251.50", "3.1"],
    ["34252.00", "0.5"]
  ]
}
```

#### 9.1.3 交易接口（私有）

awk



```awk
POST /api/v1/order/create                   # 下单
DELETE /api/v1/order/{orderId}              # 撤单
GET /api/v1/order/{orderId}                 # 订单详情
GET /api/v1/order/open                      # 当前委托
GET /api/v1/order/history                   # 历史订单
GET /api/v1/trade/history                   # 成交历史
```

**示例请求（下单）：**

json



```json
{
  "symbol": "BTC_USDT",
  "side": "buy",
  "type": "limit",
  "price": "34250.00",
  "quantity": "0.5",
  "timeInForce": "GTC"
}
```

#### 9.1.4 资产接口（私有）

awk



```awk
GET /api/v1/account/balance                 # 账户余额
GET /api/v1/account/deposits                # 充值记录
POST /api/v1/account/withdraw               # 提现申请
GET /api/v1/account/withdrawals             # 提现记录
GET /api/v1/account/address/{currency}      # 获取充值地址
```

### 9.2 WebSocket API

#### 9.2.1 连接与认证

javascript



```javascript
// 公共频道（无需认证）
ws://api.easitradecoins.com/ws/public

// 私有频道（需认证）
ws://api.easitradecoins.com/ws/private?token=xxx
```

#### 9.2.2 订阅消息格式

json



```json
{
  "method": "SUBSCRIBE",
  "params": [
    "btc_usdt@ticker",
    "btc_usdt@depth@100ms",
    "btc_usdt@trade"
  ],
  "id": 1
}
```

#### 9.2.3 推送消息示例

**实时成交：**

json



```json
{
  "e": "trade",
  "s": "BTC_USDT",
  "t": 1698825601234,
  "p": "34251.50",
  "q": "0.15",
  "m": true  // true=买方主动成交
}
```

**订单簿更新：**

json



```json
{
  "e": "depthUpdate",
  "s": "BTC_USDT",
  "b": [["34250.00", "1.5"]],  // 更新的买单
  "a": [["34251.00", "0"]]     // 删除的卖单
}
```

**个人订单更新：**

json



```json
{
  "e": "orderUpdate",
  "o": {
    "orderId": "123456789",
    "status": "FILLED",
    "filledQty": "0.5",
    "avgPrice": "34250.50"
  }
}
```

------

## 十、安全策略总结

### 10.1 多层防御体系

| 层级       | 防护措施                               | 检测手段       | 响应机制            |
| ---------- | -------------------------------------- | -------------- | ------------------- |
| **应用层** | 输入验证、SQL 注入防护、XSS 过滤       | WAF 规则匹配   | 拒绝请求 + 记录日志 |
| **认证层** | 密码强度、2FA、设备指纹、IP 白名单     | 异常登录检测   | 验证码 + 邮件确认   |
| **交易层** | 限价检查、余额验证、频率限制、反自成交 | 实时规则引擎   | 拒绝订单 + 限制账户 |
| **资产层** | 冷热分离、多签、提现审核、地址白名单   | 大额异动监控   | 人工审核 + 延迟到账 |
| **网络层** | DDoS 防护、TLS 1.3、HTTPS 强制         | 流量分析       | 黑洞路由 + CDN 缓存 |
| **数据层** | 加密存储、访问控制、审计日志、定期备份 | 数据完整性校验 | 回滚 + 灾难恢复     |

### 10.2 关键安全指标

- **密码策略**：≥12 位，包含大小写 + 数字 + 符号
- **登录失败锁定**：5 次失败锁定 30 分钟
- **Session 有效期**：Web 2 小时，Mobile 7 天
- **API 限流**：公开接口 100 次/分钟，私有接口 50 次/分钟
- **提现延迟**：大额提现（>$10k）延迟 24 小时
- **数据备份**：全量备份每日 1 次，增量备份每小时 1 次
- **安全审计**：代码审计每季度 1 次，渗透测试每半年 1 次

------

## 十一、运营支持功能

### 11.2 营销活动系统

#### 11.2.1 活动类型

- 交易挖矿
  - 按交易量奖励平台币
  - 手续费返佣
- 邀请返佣
  - 一级邀请 20% 手续费返佣
  - 二级邀请 10% 手续费返佣
- 空投活动
  - 注册空投
  - 交易空投
  - 持币空投
- VIP 等级体系
  - 30 日交易量分级
  - 手续费折扣（0.1% → 0.02%）
  - 专属客服

#### 11.2.2 活动管理

- 活动配置
  - 开始/结束时间
  - 参与条件
  - 奖励规则
  - 活动预算
- 活动监控
  - 参与人数
  - 奖励发放
  - 成本核算
  - ROI 分析

### 11.3 数据分析与报表

#### 11.3.1 运营数据

- 用户数据
  - 日活/月活用户
  - 新增用户
  - 留存率
  - 用户画像
- 交易数据
  - 交易量/交易额
  - 活跃交易对
  - 手续费收入
  - 深度分析
- 资产数据
  - 平台总资产
  - 充值/提现量
  - 币种分布
  - 冷热钱包比例

#### 11.3.2 BI 报表

- 实时大屏
  - 实时交易量
  - 在线用户数
  - 系统状态
- 定期报表
  - 日报（每日 9:00 发送）
  - 周报（每周一发送）
  - 月报（每月 1 日发送）
- 自定义报表
  - 拖拽式报表构建
  - 多维度分析
  - 导出 Excel/PDF

------

## 十二、合规与风控

### 12.1 反洗钱（AML）

#### 12.1.1 用户尽职调查（CDD）

- KYC 分级
  - Level 0：仅邮箱注册，限额 $1,000/日
  - Level 1：身份认证，限额 $10,000/日
  - Level 2：高级认证，限额 $100,000/日
- 增强尽职调查（EDD）
  - 高风险国家用户
  - 大额交易用户（>$50k）
  - 政治公众人物（PEP）

#### 12.1.2 交易监控

- **可疑交易特征**

  - 结构化交易（拆分大额为小额）
  - 快进快出（充值后立即提现）
  - 频繁小额交易
  - 资金来源不明

- **风险评分模型**

  

  ```apache
  风险分 = 交易频率 × 0.2 + 交易金额 × 0.3 + 
           关联账户 × 0.2 + 地域风险 × 0.15 + 
           历史记录 × 0.15
  ```

- **自动冻结规则**

  - 风险分 >80 分自动冻结
  - 黑名单地址交互
  - 混币器地址交互



------

## 

#### 13.1.2 网络架构



```clojure
                    ┌────────────┐
                    │   CDN      │
                    │ (Cloudflare)│
                    └─────┬──────┘
                          │
                    ┌─────▼──────┐
                    │ DDoS 防护  │
                    │  (Layer 7) │
                    └─────┬──────┘
                          │
                ┌─────────▼─────────┐
                │   负载均衡器       │
                │ (LVS/HAProxy)     │
                └─────────┬─────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
    ┌───▼───┐         ┌───▼───┐         ┌───▼───┐
    │Web服务│         │Web服务│         │Web服务│
    └───┬───┘         └───┬───┘         └───┬───┘
        │                 │                 │
        └─────────────────┼─────────────────┘
                          │
                    ┌─────▼──────┐
                    │  内网交换机 │
                    └─────┬──────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
    ┌───▼───┐         ┌───▼───┐         ┌───▼───┐
    │数据库 │         │缓存   │         │消息队列│
    └───────┘         └───────┘         └───────┘
```



------

## 十四、功能优先级与开发路线图

### 14.1 MVP- Phase 1

**核心功能：**

1. 用户系统（注册/登录/KYC）
2. 钱包系统（充值/提现/余额）
3. 现货交易（限价单/市价单）
4. 基础撮合引擎
5. 订单簿与 K 线展示
6. 基础风控（限额/频率限制）

**技术实现：**

- 单链支持（EVM）
- 单一撮合引擎
- PostgreSQL + Redis
- 基础 Web 界面

### 14.2 扩展版本 - Phase 2

**新增功能：**

1. 多链支持（Solana/TRON）
2. 代币创建工具
3. 流动性管理
4. 批量转账
5. API 接口
6. 移动端 APP

**优化：**

- 撮合引擎性能优化
- 数据库分库分表
- WebSocket 实时推送

### 14.3 专业版本 - Phase 3

**高级功能：**

1. 自动理财优化
2. 量化交易策略
3. DEX 聚合交易
4. 跨链交易
5. 社交交易
6. 合规税务报告

**企业级：**

- 多活数据中心
- 智能风控引擎
- 机构级 API
- 合规审计系统

------

## 十五、总结与核心竞争力

### 15.1 完整功能清单

| 类别         | 功能模块        | 实现难度 | 优先级 |
| ------------ | --------------- | -------- | ------ |
| **代币管理** | 标准代币创建    | 低       | P0     |
|              | 多签/时间锁代币 | 中       | P1     |
|              | 销毁/空投/预售  | 低       | P1     |
| **交易核心** | 现货交易        | 中       | P0     |
|              | 高级订单类型    | 中       | P1     |
|              | DEX 聚合        | 高       | P2     |
| **资产管理** | 充值提现        | 中       | P0     |
|              | 自动理财        | 高       | P2     |
|              | 批量转账        | 中       | P1     |
| **风控安全** | KYC/AML         | 中       | P0     |
|              | 智能风控        | 高       | P1     |
|              | 多签安全        | 中       | P1     |
| **用户体验** | K 线图表        | 低       | P0     |
|              | 价格预警        | 低       | P1     |
|              | 移动端          | 中       | P1     |
| **运营支持** | 工单系统        | 低       | P1     |
|              | 营销活动        | 中       | P2     |
|              | BI 报表         | 中       | P2     |

