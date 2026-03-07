# Alpha Pulse 架构说明

更新时间：2026-03-07  
状态：与当前服务启动链路、路由和前端状态流对齐

## 1. 总体架构

Alpha Pulse 采用 Monorepo 架构：

- `backend`：Golang + Gin + GORM + MySQL + Redis
- `frontend`：Next.js App Router + TypeScript + TailwindCSS + Zustand
- `docker`：本地容器编排与镜像构建
- `docs`：产品、接口、数据库与工程文档

系统本质上是一个分析型交易终端，而不是交易执行系统。

## 2. 运行时组件

## 2.1 Backend

主要组件：

- Gin HTTP 服务
- Binance SDK 包装层
- Collector
- Repository
- Analysis Engines
- Signal Service
- Scheduler
- Redis 快照缓存

## 2.2 Frontend

主要组件：

- Next.js App Router 页面
- Zustand `marketStore`
- `MarketSnapshotLoader`
- 分析组件：`KlineChart`、`SignalCard`、`OrderFlowPanel`、`LiquidityPanel`、`AIAnalysisPanel`

## 3. Backend 分层

## 3.1 Handler

职责：

- 解析 HTTP 请求
- 读取 query 参数
- 调用 service
- 返回统一响应结构

当前处理器：

- `MarketHandler`
- `SignalHandler`

## 3.2 Service

职责：

- 编排 Collector、Repository、Engine
- 构建聚合快照
- 管理缓存
- 提供专用时间序列/事件接口

当前核心 service：

- `MarketService`
- `SignalService`

关键事实：

- `SignalService.buildMarketSnapshot` 是当前主分析装配入口
- `MarketService` 更偏专用模块接口

## 3.3 Repository

职责：

- 封装数据访问
- 提供最近窗口、时间范围、批量写入能力
- 隔离数据库细节

当前关键 repository：

- `KlineRepository`
- `AggTradeRepository`
- `OrderBookSnapshotRepository`
- `IndicatorRepository`
- `SignalRepository`
- `MicrostructureEventRepository`

## 3.4 Collector

职责：

- 封装 Binance SDK
- 输出项目内部模型
- 提供 REST 拉取和 WebSocket 采集

当前 collector：

- `BinanceCollector`
- `BinanceStreamCollector`

## 3.5 Engine

职责：

- 承担分析逻辑
- 只关心输入数据和输出结果
- 不依赖 HTTP 层

当前核心引擎：

- `Indicator Engine`
- `Order Flow Engine`
- `Structure Engine`
- `Liquidity Engine`
- `Signal Engine`
- `AI Explain Engine`

## 4. 当前主数据流

## 4.1 主聚合链路

```text
Binance SDK
  -> Collector
  -> Repository（原始市场数据 / 历史 K 线）
  -> Indicator Engine
  -> Order Flow Engine
  -> Structure Engine
  -> Liquidity Engine
  -> Signal Engine
  -> AI Explain Engine
  -> SignalService.buildMarketSnapshot
  -> Redis Cache（可选）
  -> Frontend
```

## 4.2 订单流真实优先链路

```text
aggTrade REST / WebSocket
  -> agg_trades
  -> Order Flow Engine
  -> orderflow
  -> microstructure_events
```

回退链路：

```text
aggTrade 不足或失败
  -> 历史 OHLCV
  -> 估算型 Order Flow
```

## 4.3 流动性增强链路

```text
partial depth / depth snapshot
  -> order_book_snapshots
  -> Liquidity Engine
  -> liquidity
```

回退链路：

```text
盘口数据不可用
  -> 历史 K 线窗口
  -> 基础 Liquidity 分析
```

## 5. 当前主接口架构

## 5.1 核心设计

当前系统以 `GET /api/market-snapshot` 作为前端主接口。

这样设计的原因：

- 避免前端多接口并发带来的状态撕裂
- 保证图表、面板、信号解释使用同一分析时刻的数据
- 便于后端统一做缓存

## 5.2 `market-snapshot` 当前聚合内容

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

## 5.3 顶层微结构事件设计

当前快照中有两种微结构数据：

1. `orderflow.microstructure_events`
   - 当前订单流分析窗口的摘要

2. `market_snapshot.microstructure_events`
   - 与当前图表时间范围对齐的历史序列

架构建议：

- 面板摘要可读 `orderflow.microstructure_events`
- 图表、回放、时间线优先读顶层 `microstructure_events`

## 6. 前端状态架构

当前前端主状态源：

- `marketStore`

当前主加载器：

- `MarketSnapshotLoader`

前端数据流：

```text
MarketSnapshotLoader
  -> marketStore.refreshDashboard()
  -> GET /api/market-snapshot
  -> Zustand Store
  -> Dashboard / Chart / Signals / Market 共享渲染
```

当前原则：

- 页面尽量共享同一份快照
- 不鼓励每个组件自行重复请求模块接口

## 7. 页面架构

### 7.1 Dashboard

作用：总览页。

当前模块：

- PriceTicker
- KlineChart
- SignalCard
- OrderFlowPanel
- LiquidityPanel

### 7.2 Chart

作用：图表分析页。

当前能力：

- 多根 K 线
- 指标线
- 结构点
- 流动性轨迹
- 信号点
- 微结构事件标注

### 7.3 Signals

作用：信号与解释页。

当前模块：

- SignalCard
- AIAnalysisPanel

### 7.4 Market

作用：市场总览页。

当前模块：

- 市场概览
- 关键价位板
- 信号带
- 订单流摘要
- 流动性摘要

## 8. 图表架构

当前 `KlineChart` 不是第三方图表库，而是项目内 SVG 图层渲染。

优点：

- 可精确控制结构点、流动性区、信号点和微结构事件的叠加
- 便于跟随项目分析模型一起演进

当前已实现图层：

- 蜡烛图
- EMA20 / EMA50 / VWAP / Bollinger Bands
- 结构点：`HH / HL / LH / LL / BOS / CHOCH`
- 流动性轨迹
- 历史信号点
- 当前信号水平线
- 微结构事件：`ABS / ICE / AGR`

## 9. 缓存架构

当前 Redis 只承担一类核心职责：

- `market-snapshot` 聚合结果缓存

缓存键维度：

- `symbol`
- `interval`
- `limit`

特点：

- TTL 可配置
- Redis 不可用时自动退化为无缓存模式
- 不阻塞服务启动

当前未做：

- 多层缓存
- 实时流中间缓冲
- 热点 symbol 细粒度模块缓存

## 10. 启动与装配

服务启动阶段会完成：

1. 读取环境变量
2. 初始化 MySQL
3. 初始化 Redis
4. 初始化 Binance SDK 包装层
5. 初始化 Repository
6. 初始化各 Engine
7. 初始化 MarketService / SignalService
8. 初始化 Router
9. 启动 Scheduler
10. 启动实时 WebSocket Collector

## 11. 调度与实时流

当前运行中存在两类刷新机制：

### 11.1 Scheduler

作用：

- 定时拉取/预热基础分析数据
- 保证数据链路有基础热启动结果

### 11.2 WebSocket Collector

作用：

- 订阅 `aggTrade`
- 订阅 `partial depth`
- 为订单流和流动性分析提供更实时的输入

当前说明：

- WebSocket 已接通，但还不是完整高频回放架构

## 12. 当前已解决的关键架构问题

- 已从“多接口拼装前端状态”收敛到 `market-snapshot`
- 已从“纯 K 线估算订单流”升级到“真实 aggTrade 优先”
- 已从“纯静态价位展示”升级到“结构/流动性/信号/微结构叠加图层”
- 已把微结构事件序列持久化并纳入聚合快照

## 13. 当前仍保留的架构取舍

### 13.1 指标/结构/流动性序列不单独落表

当前这些序列是按历史窗口滚动生成，而不是预计算后持久化。

原因：

- 当前阶段更重视分析可用性和迭代效率
- 降低冗余写入

### 13.2 订单簿不是全量增量重建

当前盘口能力基于：

- 深度快照
- partial depth
- 快照增强分析

尚不是：

- 全量 order book replay 引擎

### 13.3 AI Explain 不是外部模型服务

当前 Explain Engine 是规则和模板驱动，而不是调用外部 LLM 生成。

## 14. 推荐后续架构演进

建议按以下顺序继续演进：

1. 扩展微结构图层和历史事件查询
2. 扩展 Redis 到更多热点接口和模块缓存
3. 若引入 Futures，再新增独立数据源与因子，不要污染 Spot 主链路
4. 若引入回放/训练，再考虑独立 `feature_snapshots` 或 `signal_runs`

## 15. 对后续开发的约束

1. 前端主页面默认继续以 `market-snapshot` 为主接口
2. 时间轴相关新事件必须有明确时间字段，并能映射到 candle
3. 不要重新拆出第二套 Signal 评分体系
4. 新增数据域时优先先明确原始数据层，再做分析结果层
5. 不要为了局部 UI 需求反向破坏当前服务装配方式
