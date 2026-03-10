# Alpha Pulse 实现现状与上线结论

更新时间：2026-03-10  
状态：与当前代码、测试和文档基线同步

## 1. 结论

当前系统已经完成 `Spot Analysis MVP` 主线开发。

如果目标是以下场景：

- 内部部署
- 研究型分析终端
- 演示与灰度上线
- 现货市场的观察和辅助决策

那么当前版本已经可上线。

如果目标是以下场景：

- `Futures` 数据分析
- 自动交易
- 回测平台
- 多交易所统一研究平台
- 高频订单簿逐笔重建

那么当前版本不属于“开发完成”，这些是明确延后的扩边方向。

## 2. 项目已完成项

### 2.1 总体能力

- Monorepo 基础结构
- Backend 分层架构
- Binance Spot SDK 接入
- Spot 数据链路
- REST + WebSocket 混合采集
- MySQL / Redis / Docker 基础接入
- `market-snapshot` 聚合接口
- `dev / test / prod` 运行模式

### 2.2 分析引擎

- Indicator Engine
- Order Flow Engine 真实成交优先版
- Large Trades 检测
- 微结构事件序列识别
- 更高阶微结构模式与组合评分
- Structure Engine：`internal / external` hierarchy
- Liquidity Engine：wall map / wall strength bands / wall evolution
- Signal Engine 多因子连续评分
- AI Explain Engine 基础版

### 2.3 数据与持久化

- 原始表：`kline / agg_trades / order_book_snapshots`
- 分析表：`indicators / orderflow / microstructure_events / structure / liquidity / signals`
- 扩展表：`large_trade_events / feature_snapshots`
- 顶层快照、微结构事件、大单事件、信号时间线均已进入统一分析主链路

### 2.4 前端与交互

- `/dashboard`
- `/chart`
- `/signals`
- `/market`
- `BTCUSDT / ETHUSDT` 切换
- `1m / 5m / 15m / 1h / 4h` 周期切换
- K 线、结构、流动性、信号、微结构图层
- `Microstructure Timeline`
- `AIAnalysisPanel`

### 2.5 质量与稳定性

- 后端引擎测试
- 路由集成测试
- 缓存测试与缓存失效测试
- `market-snapshot` JSON 契约测试
- 前端组件测试
- Playwright 主路径与异常态测试
- 当前代码基线已通过 `go test ./...`
- 当前代码基线已通过 `npm test`
- 当前代码基线已通过 `npm run build`

## 3. 当前可上线项

### 3.1 可上线范围

当前可直接上线的系统边界是：

- `BTCUSDT / ETHUSDT` 现货分析
- 统一 `market-snapshot` 驱动的前端工作台
- 图表、信号、订单流、结构、流动性与 AI 解释联动展示
- Redis 热点缓存与显式刷新
- `large_trade_events` / `feature_snapshots` 后台持久化

### 3.2 上线条件

建议按以下口径上线：

1. 使用 `APP_MODE=prod`
2. 使用真实 MySQL / Redis / Binance 连接
3. 保持 `ALLOW_MOCK_BINANCE_DATA=false`
4. 继续以 `market-snapshot` 作为前端主接口

### 3.3 上线预期

当前版本适合作为：

- 研究终端
- 内部分析平台
- 演示环境
- 辅助决策看盘工具

当前版本不应被描述为：

- 自动交易平台
- 完整量化研究平台
- 多资产 / 多交易所中台

## 4. 后续增强项

### 4.1 近期待增强

- 围绕 `large_trade_events` 增加查询、导出与回放链路
- 围绕 `feature_snapshots` 增加检索、导出与训练前置链路
- 在统一耗时日志之上补 metrics / monitoring
- 评估部分分析序列是否需要单独落表或冷存储

### 4.2 明确延后

- `Futures`：Funding / Open Interest / Futures 因子
- 自动下单
- 回测平台
- 用户系统 / 权限系统
- 多交易所接入
- 完整高频订单簿重放

## 5. 当前主要技术债

1. 指标、结构、流动性序列仍以实时计算和缓存为主，未全面落表
2. 盘口分析仍是快照增强，不是完整 order book replay
3. Explain Engine 仍是规则模板，不是独立模型服务
4. 当前已有统一耗时日志，但仍缺少 metrics 指标沉淀与监控面板

## 6. 后续文档口径

后续文档默认遵守以下口径：

1. 当前系统已完成 `Spot Analysis MVP`
2. `Futures` 与自动交易类能力不属于当前完成范围
3. 后续工作以“增强项”或“扩边项”表达，不再表述为主线未完成
