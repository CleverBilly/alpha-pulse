# Alpha Pulse 任务清单

更新时间：2026-03-07  
状态：与当前 PRD 和代码基线同步

## 1. 使用说明

状态定义：

- `[x]` 已完成
- `[-]` 已部分完成，当前可用但仍可增强
- `[ ]` 未开始

任务按推荐执行顺序排列。

## 2. 当前已完成基线

### 文档与范围

- [x] 将项目范围收敛为 Spot Analysis MVP
- [x] 将 Signal Engine 收敛为 `-100 ~ +100` 连续评分模型
- [x] 重写 `PRD.md` 为开发源文档
- [x] 建立并持续维护 API / Database / Architecture / Status / Task 文档

### 后端基础

- [x] Monorepo 基础结构
- [x] Backend 分层架构
- [x] MySQL / Redis / Docker 基础接入
- [x] Binance SDK 接入
- [x] REST + WebSocket 混合采集
- [x] `market-snapshot` 聚合接口

### 数据层

- [x] `kline`
- [x] `agg_trades`
- [x] `order_book_snapshots`
- [x] `indicators`
- [x] `orderflow`
- [x] `microstructure_events`
- [x] `structure`
- [x] `liquidity`
- [x] `signals`

### 引擎层

- [x] Indicator Engine
- [x] Order Flow Engine 真实成交优先版
- [x] Large Trades 检测
- [x] 微结构事件序列识别
- [x] Structure Engine 结构事件版
- [x] Liquidity Engine 盘口增强版
- [x] Signal Engine 多因子评分
- [x] AI Explain Engine 基础版

### 前端

- [x] Dashboard 页面
- [x] Chart 页面
- [x] Signals 页面
- [x] Market 页面
- [x] 1m / 5m / 15m / 1h / 4h 周期切换
- [x] 多 K 线 SVG 图
- [x] 结构点 / 流动性 / 信号 / 微结构图层
- [x] 独立 AI Analysis 面板

### 质量

- [x] 后端引擎测试
- [x] 路由集成测试
- [x] Snapshot Cache 测试
- [x] 前端组件测试
- [x] Playwright 主路径 E2E
- [x] 接口失败 / 弱网 / 切币 / 刷新异常态测试

## 3. 当前推荐任务

## 3.1 P1 图表与事件增强

- [x] 给微结构事件标注增加 tooltip
  - 验收：图表 hover 时可看到 `type / bias / score / strength / detail`
- [x] 把 `initiative_shift` 和 `large_trade_cluster` 加入可切换图层
  - 验收：用户可选择显示更多次级微结构事件
- [x] 增加独立的 `Microstructure Timeline` 视图或卡片
  - 验收：用户可按时间阅读最近事件演化，而不只看图表缩写标记
- [x] 增强 `Microstructure Timeline` 可读性
  - 验收：支持事件家族过滤、摘要统计和高阶事件高亮

## 3.2 P1 缓存与接口增强

- [x] 将 Redis 扩展到更多热点接口
  - 已覆盖：`signal-timeline`、`indicator-series`、`liquidity-series`
- [x] 明确缓存失效策略
  - 验收：切币、切周期、刷新行为下缓存表现可预测
- [x] 为关键接口增加更明确的契约测试
  - 重点：`market-snapshot` 返回字段完整性与兼容性

## 3.3 P1 可观测性

- [x] 增加关键链路日志
  - 已完成：`market-snapshot` 构建耗时、缓存命中/未命中、视图缓存读写日志、Collector、回退链路、统一结构化日志
- [x] 增加引擎耗时统计
  - 验收：可观察快照构建慢点在哪里

## 4. 第二优先级任务

## 4.1 P2 分析能力增强

- [-] 增加更高阶微结构模式
  - 已完成：连续吸收、失败拍卖、被动挂单迁移、失败拍卖扩展型、被动挂单迁移分层、组合事件评分、失败拍卖陷阱反转、流动性阶梯突破
  - 待完成：更多组合模式库与更细粒度事件表达
- [x] 继续增强 Liquidity Engine
  - 已完成：更细粒度 liquidity wall map、wall strength map 与跨周期 wall 演化
- [x] 继续增强 Structure Engine
  - 已完成：internal / external swing hierarchy、hierarchy-aware 结构事件与层级支撑阻力

## 4.2 P2 数据与平台能力

- [ ] 新增 `large_trade_events` 表
  - 适用于未来大单历史回放与聚类分析
- [ ] 引入 `signal_runs` 或 `feature_snapshots`
  - 适用于未来离线审计、训练或回测前置
- [x] 区分 `dev / test / prod` 运行模式
  - 已完成：mode 默认值、Gin mode、自动迁移、Redis、stream collector、scheduler、Binance mock fallback 分离

## 4.3 P2 Futures 方向

以下任务在明确启动 Futures 子方向后再做：

- [ ] 接入 Funding Rate
- [ ] 接入 Open Interest
- [ ] 新增 `funding_rates` 表
- [ ] 新增 `open_interest` 表
- [ ] 把 Futures 因子接入 Signal Engine

## 5. 当前不建议优先做的事

以下事情当前不建议插队：

- [ ] 自动下单
- [ ] 回测平台
- [ ] 用户系统 / 权限系统
- [ ] 多交易所同时接入
- [ ] 完整高频订单簿重建

原因：

- 这些工作会明显扩大系统边界
- 会稀释当前 Spot Analysis MVP 主线
- 当前更值得继续做的是分析能力和图层表达力

## 6. 推荐执行顺序

建议按以下顺序推进：

1. 继续扩展更高阶微结构模式库
2. 新增 `large_trade_events` 或 `feature_snapshots`
3. 如有明确业务目标，再开启 Futures 子线

## 7. 对后续 AI 工具的执行规则

后续 AI 工具继续开发时，建议遵守：

1. 默认以 `market-snapshot` 为前端主接口
2. 新增事件序列优先考虑是否需要持久化
3. 涉及时间轴的对象必须能映射到 candle
4. 不要在没有明确需求时破坏当前 Signal 分数口径
5. 扩表可以新增字段或新表，不要为了兼容旧描述而回退当前模型
