# Alpha Pulse API

更新时间：2026-03-07  
状态：与当前代码基线对齐

## 1. 总览

Base URL：

`http://localhost:8080/api`

健康检查：

- `GET /healthz`

统一响应结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

错误响应示例：

```json
{
  "code": 500,
  "message": "internal error"
}
```

## 2. 通用约定

### 2.1 支持的交易对

- `BTCUSDT`
- `ETHUSDT`

### 2.2 支持的周期

- `1m`
- `5m`
- `15m`
- `1h`
- `4h`

### 2.3 `limit` 约定

- 大多数时间序列接口默认 `48`
- 非法 `limit` 会回退到接口默认值
- 服务端会进一步做上限保护

## 3. 接口列表

### 3.1 `GET /api/price`

用途：获取实时价格。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`

返回字段：

- `symbol`
- `price`
- `time`

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "symbol": "BTCUSDT",
    "price": 65420.15,
    "time": 1772817000000
  }
}
```

### 3.2 `GET /api/kline`

用途：获取最新单根 K 线。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`

返回字段：

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

### 3.3 `GET /api/indicators`

用途：获取最新指标结果。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`

返回字段：

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

### 3.4 `GET /api/indicator-series`

用途：获取指标时间序列。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`
- `limit`：可选，默认 `48`
- `refresh`：可选，传 `1` 时显式绕过缓存并回填新结果

返回结构：

- `symbol`
- `interval`
- `points[]`

`points[]` 字段：

- `open_time`
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

### 3.5 `GET /api/orderflow`

用途：获取当前订单流分析结果。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`

返回字段：

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
- `large_trades[]`
- `microstructure_events[]`
- `created_at`

说明：

- 优先基于真实 `aggTrade` 计算
- `aggTrade` 不足时回退到 OHLCV 估算
- `microstructure_events[]` 表示当前订单流分析窗口内识别到的事件摘要

### 3.6 `GET /api/microstructure-events`

用途：获取历史微结构事件序列。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`
- `limit`：可选，默认 `20`

返回结构：

- `symbol`
- `interval`
- `events[]`

`events[]` 字段：

- `id`
- `orderflow_id`
- `symbol`
- `interval_type`
- `open_time`
- `type`
- `bias`
- `score`
- `strength`
- `price`
- `trade_time`
- `detail`
- `created_at`

当前事件类型包括：

- `absorption`
- `iceberg`
- `aggression_burst`
- `failed_auction`
- `failed_auction_high_reject`
- `failed_auction_low_reclaim`
- `initiative_shift`
- `large_trade_cluster`
- `order_book_migration`
- `order_book_migration_layered`
- `order_book_migration_accelerated`
- `auction_trap_reversal`
- `liquidity_ladder_breakout`
- `microstructure_confluence`

### 3.7 `GET /api/structure`

用途：获取当前市场结构分析结果。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`

返回字段：

- `id`
- `symbol`
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
- `created_at`

`events[]` 字段：

- `label`
- `kind`
- `tier`
- `price`
- `open_time`

### 3.8 `GET /api/market-structure-events`

用途：获取结构事件专用视图。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`

返回结构：

- `symbol`
- `interval`
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

### 3.9 `GET /api/market-structure-series`

用途：获取结构时间序列。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`
- `limit`：可选，默认 `48`
- `refresh`：可选，传 `1` 时显式绕过缓存并回填新结果

返回结构：

- `symbol`
- `interval`
- `points[]`

`points[]` 字段：

- `open_time`
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
- `event_labels[]`
- `event_tags[]`

### 3.10 `GET /api/liquidity`

用途：获取当前流动性分析结果。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`

返回字段：

- `id`
- `symbol`
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
- `created_at`

### 3.11 `GET /api/liquidity-map`

用途：获取流动性图谱专用视图。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`

返回结构：

- `symbol`
- `interval`
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

`wall_levels[]` 字段：

- `label`
- `kind`
- `side`
- `layer`
- `price`
- `quantity`
- `notional`
- `distance_bps`
- `strength`

`wall_strength_bands[]` 字段：

- `side`
- `band`
- `lower_distance_bps`
- `upper_distance_bps`
- `level_count`
- `total_notional`
- `dominant_price`
- `dominant_notional`
- `strength`

`wall_evolution[]` 字段：

- `interval`
- `buy_liquidity`
- `sell_liquidity`
- `buy_distance_bps`
- `sell_distance_bps`
- `buy_cluster_strength`
- `sell_cluster_strength`
- `buy_strength_delta`
- `sell_strength_delta`
- `order_book_imbalance`
- `sweep_type`
- `data_source`
- `dominant_side`

### 3.12 `GET /api/liquidity-series`

用途：获取流动性时间序列。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`
- `limit`：可选，默认 `48`
- `refresh`：可选，传 `1` 时显式绕过缓存并回填新结果

返回结构：

- `symbol`
- `interval`
- `points[]`

`points[]` 字段：

- `open_time`
- `buy_liquidity`
- `sell_liquidity`
- `sweep_type`
- `order_book_imbalance`
- `data_source`
- `equal_high`
- `equal_low`
- `buy_cluster_strength`
- `sell_cluster_strength`

### 3.13 `GET /api/signal`

用途：获取当前综合信号与其上游分析结果。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`

返回结构：

- `signal`
- `indicator`
- `orderflow`
- `structure`
- `liquidity`

`signal` 主要字段：

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
- `risk_reward`
- `trend_bias`
- `factors[]`
- `explain`
- `created_at`

### 3.14 `GET /api/signal-timeline`

用途：获取历史信号时间线。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`
- `limit`：可选，默认 `48`
- `refresh`：可选，传 `1` 时显式绕过缓存并回填新结果

返回结构：

- `symbol`
- `interval`
- `points[]`

`points[]` 字段：

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

### 3.15 `GET /api/market-snapshot`

用途：当前前端主聚合接口。Dashboard、Chart、Signals、Market 页面应优先依赖该接口。

查询参数：

- `symbol`：可选，默认 `BTCUSDT`
- `interval`：可选，默认 `1m`
- `limit`：可选，默认 `48`

返回结构：

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

说明：

- `orderflow.microstructure_events` 是当前订单流分析窗口内的摘要
- 顶层 `microstructure_events` 是和图表时间范围对齐的历史事件序列
- 图表和时间线组件应优先使用顶层 `microstructure_events`
- `refresh=1` 时会显式跳过缓存读取，并刷新当前 symbol 的相关缓存视图

响应骨架示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "price": {
      "symbol": "BTCUSDT",
      "price": 65420.15,
      "time": 1772817000000
    },
    "klines": [],
    "indicator": {},
    "indicator_series": [],
    "orderflow": {},
    "microstructure_events": [],
    "structure": {},
    "structure_series": [],
    "liquidity": {},
    "liquidity_series": [],
    "signal": {},
    "signal_timeline": []
  }
}
```

## 4. 前端使用建议

当前页面应遵循以下调用策略：

- Dashboard：优先只调 `market-snapshot`
- Chart：优先只调 `market-snapshot`
- Signals：优先只调 `market-snapshot`
- Market：优先只调 `market-snapshot`

以下接口保留给专用扩展视图或调试用途：

- `indicator-series`
- `market-structure-events`
- `market-structure-series`
- `liquidity-map`
- `liquidity-series`
- `microstructure-events`
- `signal-timeline`

## 5. 当前稳定契约说明

以下字段命名已经被前后端、测试和图表逻辑依赖，后续不要随意改名：

- `interval_type`
- `open_time`
- `order_book_imbalance`
- `microstructure_events`
- `signal_timeline`
- `indicator_series`
- `structure_series`
- `liquidity_series`

如果未来要扩展新字段，优先新增，不要破坏现有字段。
