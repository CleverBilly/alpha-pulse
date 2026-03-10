# Alpha Pulse 数据库设计

更新时间：2026-03-10  
状态：与当前 GORM Model、AutoMigrate 和 MySQL 注释同步

## 1. 总览

- 数据库：MySQL 8
- 默认库名：`alpha_pulse`
- 迁移方式：GORM `AutoMigrate`
- 注释策略：
  - 表注释和列注释来自 `backend/models`
  - 在 MySQL 下，启动时会同步到真实表结构

当前持久化表分为两层：

1. 原始市场数据层
2. 分析结果与归档层

## 2. 表清单

### 2.1 原始市场数据层

- `kline`
- `agg_trades`
- `order_book_snapshots`

### 2.2 分析结果与归档层

- `indicators`
- `orderflow`
- `large_trade_events`
- `microstructure_events`
- `structure`
- `liquidity`
- `signals`
- `feature_snapshots`

## 3. 表结构明细

### 3.1 `kline`

表说明：原始 K 线行情数据，供指标、结构、流动性和信号分析复用。

主要索引：

- 唯一索引：`idx_kline_symbol_interval_open_time (symbol, interval_type, open_time)`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `interval_type` | `varchar(10)` | K 线周期，如 `1m`、`5m`、`1h` |
| `open_price` | `decimal(18,8)` | 开盘价 |
| `high_price` | `decimal(18,8)` | 最高价 |
| `low_price` | `decimal(18,8)` | 最低价 |
| `close_price` | `decimal(18,8)` | 收盘价 |
| `volume` | `decimal(18,8)` | 成交量，基础资产数量 |
| `open_time` | `bigint` | K 线起始时间，Unix 毫秒 |
| `created_at` | `datetime(3)` | 入库时间 |

### 3.2 `agg_trades`

表说明：Binance 聚合成交原始数据，为订单流和微结构分析提供真实成交输入。

主要索引：

- 唯一索引：`idx_agg_trade_symbol_trade_id (symbol, agg_trade_id)`
- 普通索引：`symbol`
- 普通索引：`trade_time`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `agg_trade_id` | `bigint` | Binance 聚合成交 ID |
| `price` | `decimal(18,8)` | 成交价 |
| `quantity` | `decimal(24,8)` | 成交数量，基础资产数量 |
| `quote_quantity` | `decimal(24,8)` | 成交额，计价资产数量 |
| `first_trade_id` | `bigint` | 聚合成交覆盖的首笔原始成交 ID |
| `last_trade_id` | `bigint` | 聚合成交覆盖的末笔原始成交 ID |
| `trade_time` | `bigint` | 成交时间，Unix 毫秒 |
| `is_buyer_maker` | `tinyint(1)` | 买方是否为挂单方 |
| `is_best_price_match` | `tinyint(1)` | 是否最佳价格成交 |
| `created_at` | `datetime(3)` | 入库时间 |

### 3.3 `order_book_snapshots`

表说明：订单簿深度快照原始数据，为流动性和盘口迁移分析提供输入。

主要索引：

- 唯一索引：`idx_order_book_symbol_update_id (symbol, last_update_id)`
- 普通索引：`symbol`
- 普通索引：`event_time`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `last_update_id` | `bigint` | 盘口快照更新 ID |
| `depth_level` | `bigint` | 快照包含的深度档位数 |
| `bids_json` | `longtext` | 买盘深度原始 JSON |
| `asks_json` | `longtext` | 卖盘深度原始 JSON |
| `best_bid_price` | `decimal(18,8)` | 最优买价 |
| `best_ask_price` | `decimal(18,8)` | 最优卖价 |
| `spread` | `decimal(18,8)` | 买卖价差 |
| `event_time` | `bigint` | 快照事件时间，Unix 毫秒 |
| `created_at` | `datetime(3)` | 入库时间 |

### 3.4 `indicators`

表说明：技术指标计算快照，记录 RSI、MACD、EMA、ATR、布林带和 VWAP。

主要索引：

- 普通索引：`symbol`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `rsi` | `decimal(10,2)` | 相对强弱指标 RSI |
| `macd` | `decimal(10,4)` | MACD 快线值 |
| `macd_signal` | `decimal(10,4)` | MACD Signal 线值 |
| `macd_histogram` | `decimal(10,4)` | MACD 柱体值 |
| `ema20` | `decimal(18,8)` | 20 周期 EMA |
| `ema50` | `decimal(18,8)` | 50 周期 EMA |
| `atr` | `decimal(18,8)` | 平均真实波幅 ATR |
| `bollinger_upper` | `decimal(18,8)` | 布林带上轨 |
| `bollinger_middle` | `decimal(18,8)` | 布林带中轨 |
| `bollinger_lower` | `decimal(18,8)` | 布林带下轨 |
| `vwap` | `decimal(18,8)` | 成交量加权平均价 VWAP |
| `created_at` | `datetime(3)` | 指标计算入库时间 |

说明：

- 指标时间序列不单独落表
- 前端序列视图由后端基于历史 K 线滚动生成

### 3.5 `orderflow`

表说明：订单流分析快照，记录主动买卖量、大单统计、吸收和冰山单特征。

主要索引：

- 普通索引：`symbol`
- 普通索引：`interval_type`
- 普通索引：`open_time`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `interval_type` | `varchar(10)` | 订单流分析周期 |
| `open_time` | `bigint` | 对齐的 K 线起始时间，Unix 毫秒 |
| `buy_volume` | `decimal(18,8)` | 主动买入成交量 |
| `sell_volume` | `decimal(18,8)` | 主动卖出成交量 |
| `delta` | `decimal(18,8)` | 买卖量差值 |
| `cvd` | `decimal(18,8)` | 累积成交量差 CVD |
| `buy_large_trade_count` | `bigint` | 买方大单数量 |
| `sell_large_trade_count` | `bigint` | 卖方大单数量 |
| `buy_large_trade_notional` | `decimal(24,8)` | 买方大单成交额 |
| `sell_large_trade_notional` | `decimal(24,8)` | 卖方大单成交额 |
| `large_trade_delta` | `decimal(24,8)` | 大单成交额差值 |
| `absorption_bias` | `varchar(20)` | 吸收行为方向 |
| `absorption_strength` | `decimal(12,6)` | 吸收行为强度 |
| `iceberg_bias` | `varchar(20)` | 冰山单方向 |
| `iceberg_strength` | `decimal(12,6)` | 冰山单强度 |
| `data_source` | `varchar(20)` | 分析数据来源，如 `agg_trade` 或 `kline` |
| `created_at` | `datetime(3)` | 订单流分析入库时间 |

说明：

- `large_trades[]`、`microstructure_events[]` 会通过接口返回，但不直接落在本表
- 其持久化镜像分别写入 `large_trade_events` 和 `microstructure_events`

### 3.6 `large_trade_events`

表说明：大单事件持久化镜像，支持历史回放、聚类分析和时间轴重建。

主要索引：

- 唯一索引：`idx_large_trade_event_unique (symbol, agg_trade_id)`
- 普通索引：`orderflow_id`
- 普通索引：`agg_trade_id`
- 普通索引：`interval_type`
- 普通索引：`open_time`
- 普通索引：`notional`
- 普通索引：`trade_time`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `orderflow_id` | `bigint unsigned` | 来源订单流快照 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `agg_trade_id` | `bigint` | 来源聚合成交 ID |
| `interval_type` | `varchar(10)` | 事件所属周期 |
| `open_time` | `bigint` | 对齐的 K 线起始时间，Unix 毫秒 |
| `side` | `varchar(10)` | 成交方向，`buy` 或 `sell` |
| `price` | `decimal(18,8)` | 成交价 |
| `quantity` | `decimal(24,8)` | 成交数量，基础资产数量 |
| `notional` | `decimal(24,8)` | 成交额，计价资产数量 |
| `trade_time` | `bigint` | 成交发生时间，Unix 毫秒 |
| `created_at` | `datetime(3)` | 入库时间 |

### 3.7 `microstructure_events`

表说明：微结构事件历史序列，用于时间轴展示、图表标注和后续回放。

主要索引：

- 唯一索引：`idx_micro_event_unique (symbol, interval_type, open_time, event_type, trade_time, bias)`
- 普通索引：`orderflow_id`
- 普通索引：`symbol`
- 普通索引：`interval_type`
- 普通索引：`open_time`
- 普通索引：`trade_time`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `orderflow_id` | `bigint unsigned` | 来源订单流快照 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `interval_type` | `varchar(10)` | 事件所属周期 |
| `open_time` | `bigint` | 对齐的 K 线起始时间，Unix 毫秒 |
| `event_type` | `varchar(40)` | 微结构事件类型 |
| `bias` | `varchar(20)` | 事件方向偏向 |
| `score` | `bigint` | 事件评分 |
| `strength` | `decimal(12,6)` | 事件强度 |
| `price` | `decimal(18,8)` | 事件参考价 |
| `trade_time` | `bigint` | 事件发生时间，Unix 毫秒 |
| `detail` | `text` | 事件详细说明 |
| `created_at` | `datetime(3)` | 入库时间 |

### 3.8 `structure`

表说明：市场结构分析快照，记录趋势、关键支撑阻力以及 BOS 和 CHOCH 状态。

主要索引：

- 普通索引：`symbol`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `trend` | `varchar(20)` | 当前市场结构趋势，如 `uptrend`、`downtrend`、`range` |
| `support` | `decimal(18,8)` | 当前主支撑位 |
| `resistance` | `decimal(18,8)` | 当前主阻力位 |
| `bos` | `tinyint(1)` | 是否出现结构突破 BOS |
| `choch` | `tinyint(1)` | 是否出现角色转换 CHOCH |
| `created_at` | `datetime(3)` | 结构分析入库时间 |

说明：

- `primary_tier`、`internal_support`、`internal_resistance`、`external_support`、`external_resistance`、`events[]` 都是接口返回字段，不落库

### 3.9 `liquidity`

表说明：流动性分析快照，记录买卖流动性位、扫单类型和盘口失衡结果。

主要索引：

- 普通索引：`symbol`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `buy_liquidity` | `decimal(18,8)` | 下方主要买方流动性位 |
| `sell_liquidity` | `decimal(18,8)` | 上方主要卖方流动性位 |
| `sweep_type` | `varchar(20)` | 最近识别到的扫流动性类型 |
| `order_book_imbalance` | `decimal(12,6)` | 盘口买卖失衡值 |
| `data_source` | `varchar(20)` | 流动性分析数据来源，如 `orderbook` 或 `kline` |
| `created_at` | `datetime(3)` | 流动性分析入库时间 |

说明：

- `equal_high`、`equal_low`、`stop_clusters[]`、`wall_levels[]`、`wall_strength_bands[]`、`wall_evolution[]` 都是接口返回字段，不落库

### 3.10 `signals`

表说明：多因子交易信号快照，记录动作、分数、置信度和建议价位。

主要索引：

- 普通索引：`symbol`
- 普通索引：`interval_type`
- 普通索引：`open_time`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `interval_type` | `varchar(10)` | 信号所属周期 |
| `open_time` | `bigint` | 对齐的 K 线起始时间，Unix 毫秒 |
| `signal` | `varchar(20)` | 综合动作信号，如 `BUY`、`SELL`、`NEUTRAL` |
| `score` | `bigint` | 综合评分，范围 `-100 ~ +100` |
| `confidence` | `bigint` | 信号置信度百分比 |
| `entry_price` | `decimal(18,8)` | 建议入场价 |
| `stop_loss` | `decimal(18,8)` | 建议止损价 |
| `target_price` | `decimal(18,8)` | 建议止盈价 |
| `created_at` | `datetime(3)` | 信号生成时间 |

说明：

- `explain`、`factors[]`、`risk_reward`、`trend_bias` 为接口扩展字段，不落库

### 3.11 `feature_snapshots`

表说明：聚合特征快照归档表，保存完整 `market_snapshot` 以供审计、训练和回放。

主要索引：

- 唯一索引：`idx_feature_snapshot_unique (symbol, interval_type, open_time, snapshot_source)`
- 普通索引：`symbol`
- 普通索引：`interval_type`
- 普通索引：`open_time`
- 普通索引：`signal_action`
- 普通索引：`signal_score`

字段：

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `id` | `bigint unsigned` | 主键 ID |
| `symbol` | `varchar(20)` | 交易对代码，如 `BTCUSDT` |
| `interval_type` | `varchar(10)` | 快照所属周期 |
| `open_time` | `bigint` | 对齐的 K 线起始时间，Unix 毫秒 |
| `snapshot_source` | `varchar(32)` | 快照来源标识，默认 `market_snapshot` |
| `feature_version` | `varchar(16)` | 特征快照版本号 |
| `price` | `decimal(18,8)` | 快照时刻价格 |
| `signal_action` | `varchar(20)` | 信号方向 |
| `signal_score` | `bigint` | 信号分数 |
| `signal_confidence` | `bigint` | 信号置信度百分比 |
| `snapshot_json` | `longtext` | 完整 `market_snapshot` JSON |
| `created_at` | `datetime(3)` | 保存时间 |

## 4. 当前不落库但会通过接口返回的字段

以下字段不属于 MySQL 表结构，但会在 API 层返回：

- `indicator_series[]`
- `signal_timeline[]`
- `structure_series[]`
- `liquidity_series[]`
- `large_trades[]`
- `microstructure_events[]`
- `wall_levels[]`
- `wall_strength_bands[]`
- `wall_evolution[]`

## 5. 使用建议

1. 本地调试建议保持 `AUTO_MIGRATE=true`，这样表结构和注释会自动同步
2. 如仅使用本地 MySQL、不启 Redis，可在 `backend/.env` 中设 `ENABLE_REDIS_CACHE=false`
3. 如果要核对真实注释，可执行 `SHOW FULL COLUMNS FROM <table>;` 或 `SHOW CREATE TABLE <table>;`
