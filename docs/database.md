# Alpha Pulse 数据库设计

更新时间：2026-03-07  
状态：与当前 GORM Model 和 AutoMigrate 对齐

## 1. 数据库概览

当前项目数据库采用：

- MySQL
- 默认数据库名：`alpha_pulse`

当前通过 GORM `AutoMigrate` 管理主要业务表。

## 2. 设计原则

当前数据库设计采用两层思路：

1. 原始市场数据层
2. 分析结果层

这样做的原因：

- 原始行情数据可重复使用
- 分析引擎可重算
- 图表和信号可基于统一历史数据构建
- 降低后续接入回放、审计、二次分析的成本

## 3. 当前表清单

### 3.1 原始市场数据层

- `kline`
- `agg_trades`
- `order_book_snapshots`

### 3.2 分析结果层

- `indicators`
- `orderflow`
- `microstructure_events`
- `structure`
- `liquidity`
- `signals`

## 4. 表设计明细

## 4.1 `kline`

用途：

- 保存历史 K 线
- 为 Indicator / Structure / Liquidity / Signal 提供输入

核心字段：

- `id`
- `symbol`
- `interval_type`
- `open_price`
- `high_price`
- `low_price`
- `close_price`
- `volume`
- `open_time`
- `created_at`

约束：

- `symbol + interval_type + open_time` 唯一索引

说明：

- 当前表名为 `kline`，不是 `klines`
- 前端图表时间轴默认依赖 `open_time`

## 4.2 `agg_trades`

用途：

- 保存 Binance 聚合成交原始数据
- 为 Order Flow 与微结构事件识别提供输入

核心字段：

- `id`
- `symbol`
- `agg_trade_id`
- `price`
- `quantity`
- `quote_quantity`
- `first_trade_id`
- `last_trade_id`
- `trade_time`
- `is_buyer_maker`
- `is_best_price_match`
- `created_at`

约束：

- `symbol + agg_trade_id` 唯一索引

说明：

- 当前订单流真实分析优先使用该表
- 回退路径才会使用 OHLCV 估算

## 4.3 `order_book_snapshots`

用途：

- 保存盘口快照原始数据
- 为 Liquidity Engine 提供盘口墙与盘口失衡输入

核心字段：

- `id`
- `symbol`
- `last_update_id`
- `depth_level`
- `bids_json`
- `asks_json`
- `best_bid_price`
- `best_ask_price`
- `spread`
- `event_time`
- `created_at`

约束：

- `symbol + last_update_id` 唯一索引

说明：

- `bids_json`、`asks_json` 保存原始深度结构
- 当前用于盘口快照增强分析，不是完整增量簿引擎

## 4.4 `indicators`

用途：

- 保存某次指标计算的最新结果快照

核心字段：

- `id`
- `symbol`
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
- `created_at`

说明：

- 指标时间序列当前不单独落表
- 时间序列由后端基于历史 K 线滚动生成

## 4.5 `orderflow`

用途：

- 保存当前订单流分析结果快照

核心字段：

- `id`
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
- `created_at`

不落库但会通过接口返回的字段：

- `large_trades[]`
- `microstructure_events[]`

说明：

- `large_trades[]` 是当前窗口摘要，不是独立持久化事件表
- `microstructure_events[]` 会在持久化时映射写入 `microstructure_events` 表

## 4.6 `microstructure_events`

用途：

- 持久化微结构事件历史序列
- 支持图表标注、历史查询和未来事件回放

核心字段：

- `id`
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
- `created_at`

当前事件类型：

- `absorption`
- `iceberg`
- `aggression_burst`
- `initiative_shift`
- `large_trade_cluster`

唯一约束：

- `symbol + interval_type + open_time + event_type + trade_time + bias`

说明：

- 当前 `market-snapshot` 顶层 `microstructure_events` 就是基于该表时间窗口查询得到

## 4.7 `structure`

用途：

- 保存市场结构分析结果快照

核心字段：

- `id`
- `symbol`
- `trend`
- `support`
- `resistance`
- `bos`
- `choch`
- `created_at`

不落库但会通过接口返回：

- `events[]`

说明：

- `events[]` 包括 `HH / HL / LH / LL / BOS / CHOCH` 等结构事件
- 当前结构时间序列由后端滚动生成，不单独落表

## 4.8 `liquidity`

用途：

- 保存流动性分析结果快照

核心字段：

- `id`
- `symbol`
- `buy_liquidity`
- `sell_liquidity`
- `sweep_type`
- `order_book_imbalance`
- `data_source`
- `created_at`

不落库但会通过接口返回：

- `equal_high`
- `equal_low`
- `stop_clusters[]`

说明：

- 这些字段当前通过分析结果动态输出，而不是存进 `liquidity` 表

## 4.9 `signals`

用途：

- 保存最终交易信号快照

核心字段：

- `id`
- `symbol`
- `interval_type`
- `open_time`
- `signal`
- `score`
- `confidence`
- `entry_price`
- `stop_loss`
- `target_price`
- `created_at`

不落库但会通过接口返回：

- `explain`
- `factors[]`
- `risk_reward`
- `trend_bias`

说明：

- 历史信号时间线由该表读取并压缩得到

## 5. 当前不单独落表但已存在的逻辑对象

这些对象当前已参与业务，但没有独立持久化表：

- `indicator_series`
- `structure_series`
- `liquidity_series`
- `signal_timeline`
- `large_trades[]`
- `structure.events[]`
- `liquidity.stop_clusters[]`

这不是缺陷，而是当前阶段的取舍。

## 6. 已知设计取舍

### 6.1 时间序列不全量落库

当前没有为以下对象单独建表：

- 指标序列
- 结构序列
- 流动性序列

原因：

- 当前版本更偏分析型 MVP
- 这些序列可由历史 K 线和当前引擎滚动生成
- 避免过早引入大量冗余分析结果表

### 6.2 `large_trades` 未独立建表

当前大单事件仅作为订单流结果摘要返回，没有独立持久化表。

如果后续需要：

- 大单历史回放
- 大单图层渲染
- 大单聚类统计

建议新增：

- `large_trade_events`

### 6.3 Futures 相关表未引入

当前未引入：

- `funding_rates`
- `open_interest`

这是明确的当前版本边界，不是遗漏。

## 7. 后续推荐扩展表

建议按照真实需要再扩：

- `large_trade_events`
- `signal_runs`
- `funding_rates`
- `open_interest`
- `feature_snapshots`（如果后续需要离线训练或因子回放）

## 8. 迁移与约束建议

后续开发如果需要扩表，建议遵循：

1. 优先新增字段或新表，不要破坏现有字段名
2. 所有时间轴相关表必须有明确时间字段：`open_time` 或 `trade_time` 或 `event_time`
3. 原始数据表和分析结果表继续分层
4. 对图表直接消费的字段，优先保持 JSON 字段名稳定
5. 需要历史重放的事件优先单独落表，不要只保存在临时响应体里
