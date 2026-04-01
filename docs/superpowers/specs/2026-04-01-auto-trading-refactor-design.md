# Alpha Pulse V2 — 系统架构参考 PRD

> **读者**：系统作者（自用）+ AI 开发助手（驱动精准实现）
> **时态**：As-Is（现状）→ Gap Analysis（债务）→ To-Be（目标）
> **维护约定**：每次模块接口变更后同步更新对应契约卡片

---

## 1. 系统定位

### 一句话定义

Alpha Pulse V2 是 AI 驱动的加密货币合约方向判断与自动交易终端，以多周期市场结构分析为核心，实现从行情采集到自动下单的全链路自动化。

### 核心能力边界

| 能力 | 状态 |
|------|------|
| 多周期市场快照分析（1m/5m/15m/1h/4h） | ✅ 已实现 |
| 多因子信号评分（技术/订单流/结构/流动性） | ✅ 已实现 |
| Direction Copilot A 级信号判断 | ✅ 已实现 |
| 自动限价开仓 + 止损止盈保护单 | ✅ 已实现 |
| 飞书机器人告警推送 | ✅ 已实现 |
| 信号配置热更新（无需重启） | ❌ 待实现（G-04） |
| 事件驱动分析（<1s 延迟） | ❌ 待实现（G-02） |
| Trade Runtime 退避熔断 | ❌ 待实现（G-09） |

### 不做什么

- 不支持现货交易（仅合约）
- 不支持多账户管理
- 不提供策略回测框架
- 不支持多交易所（仅 Binance Futures）

---

## 2. As-Is 模块契约

> 每个模块用统一格式描述，AI 修改前必须先读对应契约卡片。

---

### 2.1 Collector（采集层）

**文件**：`backend/internal/collector/`

| 字段 | 描述 |
|------|------|
| **职责** | 从 Binance 采集原始市场数据并落库，不做任何分析 |
| **输入** | Binance REST API（K线）+ WebSocket（AggTrade、PartialDepth） |
| **输出** | MySQL 表：`kline`、`agg_trades`、`order_book_snapshots` |
| **前置条件** | Binance 连接可用；dev 模式下自动切换 Mock 数据 |
| **副作用** | DB 写入；WebSocket 断线自动重连 |
| **禁止** | 不做任何指标计算；不调用其他引擎 |
| **goroutine** | 2 个长期 goroutine（AggTrade loop + PartialDepth loop），均受 context 控制 |

---

### 2.2 Indicator Engine（指标引擎）

**文件**：`backend/internal/indicator/indicator_engine.go`

| 字段 | 描述 |
|------|------|
| **职责** | 基于 K 线序列计算技术指标 |
| **输入** | 120 根 K 线（从 DB 查询） |
| **输出** | `models.Indicator`：EMA20/50、VWAP、Bollinger Band、RSI-14、MACD(12,26,9)、ATR-14 |
| **前置条件** | 最少 50 根 K 线，否则返回空结构 |
| **硬编码参数** | RSI=14, MACD=(12,26,9), EMA=(20,50), ATR=14（`indicator_engine.go:56-80`） |
| **禁止** | 不访问订单流数据；不访问盘口数据 |

---

### 2.3 OrderFlow Engine（订单流引擎）

**文件**：`backend/internal/orderflow/orderflow_engine.go`

| 字段 | 描述 |
|------|------|
| **职责** | 分析成交数据识别主动买卖、大单、冰山单 |
| **输入** | 最近 25 条聚合成交 + 盘口快照 |
| **输出** | `models.OrderFlow`：Delta、CVD、大单事件、吸收信号、冰山检测 |
| **前置条件** | 最少 60 条成交记录；大单阈值 100K USDT（`orderflow_engine.go:77-80`，硬编码） |
| **禁止** | 不访问 K 线数据 |

---

### 2.4 Structure Engine（结构引擎）

**文件**：`backend/internal/structure/structure_engine.go`

| 字段 | 描述 |
|------|------|
| **职责** | 识别市场结构：高低点序列、结构突破、换手信号 |
| **输入** | 30 根 K 线 |
| **输出** | `models.Structure`：HH/HL/LH/LL 序列、BOS（结构突破）、CHOCH（换手） |
| **前置条件** | 最少 30 根 K 线，否则返回空 |

---

### 2.5 Liquidity Engine（流动性引擎）

**文件**：`backend/internal/liquidity/liquidity_engine.go`

| 字段 | 描述 |
|------|------|
| **职责** | 识别流动性区域：相等价位、止损簇、扫单事件 |
| **输入** | 25 根 K 线 + 盘口 20 档深度 |
| **输出** | `models.Liquidity`：流动性区域列表、扫单事件、深度不对称 |

---

### 2.6 Signal Engine（信号引擎）

**文件**：`backend/internal/signal/signal_engine.go`

| 字段 | 描述 |
|------|------|
| **职责** | 聚合 4 个引擎结果，输出综合方向信号和置信度 |
| **输入** | Indicator + OrderFlow + Structure + Liquidity + 当前价格 |
| **输出** | `models.Signal`：score∈[-100,100]、direction(BUY/SELL/NEUTRAL)、confidence∈[5,95] |
| **阈值（硬编码）** | buyThreshold=35、sellThreshold=-35（`signal_engine.go:12-15`） |
| **置信度公式** | base=46+min(∣score∣/2,28)；alignmentBonus=min(同向因子数×4,20)；oppositionPenalty=min(逆向因子数×5,15)；final=clamp(base+bonus-penalties,5,95)（`signal_engine.go:514-548`） |
| **禁止** | 不直接访问 DB；不调用 Binance API |

**因子权重（硬编码）**：

| 因子 | 最大贡献值 |
|------|-----------|
| scoreTrend（EMA/价格结构） | ±25 |
| scoreMomentum（RSI/MACD） | ±25 |
| scoreOrderFlow（Delta/CVD/大单） | ±26 |
| scoreStructure（BOS/CHOCH） | ±20 |
| scoreLiquidity（流动性区域） | ±15 |
| scoreMicrostructure（盘口/微结构） | ±18 |

---

### 2.7 Direction Copilot

**文件**：`backend/internal/service/direction_copilot.go`

| 字段 | 描述 |
|------|------|
| **职责** | 四周期联合决策，过滤低质量信号，输出可交易方向 |
| **输入** | 4h/1h/15m/5m 四个 Signal（必须齐全） |
| **输出** | `DirectionDecision`：state/weighted_bias/tradeability/risk_label/no_trade_reason |
| **A 级判定条件** | confidence≥72 AND ∣weighted_bias∣≥2.7（`direction_copilot.go:157-159`） |
| **加权公式** | weighted_bias = 1h×1.35 + 4h×0.85 + 15m×0.55 + 5m×0.35 + futures_support |

**No-Trade 六层过滤（顺序执行，任一触发即返回）**：

| 层级 | 条件 | 原因标签 |
|------|------|---------|
| 1 | 缺少任一周期快照 | 数据不完整 |
| 2 | 1h confidence<55 OR ∣score∣<12 | 主周期方向不明确 |
| 3 | 4h 与 1h 方向冲突 | 逆大级别风险 |
| 4 | 15m 与 1h 方向冲突 | 提前动手风险 |
| 5 | 5m 与 15m 方向冲突 | 最后一脚反转 |
| 6 | 期货拥挤度异常 | long-squeeze/short-squeeze/funding堆积 |

---

### 2.8 Alert Service

**文件**：`backend/internal/service/alert_service.go`

| 字段 | 描述 |
|------|------|
| **职责** | 消费 DirectionDecision，生成告警事件，触发下单或推送 |
| **输入** | DirectionDecision + 告警偏好配置 |
| **输出** | AlertEvent（类型：setup_ready/no_trade/outcome） |
| **副作用** | 写 DB（alerts 表）；飞书推送；触发 AutoTradeCoordinator |
| **触发下单条件** | AlertEvent.Type == "setup_ready" AND TradeEnabled AND TradeAutoExecute |

---

### 2.9 Trade Executor + Trade Runtime

**文件**：`backend/internal/service/trade_executor.go`、`trade_runtime.go`、`auto_trade_coordinator.go`

| 字段 | 描述 |
|------|------|
| **职责** | 执行开仓、保护单、持仓盯单 |
| **输入** | AlertEvent(setup_ready) + 交易配置（TradeConfig） |
| **开仓流程** | 校验风险比（RiskReward≥MinRiskReward）→ 拉账户状态 → 计算仓量 → PlaceFuturesLimitOrder |
| **仓量公式** | qty = (balance × riskPct% × leverage) / entryPrice |
| **保护单** | FILLED 后立即下 STOP_MARKET(止损) + TAKE_PROFIT_MARKET(止盈) |
| **盯单** | 每 3s ReconcilePendingEntries；每 15s SyncPositions |
| **禁止** | 不直接读取 Signal 数据；不调用分析引擎 |

**关键配置变量**（`config/config.go:52-56`）：

| 变量 | 说明 |
|------|------|
| `TRADE_ENABLED` | 全局开关 |
| `TRADE_AUTO_EXECUTE` | 自动触发开关 |
| `TRADE_ALLOWED_SYMBOLS` | 白名单 |
| `TRADE_WATCHER_INTERVAL_SECONDS` | 盯单周期（默认 3s） |
| `TRADE_SYNC_INTERVAL_SECONDS` | 持仓同步周期（默认 15s） |

---

### 2.10 Scheduler

**文件**：`backend/internal/scheduler/jobs.go`

| 字段 | 描述 |
|------|------|
| **职责** | 定时驱动完整分析→信号→告警→下单链路 |
| **触发间隔** | 15s（`SCHEDULER_INTERVAL_SECONDS`，最小 1s） |
| **任务序列** | WarmupSymbol × 3符号 → GetSignal(1m) → EvaluateAll → OutcomeTracker |
| **goroutine** | 1 个长期 goroutine，受 context 控制 |
| **当前问题** | 轮询驱动导致最坏 15s 信号延迟（G-02） |

---

## 3. As-Is 数据流

```
Binance WebSocket
    ↓ AggTrade / PartialDepth（实时）
Collector → MySQL（kline / agg_trades / order_book_snapshots）

Scheduler（每 15s）
    ↓
MarketService.WarmupSymbol(symbol) × 3符号
    ↓
[串行] Indicator → OrderFlow → Structure → Liquidity → Signal Engine
    ↓ 写 Redis（Snapshot TTL 5s / Analysis TTL 15s）
    ↓
Direction Copilot（4h/1h/15m/5m 联合决策）
    ↓
Alert Service（生成 setup_ready 事件）
    ↓ [条件：TradeEnabled + TradeAutoExecute]
AutoTradeCoordinator → Trade Executor
    ↓ [3次同步 API：GetBalance + GetLeverage + GetRules]
Binance Futures API → PlaceFuturesLimitOrder
    ↓
Trade Runtime（每3s盯单 / 每15s持仓同步）
    ↓
FILLED → PlaceStopLoss + PlaceTakeProfit
```

**端到端延迟**：调度间隔（最坏 15s）+ 快照构建（1-3s）+ 下单 API（300-600ms）≈ **17-19s**

---

## 4. Gap Analysis — 架构债务清单

| ID | 严重性 | 描述 | 位置 | 推荐修复 |
|----|--------|------|------|---------|
| G-01 | 🔴 P0 | 快照构建 6 引擎串行，1-3s 耗时 | `market_service.go` | `errgroup` 并发调用 |
| G-02 | 🔴 P0 | Scheduler 轮询，最坏 15s 延迟 | `scheduler/jobs.go:49` | WebSocket 新K线事件驱动 |
| G-03 | 🔴 P0 | 下单前 3 次同步 API，+300-600ms | `trade_executor.go:63-74` | 预取账户状态缓存 |
| G-04 | 🟡 P1 | 信号阈值 + 因子权重硬编码 | `signal_engine.go:12-15` | DB 配置表 + 热更新 |
| G-05 | 🟡 P1 | 新K线后全量清缓存，首次访问走全表 | `market_service.go` | Write-through 主动回填 |
| G-06 | 🟡 P1 | 持仓对账 N+1 查询 | `trade_runtime.go` | `WHERE symbol IN (...)` 批量 |
| G-07 | 🟡 P1 | 大单查询无分页，可能全表扫描 | repository 层 | 强制 LIMIT + 组合索引 |
| G-08 | 🟢 P2 | 无 Prometheus 指标导出 | `observability/` | 引擎耗时/缓存命中率埋点 |
| G-09 | 🟢 P2 | Trade Runtime 无退避熔断 | `trade_runtime.go` | 指数退避 + 熔断器 |
| G-10 | 🟢 P2 | 置信度公式无文档，AI 修改易破坏 | `signal_engine.go:514-548` | PRD 文档化 + 代码注释引用 |

---

## 5. To-Be 目标架构

### 5.1 核心变更

**变更 1：事件驱动替代轮询（解决 G-02）**

```
旧：Scheduler 每 15s 触发 → 最坏延迟 15s
新：WebSocket 新K线事件 → chan KlineEvent → 立即触发分析 → 延迟 <1s
调度器：保留为兜底补偿机制（网络中断恢复后的全量刷新）
```

实现要点：
- `BinanceStreamCollector` 在检测到新 K 线收盘时发布 `KlineEvent` 到 `chan KlineEvent`
- `MarketService` 订阅该 channel，触发对应 symbol+interval 的快照重建
- Scheduler 降级为 60s 兜底全量刷新

**变更 2：并发快照构建（解决 G-01）**

```
旧：6 引擎串行调用，1-3s
新：errgroup 并发调用，耗时 = 最慢引擎，约 200-500ms
```

实现要点：
- 各引擎输入数据相互独立（K线/成交/盘口分别查询），无竞争
- `errgroup.WithContext` 管理，任一引擎失败降级为空结果，不阻断整体

**变更 3：账户状态预取缓存（解决 G-03）**

```
旧：下单时 3 次同步 API，+300-600ms
新：后台每 30s 预取 Balance/Leverage/Rules → 写内存缓存
    下单时读缓存，延迟 <1ms
```

实现要点：
- 新增 `AccountStateCache` 结构体，`sync.RWMutex` 保护
- 下单前校验缓存时效（超过 60s 强制刷新）

**变更 4：信号配置表（解决 G-04）**

```
旧：buyThreshold=35 硬编码，调优需重新编译
新：DB 表 signal_configs(id, symbol, interval, key, value, updated_at)
    启动时加载到内存；HTTP API PATCH /api/signal-configs 热更新
    Signal Engine 通过 ConfigProvider 接口读取
```

**变更 5：Write-through 缓存（解决 G-05）**

```
旧：新K线 → 清空缓存 → 首次访问重算（DB 压力峰值）
新：新K线 → 并发重算 → 主动写入 Redis → 后续请求 100% 命中
```

**变更 6：Trade Runtime 退避熔断（解决 G-09）**

```
旧：每 3s 无限重试，API 故障时触发 IP 封禁风险
新：指数退避（1s → 2s → 4s → 8s，上限 60s）
    连续失败 5 次 → 熔断，发飞书告警
    Binance API 恢复后自动重置
```

---

### 5.2 To-Be 数据流

```
Binance WebSocket
    ↓ 新K线收盘事件
EventBus（chan KlineEvent）
    ↓ 立即触发（<100ms）
MarketService.RebuildSnapshot(symbol, interval)
    ↓
[并发 errgroup] Indicator ∥ OrderFlow ∥ Structure ∥ Liquidity
    ↓ 汇聚
Signal Engine（读配置表阈值）
    ↓ Write-through
Redis Cache（Snapshot TTL 5s）
    ↓
Direction Copilot（4h/1h/15m/5m）
    ↓
Alert Service → AutoTradeCoordinator
    ↓ [读账户状态缓存，无额外 API 调用]
Trade Executor → PlaceFuturesLimitOrder
    ↓
Trade Runtime（指数退避盯单 + 熔断保护）
    ↓
FILLED → StopLoss + TakeProfit
```

**目标端到端延迟**：事件触发（<100ms）+ 并发快照（200-500ms）+ 下单（<50ms）≈ **<1s**

---

### 5.3 To-Be 模块边界约定

| 模块 | 可以做 | 禁止做 |
|------|--------|--------|
| 各分析引擎（Indicator/OrderFlow/Structure/Liquidity） | 接收参数，纯计算，返回结果 | 直接访问 DB；调用其他引擎；调用 Binance API |
| MarketService | 编排引擎调用、管理缓存、发布事件 | 包含业务决策逻辑（方向判断） |
| Signal Engine | 读 ConfigProvider 接口获取阈值 | 硬编码任何数值 |
| Direction Copilot | 多周期联合决策 | 直接访问 DB 或 Binance API |
| Trade Executor | 下单、读账户状态缓存 | 直接读取 Signal/Analysis 数据 |
| Alert Service | 生成告警、触发下单链路 | 直接调用 Binance API |
| AccountStateCache | 缓存账户余额/杠杆/规则 | 不参与交易决策 |

---

## 6. AI 开发约定

> AI 在修改任何模块前，必须先读对应的契约卡片，确认输入/输出/禁止事项后再动手。

### 6.1 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 引擎接口 | `XxxEngine` | `IndicatorEngine`、`SignalEngine` |
| 服务层 | `XxxService` | `MarketService`、`AlertService` |
| 仓储层 | `XxxRepository` | `SignalRepository`、`OrderRepository` |
| 配置提供者 | `XxxProvider` | `ConfigProvider`、`AccountStateCache` |
| 事件类型 | `XxxEvent` | `KlineEvent`、`AlertEvent` |

### 6.2 变更守则

1. **新增引擎**：必须实现无状态纯函数接口；必须在本 PRD 添加契约卡片
2. **修改信号阈值**：通过 ConfigProvider 接口，不得硬编码
3. **新增 DB 表**：在 CLAUDE.md 数据库章节更新表清单；dev 模式 AutoMigrate 自动执行
4. **修改下单逻辑**：必须有对应单元测试（mock Binance API）；必须考虑重复下单防护
5. **单文件超 500 行**：拆分，按职责分文件，不允许例外

### 6.3 测试约定

| 场景 | 测试类型 | Mock 范围 |
|------|---------|---------|
| 分析引擎计算逻辑 | 单元测试 | 无需 Mock，纯函数 |
| Signal 评分 | 单元测试 | Mock 引擎输出 |
| Direction Copilot 决策 | 单元测试 | Mock Signal 数据 |
| Trade Executor 下单 | 单元测试 | Mock Binance API |
| 调度→快照→信号完整链路 | 集成测试 | APP_MODE=test |

### 6.4 性能目标（To-Be）

| 指标 | 当前值 | 目标值 |
|------|--------|--------|
| 新K线→信号生成延迟 | 15-19s | <1s |
| 单次快照构建耗时 | 1-3s | <500ms |
| 信号→下单延迟 | 300-600ms | <100ms |
| Redis 缓存命中率 | ~60%（估算） | >95% |
| Binance API 调用（下单路径） | 3次同步 | 0次（读缓存） |

---

*文档版本：2026-04-01 | 下次更新触发条件：模块接口变更 / 新增模块 / Gap 修复完成*
