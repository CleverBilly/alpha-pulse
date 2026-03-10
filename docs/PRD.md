# Alpha Pulse PRD

版本：V2.0 Source of Truth  
状态：Active  
更新日期：2026-03-10

## 0. 文档定位

本文件是 `alpha-pulse` 当前阶段的产品与开发源文档，供以下角色直接使用：

- 产品设计
- 全栈开发
- 架构设计
- 测试编写
- 其他 AI coding agent / 自动化开发工具

目标不是写“概念性 PRD”，而是给出一份足够接近真实代码与真实范围的可执行规范。

如果其他旧文档与本 PRD 冲突，以本 PRD 为准。

如果本 PRD 与代码细节出现轻微偏差，以当前仓库中的接口字段名、模型字段名、路由和测试为最终实现参考。

## 1. 项目概述

### 1.1 项目名称

`alpha-pulse`

### 1.2 产品一句话定义

Alpha Pulse 是一个面向 `BTCUSDT` 和 `ETHUSDT` 的 AI Crypto Trading Dashboard，用统一界面整合价格、K 线、技术指标、订单流、市场结构、流动性、交易信号和 AI 解释。

### 1.3 核心价值

Alpha Pulse 不追求“列出更多指标”，而是解决以下问题：

- 交易者需要在多个平台来回切换才能看到完整上下文
- 单一指标不能解释为什么要买或卖
- 结构、流动性、订单流、信号通常分散在不同工具中
- 现有信号系统经常只有结论，没有可解释原因

Alpha Pulse 的核心价值是把多模块分析结果汇总成一条可解释的交易分析链路：

`市场数据 -> 指标/订单流/结构/流动性 -> 多因子信号 -> AI 解释 -> 前端可视化`

## 2. 当前阶段定义

### 2.1 当前阶段结论

当前项目可视为：

- 已完成可运行的分析型 MVP
- 已建立完整的后端分析主链路
- 已建立统一聚合接口 `GET /api/market-snapshot`
- 已具备前端多页面展示能力
- 已具备基础测试体系

### 2.2 当前里程碑命名

建议将当前代码状态定义为：

`V1.2 Analysis MVP`

说明：

- 不再使用最初“目录骨架”视角定义版本
- 当前已明显超出最初 V1 骨架阶段
- 但尚未达到 Futures、多交易所、自动交易、回测平台等更高阶段

### 2.3 当前里程碑范围

当前版本已包含：

- Spot 市场数据接入
- Binance SDK 接入
- REST + WebSocket 混合采集
- MySQL 持久化
- Redis 快照缓存
- Dashboard / Chart / Signals / Market 页面
- K 线图、结构、流动性、信号和微结构标注
- 多因子 Signal Engine
- AI Explain Engine
- 前端组件测试与 Playwright E2E

## 3. 产品目标与非目标

### 3.1 当前阶段目标

1. 支持 `BTCUSDT`、`ETHUSDT`
2. 支持 `1m / 5m / 15m / 1h / 4h`
3. 提供一个统一的市场分析快照接口给前端使用
4. 形成真实可用的交易分析主链路
5. 输出可解释的交易信号，而不是只输出 `BUY/SELL`
6. 在图表上直接展示结构点、流动性区、历史信号和微结构事件

### 3.2 当前阶段不做

以下能力明确不属于当前版本必交付项：

- 自动下单
- 账户资产管理
- 策略回测
- 交易绩效归因
- 链上数据分析
- 多交易所接入
- Futures Funding / Open Interest 完整接入
- 高频订单簿重建与逐笔回放系统
- 机器学习训练平台

## 4. 目标用户

### 4.1 短线交易者

关注：

- `1m / 5m / 15m`
- 结构突破、扫流动性、订单流变化
- 进场点、止损位、目标位

### 4.2 波段交易者

关注：

- `15m / 1h / 4h`
- 趋势强弱
- EMA、VWAP、Bollinger、结构共振
- 风险收益比与置信度

### 4.3 研究型用户 / 量化观察者

关注：

- 指标序列
- 流动性轨迹
- 结构事件流
- 微结构事件时间序列
- 聚合接口的稳定契约

## 5. 支持范围

### 5.1 交易对

当前正式支持：

- `BTCUSDT`
- `ETHUSDT`

### 5.2 周期

当前正式支持：

- `1m`
- `5m`
- `15m`
- `1h`
- `4h`

### 5.3 数据源

当前主数据源：

- `github.com/adshao/go-binance/v2`
- Spot 公共市场数据

当前采集方式：

- REST：`ticker / klines / aggTrades / depth snapshot`
- WebSocket：`aggTrade / partial depth`

### 5.4 回退策略

为保证本地开发、测试和离线场景可运行，当前代码允许在 Binance SDK 请求失败时回退到 mock 数据。

该回退策略的目标是：

- 保证服务不因外部网络失败而完全不可用
- 保证测试链路稳定
- 保证分析引擎在无外网环境下仍可验证

注意：

- 该回退只用于开发稳定性，不应被误解为真实生产信号质量保证
- 如果未来进入真实交易环境，必须进一步区分 `dev / test / prod`

## 6. Monorepo 结构

```text
alpha-pulse
├── backend
├── frontend
├── docker
├── scripts
└── docs
```

### 6.1 backend

- Golang
- Gin
- GORM
- MySQL
- Redis

### 6.2 frontend

- Next.js App Router
- TypeScript
- TailwindCSS
- Zustand

### 6.3 docker

- `docker-compose.yml`
- `backend.Dockerfile`
- `frontend.Dockerfile`

### 6.4 docs

- `PRD.md`
- `api.md`
- `architecture.md`
- `database.md`
- `implementation-status.md`
- `task-list.md`

## 7. 技术架构

## 7.1 Backend 架构

当前采用分层架构：

- `Handler`
- `Service`
- `Repository`
- `Models`
- `Collector`
- `Engine`
- `Scheduler`
- `pkg/database`
- `pkg/binance`

### 7.1.1 Handler

职责：

- 接收 HTTP 请求
- 解析 query 参数
- 调用 service
- 返回统一 JSON 响应结构

### 7.1.2 Service

职责：

- 编排 collector、repository、engine
- 统一构建聚合结果
- 管理缓存逻辑
- 处理主要业务流，如 `market-snapshot`

### 7.1.3 Repository

职责：

- 封装 GORM 查询与写入
- 隔离数据库访问细节
- 支持最近窗口、时间范围、批量 upsert 等行为

### 7.1.4 Collector

职责：

- 封装 Binance SDK 调用
- 转换成项目内部模型
- 提供 REST 拉取和 WebSocket 采集

### 7.1.5 Engine

职责：

- 承担具体分析逻辑
- 不直接依赖 HTTP 层
- 输入历史数据，输出分析结果

## 7.2 Frontend 架构

当前前端数据流为：

`MarketSnapshotLoader -> Zustand marketStore -> /api/market-snapshot -> 页面组件共享渲染`

前端核心原则：

- `market-snapshot` 是主接口
- 页面尽量共享同一份快照状态
- 组件不要各自重复拉多个接口，除非是专用扩展视图

## 7.3 当前主链路

```text
Binance SDK
  -> Collector
  -> Repository（原始数据 / 历史数据）
  -> Indicator / OrderFlow / Structure / Liquidity Engines
  -> Signal Engine
  -> AI Explain Engine
  -> SignalService.buildMarketSnapshot
  -> Redis Cache（可选）
  -> Frontend
```

## 8. 核心模块要求

## 8.1 Market Collector

职责：

- 拉取实时价格
- 拉取历史 K 线
- 拉取聚合成交
- 拉取盘口快照
- 建立 `aggTrade / partial depth` WebSocket 订阅

当前状态：已完成基础版。

输入：

- `symbol`
- `interval`
- `limit`

输出：

- `price`
- `klines`
- `agg trades`
- `order book snapshot`

## 8.2 Indicator Engine

当前指标范围：

- `RSI`
- `MACD`
- `MACD Signal`
- `MACD Histogram`
- `EMA20`
- `EMA50`
- `ATR`
- `Bollinger Upper`
- `Bollinger Middle`
- `Bollinger Lower`
- `VWAP`

要求：

- 支持基于历史 K 线计算
- 返回最新指标值
- 支持返回指标时间序列
- 时间序列与最新值必须口径一致

## 8.3 Order Flow Engine

当前输出：

- `buy_volume`
- `sell_volume`
- `delta`
- `cvd`
- `buy_large_trade_count`
- `sell_large_trade_count`
- `buy_large_trade_notional`
- `sell_large_trade_notional`
- `large_trade_delta`
- `absorption_bias`
- `absorption_strength`
- `iceberg_bias`
- `iceberg_strength`
- `large_trades[]`
- `microstructure_events[]`
- `data_source`

当前实现策略：

- 优先使用真实 `aggTrade`
- 样本不足时自动回退到 OHLCV 估算

当前微结构事件类型：

- `absorption`
- `iceberg`
- `aggression_burst`
- `initiative_shift`
- `large_trade_cluster`

说明：

- 当前快照展示和历史持久化都已接通
- `microstructure_events` 当前已可持久化并可做历史查询

## 8.4 Market Structure Engine

当前输出：

- `trend`
- `primary_tier`
- `support`
- `resistance`
- `internal_support`
- `internal_resistance`
- `external_support`
- `external_resistance`
- `bos`
- `choch`
- `events[]`

当前结构事件标签：

- `HH`
- `HL`
- `LH`
- `LL`
- `BOS`
- `CHOCH`

要求：

- 支持 swing point 识别
- 支持 internal / external swing hierarchy
- 支持结构时间序列
- 前端图表必须可直接消费 `events[]`

## 8.5 Liquidity Engine

当前输出：

- `buy_liquidity`
- `sell_liquidity`
- `sweep_type`
- `order_book_imbalance`
- `data_source`
- `equal_high`
- `equal_low`
- `stop_clusters[]`
- `wall_levels[]`
- `wall_strength_bands[]`
- `wall_evolution[]`

要求：

- 优先使用盘口快照增强分析
- 无盘口数据时允许回退 K 线逻辑
- 支持流动性时间序列

## 8.6 Signal Engine

Signal Engine 当前为多因子连续评分模型。

### 8.6.1 分数范围

- 总分范围：`-100 ~ +100`

### 8.6.2 动作阈值

当前阈值：

- `score >= 35` -> `BUY`
- `score <= -35` -> `SELL`
- 其余 -> `NEUTRAL`

### 8.6.3 当前输出字段

- `signal`
- `score`
- `confidence`
- `entry_price`
- `stop_loss`
- `target_price`
- `risk_reward`
- `trend_bias`
- `factors[]`
- `explain`

### 8.6.4 当前评分因子

当前因子共 7 类：

- `Trend`
- `Momentum`
- `Order Flow`
- `Structure`
- `Liquidity`
- `Microstructure`
- `Volatility`

### 8.6.5 重要约束

后续开发必须遵守：

- 不要重新引入旧版 `-12 ~ +12` 离散模型
- 不要再定义第二套动作阈值体系
- 如果需要 `Strong Buy / Strong Sell`，只能基于当前连续分数做映射层

## 8.7 AI Explain Engine

当前作用：

- 读取 Signal Engine 输出
- 基于 `factors[]`、趋势、风险收益比等生成中文解释
- 供 `SignalCard` 与 `AI Analysis` 面板直接展示

当前特点：

- 以规则模板为主
- 解释文本不是独立外部大模型服务返回

## 9. 数据模型与数据库

当前数据库默认：

- MySQL
- DB 名称：`alpha_pulse`

## 9.1 原始数据层

### 9.1.1 `kline`

用途：

- 历史 K 线
- 所有分析引擎基础输入

关键字段：

- `symbol`
- `interval_type`
- `open_price`
- `high_price`
- `low_price`
- `close_price`
- `volume`
- `open_time`

关键约束：

- `symbol + interval_type + open_time` 唯一索引

### 9.1.2 `agg_trades`

用途：

- 逐笔/聚合成交历史
- 订单流与微结构分析输入

关键字段：

- `symbol`
- `agg_trade_id`
- `price`
- `quantity`
- `quote_quantity`
- `trade_time`
- `is_buyer_maker`

关键约束：

- `symbol + agg_trade_id` 唯一索引

### 9.1.3 `order_book_snapshots`

用途：

- 盘口快照历史
- 流动性与盘口失衡分析输入

关键字段：

- `symbol`
- `last_update_id`
- `depth_level`
- `bids_json`
- `asks_json`
- `best_bid_price`
- `best_ask_price`
- `spread`
- `event_time`

关键约束：

- `symbol + last_update_id` 唯一索引

## 9.2 分析结果层

### 9.2.1 `indicators`

关键字段：

- `rsi`
- `macd`
- `macd_signal`
- `macd_histogram`
- `ema20`
- `ema50`
- `atr`
- `bollinger_upper`
- `bollinger_middle`
- `bollinger_lower`
- `vwap`

### 9.2.2 `orderflow`

关键字段：

- `symbol`
- `interval_type`
- `open_time`
- `buy_volume`
- `sell_volume`
- `delta`
- `cvd`
- `buy_large_trade_count`
- `sell_large_trade_count`
- `buy_large_trade_notional`
- `sell_large_trade_notional`
- `large_trade_delta`
- `absorption_bias`
- `absorption_strength`
- `iceberg_bias`
- `iceberg_strength`
- `data_source`
- `large_trades[]`

说明：

- `large_trades[]` 当前元素包含 `agg_trade_id / side / price / quantity / notional / trade_time`
- 大单摘要会同步持久化到 `large_trade_events`

### 9.2.3 `microstructure_events`

关键字段：

- `orderflow_id`
- `symbol`
- `interval_type`
- `open_time`
- `event_type`
- `bias`
- `score`
- `strength`
- `price`
- `trade_time`
- `detail`

用途：

- 保存微结构事件历史
- 支持按 `symbol + interval + trade_time` 查询
- 支持快照图表标注和未来事件回放

### 9.2.4 `structure`

关键字段：

- `trend`
- `support`
- `resistance`
- `bos`
- `choch`

### 9.2.5 `liquidity`

关键字段：

- `buy_liquidity`
- `sell_liquidity`
- `sweep_type`
- `order_book_imbalance`
- `equal_high`
- `equal_low`
- `data_source`
- `wall_levels[]`
- `wall_strength_bands[]`
- `wall_evolution[]`

### 9.2.6 `signals`

关键字段：

- `symbol`
- `interval_type`
- `open_time`
- `signal`
- `score`
- `confidence`
- `entry_price`
- `stop_loss`
- `target_price`

### 9.2.7 `large_trade_events`

关键字段：

- `orderflow_id`
- `symbol`
- `agg_trade_id`
- `interval_type`
- `open_time`
- `side`
- `price`
- `quantity`
- `notional`
- `trade_time`

用途：

- 保存真实大单事件历史
- 支持未来大单回放、聚类分析与时间轴查询

### 9.2.8 `feature_snapshots`

关键字段：

- `symbol`
- `interval_type`
- `open_time`
- `snapshot_source`
- `feature_version`
- `price`
- `signal_action`
- `signal_score`
- `signal_confidence`
- `snapshot_json`

用途：

- 保存 `market-snapshot` 的完整离线特征快照
- 支持审计、训练前置与回测前上下文回放

说明：

以下字段当前不落库，仅用于接口返回：

- `explain`
- `factors[]`
- `risk_reward`
- `trend_bias`

## 10. API 设计

统一响应结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

Base URL：

`http://localhost:8080/api`

## 10.1 当前接口列表

### 基础接口

- `GET /api/price`
- `GET /api/kline`
- `GET /api/indicators`
- `GET /api/indicator-series`
- `GET /api/orderflow`
- `GET /api/microstructure-events`
- `GET /api/structure`
- `GET /api/market-structure-events`
- `GET /api/market-structure-series`
- `GET /api/liquidity`
- `GET /api/liquidity-map`
- `GET /api/liquidity-series`
- `GET /api/signal`
- `GET /api/signal-timeline`

### 主聚合接口

- `GET /api/market-snapshot`

## 10.2 `market-snapshot` 设计原则

`market-snapshot` 是当前前端主接口。

设计原则：

- 前端主页面优先只请求这一条接口
- 聚合当前图表与分析面板所需的所有核心数据
- 避免前端重复请求多个模块接口造成状态不一致

### 10.2.1 查询参数

- `symbol`，默认 `BTCUSDT`
- `interval`，默认 `1m`
- `limit`，默认 `48`

### 10.2.2 当前返回字段

- `price`
- `klines`
- `indicator`
- `indicator_series`
- `orderflow`
- `microstructure_events`
- `structure`
- `structure_series`
- `liquidity`
- `liquidity_series`
- `signal`
- `signal_timeline`

### 10.2.3 顶层 `microstructure_events` 的设计原因

当前快照中保留两层微结构信息：

1. `orderflow.microstructure_events`
   - 表示当前订单流分析窗口识别出的结果
   - 更偏“当前窗口分析摘要”

2. `market_snapshot.microstructure_events`
   - 表示与当前可见图表时间范围对齐的历史事件序列
   - 用于图表标注、历史回看、面板展示

后续开发中，图表与时间线组件应优先使用顶层 `microstructure_events`。

## 11. 前端页面需求

## 11.1 页面列表

当前页面：

- `/dashboard`
- `/chart`
- `/signals`
- `/market`

## 11.2 Dashboard 页面

必须展示：

- `PriceTicker`
- `KlineChart`
- `SignalCard`
- `OrderFlowPanel`
- `LiquidityPanel`

必须支持：

- 切币 `BTCUSDT / ETHUSDT`
- 切周期 `1m / 5m / 15m / 1h / 4h`
- 手动刷新
- 自动刷新
- 异常态展示

## 11.3 Chart 页面

必须展示：

- 多根 K 线图
- 指标线
- 结构点
- 流动性区
- 信号点
- 微结构事件点

当前图表标注要求：

- 结构点：`HH / HL / LH / LL / BOS / CHOCH`
- 信号点：`BUY / SELL`
- 微结构点：`ABS / ICE / AGR`
- 价位线：`Support / Resistance / Buy Liquidity / Sell Liquidity / Equal High / Equal Low / Stop Clusters / Entry / Target / Stop`

## 11.4 Signals 页面

必须展示：

- `SignalCard`
- `AIAnalysisPanel`

AI Analysis 必须包含：

- `Decision Memo`
- `Bullish Drivers`
- `Risk Factors`
- `Execution Plan`
- `Recent Signal Tape`
- `Microstructure Tape`

## 11.5 Market 页面

必须展示：

- 市场概览
- 关键价位
- Signal Tape
- 订单流摘要
- 流动性摘要

## 12. 图表要求

图表是当前产品的核心呈现层，不允许退化为“纯数值面板”。

当前图表必须支持：

- 多根蜡烛图
- `EMA20 / EMA50 / VWAP / Bollinger Bands`
- 结构事件标注
- 流动性轨迹
- 历史信号点位
- 当前信号水平线
- 微结构事件标注

### 12.1 微结构图层规范

当前只对以下事件做图层标注：

- `absorption` -> `ABS`
- `iceberg` -> `ICE`
- `aggression_burst` -> `AGR`

要求：

- 使用 `trade_time` 映射到对应 candle
- 同一 candle 上多个事件需要自动错位避免完全重叠
- 图层必须来自顶层 `microstructure_events` 历史序列

## 13. 缓存与刷新策略

## 13.1 Redis 缓存

当前 Redis 已用于缓存：

- `GET /api/market-snapshot`

缓存键维度：

- `symbol`
- `interval`
- `limit`

TTL：

- 环境变量 `MARKET_SNAPSHOT_CACHE_TTL`
- 当前默认 `5s`

要求：

- Redis 不可用时服务不能崩溃
- 应自动退化为无缓存模式

## 13.2 前端刷新策略

当前前端刷新方式：

- 页面加载自动拉一次 `market-snapshot`
- 定时轮询刷新
- 用户可手动刷新
- 切币 / 切周期后立即拉取新快照

## 14. 调度器与实时流

当前后端包括：

- 启动时 WebSocket 订阅 `aggTrade / partial depth`
- 调度器刷新核心分析数据

要求：

- 启动后能持续维持 BTC / ETH 的基础实时输入流
- 定时任务不要与前端快照请求相互打断
- 调度器失败不应导致服务主链路不可用

## 15. 部署与环境变量

## 15.1 Backend 环境变量

- `APP_PORT`
- `MYSQL_DSN`
- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `MARKET_SNAPSHOT_CACHE_TTL`
- `BINANCE_API_KEY`
- `BINANCE_SECRET_KEY`

## 15.2 Frontend 环境变量

- `NEXT_PUBLIC_API_BASE_URL`

## 15.3 Docker 部署要求

当前 `docker-compose` 必须包含：

- `mysql`
- `redis`
- `backend`
- `frontend`

## 16. 测试与质量门槛

## 16.1 Backend

当前必须具备：

- 引擎单元测试
- 聚合接口测试
- 缓存行为测试

重点覆盖：

- Indicator Engine
- Order Flow Engine
- Structure Engine
- Liquidity Engine
- Signal Engine
- `GET /api/market-snapshot`

## 16.2 Frontend

当前必须具备：

- 组件测试
- E2E 测试
- 生产构建通过

当前已覆盖的组件测试：

- `PriceTicker`
- `OrderFlowPanel`
- `SignalCard`
- `AIAnalysisPanel`
- `KlineChart`

当前已覆盖的 E2E 场景：

- Dashboard 主路径
- Signals 页面
- Market 页面
- 接口失败
- 弱网加载
- 切币
- 手动刷新

## 16.3 合格标准

任何后续提交，如果涉及主链路，必须至少满足：

1. 后端 `go test ./...` 通过
2. 前端 `npm test` 通过
3. 前端 `npm run test:e2e` 通过
4. 前端 `npm run build` 通过

## 17. 当前实现状态总结

当前项目的 `Spot Analysis MVP` 主线开发已经完成。

当前已完成的关键能力：

- Spot SDK 接入
- 聚合快照主链路
- 多因子信号系统
- 订单流大单与微结构事件
- 结构事件与流动性时间序列
- 高阶微结构模式与组合事件评分
- `large_trade_events` 与 `feature_snapshots` 持久化
- Market 页面与独立 AI Analysis 面板
- Redis 快照缓存与显式刷新
- 基础自动化测试

当前版本可作为研究型分析终端上线，但不包含 `Futures`、自动下单、回测平台和多交易所能力。

## 18. 后续路线图

## 18.1 P1 继续增强

- `initiative_shift / large_trade_cluster` 图表标注层
- 微结构事件 tooltip / hover 详情
- 更完整的结构/流动性历史序列接口
- Redis 扩展到更多热点接口
- 更细的页面级 E2E

## 18.2 P2 扩展方向

- Futures `Funding Rate`
- Futures `Open Interest`
- 更多微结构模式识别
- 更强的 Explain Engine 模板系统
- 更细粒度的缓存和可观测性

## 18.3 明确暂缓

以下方向暂不作为当前版本执行目标：

- 自动交易
- 回测系统
- 多交易所适配层
- 策略市场 / 用户系统 / 账户系统

## 19. 对其他 AI 工具的开发约束

如果后续由其他 AI 工具继续开发，必须遵守以下规则：

1. 默认以 `market-snapshot` 为前端主接口，不要随意拆回多接口并发拉取
2. 默认以 Spot 分析为当前里程碑，不要在未明确需求下引入 Futures 大改
3. Signal Engine 继续沿用当前 `-100 ~ +100` 连续评分模型
4. 需要扩表时，可以新增字段或新表，不要为了兼容旧 PRD 而强行复用错误字段名
5. 图表时间轴相关的新事件，优先用时间映射到 candle，不要简单挂到最新一根
6. 如果需要新增新型事件序列，优先考虑：
   - 是否需要持久化
   - 是否需要独立查询接口
   - 是否需要并入 `market-snapshot`
7. 当前页面和图表已经存在明确视觉结构，新增组件应保持现有信息密度和布局意图，不要退化成普通后台表格页

## 20. 开发完成定义（Definition of Done）

一个新功能只有在同时满足以下条件时才算完成：

1. 代码已落地，不只是文档或 TODO
2. 后端/前端契约已同步
3. 至少有对应测试覆盖主路径
4. `market-snapshot` 或相关专用接口已纳入真实链路
5. 前端已有对应展示或消费逻辑
6. 文档已同步更新

## 21. 推荐阅读入口

对于新的开发者或 AI 工具，建议阅读顺序：

1. 本文档 `docs/PRD.md`
2. `docs/api.md`
3. `docs/architecture.md`
4. `docs/database.md`
5. `docs/implementation-status.md`
6. `docs/task-list.md`

同时重点查看以下代码入口：

- `backend/cmd/server/main.go`
- `backend/router/router.go`
- `backend/internal/service/signal_service.go`
- `backend/internal/service/market_service.go`
- `backend/internal/signal/signal_engine.go`
- `backend/internal/orderflow/orderflow_engine.go`
- `frontend/store/marketStore.ts`
- `frontend/components/chart/KlineChart.tsx`
- `frontend/components/signal/SignalCard.tsx`
- `frontend/components/analysis/AIAnalysisPanel.tsx`
