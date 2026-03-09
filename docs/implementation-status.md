# Alpha Pulse 实现现状盘点

更新时间：2026-03-07  
状态：与当前代码、测试和文档基线同步

## 1. 结论

当前项目已经不是“项目骨架”，而是一个可运行的分析型交易终端 MVP。

按当前能力评估：

- 按“可演示 MVP”口径：约 `80% ~ 85%`
- 按“完整交易研究终端”口径：约 `65% ~ 70%`

## 2. 当前总体完成度

### 已完成

- Monorepo 基础结构
- Backend 分层架构
- Binance SDK 接入
- Spot 数据链路
- REST + WebSocket 混合采集
- MySQL / Redis / Docker 基础接入
- Indicator Engine
- Order Flow Engine 真实成交优先版本
- Structure Engine 结构事件版
- Liquidity Engine 盘口增强版
- Signal Engine 多因子评分模型
- AI Explain Engine
- `GET /api/market-snapshot`
- Dashboard / Chart / Signals / Market 页面
- 结构 / 流动性 / 信号 / 微结构图层标注
- 组件测试与主路径 E2E

### 未完成但明确不属于当前主线

- Futures Funding / Open Interest
- 自动交易
- 回测系统
- 多交易所接入
- 完整高频订单簿重放

## 3. 模块对照

| 模块 | 当前状态 | 说明 |
| --- | --- | --- |
| Binance Spot 数据接入 | 已完成 | 采用 `go-binance/v2`，已支持 REST 与基础 WebSocket |
| Binance Futures 数据接入 | 未开始 | 不在当前主线范围 |
| 币种切换 | 已完成 | 当前支持 `BTCUSDT`、`ETHUSDT` |
| 周期切换 | 已完成 | 当前支持 `1m / 5m / 15m / 1h / 4h` |
| Indicator Engine | 已完成 | 最新值与序列均已可用 |
| Order Flow Engine | 已完成增强版 | 真实 `aggTrade` 优先，支持 large trades 与微结构事件 |
| 微结构事件持久化 | 已完成 | `microstructure_events` 已落库并可查询 |
| Structure Engine | 已完成增强版 | 支持 internal / external swing hierarchy、HH/HL/LH/LL/BOS/CHOCH 与序列接口 |
| Liquidity Engine | 已完成增强版 | 支持盘口失衡、equal high/low、stop clusters、细粒度 wall map、wall strength bands 与跨周期 wall 演化 |
| Signal Engine | 已完成增强版 | 7 因子连续评分模型 |
| AI Explain Engine | 已完成基础版 | 基于规则模板输出中文解释 |
| 聚合快照接口 | 已完成 | 当前前端主接口 |
| Dashboard 页面 | 已完成 | 主分析工作台 |
| Chart 页面 | 已完成 | 图表分析与多图层标注 |
| Signals 页面 | 已完成 | SignalCard + AI Analysis |
| Market 页面 | 已完成基础版 | 市场概览、关键价位、信号带 |
| Redis 缓存 | 已完成增强版 | 已覆盖 `market-snapshot / signal-timeline / indicator-series / liquidity-series`，并支持 symbol 级全周期失效与 `refresh=1` 绕过缓存 |
| Runtime Modes | 已完成基础版 | 已区分 `dev / test / prod`，并接入 Gin mode、自动迁移、后台任务和 mock fallback 开关 |
| 后端测试 | 已完成基础版 | 引擎测试、缓存测试、路由测试已具备 |
| 前端组件测试 | 已完成基础版 | 关键组件已覆盖 |
| 前端 E2E | 已完成基础版 | 主路径和异常态已覆盖 |

## 4. 后端完成情况

### 4.1 数据层

已落地原始数据表：

- `kline`
- `agg_trades`
- `order_book_snapshots`

已落地分析结果表：

- `indicators`
- `orderflow`
- `microstructure_events`
- `structure`
- `liquidity`
- `signals`

### 4.2 服务层

已完成主要服务：

- `MarketService`
- `SignalService`

其中：

- `SignalService.buildMarketSnapshot` 已成为主装配入口
- `market-snapshot` 已集成指标序列、结构序列、流动性序列、信号时间线和微结构历史序列

### 4.3 路由层

当前对外 API 已覆盖：

- `price`
- `kline`
- `indicators`
- `indicator-series`
- `orderflow`
- `microstructure-events`
- `structure`
- `market-structure-events`
- `market-structure-series`
- `liquidity`
- `liquidity-map`
- `liquidity-series`
- `signal`
- `signal-timeline`
- `market-snapshot`

## 5. 前端完成情况

### 5.1 页面

已完成页面：

- `/dashboard`
- `/chart`
- `/signals`
- `/market`

### 5.2 图表

`KlineChart` 当前已支持：

- 多根蜡烛图
- 指标线
- 结构点标注
- 流动性轨迹
- 历史信号点
- Entry / Target / Stop 水平线
- 微结构事件标注：`ABS / ICE / AGR / FAU / FAH / FAL / OBM / OBL / OBA / MCF`
- 配置驱动的次级微结构图层开关：`initiative_shift / large_trade_cluster / failed_auction / order_book_migration / microstructure_confluence`
- 微结构事件 tooltip：`type / bias / score / strength / detail`

### 5.3 面板

已完成组件：

- `PriceTicker`
- `SignalCard`
- `OrderFlowPanel`
- `LiquidityPanel`
  - 支持 bid / ask 分层 wall map、wall strength bands 与跨周期 wall 演化展示
- `KlineChart`
  - 支持 structure hierarchy：主层级与 internal support / resistance 同图表达
- `AIAnalysisPanel`
- `MarketOverviewBoard`
- `MarketLevelsBoard`
- `SignalTape`
- `MicrostructureTimeline`
  - 支持事件家族过滤、摘要统计和高阶事件高亮

## 6. 测试完成情况

### 6.1 Backend

已覆盖：

- Indicator Engine
- Order Flow Engine
- Structure Engine
- Liquidity Engine
- Signal Engine
- Market Snapshot 路由
- Snapshot Cache 行为
- Analysis View Cache 行为
- 缓存失效策略
- `market-snapshot` JSON 契约
- 统一耗时日志

### 6.2 Frontend

组件测试已覆盖：

- `PriceTicker`
- `OrderFlowPanel`
- `SignalCard`
- `AIAnalysisPanel`
- `KlineChart`
- `MicrostructureTimeline`

E2E 已覆盖：

- Dashboard 主路径
- Signals 页面
- Market 页面
- 接口失败
- 弱网加载
- 切币
- 手动刷新

### 6.3 当前验证结果

当前代码基线下已验证通过：

- `go test ./...`
- `npm test`
- `npm run build`

## 7. 当前最重要的能力边界

### 当前已具备

- 真实成交优先的订单流
- 盘口增强流动性分析
- 多因子信号解释
- 微结构事件持久化与图表展示
- 微结构时间线卡片、事件过滤与图表 tooltip
- 多端点 Redis 缓存、symbol 级失效与显式刷新
- `dev / test / prod` 运行模式与启动期开关
- Liquidity Engine 细粒度 wall map / wall strength bands / 跨周期 wall 演化
- Structure Engine internal / external swing hierarchy
- 高阶微结构模式：连续吸收、失败拍卖扩展型、挂单迁移分层、失败拍卖陷阱反转、流动性阶梯突破、组合事件评分
- Collector / fallback / engine stage 统一耗时日志
- 页面级统一快照驱动

### 当前仍缺失

- Futures 数据域
- 大单事件独立持久化表
- 更丰富的高阶微结构模式库
- 更细粒度的可观测性与缓存分层

## 8. 当前主要风险与技术债

1. 指标/结构/流动性序列没有单独落表，重算成本仍在服务端承担
2. 盘口仍是快照增强分析，不是完整 order book replay
3. Explain Engine 仍是规则模板，不是独立模型服务
4. 当前已有统一耗时日志，但仍缺少 metrics 指标沉淀与监控面板

## 9. 推荐下一步

当前最合理的下一阶段方向：

1. 继续扩展更高阶微结构模式库
2. 按需继续扩展更高阶微结构模式库
3. 为日志继续接入 metrics/monitoring，并考虑 `feature_snapshots`
4. 如果要继续扩数据域，再单独开 Futures 方向，不要和当前 Spot 主链路混改
