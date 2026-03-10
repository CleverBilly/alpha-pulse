# Alpha Pulse 项目收尾清单

更新时间：2026-03-10  
状态：与当前代码、测试和文档基线同步

状态定义：

- `[x]` 已完成
- `[-]` 主线已完成，但仍建议增强
- `[ ]` 明确延后或不在当前范围内

## 1. 结论

- [x] 当前 `Spot Analysis MVP` 主线开发已完成
- [x] 当前系统可作为现货分析终端上线，用于内部演示、研究和日常观察
- [x] 当前版本不包含 `Futures`、自动下单、回测平台、多交易所和完整高频订单簿重建
- [x] 当前剩余工作主要是运营化和研究增强，不阻断上线

## 2. 项目已完成项

### 2.1 范围与文档

- [x] 项目范围已收敛为 `Spot Analysis MVP`
- [x] Signal Engine 已收敛为 `-100 ~ +100` 连续评分模型
- [x] `PRD / API / Database / Architecture / Status / Task` 文档已建立并同步

### 2.2 后端与数据链路

- [x] Monorepo、Backend 分层架构、MySQL、Redis、Docker 已接通
- [x] Binance Spot SDK、REST + WebSocket 混合采集已接通
- [x] `market-snapshot` 聚合接口已成为前端主接口
- [x] `dev / test / prod` 运行模式已区分
- [x] `prod` 模式默认禁用 mock 市场数据回退

### 2.3 数据层与持久化

- [x] 原始市场数据表：`kline / agg_trades / order_book_snapshots`
- [x] 分析结果表：`indicators / orderflow / microstructure_events / structure / liquidity / signals`
- [x] 扩展数据表：`large_trade_events / feature_snapshots`
- [x] `large_trade_events` 已保留 `orderflow` 语义上下文，适合历史回放与聚类分析
- [x] `feature_snapshots` 已保留聚合快照 JSON 与关键摘要字段，适合离线审计与训练前置

### 2.4 引擎与分析能力

- [x] Indicator Engine
- [x] Order Flow Engine 真实成交优先版
- [x] Large Trades 检测
- [x] 微结构事件序列识别
- [x] 更高阶微结构模式库与组合事件评分
- [x] Structure Engine：`internal / external` hierarchy
- [x] Liquidity Engine：wall map / wall strength bands / wall evolution
- [x] Signal Engine 多因子评分
- [x] AI Explain Engine 基础版

### 2.5 前端与交互

- [x] `Dashboard / Chart / Signals / Market` 页面
- [x] `BTCUSDT / ETHUSDT` 切换
- [x] `1m / 5m / 15m / 1h / 4h` 周期切换
- [x] 结构、流动性、信号、微结构图层
- [x] 微结构 tooltip、图层开关与 `Microstructure Timeline`
- [x] 独立 AI Analysis 面板

### 2.6 质量保障

- [x] 后端引擎测试
- [x] 路由集成测试
- [x] 缓存测试与缓存失效测试
- [x] `market-snapshot` 契约测试
- [x] 前端组件测试
- [x] Playwright 主路径与异常态 E2E
- [x] `go test ./...`
- [x] `npm test`
- [x] `npm run build`
- [x] `npm run lint`

## 3. 当前可上线项

### 3.1 可上线范围

- [x] 作为现货研究型分析终端上线
- [x] 支持 `BTCUSDT / ETHUSDT`
- [x] 支持 `1m / 5m / 15m / 1h / 4h`
- [x] 支持统一快照驱动的图表、信号、订单流、结构、流动性和 AI 解释
- [x] 支持 Redis 热点缓存、symbol 级全周期失效与 `refresh=1`
- [x] 支持 `large_trade_events` 与 `feature_snapshots` 的后台持久化

### 3.2 可直接交付给用户的能力

- [x] Dashboard 总览分析
- [x] Chart 多图层图表分析
- [x] Signals 信号列表与解释
- [x] Market 市场快照与关键价位
- [x] Microstructure Timeline 事件演化阅读
- [x] Liquidity wall map / wall evolution 展示

### 3.3 上线口径说明

- [x] 当前版本适合内部部署、灰度发布或研究终端场景
- [x] 当前版本不是自动交易系统，也不是完整量化研究平台
- [x] 如以真实生产模式部署，应使用 `APP_MODE=prod` 并提供真实 Binance 连接与数据库
- [-] 如需更强生产可观测性，建议在上线后补 metrics / monitoring

## 4. 后续增强项

### 4.1 近期待增强

- [-] 围绕 `large_trade_events` 增加查询、导出与历史回放接口
- [-] 围绕 `feature_snapshots` 增加检索、导出与训练前置链路
- [-] 在统一耗时日志之上增加 metrics、告警与监控面板
- [-] 视重算成本决定是否将部分分析序列进一步落表或冷存储

### 4.2 明确延后，不影响当前上线

- [ ] `Futures`：`Funding Rate / Open Interest / Futures 因子`
- [ ] 自动下单
- [ ] 回测平台
- [ ] 用户系统 / 权限系统
- [ ] 多交易所同时接入
- [ ] 完整高频订单簿重建

## 5. 对后续开发的约束

1. 默认继续以 `market-snapshot` 为前端主接口
2. 新增事件序列优先判断是否需要持久化
3. 涉及时间轴的对象必须能映射到 candle
4. 不要在没有明确目标时破坏当前 Signal 分数口径
5. `Futures` 必须作为独立方向推进，不与当前 Spot 主链路混改
